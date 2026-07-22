package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
)

type mockUserRepo struct {
	users map[int64]*domain.User
	seq   int64
	err   error
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if m.err != nil { return m.err }
	if m.users == nil { m.users = make(map[int64]*domain.User) }
	m.seq++
	user.ID = m.seq
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	if m.err != nil { return nil, m.err }
	u, ok := m.users[id]
	if !ok { return nil, domain.ErrUserNotFound }
	return u, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if m.err != nil { return nil, m.err }
	for _, u := range m.users {
		if u.Username == username { return u, nil }
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.err != nil { return nil, m.err }
	for _, u := range m.users {
		if u.Email == email { return u, nil }
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]domain.User, int64, error) { return nil, 0, nil }
func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	if m.err != nil { return m.err }
	m.users[user.ID] = user
	return nil
}
func (m *mockUserRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil { return m.err }
	delete(m.users, id)
	return nil
}
func (m *mockUserRepo) AddRole(_ context.Context, userID int64, roleName string) error { return nil }
func (m *mockUserRepo) RemoveRole(_ context.Context, userID int64, roleName string) error { return nil }

type mockRoleRepo struct {
	roles map[int64]*domain.Role
	err   error
}

func (m *mockRoleRepo) Create(_ context.Context, role *domain.Role) error {
	if m.err != nil { return m.err }
	if m.roles == nil { m.roles = make(map[int64]*domain.Role) }
	role.ID = int64(len(m.roles) + 1)
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleRepo) GetByID(_ context.Context, id int64) (*domain.Role, error) {
	if m.err != nil { return nil, m.err }
	if r, ok := m.roles[id]; ok { return r, nil }
	return nil, domain.ErrInvalidRole
}

func (m *mockRoleRepo) GetByName(_ context.Context, name string) (*domain.Role, error) {
	if m.err != nil { return nil, m.err }
	for _, r := range m.roles {
		if r.Name == name { return r, nil }
	}
	return nil, domain.ErrInvalidRole
}

func (m *mockRoleRepo) List(_ context.Context, _ *int64) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepo) Update(_ context.Context, role *domain.Role) error {
	if m.err != nil { return m.err }
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) Delete(_ context.Context, id int64) error { return nil }
func (m *mockRoleRepo) AddPermission(_ context.Context, roleID int64, permission domain.Permission) error { return nil }
func (m *mockRoleRepo) RemovePermission(_ context.Context, roleID int64, permission domain.Permission) error { return nil }
func (m *mockRoleRepo) GetPermissions(_ context.Context, roleID int64) ([]domain.Permission, error) {
	if m.err != nil { return nil, m.err }
	r, ok := m.roles[roleID]
	if !ok { return nil, domain.ErrInvalidRole }
	return r.Permissions, nil
}

type mockStaffRepo struct {
	err error
}

func (m *mockStaffRepo) ListByUserID(_ context.Context, _ int64) ([]domain.UserRoleAssignment, error) {
	return nil, m.err
}
func (m *mockStaffRepo) Create(_ context.Context, _ int64, _ int64, _, _ string) error {
	return m.err
}
func (m *mockStaffRepo) SetDefaultRole(_ context.Context, _, _ int64) error {
	return m.err
}

type mockPermissionRepo struct {
	permissions map[domain.Permission]bool
	err         error
}

func (m *mockPermissionRepo) Create(_ context.Context, p domain.Permission) error { return nil }
func (m *mockPermissionRepo) GetAll(_ context.Context) ([]domain.Permission, error) {
	result := make([]domain.Permission, 0, len(m.permissions))
	for p := range m.permissions { result = append(result, p) }
	return result, nil
}
func (m *mockPermissionRepo) Delete(_ context.Context, p domain.Permission) error { return nil }

type mockSessionStore struct {
	sessions map[string]*domain.Session
}

func (m *mockSessionStore) Create(_ context.Context, session *domain.Session) error {
	if m.sessions == nil { m.sessions = make(map[string]*domain.Session) }
	m.sessions[session.ID] = session
	return nil
}
func (m *mockSessionStore) Get(_ context.Context, id string) (*domain.Session, error) {
	if m.sessions == nil { return nil, domain.ErrSessionNotFound }
	s, ok := m.sessions[id]
	if !ok { return nil, domain.ErrSessionNotFound }
	return s, nil
}
func (m *mockSessionStore) Update(_ context.Context, session *domain.Session) error {
	if m.sessions == nil { return domain.ErrSessionNotFound }
	m.sessions[session.ID] = session
	return nil
}
func (m *mockSessionStore) Delete(_ context.Context, id string) error {
	if m.sessions == nil { return domain.ErrSessionNotFound }
	delete(m.sessions, id)
	return nil
}

type mockLoginAuditRepo struct { err error }
func (m *mockLoginAuditRepo) Create(_ context.Context, audit *domain.LoginAudit) error { return nil }
func (m *mockLoginAuditRepo) List(_ context.Context, offset, limit int) ([]domain.LoginAudit, int64, error) { return nil, 0, nil }

type mockEventPublisher struct { err error }
func (m *mockEventPublisher) Publish(_ context.Context, _ string, _ interface{}) error { return nil }

type mockPasswordHasher struct {
	hashErr    error
	compareErr error
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.hashErr != nil { return "", m.hashErr }
	return "hashed:" + password, nil
}
func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	if m.compareErr != nil { return m.compareErr }
	if hashedPassword != "hashed:"+password { return domain.ErrInvalidCredentials }
	return nil
}

type mockTokenSigner struct {
	signedToken      string
	claims           *port.TokenClaims
	signErr          error
	verifyErr        error
	refreshUserID    int64
	refreshSignErr   error
	refreshVerifyErr error
}

func (m *mockTokenSigner) SignAccessToken(_ port.TokenClaims) (string, error) {
	if m.signErr != nil { return "", m.signErr }
	return m.signedToken, nil
}
func (m *mockTokenSigner) SignRefreshToken(_ int64, _ string) (string, error) {
	if m.refreshSignErr != nil { return "", m.refreshSignErr }
	return "refresh:" + m.signedToken, nil
}
func (m *mockTokenSigner) VerifyAccessToken(_ string) (*port.TokenClaims, error) {
	if m.verifyErr != nil { return nil, m.verifyErr }
	return m.claims, nil
}
func (m *mockTokenSigner) VerifyRefreshToken(_ string) (int64, string, error) {
	if m.refreshVerifyErr != nil { return 0, "", m.refreshVerifyErr }
	return m.refreshUserID, "test-session-id", nil
}

func newTestAuthUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordHasher port.PasswordHasher,
	tokenSigner port.TokenSigner,
	sessionStore port.SessionStore,
) *authUsecase {
	if sessionStore == nil { sessionStore = &mockSessionStore{} }
	return NewAuthUsecase(
		userRepo, roleRepo, &mockPermissionRepo{}, &mockLoginAuditRepo{},
		&mockStaffRepo{}, sessionStore, &mockEventPublisher{},
		passwordHasher, tokenSigner, &mockTokenHasher{}, nil, nil,
	)
}

type mockTokenHasher struct{}
func (m *mockTokenHasher) HashToken(token string) string {
	if token == "" { return "" }
	return "hashed:" + token
}
