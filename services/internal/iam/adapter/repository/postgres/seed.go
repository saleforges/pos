package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"time"

	"github.com/saleforge/pos/services/pkg/otel"
	"golang.org/x/crypto/argon2"
)

func devPasswordHash(password string) string {
	salt := []byte("pos-dev-salt-1234") // fixed for reproducibility
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	var buf bytes.Buffer
	buf.Write(salt)
	buf.Write(hash)
	return base64.RawStdEncoding.EncodeToString(buf.Bytes())
}

func roleDisplayID(name string) string {
	switch name {
	case "owner":
		return "role_owner"
	case "admin":
		return "role_admin"
	case "supervisor":
		return "role_supervisor"
	case "cashier":
		return "role_cashier"
	case "viewer":
		return "role_viewer"
	default:
		return "role_" + name
	}
}

var defaultPermissions = []string{
	"catalog.read", "catalog.create", "catalog.update", "catalog.delete",
	"sales.create", "sales.read", "sales.update", "sales.delete", "sales.refund",
	"inventory.read", "inventory.write", "inventory.adjust",
	"merchant.manage",
	"user.create", "user.read", "user.update", "user.delete", "user.list",
	"role.create", "role.read", "role.update", "role.delete", "role.assign",
	"permission.create", "permission.read", "permission.update", "permission.delete", "permission.assign",
	"session.manage", "apikey.manage",
	"audit.view",
}

var defaultRoles = []struct {
	Name        string
	Description string
	Permissions []string
}{
	{
		Name: "superadmin", Description: "Platform super administrator",
		Permissions: defaultPermissions,
	},
	{
		Name: "owner", Description: "Full system access",
		Permissions: defaultPermissions,
	},
	{
		Name: "admin", Description: "Administrative access",
		Permissions: defaultPermissions,
	},
	{
		Name: "supervisor", Description: "Supervisory access",
		Permissions: []string{
			"catalog.read", "catalog.create", "catalog.update",
			"sales.create", "sales.read", "sales.update", "sales.refund",
			"inventory.read", "inventory.write", "inventory.adjust",
			"user.read", "user.list",
		},
	},
	{
		Name: "cashier", Description: "Cashier access",
		Permissions: []string{
			"catalog.read",
			"sales.create", "sales.read",
			"inventory.read",
		},
	},
	{
		Name: "viewer", Description: "Read-only access",
		Permissions: []string{
			"catalog.read",
			"sales.read",
			"inventory.read",
			"user.read", "user.list",
		},
	},
}

func SeedData(ctx context.Context, pool *otel.TracedPool) error {
	now := time.Now().UTC()

	for _, p := range defaultPermissions {
		_, err := pool.Exec(ctx,
			`INSERT INTO permissions (name, created_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			p, now,
		)
		if err != nil {
			return err
		}
	}

	for _, r := range defaultRoles {
		_, err := pool.Exec(ctx,
			`INSERT INTO roles (id, display_id, name, description, is_system, created_at, updated_at) VALUES (gen_random_uuid(), $1, $2, $3, true, $4, $4) ON CONFLICT DO NOTHING`,
			roleDisplayID(r.Name), r.Name, r.Description, now,
		)
		if err != nil {
			return err
		}

		for _, p := range r.Permissions {
			_, err = pool.Exec(ctx,
				`INSERT INTO role_permissions (role_id, permission_name) SELECT id, $2 FROM roles WHERE name = $1 ON CONFLICT DO NOTHING`,
				r.Name, p,
			)
			if err != nil {
				return err
			}
		}
	}

	seedUsers := []struct {
		ID       string
		Username string
		Email    string
		Password string
		Type     string
		Roles    []string
	}{
		{
			ID: "usr_superadmin", Username: "superadmin", Email: "superadmin@pos.com",
			Password: devPasswordHash("Admin123"),
			Type:     "platform",
			Roles:    []string{"superadmin"},
		},
		{
			ID: "usr_owner", Username: "owner", Email: "owner@merchant.com",
			Password: devPasswordHash("Owner123"),
			Type:     "merchant",
			Roles:    []string{"owner"},
		},
	}

	for _, u := range seedUsers {
		_, err := pool.Exec(ctx,
			`INSERT INTO users (id, username, email, password, type, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, 'active', $6, $6) ON CONFLICT DO NOTHING`,
			u.ID, u.Username, u.Email, u.Password, u.Type, now,
		)
		if err != nil {
			return err
		}

		for _, role := range u.Roles {
			_, err = pool.Exec(ctx,
				`INSERT INTO user_roles (user_id, role_id) SELECT $1, id FROM roles WHERE name = $2 ON CONFLICT DO NOTHING`,
				u.ID, role,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
