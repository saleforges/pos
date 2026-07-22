package postgres

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

type LoginAuditRepository struct {
	pool *otel.TracedPool
}

func NewLoginAuditRepository(pool *otel.TracedPool) *LoginAuditRepository {
	return &LoginAuditRepository{pool: pool}
}

func (r *LoginAuditRepository) Create(ctx context.Context, audit *domain.LoginAudit) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO login_audits (user_id, email, success, ip_address, user_agent, reason, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		audit.UserID, audit.Email, audit.Success, audit.IPAddress, audit.UserAgent, audit.Reason, time.Now().UTC(),
	).Scan(&audit.ID)
	return err
}

func (r *LoginAuditRepository) List(ctx context.Context, offset, limit int) ([]domain.LoginAudit, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM login_audits`).Scan(&total); err != nil {
		return nil, 0, err
	}

	if limit == -1 {
		limit = int(total)
		if limit == 0 { limit = 1 }
		offset = 0
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email, success, ip_address, user_agent, reason, created_at FROM login_audits ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var audits []domain.LoginAudit
	for rows.Next() {
		var a domain.LoginAudit
		if err := rows.Scan(&a.ID, &a.UserID, &a.Email, &a.Success, &a.IPAddress, &a.UserAgent, &a.Reason, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		audits = append(audits, a)
	}
	return audits, total, rows.Err()
}
