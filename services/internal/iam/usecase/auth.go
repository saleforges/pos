package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type RegisterParams struct {
	Username  string
	Email     string
	Password  string
	Roles     []string
	UserType  domain.UserType
	IPAddress string
	UserAgent string
}

type AuthResult struct {
	port.TokenPair
}

type LoginParams struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginResult struct {
	port.TokenPair
}

type RefreshTokenParams struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

type LogoutParams struct {
	SessionID string
}

type IntrospectResult struct {
	Active      bool                        `json:"active"`
	UserID      int64                       `json:"user_id,omitempty"`
	UserType    domain.UserType             `json:"user_type"`
	RoleName    string                      `json:"role_name,omitempty"`
	Staff       []domain.UserRoleAssignment `json:"staff,omitempty"`
	Permissions []domain.Permission         `json:"permissions,omitempty"`
}

type authUsecase struct {
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	loginAuditRepo repository.LoginAuditRepository
	staffRepo      repository.StaffRepository
	sessionStore   port.SessionStore
	eventPublisher port.EventPublisher
	passwordHasher port.PasswordHasher
	tokenSigner    port.TokenSigner
	tokenHasher    port.TokenHasher
	userCache      port.UserCache
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	loginAuditRepo repository.LoginAuditRepository,
	staffRepo repository.StaffRepository,
	sessionStore port.SessionStore,
	eventPublisher port.EventPublisher,
	passwordHasher port.PasswordHasher,
	tokenSigner port.TokenSigner,
	tokenHasher port.TokenHasher,
	userCache port.UserCache,
) *authUsecase {
	return &authUsecase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		loginAuditRepo: loginAuditRepo,
		staffRepo:      staffRepo,
		sessionStore:   sessionStore,
		eventPublisher: eventPublisher,
		passwordHasher: passwordHasher,
		tokenSigner:    tokenSigner,
		tokenHasher:    tokenHasher,
		userCache:      userCache,
	}
}

