package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.ShiftRepository = (*ShiftRepository)(nil)

const shiftCols = `id, merchant_id, branch_id, opened_by, closed_by, status, starting_cash, expected_cash, actual_cash, variance, note, opened_at, closed_at`

type ShiftRepository struct {
	pool *otel.TracedPool
}

func NewShiftRepository(pool *otel.TracedPool) *ShiftRepository {
	return &ShiftRepository{pool: pool}
}

func (r *ShiftRepository) Create(ctx context.Context, s *domain.Shift) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO shifts (merchant_id, branch_id, opened_by, status, starting_cash, note, opened_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		s.MerchantID, s.BranchID, s.OpenedBy, s.Status, s.StartingCash, nullIfEmpty(s.Note), s.OpenedAt,
	).Scan(&s.ID)
}

func (r *ShiftRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Shift, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+shiftCols+` FROM shifts WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return scanShift(row)
}

func (r *ShiftRepository) GetOpenByBranch(ctx context.Context, merchantID, branchID int64) (*domain.Shift, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+shiftCols+` FROM shifts WHERE merchant_id = $1 AND branch_id = $2 AND status = 'open'`, merchantID, branchID)
	return scanShift(row)
}

func (r *ShiftRepository) Update(ctx context.Context, s *domain.Shift) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE shifts SET closed_by=$1, status=$2, expected_cash=$3, actual_cash=$4, variance=$5, note=$6, closed_at=$7
		 WHERE id=$8 AND merchant_id=$9`,
		s.ClosedBy, s.Status, s.ExpectedCash, s.ActualCash, s.Variance, nullIfEmpty(s.Note), s.ClosedAt, s.ID, s.MerchantID)
	return err
}

func (r *ShiftRepository) List(ctx context.Context, merchantID, branchID int64) ([]domain.Shift, error) {
	query := `SELECT ` + shiftCols + ` FROM shifts WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if branchID > 0 {
		query += ` AND branch_id = $2`
		args = append(args, branchID)
	}
	query += ` ORDER BY opened_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShifts(rows)
}

func (r *ShiftRepository) SumCashPayments(ctx context.Context, merchantID, branchID int64, from, to time.Time) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(pr.amount), 0)
		 FROM payment_records pr
		 JOIN orders o ON o.id = pr.order_id
		 WHERE o.merchant_id = $1 AND o.branch_id = $2
		   AND pr.paid_at >= $3 AND pr.paid_at < $4
		   AND LOWER(pr.method) IN ('cash', 'tunai')`,
		merchantID, branchID, from, to,
	).Scan(&total)
	return total, err
}

func scanShift(row pgx.Row) (*domain.Shift, error) {
	var s domain.Shift
	var note *string
	if err := row.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.OpenedBy, &s.ClosedBy, &s.Status,
		&s.StartingCash, &s.ExpectedCash, &s.ActualCash, &s.Variance, &note, &s.OpenedAt, &s.ClosedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrShiftNotFound
		}
		return nil, err
	}
	s.Note = deref(note)
	return &s, nil
}

func scanShifts(rows pgx.Rows) ([]domain.Shift, error) {
	var result []domain.Shift
	for rows.Next() {
		var s domain.Shift
		var note *string
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.BranchID, &s.OpenedBy, &s.ClosedBy, &s.Status,
			&s.StartingCash, &s.ExpectedCash, &s.ActualCash, &s.Variance, &note, &s.OpenedAt, &s.ClosedAt); err != nil {
			return nil, err
		}
		s.Note = deref(note)
		result = append(result, s)
	}
	return result, rows.Err()
}
