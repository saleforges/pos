package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.StockRepository = (*StockRepository)(nil)

const stockCols = `id, merchant_id, branch_id, product_item_id, available, created_at, updated_at`

type StockRepository struct {
	pool *otel.TracedPool
}

func NewStockRepository(pool *otel.TracedPool) *StockRepository {
	return &StockRepository{pool: pool}
}

func (r *StockRepository) Create(ctx context.Context, stock *domain.Stock) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO stocks (merchant_id, branch_id, product_item_id, available, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		stock.MerchantID, stock.BranchID, stock.ProductItemID, stock.Available, stock.CreatedAt, stock.UpdatedAt,
	).Scan(&stock.ID)
}

func (r *StockRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+stockCols+` FROM stocks WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return scanStock(row)
}

func (r *StockRepository) List(ctx context.Context, merchantID int64) ([]domain.Stock, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+stockCols+` FROM stocks WHERE merchant_id = $1 ORDER BY id`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStocks(rows)
}

func (r *StockRepository) Update(ctx context.Context, stock *domain.Stock) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE stocks SET available=$1, updated_at=$2 WHERE id=$3 AND merchant_id=$4`,
		stock.Available, stock.UpdatedAt, stock.ID, stock.MerchantID)
	return err
}

func scanStock(row pgx.Row) (*domain.Stock, error) {
	var s domain.Stock
	err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.ProductItemID, &s.Available, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func scanStocks(rows pgx.Rows) ([]domain.Stock, error) {
	var result []domain.Stock
	for rows.Next() {
		var s domain.Stock
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.ProductItemID, &s.Available, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
