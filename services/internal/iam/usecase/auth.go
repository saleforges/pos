package usecase

import (
	"context"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type AuthUsecase struct {
	userRepo         repository.UserRepository
	roleRepo         repository.RoleRepository
	permissionRepo   repository.PermissionRepository
	refreshTokenRepo repository.RefreshTokenRepository
	loginAuditRepo   repository.LoginAuditRepository
	eventPublisher   port.EventPublisher
	passwordHasher   port.PasswordHasher
	tokenSigner      port.TokenSigner
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	loginAuditRepo repository.LoginAuditRepository,
	eventPublisher port.EventPublisher,
	passwordHasher port.PasswordHasher,
	tokenSigner port.TokenSigner,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		permissionRepo:   permissionRepo,
		refreshTokenRepo: refreshTokenRepo,
		loginAuditRepo:   loginAuditRepo,
		eventPublisher:   eventPublisher,
		passwordHasher:   passwordHasher,
		tokenSigner:      tokenSigner,
	}
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*LoginResult, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*LoginResult, error)
	Logout(ctx context.Context, input LogoutInput) error
	Introspect(ctx context.Context, tokenString string) (*IntrospectResult, error)
	ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error)
	HasPermission(claims *port.TokenClaims, required domain.Permission) bool

	ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, input UpdateUserInput) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error

	ListRoles(ctx context.Context) ([]domain.Role, error)
	CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error)
	GetRole(ctx context.Context, name string) (*domain.Role, error)
	UpdateRole(ctx context.Context, input UpdateRoleInput) (*domain.Role, error)
	DeleteRole(ctx context.Context, name string) error

	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	CreatePermission(ctx context.Context, permission domain.Permission) error
	DeletePermission(ctx context.Context, permission domain.Permission) error

	AssignRole(ctx context.Context, userID, roleName string) error
	RemoveRole(ctx context.Context, userID, roleName string) error

	AssignPermission(ctx context.Context, roleName string, permission domain.Permission) error
	RemovePermission(ctx context.Context, roleName string, permission domain.Permission) error
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
	Roles    []string
}

type RegisterResult struct {
	User domain.User
	port.TokenPair
}

type AuthResult struct {
	User  domain.User
	Token string
}

type LoginInput struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginResult struct {
	User domain.User
	port.TokenPair
}

type RefreshTokenInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

type LogoutInput struct {
	RefreshToken string
	UserID       string
}

type IntrospectResult struct {
	Active      bool                `json:"active"`
	UserID      string              `json:"user_id,omitempty"`
	Roles       []string            `json:"roles,omitempty"`
	Permissions []domain.Permission `json:"permissions,omitempty"`
}

