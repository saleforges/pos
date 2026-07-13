package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[int64]*domain.User
	seq   int64
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int64]*domain.User),
	}
}

func (r *UserRepository) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	user.ID = r.seq
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetByID(_ context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) List(_ context.Context, offset, limit int) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		all = append(all, *u)
	}
	if offset >= len(all) {
		return []domain.User{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *UserRepository) Update(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *UserRepository) AddRole(_ context.Context, userID int64, roleName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	for _, rl := range user.Roles {
		if rl.Role.Name == roleName && rl.MerchantID == 0 {
			return nil
		}
	}
	user.Roles = append(user.Roles, domain.UserRoleAssignment{Role: domain.Role{Name: roleName}})
	return nil
}

func (r *UserRepository) RemoveRole(_ context.Context, userID int64, roleName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	roles := make([]domain.UserRoleAssignment, 0, len(user.Roles))
	for _, rl := range user.Roles {
		if rl.Role.Name != roleName || rl.MerchantID != 0 {
			roles = append(roles, rl)
		}
	}
	user.Roles = roles
	return nil
}

type RoleRepository struct {
	mu       sync.RWMutex
	roles    map[int64]*domain.Role
	nameToID map[string]int64
	seq      int64
}

var _ repository.RoleRepository = (*RoleRepository)(nil)

func NewRoleRepository() *RoleRepository {
	r := &RoleRepository{
		roles:    make(map[int64]*domain.Role),
		nameToID: make(map[string]int64),
	}
	for name, role := range domain.DefaultRoles {
		cp := role
		r.seq++
		cp.ID = r.seq
		r.roles[cp.ID] = &cp
		r.nameToID[name] = cp.ID
	}
	return r
}

func (r *RoleRepository) Create(_ context.Context, role *domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	role.ID = r.seq
	r.roles[role.ID] = role
	r.nameToID[role.Name] = role.ID
	return nil
}

func (r *RoleRepository) GetByID(_ context.Context, id int64) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, ok := r.roles[id]
	if !ok {
		return nil, domain.ErrInvalidRole
	}
	return role, nil
}

func (r *RoleRepository) GetByName(_ context.Context, name string) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.nameToID[name]
	if !ok {
		return nil, domain.ErrInvalidRole
	}
	role, ok := r.roles[id]
	if !ok {
		return nil, domain.ErrInvalidRole
	}
	return role, nil
}

func (r *RoleRepository) List(_ context.Context, _ *int64) ([]domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Role, 0, len(r.roles))
	for _, role := range r.roles {
		result = append(result, *role)
	}
	return result, nil
}

func (r *RoleRepository) Update(_ context.Context, role *domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if role.IsSystem {
		return nil
	}
	r.roles[role.ID] = role
	return nil
}

func (r *RoleRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	role, ok := r.roles[id]
	if !ok {
		return domain.ErrInvalidRole
	}
	if role.IsSystem {
		return nil
	}
	delete(r.nameToID, role.Name)
	delete(r.roles, id)
	return nil
}

func (r *RoleRepository) AddPermission(_ context.Context, roleID int64, permission domain.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	role, ok := r.roles[roleID]
	if !ok {
		return domain.ErrInvalidRole
	}
	for _, p := range role.Permissions {
		if p == permission {
			return nil
		}
	}
	role.Permissions = append(role.Permissions, permission)
	return nil
}

func (r *RoleRepository) RemovePermission(_ context.Context, roleID int64, permission domain.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	role, ok := r.roles[roleID]
	if !ok {
		return domain.ErrInvalidRole
	}
	perms := make([]domain.Permission, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		if p != permission {
			perms = append(perms, p)
		}
	}
	role.Permissions = perms
	return nil
}

func (r *RoleRepository) GetPermissions(_ context.Context, roleID int64) ([]domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	role, ok := r.roles[roleID]
	if !ok {
		return nil, domain.ErrInvalidRole
	}
	return role.Permissions, nil
}

type PermissionRepository struct {
	mu          sync.RWMutex
	permissions map[domain.Permission]bool
}

var _ repository.PermissionRepository = (*PermissionRepository)(nil)

func NewPermissionRepository() *PermissionRepository {
	r := &PermissionRepository{
		permissions: make(map[domain.Permission]bool),
	}
	for _, role := range domain.DefaultRoles {
		for _, p := range role.Permissions {
			r.permissions[p] = true
		}
	}
	return r
}

func (r *PermissionRepository) Create(_ context.Context, permission domain.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[permission] = true
	return nil
}

func (r *PermissionRepository) GetAll(_ context.Context) ([]domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Permission, 0, len(r.permissions))
	for p := range r.permissions {
		result = append(result, p)
	}
	return result, nil
}

func (r *PermissionRepository) Delete(_ context.Context, permission domain.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.permissions, permission)
	return nil
}

type RefreshTokenRepository struct {
	mu     sync.RWMutex
	tokens map[int64]*domain.RefreshToken
	seq    int64
}

var _ repository.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{
		tokens: make(map[int64]*domain.RefreshToken),
	}
}

func (r *RefreshTokenRepository) Create(_ context.Context, token *domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	token.ID = r.seq
	r.tokens[token.ID] = token
	return nil
}

func (r *RefreshTokenRepository) GetByToken(_ context.Context, token string) (*domain.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, domain.ErrInvalidRefreshToken
}

func (r *RefreshTokenRepository) Revoke(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if !ok {
		return domain.ErrInvalidRefreshToken
	}
	now := nowFunc()
	t.RevokedAt = &now
	return nil
}

func (r *RefreshTokenRepository) RevokeByUser(_ context.Context, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := nowFunc()
	for _, t := range r.tokens {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (r *RefreshTokenRepository) HasActiveTokens(_ context.Context, userID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := nowFunc()
	for _, t := range r.tokens {
		if t.UserID == userID && t.RevokedAt == nil && t.ExpiresAt.After(now) {
			return true, nil
		}
	}
	return false, nil
}

type LoginAuditRepository struct {
	mu     sync.RWMutex
	audits []domain.LoginAudit
}

var _ repository.LoginAuditRepository = (*LoginAuditRepository)(nil)

func NewLoginAuditRepository() *LoginAuditRepository {
	return &LoginAuditRepository{}
}

var nowFunc = func() time.Time {
	return time.Now().UTC()
}

func (r *LoginAuditRepository) Create(_ context.Context, audit *domain.LoginAudit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.audits = append(r.audits, *audit)
	return nil
}

func (r *LoginAuditRepository) List(_ context.Context, offset, limit int) ([]domain.LoginAudit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if offset >= len(r.audits) {
		return []domain.LoginAudit{}, nil
	}
	end := offset + limit
	if end > len(r.audits) {
		end = len(r.audits)
	}
	return r.audits[offset:end], nil
}
