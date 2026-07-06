package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saleforge/pos/services/internal/iam/domain"
)

type LoginAuditRepository struct {
	pool *pgxpool.Pool
}

func NewLoginAuditRepository(pool *pgxpool.Pool) *LoginAuditRepository {
	return &LoginAuditRepository{pool: pool}
}

func (r *LoginAuditRepository) Create(ctx context.Context, audit *domain.LoginAudit) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO login_audits (id, user_id, email, success, ip_address, user_agent, reason, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		audit.ID, audit.UserID, audit.Email, audit.Success, audit.IPAddress, audit.UserAgent, audit.Reason, time.Now().UTC(),
	)
	return err
}

func (r *LoginAuditRepository) List(ctx context.Context, offset, limit int) ([]domain.LoginAudit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email, success, ip_address, user_agent, reason, created_at FROM login_audits ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []domain.LoginAudit
	for rows.Next() {
		var a domain.LoginAudit
		if err := rows.Scan(&a.ID, &a.UserID, &a.Email, &a.Success, &a.IPAddress, &a.UserAgent, &a.Reason, &a.CreatedAt); err != nil {
			return nil, err
		}
		audits = append(audits, a)
	}
	return audits, rows.Err()
}
