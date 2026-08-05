package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.PaymentRepository = (*PaymentRepository)(nil)

const paymentCols = `id, merchant_id, branch_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, payment_ref, created_at, updated_at`

type PaymentRepository struct {
	pool *otel.TracedPool
}

func NewPaymentRepository(pool *otel.TracedPool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.PaymentTransaction) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO payment_transactions (merchant_id, branch_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, payment_ref, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`,
		p.MerchantID, p.BranchID, p.OrderID, p.Gateway, p.Status, p.Amount, nullIfEmpty(p.PaymentURL),
		nullIfEmpty(p.PaymentNo), nullIfEmpty(p.QrString), nullIfEmpty(p.QrImage), nullIfEmpty(p.ExpiredAt),
		nullIfEmpty(p.SessionID), nullIfEmpty(p.PaymentRef), p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+paymentCols+` FROM payment_transactions WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return scanPayment(row)
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID int64) (*domain.PaymentTransaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+paymentCols+` FROM payment_transactions WHERE order_id = $1 ORDER BY id DESC LIMIT 1`, orderID)
	return scanPayment(row)
}

func (r *PaymentRepository) UpdatePaymentURL(ctx context.Context, id int64, paymentURL, sessionID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET payment_url = $1, session_id = $2, updated_at = NOW() WHERE id = $3`,
		nullIfEmpty(paymentURL), nullIfEmpty(sessionID), id)
	return err
}

func (r *PaymentRepository) UpdateDetails(ctx context.Context, id int64, p *domain.PaymentTransaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET payment_url = $1, payment_no = $2, qr_string = $3, qr_image = $4, expired_at = $5, session_id = $6, payment_ref = $7, updated_at = NOW() WHERE id = $8`,
		nullIfEmpty(p.PaymentURL), nullIfEmpty(p.PaymentNo), nullIfEmpty(p.QrString), nullIfEmpty(p.QrImage),
		nullIfEmpty(p.ExpiredAt), nullIfEmpty(p.SessionID), nullIfEmpty(p.PaymentRef), id)
	return err
}

func (r *PaymentRepository) GetByPaymentRef(ctx context.Context, paymentRef string) (*domain.PaymentTransaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+paymentCols+` FROM payment_transactions WHERE payment_ref = $1`, paymentRef)
	return scanPayment(row)
}

func (r *PaymentRepository) MarkPaid(ctx context.Context, id int64, paymentRef string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET status = 'paid', payment_ref = COALESCE($1, payment_ref), updated_at = NOW() WHERE id = $2`,
		nullIfEmpty(paymentRef), id)
	return err
}

func (r *PaymentRepository) MarkExpired(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET status = 'expired', updated_at = NOW() WHERE id = $1`, id)
	return err
}

func scanPayment(row pgx.Row) (*domain.PaymentTransaction, error) {
	var p domain.PaymentTransaction
	var paymentURL, paymentNo, qrString, qrImage, expiredAt, sessionID, paymentRef *string
	err := row.Scan(&p.ID, &p.MerchantID, &p.BranchID, &p.OrderID, &p.Gateway, &p.Status, &p.Amount,
		&paymentURL, &paymentNo, &qrString, &qrImage, &expiredAt, &sessionID, &paymentRef,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	p.PaymentURL = deref(paymentURL)
	p.PaymentNo = deref(paymentNo)
	p.QrString = deref(qrString)
	p.QrImage = deref(qrImage)
	p.ExpiredAt = deref(expiredAt)
	p.SessionID = deref(sessionID)
	p.PaymentRef = deref(paymentRef)
	return &p, nil
}

func (r *PaymentRepository) GetStaticQR(ctx context.Context, merchantID, branchID int64) (*domain.StaticQR, error) {
	var qr domain.StaticQR
	err := r.pool.QueryRow(ctx,
		`SELECT merchant_id, branch_id, payment_no, qr_string, qr_image FROM payment_static_qrs WHERE merchant_id = $1 AND branch_id = $2`, merchantID, branchID,
	).Scan(&qr.MerchantID, &qr.BranchID, &qr.PaymentNo, &qr.QrString, &qr.QrImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return &qr, nil
}

func (r *PaymentRepository) UpsertStaticQR(ctx context.Context, qr *domain.StaticQR) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_static_qrs (merchant_id, branch_id, payment_no, qr_string, qr_image, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 ON CONFLICT (merchant_id, branch_id) DO UPDATE SET payment_no = $3, qr_string = $4, qr_image = $5, updated_at = NOW()`,
		qr.MerchantID, qr.BranchID, qr.PaymentNo, qr.QrString, qr.QrImage)
	return err
}

// ListChangedSince returns payments updated after the given time — the
// incremental payload for mobile payment sync.
func (r *PaymentRepository) ListChangedSince(ctx context.Context, merchantID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	return r.listChangedSince(ctx, merchantID, 0, since)
}

func (r *PaymentRepository) SyncByBranch(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	return r.listChangedSince(ctx, merchantID, branchID, since)
}

func (r *PaymentRepository) listChangedSince(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	query := `SELECT ` + paymentCols + ` FROM payment_transactions WHERE merchant_id = $1`
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

	var result []domain.PaymentTransaction
	for rows.Next() {
		var p domain.PaymentTransaction
		var paymentURL, paymentNo, qrString, qrImage, expiredAt, sessionID, paymentRef *string
		if err := rows.Scan(&p.ID, &p.MerchantID, &p.OrderID, &p.Gateway, &p.Status, &p.Amount,
			&paymentURL, &paymentNo, &qrString, &qrImage, &expiredAt, &sessionID, &paymentRef,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.PaymentURL = deref(paymentURL)
		p.PaymentNo = deref(paymentNo)
		p.QrString = deref(qrString)
		p.QrImage = deref(qrImage)
		p.ExpiredAt = deref(expiredAt)
		p.SessionID = deref(sessionID)
		p.PaymentRef = deref(paymentRef)
		result = append(result, p)
	}
	return result, rows.Err()
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
