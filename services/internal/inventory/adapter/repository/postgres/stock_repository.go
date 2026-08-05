package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.StockRepository = (*StockRepository)(nil)
var _ repository.StockAdjustmentRepository = (*StockRepository)(nil)

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

// ListChangedSince returns stocks updated after the given time — the
// incremental payload for mobile stock sync.
func (r *StockRepository) ListChangedSince(ctx context.Context, merchantID int64, since *time.Time) ([]domain.Stock, error) {
	return r.listChangedSince(ctx, merchantID, 0, since)
}

func (r *StockRepository) SyncByBranch(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Stock, error) {
	return r.listChangedSince(ctx, merchantID, branchID, since)
}

func (r *StockRepository) listChangedSince(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Stock, error) {
	query := `SELECT ` + stockCols + ` FROM stocks WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if branchID > 0 {
		query += ` AND branch_id = $2`
		args = append(args, branchID)
	}
	if since != nil {
		query += ` AND updated_at > $3`
		args = append(args, *since)
	}
	query += ` ORDER BY id`
	rows, err := r.pool.Query(ctx, query, args...)
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

// Deduct applies a batch of stock deductions atomically and records a
// stock_out movement per line. Fails the whole batch if any item has
// insufficient available stock.
func (r *StockRepository) Deduct(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return r.adjust(ctx, merchantID, branchID, referenceType, referenceID, items, domain.MovementTypeStockOut, -1)
}

// Restore applies a batch of stock restores atomically and records a
// stock_in movement per line (e.g. when an order is cancelled).
func (r *StockRepository) Restore(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return r.adjust(ctx, merchantID, branchID, referenceType, referenceID, items, domain.MovementTypeStockIn, 1)
}

func (r *StockRepository) adjust(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem, movementType domain.MovementType, sign int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, it := range items {
		if it.ProductItemID == 0 || it.Quantity <= 0 {
			return domain.ErrInvalidStock
		}
		// Lock the stock row for this branch+item.
		var available int64
		err := tx.QueryRow(ctx,
			`SELECT available FROM stocks WHERE merchant_id=$1 AND branch_id=$2 AND product_item_id=$3 FOR UPDATE`,
			merchantID, branchID, it.ProductItemID).Scan(&available)
		if err != nil {
			return domain.ErrStockNotFound
		}
		if sign < 0 && available < it.Quantity {
			return domain.ErrInsufficientStock
		}

		delta := sign * it.Quantity
		if _, err := tx.Exec(ctx,
			`UPDATE stocks SET available = available + $1, updated_at = NOW()
			 WHERE merchant_id=$2 AND branch_id=$3 AND product_item_id=$4`,
			delta, merchantID, branchID, it.ProductItemID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO stock_movements (merchant_id, branch_id, product_item_id, type, quantity, reference_type, reference_id, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
			merchantID, branchID, it.ProductItemID, movementType, it.Quantity, referenceType, referenceID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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
