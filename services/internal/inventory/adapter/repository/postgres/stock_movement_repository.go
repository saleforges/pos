package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.StockMovementRepository = (*StockMovementRepository)(nil)

const smCols = `id, merchant_id, branch_id, product_item_id, type, quantity, reference_type, reference_id, note, created_at`

type StockMovementRepository struct {
	pool *otel.TracedPool
}

func NewStockMovementRepository(pool *otel.TracedPool) *StockMovementRepository {
	return &StockMovementRepository{pool: pool}
}

func (r *StockMovementRepository) Create(ctx context.Context, m *domain.StockMovement) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO stock_movements (merchant_id, branch_id, product_item_id, type, quantity, reference_type, reference_id, note, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		m.MerchantID, m.BranchID, m.ProductItemID, m.Type, m.Quantity, nullIfEmpty(m.ReferenceType), m.ReferenceID, m.Note, m.CreatedAt,
	).Scan(&m.ID)
}

func (r *StockMovementRepository) ListByProductItem(ctx context.Context, productItemID int64, merchantID int64) ([]domain.StockMovement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+smCols+` FROM stock_movements WHERE product_item_id = $1 AND merchant_id = $2 ORDER BY created_at DESC`,
		productItemID, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStockMovements(rows)
}

func scanStockMovement(row pgx.Row) (*domain.StockMovement, error) {
	var m domain.StockMovement
	err := row.Scan(&m.ID, &m.MerchantID, &m.BranchID, &m.ProductItemID, &m.Type, &m.Quantity, &m.ReferenceType, &m.ReferenceID, &m.Note, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func scanStockMovements(rows pgx.Rows) ([]domain.StockMovement, error) {
	var result []domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		if err := rows.Scan(&m.ID, &m.MerchantID, &m.BranchID, &m.ProductItemID, &m.Type, &m.Quantity, &m.ReferenceType, &m.ReferenceID, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
