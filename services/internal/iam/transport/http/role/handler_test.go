package role

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type mockRoleService struct {
	listRolesFn    func(ctx context.Context, merchantID *int64) ([]domain.Role, error)
	createRoleFn   func(ctx context.Context, params usecase.CreateRoleParams) (*domain.Role, error)
	getRoleFn      func(ctx context.Context, id int64) (*domain.Role, error)
	updateRoleFn   func(ctx context.Context, params usecase.UpdateRoleParams) (*domain.Role, error)
	deleteRoleFn   func(ctx context.Context, id int64) error
	assignRoleFn   func(ctx context.Context, userID int64, roleName string) error
	removeRoleFn   func(ctx context.Context, userID int64, roleName string) error
	assignPermFn   func(ctx context.Context, roleID int64, permission domain.Permission) error
	removePermFn   func(ctx context.Context, roleID int64, permission domain.Permission) error
}

func (m *mockRoleService) ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	if m.listRolesFn != nil { return m.listRolesFn(ctx, merchantID) }
	return []domain.Role{{ID: 1, Name: "admin"}}, nil
}
func (m *mockRoleService) CreateRole(ctx context.Context, params usecase.CreateRoleParams) (*domain.Role, error) {
	if m.createRoleFn != nil { return m.createRoleFn(ctx, params) }
	return &domain.Role{ID: 1, Name: params.Name, Description: params.Description}, nil
}
func (m *mockRoleService) GetRole(ctx context.Context, id int64) (*domain.Role, error) {
	if m.getRoleFn != nil { return m.getRoleFn(ctx, id) }
	return &domain.Role{ID: id, Name: "viewer"}, nil
}
func (m *mockRoleService) UpdateRole(ctx context.Context, params usecase.UpdateRoleParams) (*domain.Role, error) {
	if m.updateRoleFn != nil { return m.updateRoleFn(ctx, params) }
	return &domain.Role{ID: params.ID, Description: *params.Description}, nil
}
func (m *mockRoleService) DeleteRole(ctx context.Context, id int64) error {
	if m.deleteRoleFn != nil { return m.deleteRoleFn(ctx, id) }
	return nil
}
func (m *mockRoleService) AssignRole(ctx context.Context, userID int64, roleName string) error {
	if m.assignRoleFn != nil { return m.assignRoleFn(ctx, userID, roleName) }
	return nil
}
func (m *mockRoleService) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	if m.removeRoleFn != nil { return m.removeRoleFn(ctx, userID, roleName) }
	return nil
}
func (m *mockRoleService) AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	if m.assignPermFn != nil { return m.assignPermFn(ctx, roleID, permission) }
	return nil
}
func (m *mockRoleService) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	if m.removePermFn != nil { return m.removePermFn(ctx, roleID, permission) }
	return nil
}

func withClaims(c echo.Context) echo.Context {
	c.Set(common.ClaimsKey, &port.TokenClaims{UserID: 1, MerchantID: 0})
	return c
}

func TestListRoles(t *testing.T) {
	t.Parallel()

	t.Run("returns role list", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/roles", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles")
		withClaims(c)

		h.ListRoles(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }

		var wrapped struct { Data []domain.Role `json:"data"` }
		json.Unmarshal(rec.Body.Bytes(), &wrapped)
		if len(wrapped.Data) != 1 { t.Errorf("expected 1 role, got %d", len(wrapped.Data)) }
	})
}

func TestCreateRole(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		body := `{"name":"custom_role","description":"desc","permissions":["user.read"]}`
		req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles")

		h.CreateRole(c)
		if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }

		var wrapped struct { Data domain.Role `json:"data"` }
		json.Unmarshal(rec.Body.Bytes(), &wrapped)
		if wrapped.Data.Name != "custom_role" { t.Errorf("expected name 'custom_role', got %q", wrapped.Data.Name) }
	})

	t.Run("invalid role maps to 409", func(t *testing.T) {
		rs := &mockRoleService{createRoleFn: func(_ context.Context, _ usecase.CreateRoleParams) (*domain.Role, error) {
			return nil, domain.ErrInvalidRole
		}}
		h := NewHandler(rs)
		e := echo.New()
		body := `{"name":"admin","description":"","permissions":[]}`
		req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles")

		h.CreateRole(c)
		if rec.Code != http.StatusConflict { t.Errorf("expected 409, got %d", rec.Code) }
	})
}

func TestGetRole(t *testing.T) {
	t.Parallel()

	t.Run("existing role returns 200", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.GetRole(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("non-existent role returns 404", func(t *testing.T) {
		rs := &mockRoleService{getRoleFn: func(_ context.Context, _ int64) (*domain.Role, error) {
			return nil, domain.ErrInvalidRole
		}}
		h := NewHandler(rs)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/roles/999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id")
		c.SetParamNames("id")
		c.SetParamValues("999")

		h.GetRole(c)
		if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
	})
}

func TestUpdateRole(t *testing.T) {
	t.Parallel()

	h := NewHandler(&mockRoleService{})
	e := echo.New()
	body := `{"description":"updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/roles/1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	h.UpdateRole(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestDeleteRole(t *testing.T) {
	t.Parallel()

	h := NewHandler(&mockRoleService{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/roles/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	h.DeleteRole(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestAssignAndRemoveRole(t *testing.T) {
	t.Parallel()

	t.Run("assign role with invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/users/1/roles", strings.NewReader(`invalid`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id/roles")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.AssignRole(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("assign role returns 200", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		body := `{"role":"viewer"}`
		req := httptest.NewRequest(http.MethodPost, "/users/1/roles", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id/roles")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.AssignRole(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("remove role returns 200", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/users/1/roles/viewer", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id/roles/:roleId")
		c.SetParamNames("id", "roleId")
		c.SetParamValues("1", "viewer")

		h.RemoveRole(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("create role with invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{bad`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles")

		h.CreateRole(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("update role with invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/roles/1", strings.NewReader(`{bad`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.UpdateRole(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("invalid role in assign maps to 404", func(t *testing.T) {
		rs := &mockRoleService{assignRoleFn: func(_ context.Context, _ int64, _ string) error {
			return domain.ErrInvalidRole
		}}
		h := NewHandler(rs)
		e := echo.New()
		body := `{"role":"nonexistent"}`
		req := httptest.NewRequest(http.MethodPost, "/users/1/roles", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id/roles")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.AssignRole(c)
		if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
	})
}

func TestAssignAndRemovePermission(t *testing.T) {
	t.Parallel()

	t.Run("assign permission with invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/roles/1/permissions", strings.NewReader(`bad`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id/permissions")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.AssignPermission(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("assign permission returns 200", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		body := `{"permission":"user.read"}`
		req := httptest.NewRequest(http.MethodPost, "/roles/1/permissions", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id/permissions")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.AssignPermission(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("remove permission returns 200", func(t *testing.T) {
		h := NewHandler(&mockRoleService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/roles/1/permissions/user.read", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/roles/:id/permissions/:permissionId")
		c.SetParamNames("id", "permissionId")
		c.SetParamValues("1", "user.read")

		h.RemovePermission(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})
}
