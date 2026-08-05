package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.PaymentRepository = (*PaymentRepository)(nil)

const paymentCols = `id, merchant_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, payment_ref, created_at, updated_at`

type PaymentRepository struct {
	pool *otel.TracedPool
}

func NewPaymentRepository(pool *otel.TracedPool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.PaymentTransaction) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO payment_transactions (merchant_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, payment_ref, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
		p.MerchantID, p.OrderID, p.Gateway, p.Status, p.Amount, nullIfEmpty(p.PaymentURL),
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
	err := row.Scan(&p.ID, &p.MerchantID, &p.OrderID, &p.Gateway, &p.Status, &p.Amount,
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

func (r *PaymentRepository) GetStaticQR(ctx context.Context, merchantID int64) (*domain.StaticQR, error) {
	var qr domain.StaticQR
	err := r.pool.QueryRow(ctx,
		`SELECT merchant_id, payment_no, qr_string, qr_image FROM payment_static_qrs WHERE merchant_id = $1`, merchantID,
	).Scan(&qr.MerchantID, &qr.PaymentNo, &qr.QrString, &qr.QrImage)
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
		`INSERT INTO payment_static_qrs (merchant_id, payment_no, qr_string, qr_image, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (merchant_id) DO UPDATE SET payment_no = $2, qr_string = $3, qr_image = $4, updated_at = NOW()`,
		qr.MerchantID, qr.PaymentNo, qr.QrString, qr.QrImage)
	return err
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
