package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/otel"
)

type MerchantRepository struct {
	pool *otel.TracedPool
}

func NewMerchantRepository(pool *otel.TracedPool) *MerchantRepository {
	return &MerchantRepository{pool: pool}
}

func (r *MerchantRepository) Create(ctx context.Context, merchant *domain.Merchant) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO merchants (name, legal_name, address, phone, email, logo_url, tax_id,
		                       status, tax_rate, currency, timezone, receipt_footer, receipt_logo,
		                       order_prefix, low_stock_threshold, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7,
		        $8, $9, $10, $11, $12, $13,
		        $14, $15, $16, $17)
		RETURNING id`,
		merchant.Name, merchant.LegalName, merchant.Address,
		merchant.Phone, merchant.Email, merchant.LogoURL, merchant.TaxID,
		merchant.Status, merchant.Settings.TaxRate, merchant.Settings.Currency,
		merchant.Settings.Timezone, merchant.Settings.ReceiptFooter,
		merchant.Settings.ReceiptLogo, merchant.Settings.OrderPrefix,
		merchant.Settings.LowStockThreshold, merchant.CreatedAt, merchant.UpdatedAt,
	).Scan(&merchant.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrMerchantExists
		}
		return fmt.Errorf("failed to create merchant: %w", err)
	}
	return nil
}

func (r *MerchantRepository) GetByID(ctx context.Context, id int64) (*domain.Merchant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, legal_name, address, phone, email, logo_url, tax_id,
		       status, tax_rate, currency, timezone, receipt_footer, receipt_logo,
		       order_prefix, low_stock_threshold, created_at, updated_at
		FROM merchants WHERE id = $1`, id)

	m := &domain.Merchant{}
	err := row.Scan(
		&m.ID, &m.Name, &m.LegalName, &m.Address,
		&m.Phone, &m.Email, &m.LogoURL, &m.TaxID,
		&m.Status, &m.Settings.TaxRate, &m.Settings.Currency,
		&m.Settings.Timezone, &m.Settings.ReceiptFooter,
		&m.Settings.ReceiptLogo, &m.Settings.OrderPrefix,
		&m.Settings.LowStockThreshold, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}
	return m, nil
}

func (r *MerchantRepository) List(ctx context.Context, offset, limit int) ([]domain.Merchant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, legal_name, address, phone, email, logo_url, tax_id,
		       status, tax_rate, currency, timezone, receipt_footer, receipt_logo,
		       order_prefix, low_stock_threshold, created_at, updated_at
		FROM merchants ORDER BY created_at DESC OFFSET $1 LIMIT $2`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list merchants: %w", err)
	}
	defer rows.Close()

	var result []domain.Merchant
	for rows.Next() {
		var m domain.Merchant
		if err := rows.Scan(
			&m.ID, &m.Name, &m.LegalName, &m.Address,
			&m.Phone, &m.Email, &m.LogoURL, &m.TaxID,
			&m.Status, &m.Settings.TaxRate, &m.Settings.Currency,
			&m.Settings.Timezone, &m.Settings.ReceiptFooter,
			&m.Settings.ReceiptLogo, &m.Settings.OrderPrefix,
			&m.Settings.LowStockThreshold, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		result = append(result, m)
	}
	if result == nil {
		return []domain.Merchant{}, nil
	}
	return result, nil
}

func (r *MerchantRepository) Update(ctx context.Context, merchant *domain.Merchant) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE merchants SET name=$1, legal_name=$2, address=$3, phone=$4, email=$5,
		                     logo_url=$6, tax_id=$7, status=$8, tax_rate=$9, currency=$10,
		                     timezone=$11, receipt_footer=$12, receipt_logo=$13,
		                     order_prefix=$14, low_stock_threshold=$15, updated_at=$16
		WHERE id=$17`,
		merchant.Name, merchant.LegalName, merchant.Address,
		merchant.Phone, merchant.Email, merchant.LogoURL, merchant.TaxID,
		merchant.Status, merchant.Settings.TaxRate, merchant.Settings.Currency,
		merchant.Settings.Timezone, merchant.Settings.ReceiptFooter,
		merchant.Settings.ReceiptLogo, merchant.Settings.OrderPrefix,
		merchant.Settings.LowStockThreshold, merchant.UpdatedAt, merchant.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrMerchantNotFound
	}
	return nil
}

func (r *MerchantRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM merchants WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete merchant: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrMerchantNotFound
	}
	return nil
}
