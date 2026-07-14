package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type AuthUsecase struct {
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	loginAuditRepo repository.LoginAuditRepository
	staffRepo      repository.StaffRepository
	sessionStore   port.SessionStore
	eventPublisher port.EventPublisher
	passwordHasher port.PasswordHasher
	tokenSigner    port.TokenSigner
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
	userCache port.UserCache,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		loginAuditRepo: loginAuditRepo,
		staffRepo:      staffRepo,
		sessionStore:   sessionStore,
		eventPublisher: eventPublisher,
		passwordHasher: passwordHasher,
		tokenSigner:    tokenSigner,
		userCache:      userCache,
	}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*LoginResult, error)
	Logout(ctx context.Context, input LogoutInput) error
	Introspect(ctx context.Context, tokenString string) (*IntrospectResult, error)
	ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error)
	HasPermission(claims *port.TokenClaims, required domain.Permission) bool

	ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error)

	ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error)
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, input UpdateUserInput) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error

	ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error)
	CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error)
	GetRole(ctx context.Context, id int64) (*domain.Role, error)
	UpdateRole(ctx context.Context, input UpdateRoleInput) (*domain.Role, error)
	DeleteRole(ctx context.Context, id int64) error

	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	CreatePermission(ctx context.Context, permission domain.Permission) error
	DeletePermission(ctx context.Context, permission domain.Permission) error

	AssignRole(ctx context.Context, userID int64, roleName string) error
	RemoveRole(ctx context.Context, userID int64, roleName string) error

	AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error
	RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error
}

type RegisterInput struct {
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

type LoginInput struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginResult struct {
	port.TokenPair
}

type RefreshTokenInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

type LogoutInput struct {
	SessionID string
}

type IntrospectResult struct {
	Active      bool                      `json:"active"`
	UserID      int64                     `json:"user_id,omitempty"`
	UserType    domain.UserType           `json:"user_type"`
	RoleName    string                    `json:"role_name,omitempty"`
	Staff       []domain.UserRoleAssignment `json:"staff,omitempty"`
	Permissions []domain.Permission       `json:"permissions,omitempty"`
}

type UpdateUserInput struct {
	ID       int64
	Username *string
	Email    *string
	Status   *domain.UserStatus
}

type CreateRoleInput struct {
	Name        string
	Description string
	Permissions []domain.Permission
}

type UpdateRoleInput struct {
	ID          int64
	Description *string
}

var (
	passwordLower = regexp.MustCompile(`[a-z]`)
	passwordUpper = regexp.MustCompile(`[A-Z]`)
	passwordDigit = regexp.MustCompile(`[0-9]`)
)

func validatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordPolicy
	}
	if !passwordLower.MatchString(password) {
		return domain.ErrPasswordPolicy
	}
	if !passwordUpper.MatchString(password) {
		return domain.ErrPasswordPolicy
	}
	if !passwordDigit.MatchString(password) {
		return domain.ErrPasswordPolicy
	}
	return nil
}

