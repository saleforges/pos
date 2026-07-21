package postgres

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
	"golang.org/x/crypto/argon2"
)

func devPasswordHash(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic("seed: failed to generate random salt: " + err.Error())
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	var buf bytes.Buffer
	buf.Write(salt)
	buf.Write(hash)
	return base64.RawStdEncoding.EncodeToString(buf.Bytes())
}

// defaultPermissions returns the full set of known permissions from domain.DefaultRoles.
func defaultPermissions() []string {
	seen := make(map[string]bool)
	var perms []string
	for _, role := range domain.DefaultRoles {
		for _, p := range role.Permissions {
			s := string(p)
			if !seen[s] {
				seen[s] = true
				perms = append(perms, s)
			}
		}
	}
	return perms
}

// seedRoles produces the role/permission data for seeding, driven from domain.DefaultRoles.
func seedRoles() []struct {
	Name        string
	Description string
	Permissions []string
} {
	var roles []struct {
		Name        string
		Description string
		Permissions []string
	}
	for _, r := range domain.DefaultRoles {
		perms := make([]string, len(r.Permissions))
		for i, p := range r.Permissions {
			perms[i] = string(p)
		}
		roles = append(roles, struct {
			Name        string
			Description string
			Permissions []string
		}{Name: r.Name, Description: r.Description, Permissions: perms})
	}
	return roles
}

func SeedData(ctx context.Context, pool *otel.TracedPool) error {
	now := time.Now().UTC()

	for _, p := range defaultPermissions() {
		_, err := pool.Exec(ctx,
			`INSERT INTO permissions (name, created_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			p, now,
		)
		if err != nil {
			return err
		}
	}

	for _, r := range seedRoles() {
		_, err := pool.Exec(ctx,
			`INSERT INTO roles (name, description, is_system, created_at, updated_at) VALUES ($1, $2, true, $3, $3) ON CONFLICT DO NOTHING`,
			r.Name, r.Description, now,
		)
		if err != nil {
			return err
		}

		for _, p := range r.Permissions {
			_, err = pool.Exec(ctx,
				`INSERT INTO role_permissions (role_id, permission_id) SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = $1 AND p.name = $2 ON CONFLICT DO NOTHING`,
				r.Name, p,
			)
			if err != nil {
				return err
			}
		}
	}

	seedUsers := []struct {
		Username   string
		Email      string
		Password   string
		Type       string
		RoleName   string
		MerchantID *int64
		BranchID   *int64
	}{
		{
			Username: "superadmin", Email: "superadmin@pos.com",
			Password: devPasswordHash("Admin123"),
			Type:     "platform",
			RoleName: "superadmin",
		},
		{
			Username: "owner", Email: "owner@merchant.com",
			Password: devPasswordHash("Owner123"),
			Type:     "merchant",
			RoleName: "owner",
			MerchantID: ptrInt64(1),
		},
		{
			Username: "cashier1", Email: "cashier1@merchant.com",
			Password: devPasswordHash("Cashier123"),
			Type:     "merchant",
			RoleName: "cashier",
			MerchantID: ptrInt64(1),
			BranchID:   ptrInt64(1),
		},
	}

	for _, u := range seedUsers {
		_, err := pool.Exec(ctx,
			`INSERT INTO users (username, email, password, type, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'active', $5, $5)
			 ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password, updated_at = NOW()`,
			u.Username, u.Email, u.Password, u.Type, now,
		)
		if err != nil {
			return err
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id, merchant_id, branch_id)
			 SELECT u.id, r.id, $3, $4 FROM users u, roles r WHERE u.username = $1 AND r.name = $2
			 ON CONFLICT DO NOTHING`,
			u.Username, u.RoleName, u.MerchantID, u.BranchID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ptrInt64(v int64) *int64 {
	return &v
}