func (uc *authUsecase) Register(ctx context.Context, input RegisterParams) (*AuthResult, error) {
	ctx, span := otel.StartSpan(ctx, "auth.Register")
	defer span.End()

	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	if len(input.Roles) == 0 {
		input.Roles = []string{"viewer"}
	}

	for _, role := range input.Roles {
		if _, ok := domain.DefaultRoles[role]; !ok {
			_, err := uc.roleRepo.GetByName(ctx, role)
			if err != nil {
				return nil, domain.ErrInvalidRole
			}
		}
	}

	existing, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		logger.Error("register: check username failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	existing, err = uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		logger.Error("register: check email failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashed, err := uc.passwordHasher.Hash(input.Password)
	if err != nil {
		logger.Error("register: password hash failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	now := time.Now().UTC()
	userType := input.UserType
	if userType == "" {
		userType = domain.UserTypeMerchant
	}
	user := &domain.User{
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashed),
		Type:      userType,
		Status:    domain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		logger.Error("register: create user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	for _, r := range input.Roles {
		if err := uc.userRepo.AddRole(ctx, user.ID, r); err != nil {
			logger.Error("register: add role failed", "error", err.Error())
			return nil, domain.ErrInternal
		}
	}

	user, err = uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	uc.cacheSet(ctx, user)

	systemRoleName := ""
	if user.SystemRole != nil {
		systemRoleName = user.SystemRole.Name
	}

	permissions, err := uc.collectPermissions(ctx, systemRoleName)
	if err != nil {
		logger.Error("register: collect permissions failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	sessionID := uuid.New().String()
	refreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID, sessionID)
	if err != nil {
		logger.Error("register: sign refresh token failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	session := &domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: uc.tokenHasher.HashToken(refreshTokenStr),
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		LastUsedAt:       now,
		ExpiresAt:        now.Add(30 * 24 * time.Hour),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := uc.sessionStore.Create(ctx, session); err != nil {
		return nil, domain.ErrInternal
	}

	accessToken, err := uc.tokenSigner.SignAccessToken(port.TokenClaims{
		Subject:     strconv.FormatInt(user.ID, 10),
		SessionID:   sessionID,
		UserID:      user.ID,
		UserType:    user.Type,
		RoleName:    systemRoleName,
		Permissions: permissions,
	})
	if err != nil {
		logger.Error("register: sign access token failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	uc.publishEvent(ctx, "UserCreated", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})

	return &AuthResult{
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenStr,
			ExpiresIn:    3600,
		},
	}, nil
}

func (uc *authUsecase) Login(ctx context.Context, input LoginParams) (*LoginResult, error) {
	ctx, span := otel.StartSpan(ctx, "auth.Login")
	defer span.End()

	user, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		uc.auditLogin(ctx, 0, input.Username, false, input.IPAddress, input.UserAgent, "user not found")
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		logger.Error("login: get user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	if err := uc.passwordHasher.Compare(user.Password, input.Password); err != nil {
		uc.auditLogin(ctx, user.ID, input.Username, false, input.IPAddress, input.UserAgent, "invalid password")
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status == domain.UserStatusDisabled {
		uc.auditLogin(ctx, user.ID, input.Username, false, input.IPAddress, input.UserAgent, "user disabled")
		return nil, domain.ErrUserDisabled
	}

	systemRoleName := ""
	if user.SystemRole != nil {
		systemRoleName = user.SystemRole.Name
	}

	permissions, err := uc.collectPermissions(ctx, systemRoleName)
	if err != nil {
		logger.Error("login: collect permissions failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	now := time.Now().UTC()
	sessionID := uuid.New().String()
	refreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	session := &domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: uc.tokenHasher.HashToken(refreshTokenStr),
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		LastUsedAt:       now,
		ExpiresAt:        now.Add(30 * 24 * time.Hour),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := uc.sessionStore.Create(ctx, session); err != nil {
		return nil, domain.ErrInternal
	}

	roleID, merchantID, branchID := uc.resolveActiveRole(ctx, user.ID)
	if roleID != 0 {
		session.ActiveUserRoleID = roleID
		session.UpdatedAt = now
		if err := uc.sessionStore.Update(ctx, session); err != nil {
			logger.Error("login: update session with active role failed", "error", err.Error())
		}
	}

	claims := port.TokenClaims{
		Subject:     strconv.FormatInt(user.ID, 10),
		SessionID:   sessionID,
		RoleID:      roleID,
		MerchantID:  merchantID,
		BranchID:    branchID,
		UserID:      user.ID,
		UserType:    user.Type,
		RoleName:    systemRoleName,
		Permissions: permissions,
	}

	accessToken, err := uc.tokenSigner.SignAccessToken(claims)
	if err != nil {
		return nil, domain.ErrInternal
	}

	uc.auditLogin(ctx, user.ID, input.Username, true, input.IPAddress, input.UserAgent, "")

	uc.cacheSet(ctx, user)

	return &LoginResult{
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenStr,
			ExpiresIn:    3600,
		},
	}, nil
}

func (uc *authUsecase) RefreshToken(ctx context.Context, input RefreshTokenParams) (*LoginResult, error) {
	userID, sessionID, err := uc.tokenSigner.VerifyRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	session, err := uc.sessionStore.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, domain.ErrInvalidRefreshToken
		}
		logger.Error("refresh: get session failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	if session.RefreshTokenHash != uc.tokenHasher.HashToken(input.RefreshToken) {
		return nil, domain.ErrInvalidRefreshToken
	}

	if session.RevokedAt != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, domain.ErrInvalidRefreshToken
	}

	if session.UserID != userID {
		return nil, domain.ErrInvalidRefreshToken
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidRefreshToken
		}
		logger.Error("refresh: get user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	if user.Status == domain.UserStatusDisabled {
		return nil, domain.ErrUserDisabled
	}

	systemRoleName := ""
	if user.SystemRole != nil {
		systemRoleName = user.SystemRole.Name
	}

	permissions, err := uc.collectPermissions(ctx, systemRoleName)
	if err != nil {
		logger.Error("refresh: collect permissions failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	now := time.Now().UTC()
	newRefreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	session.RefreshTokenHash = uc.tokenHasher.HashToken(newRefreshTokenStr)
	session.LastUsedAt = now
	session.UpdatedAt = now
	if err := uc.sessionStore.Update(ctx, session); err != nil {
		logger.Error("refresh: update session failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	roleID, merchantID, branchID := uc.resolveActiveRole(ctx, user.ID)

	accessToken, err := uc.tokenSigner.SignAccessToken(port.TokenClaims{
		Subject:     strconv.FormatInt(user.ID, 10),
		SessionID:   sessionID,
		RoleID:      roleID,
		MerchantID:  merchantID,
		BranchID:    branchID,
		UserID:      user.ID,
		UserType:    user.Type,
		RoleName:    systemRoleName,
		Permissions: permissions,
	})
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &LoginResult{
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: newRefreshTokenStr,
			ExpiresIn:    3600,
		},
	}, nil
}

func (uc *authUsecase) Logout(ctx context.Context, input LogoutParams) error {
	return uc.sessionStore.Delete(ctx, input.SessionID)
}

func (uc *authUsecase) SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*AuthResult, error) {
	session, err := uc.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrForbidden
		}
		logger.Error("switch context: get user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	var found *domain.UserRoleAssignment
	for i := range user.Roles {
		r := &user.Roles[i]
		if r.Role.ID == userRoleID {
			found = r
			break
		}
	}
	if found == nil {
		return nil, domain.ErrForbidden
	}

	session.ActiveUserRoleID = found.Role.ID
	session.UpdatedAt = time.Now().UTC()
	if err := uc.sessionStore.Update(ctx, session); err != nil {
		return nil, domain.ErrInternal
	}

	systemRoleName := ""
	if user.SystemRole != nil {
		systemRoleName = user.SystemRole.Name
	}

	permissions, err := uc.collectPermissions(ctx, systemRoleName)
	if err != nil {
		logger.Error("switch context: collect permissions failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	var branchID int64
	if found.BranchID != nil {
		branchID = *found.BranchID
	}

	accessToken, err := uc.tokenSigner.SignAccessToken(port.TokenClaims{
		Subject:     strconv.FormatInt(user.ID, 10),
		SessionID:   sessionID,
		RoleID:      found.Role.ID,
		MerchantID:  found.MerchantID,
		BranchID:    branchID,
		UserID:      user.ID,
		UserType:    user.Type,
		RoleName:    systemRoleName,
		Permissions: permissions,
	})
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &AuthResult{
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: "",
			ExpiresIn:    3600,
		},
	}, nil
}

func (uc *authUsecase) SetDefaultRole(ctx context.Context, userID, roleID int64) error {
	return uc.staffRepo.SetDefaultRole(ctx, userID, roleID)
}

func (uc *authUsecase) Introspect(ctx context.Context, tokenString string) (*IntrospectResult, error) {
	claims, err := uc.tokenSigner.VerifyAccessToken(tokenString)
	if err != nil {
		return &IntrospectResult{Active: false}, nil
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user.Status == domain.UserStatusDisabled {
		return &IntrospectResult{Active: false}, nil
	}

	staff, err := uc.staffRepo.ListByUserID(ctx, claims.UserID)
	if err != nil {
		return &IntrospectResult{Active: false}, nil
	}

	return &IntrospectResult{
		Active:      true,
		UserID:      claims.UserID,
		UserType:    user.Type,
		RoleName:    claims.RoleName,
		Staff:       staff,
		Permissions: claims.Permissions,
	}, nil
}

func (uc *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	claims, err := uc.tokenSigner.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := uc.cacheGet(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidToken
		}
		logger.Error("validate token: get user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	if user.Status == domain.UserStatusDisabled {
		return nil, domain.ErrUserDisabled
	}

	if claims.SessionID != "" {
		session, err := uc.sessionStore.Get(ctx, claims.SessionID)
		if err != nil {
			if errors.Is(err, domain.ErrSessionNotFound) {
				return nil, domain.ErrInvalidToken
			}
			logger.Error("validate token: get session failed", "error", err.Error())
			return nil, domain.ErrInternal
		}
		if session.RevokedAt != nil {
			return nil, domain.ErrInvalidToken
		}
		if time.Now().UTC().After(session.ExpiresAt) {
			return nil, domain.ErrInvalidToken
		}
	}

	return claims, nil
}

func (uc *authUsecase) HasPermission(claims *port.TokenClaims, required domain.Permission) bool {
	for _, p := range claims.Permissions {
		if p == required {
			return true
		}
	}
	return false
}

// collectPermissions retrieves permissions for the given role name.
func (uc *authUsecase) collectPermissions(ctx context.Context, roleName string) ([]domain.Permission, error) {
	if roleName == "" {
		return nil, nil
	}
	role, err := uc.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// resolveActiveRole finds the user's active role, returning roleID, merchantID, and branchID.
func (uc *authUsecase) resolveActiveRole(ctx context.Context, userID int64) (roleID, merchantID, branchID int64) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil || len(user.Roles) == 0 {
		return 0, 0, 0
	}
	for _, r := range user.Roles {
		if r.IsDefault {
			if r.BranchID != nil {
				return r.Role.ID, r.MerchantID, *r.BranchID
			}
			return r.Role.ID, r.MerchantID, 0
		}
	}
	r := user.Roles[0]
	if r.BranchID != nil {
		return r.Role.ID, r.MerchantID, *r.BranchID
	}
	return r.Role.ID, r.MerchantID, 0
}

// auditLogin persists a login audit event.
func (uc *authUsecase) auditLogin(ctx context.Context, userID int64, email string, success bool, ip, userAgent, reason string) {
	audit := &domain.LoginAudit{
		UserID:    userID,
		Email:     email,
		Success:   success,
		IPAddress: ip,
		UserAgent: userAgent,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
	}
	uc.loginAuditRepo.Create(ctx, audit)
}

// publishEvent publishes a domain event through the event publisher.
func (uc *authUsecase) publishEvent(ctx context.Context, eventName string, payload interface{}) {
	uc.eventPublisher.Publish(ctx, eventName, payload)
}