func (uc *AuthUsecase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	ctx, span := otel.StartSpan(ctx, "auth.Register")
	defer span.End()

	if err := validatePassword(input.Password); err != nil {
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

	existing, _ := uc.userRepo.GetByUsername(ctx, input.Username)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	existing, _ = uc.userRepo.GetByEmail(ctx, input.Email)
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
		uc.userRepo.AddRole(ctx, user.ID, r)
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

	permissions, _ := uc.collectPermissions(ctx, systemRoleName)

	sessionID := uuid.New().String()
	refreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID, sessionID)
	if err != nil {
		logger.Error("register: sign refresh token failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	session := &domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshTokenStr),
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

func (uc *AuthUsecase) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	ctx, span := otel.StartSpan(ctx, "auth.Login")
	defer span.End()

	user, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		uc.auditLogin(ctx, 0, input.Username, false, input.IPAddress, input.UserAgent, "user not found")
		return nil, domain.ErrInvalidCredentials
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
		RefreshTokenHash: hashToken(refreshTokenStr),
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
		uc.sessionStore.Update(ctx, session)
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

func (uc *AuthUsecase) RefreshToken(ctx context.Context, input RefreshTokenInput) (*LoginResult, error) {
	userID, sessionID, err := uc.tokenSigner.VerifyRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	session, err := uc.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if session.RefreshTokenHash != hashToken(input.RefreshToken) {
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
		return nil, domain.ErrInvalidRefreshToken
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
		return nil, domain.ErrInternal
	}

	now := time.Now().UTC()
	newRefreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	session.RefreshTokenHash = hashToken(newRefreshTokenStr)
	session.LastUsedAt = now
	session.UpdatedAt = now
	if err := uc.sessionStore.Update(ctx, session); err != nil {
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

func (uc *AuthUsecase) Logout(ctx context.Context, input LogoutInput) error {
	return uc.sessionStore.Delete(ctx, input.SessionID)
}

func (uc *AuthUsecase) Introspect(ctx context.Context, tokenString string) (*IntrospectResult, error) {
	claims, err := uc.tokenSigner.VerifyAccessToken(tokenString)
	if err != nil {
		if strings.Contains(err.Error(), "AUTH003") {
			return &IntrospectResult{Active: false}, nil
		}
		return &IntrospectResult{Active: false}, nil
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user.Status == domain.UserStatusDisabled {
		return &IntrospectResult{Active: false}, nil
	}

	staff, _ := uc.staffRepo.ListByUserID(ctx, claims.UserID)

	return &IntrospectResult{
		Active:      true,
		UserID:      claims.UserID,
		UserType:    user.Type,
		RoleName:    claims.RoleName,
		Staff:       staff,
		Permissions: claims.Permissions,
	}, nil
}

func (uc *AuthUsecase) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	claims, err := uc.tokenSigner.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := uc.cacheGet(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if user.Status == domain.UserStatusDisabled {
		return nil, domain.ErrUserDisabled
	}

	if claims.SessionID != "" {
		session, err := uc.sessionStore.Get(ctx, claims.SessionID)
		if err != nil {
			return nil, domain.ErrInvalidToken
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

func (uc *AuthUsecase) HasPermission(claims *port.TokenClaims, required domain.Permission) bool {
	for _, p := range claims.Permissions {
		if p == required {
			return true
		}
	}
	return false
}

func (uc *AuthUsecase) ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error) {
	return uc.userRepo.List(ctx, offset, limit)
}

func (uc *AuthUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *AuthUsecase) UpdateUser(ctx context.Context, input UpdateUserInput) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Status != nil {
		user.Status = *input.Status
	}

	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, domain.ErrInternal
	}

	uc.cacheDel(ctx, user.ID)

	uc.publishEvent(ctx, "UserUpdated", map[string]interface{}{
		"user_id": user.ID,
	})

	return user, nil
}

func (uc *AuthUsecase) DeleteUser(ctx context.Context, id int64) error {
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return domain.ErrUserNotFound
	}

	uc.cacheDel(ctx, id)

	uc.publishEvent(ctx, "UserDeleted", map[string]interface{}{
		"user_id": id,
	})

	return nil
}

func (uc *AuthUsecase) ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	return uc.roleRepo.List(ctx, merchantID)
}

func (uc *AuthUsecase) CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error) {
	if _, ok := domain.DefaultRoles[input.Name]; ok {
		return nil, domain.ErrInvalidRole
	}

	role := &domain.Role{
		Name:        input.Name,
		Description: input.Description,
		Permissions: input.Permissions,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, domain.ErrInternal
	}

	return role, nil
}

func (uc *AuthUsecase) GetRole(ctx context.Context, id int64) (*domain.Role, error) {
	return uc.roleRepo.GetByID(ctx, id)
}

func (uc *AuthUsecase) UpdateRole(ctx context.Context, input UpdateRoleInput) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, domain.ErrInvalidRole
	}

	if input.Description != nil {
		role.Description = *input.Description
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, domain.ErrInternal
	}

	return role, nil
}

func (uc *AuthUsecase) DeleteRole(ctx context.Context, id int64) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrInvalidRole
	}
	if _, ok := domain.DefaultRoles[role.Name]; ok {
		return domain.ErrInvalidRole
	}

	return uc.roleRepo.Delete(ctx, role.ID)
}

