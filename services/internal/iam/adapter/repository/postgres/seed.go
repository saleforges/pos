package postgres

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/pkg/otel"
)

var defaultPermissions = []string{
	"catalog.read", "catalog.create", "catalog.update", "catalog.delete",
	"sales.create", "sales.read", "sales.update", "sales.delete", "sales.refund",
	"inventory.read", "inventory.write", "inventory.adjust",
	"merchant.manage",
	"user.create", "user.read", "user.update", "user.delete", "user.list",
	"role.create", "role.read", "role.update", "role.delete", "role.assign",
	"permission.create", "permission.read", "permission.update", "permission.delete", "permission.assign",
	"audit.view",
}

var defaultRoles = []struct {
	Name        string
	Description string
	Permissions []string
}{
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
			`INSERT INTO roles (name, description, is_system, created_at, updated_at) VALUES ($1, $2, true, $3, $3) ON CONFLICT DO NOTHING`,
			r.Name, r.Description, now,
		)
		if err != nil {
			return err
		}

		for _, p := range r.Permissions {
			_, err = pool.Exec(ctx,
				`INSERT INTO role_permissions (role_name, permission_name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				r.Name, p,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