type UpdateUserInput struct {
	ID       string
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
	Name        string
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
	user := &domain.User{
		ID:        generateID(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashed),
		Roles:     input.Roles,
		Status:    domain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		logger.Error("register: create user failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	permissions, err := uc.collectPermissions(ctx, user.Roles)
	if err != nil {
		logger.Error("register: collect permissions failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	token, err := uc.tokenSigner.SignAccessToken(port.TokenClaims{
		UserID:      user.ID,
		Roles:       user.Roles,
		Permissions: permissions,
	})
	if err != nil {
		logger.Error("register: sign token failed", "error", err.Error())
		return nil, domain.ErrInternal
	}

	uc.publishEvent(ctx, "UserCreated", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"roles":    user.Roles,
	})

	return &AuthResult{User: *user, Token: token}, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	ctx, span := otel.StartSpan(ctx, "auth.Login")
	defer span.End()

	user, err := uc.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		uc.auditLogin(ctx, "", input.Username, false, input.IPAddress, input.UserAgent, "user not found")
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

	permissions, err := uc.collectPermissions(ctx, user.Roles)
	if err != nil {
		return nil, domain.ErrInternal
	}

	claims := port.TokenClaims{
		UserID:      user.ID,
		Roles:       user.Roles,
		Permissions: permissions,
	}

	accessToken, err := uc.tokenSigner.SignAccessToken(claims)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	now := time.Now().UTC()
	rt := &domain.RefreshToken{
		ID:        generateID(),
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: now.Add(30 * 24 * time.Hour),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, domain.ErrInternal
	}

	uc.auditLogin(ctx, user.ID, input.Username, true, input.IPAddress, input.UserAgent, "")

	return &LoginResult{
		User: *user,
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenStr,
			ExpiresIn:    900,
		},
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, input RefreshTokenInput) (*LoginResult, error) {
	userID, err := uc.tokenSigner.VerifyRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	stored, err := uc.refreshTokenRepo.GetByToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if stored.UserID != userID {
		return nil, domain.ErrInvalidRefreshToken
	}

	if stored.RevokedAt != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if time.Now().UTC().After(stored.ExpiresAt) {
		return nil, domain.ErrInvalidRefreshToken
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	if user.Status == domain.UserStatusDisabled {
		return nil, domain.ErrUserDisabled
	}

	now := time.Now().UTC()
	stored.RevokedAt = &now
	uc.refreshTokenRepo.Revoke(ctx, stored.ID)

	permissions, err := uc.collectPermissions(ctx, user.Roles)
	if err != nil {
		return nil, domain.ErrInternal
	}

	claims := port.TokenClaims{
		UserID:      user.ID,
		Roles:       user.Roles,
		Permissions: permissions,
	}

	accessToken, err := uc.tokenSigner.SignAccessToken(claims)
	if err != nil {
		return nil, domain.ErrInternal
	}

	newRefreshTokenStr, err := uc.tokenSigner.SignRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	rt := &domain.RefreshToken{
		ID:        generateID(),
		UserID:    user.ID,
		Token:     newRefreshTokenStr,
		ExpiresAt: now.Add(30 * 24 * time.Hour),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, domain.ErrInternal
	}

	uc.publishEvent(ctx, "PasswordChanged", nil)

	return &LoginResult{
		User: *user,
		TokenPair: port.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: newRefreshTokenStr,
			ExpiresIn:    900,
		},
	}, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken != "" {
		stored, err := uc.refreshTokenRepo.GetByToken(ctx, input.RefreshToken)
		if err == nil && stored.UserID == input.UserID {
			now := time.Now().UTC()
			stored.RevokedAt = &now
			uc.refreshTokenRepo.Revoke(ctx, stored.ID)
		}
	}

	uc.refreshTokenRepo.RevokeByUser(ctx, input.UserID)
	return nil
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

	return &IntrospectResult{
		Active:      true,
		UserID:      claims.UserID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}

func (uc *AuthUsecase) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	claims, err := uc.tokenSigner.VerifyAccessToken(tokenString)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if user.Status == domain.UserStatusDisabled {
		return nil, domain.ErrUserDisabled
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

func (uc *AuthUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
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

	uc.publishEvent(ctx, "UserUpdated", map[string]interface{}{
		"user_id": user.ID,
	})

	return user, nil
}

func (uc *AuthUsecase) DeleteUser(ctx context.Context, id string) error {
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return domain.ErrUserNotFound
	}

	uc.publishEvent(ctx, "UserDeleted", map[string]interface{}{
		"user_id": id,
	})

	return nil
}

func (uc *AuthUsecase) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return uc.roleRepo.List(ctx)
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

func (uc *AuthUsecase) GetRole(ctx context.Context, name string) (*domain.Role, error) {
	return uc.roleRepo.GetByName(ctx, name)
}

func (uc *AuthUsecase) UpdateRole(ctx context.Context, input UpdateRoleInput) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByName(ctx, input.Name)
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

func (uc *AuthUsecase) DeleteRole(ctx context.Context, name string) error {
	if _, ok := domain.DefaultRoles[name]; ok {
		return domain.ErrInvalidRole
	}

	return uc.roleRepo.Delete(ctx, name)
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

func (uc *AuthUsecase) AssignRole(ctx context.Context, userID, roleName string) error {
	if _, ok := domain.DefaultRoles[roleName]; !ok {
		if _, err := uc.roleRepo.GetByName(ctx, roleName); err != nil {
			return domain.ErrInvalidRole
		}
	}

	if err := uc.userRepo.AddRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.publishEvent(ctx, "RoleAssigned", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *AuthUsecase) RemoveRole(ctx context.Context, userID, roleName string) error {
	if err := uc.userRepo.RemoveRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.publishEvent(ctx, "RoleRevoked", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *AuthUsecase) AssignPermission(ctx context.Context, roleName string, permission domain.Permission) error {
	return uc.roleRepo.AddPermission(ctx, roleName, permission)
}

func (uc *AuthUsecase) RemovePermission(ctx context.Context, roleName string, permission domain.Permission) error {
	return uc.roleRepo.RemovePermission(ctx, roleName, permission)
}

func (uc *AuthUsecase) collectPermissions(ctx context.Context, roles []string) ([]domain.Permission, error) {
	permSet := make(map[domain.Permission]bool)
	for _, roleName := range roles {
		perms, err := uc.roleRepo.GetPermissions(ctx, roleName)
		if err != nil {
			continue
		}
		for _, p := range perms {
			permSet[p] = true
		}
	}
	result := make([]domain.Permission, 0, len(permSet))
	for p := range permSet {
		result = append(result, p)
	}
	return result, nil
}

func (uc *AuthUsecase) auditLogin(ctx context.Context, userID, email string, success bool, ip, userAgent, reason string) {
	audit := &domain.LoginAudit{
		ID:        generateID(),
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

var idCounter atomic.Int64

func generateID() string {
	idCounter.Add(1)
	b := make([]byte, 8)
	now := time.Now().UnixNano()
	for i := 0; i < 8; i++ {
		b[i] = byte(now >> (i * 8))
	}
	return formatID(b)
}

func formatID(b []byte) string {
	const hex = "0123456789abcdef"
	buf := make([]byte, 16)
	for i, v := range b {
		buf[i*2] = hex[v>>4]
		buf[i*2+1] = hex[v&0x0f]
	}
	return string(buf)
}