func (uc *AuthUsecase) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return uc.permissionRepo.GetAll(ctx)
}

func (uc *AuthUsecase) CreatePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Create(ctx, permission)
}

func (uc *AuthUsecase) DeletePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Delete(ctx, permission)
}

func (uc *AuthUsecase) AssignRole(ctx context.Context, userID int64, roleName string) error {
	if _, ok := domain.DefaultRoles[roleName]; !ok {
		if _, err := uc.roleRepo.GetByName(ctx, roleName); err != nil {
			return domain.ErrInvalidRole
		}
	}

	if err := uc.userRepo.AddRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.cacheDel(ctx, userID)

	uc.publishEvent(ctx, "RoleAssigned", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *AuthUsecase) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	if err := uc.userRepo.RemoveRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.cacheDel(ctx, userID)

	uc.publishEvent(ctx, "RoleRevoked", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *AuthUsecase) AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.AddPermission(ctx, roleID, permission)
}

func (uc *AuthUsecase) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.RemovePermission(ctx, roleID, permission)
}

func (uc *AuthUsecase) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return uc.staffRepo.ListByUserID(ctx, userID)
}

func (uc *AuthUsecase) collectPermissions(ctx context.Context, roleName string) ([]domain.Permission, error) {
	if roleName == "" {
		return nil, nil
	}
	role, err := uc.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return nil, nil
	}
	return role.Permissions, nil
}

func (uc *AuthUsecase) resolveActiveRole(ctx context.Context, userID int64) (roleID, merchantID, branchID int64) {
	staff, err := uc.staffRepo.ListByUserID(ctx, userID)
	if err != nil || len(staff) == 0 {
		return 0, 0, 0
	}
	for _, s := range staff {
		if s.IsDefault {
			if s.BranchID != nil {
				return s.ID, s.MerchantID, *s.BranchID
			}
			return s.ID, s.MerchantID, 0
		}
	}
	s := staff[0]
	if s.BranchID != nil {
		return s.ID, s.MerchantID, *s.BranchID
	}
	return s.ID, s.MerchantID, 0
}

func (uc *AuthUsecase) auditLogin(ctx context.Context, userID int64, email string, success bool, ip, userAgent, reason string) {
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

func (uc *AuthUsecase) publishEvent(ctx context.Context, eventName string, payload interface{}) {
	uc.eventPublisher.Publish(ctx, eventName, payload)
}

func (uc *AuthUsecase) cacheGet(ctx context.Context, id int64) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)

	if uc.userCache != nil {
		if u, ok := uc.userCache.Get(ctx, id); ok {
			span.AddEvent("cache.hit", trace.WithAttributes(
				attribute.Int64("cache.key", id),
			))
			return u, nil
		}
		span.AddEvent("cache.miss", trace.WithAttributes(
			attribute.Int64("cache.key", id),
		))
	}

	u, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if uc.userCache != nil {
		uc.userCache.Set(ctx, u, 0)
	}
	return u, nil
}

func (uc *AuthUsecase) cacheSet(ctx context.Context, u *domain.User) {
	if uc.userCache != nil {
		uc.userCache.Set(ctx, u, 0)
	}
}

func (uc *AuthUsecase) cacheDel(ctx context.Context, id int64) {
	if uc.userCache != nil {
		uc.userCache.Delete(ctx, id)
	}
}
