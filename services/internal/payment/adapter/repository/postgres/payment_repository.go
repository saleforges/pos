package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.PaymentRepository = (*PaymentRepository)(nil)

const paymentCols = `id, merchant_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, gateway_ref, transaction_id, created_at, updated_at`

type PaymentRepository struct {
	pool *otel.TracedPool
}

func NewPaymentRepository(pool *otel.TracedPool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.PaymentTransaction) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO payment_transactions (merchant_id, order_id, gateway, status, amount, payment_url, payment_no, qr_string, qr_image, expired_at, session_id, gateway_ref, transaction_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`,
		p.MerchantID, p.OrderID, p.Gateway, p.Status, p.Amount, nullIfEmpty(p.PaymentURL),
		nullIfEmpty(p.PaymentNo), nullIfEmpty(p.QrString), nullIfEmpty(p.QrImage), nullIfEmpty(p.ExpiredAt),
		nullIfEmpty(p.SessionID), nullIfEmpty(p.GatewayRef), nullIfEmpty(p.TransactionID), p.CreatedAt, p.UpdatedAt,
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
		`UPDATE payment_transactions SET payment_url = $1, payment_no = $2, qr_string = $3, qr_image = $4, expired_at = $5, session_id = $6, transaction_id = $7, updated_at = NOW() WHERE id = $8`,
		nullIfEmpty(p.PaymentURL), nullIfEmpty(p.PaymentNo), nullIfEmpty(p.QrString), nullIfEmpty(p.QrImage),
		nullIfEmpty(p.ExpiredAt), nullIfEmpty(p.SessionID), nullIfEmpty(p.TransactionID), id)
	return err
}

func (r *PaymentRepository) GetByGatewayRef(ctx context.Context, gatewayRef string) (*domain.PaymentTransaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+paymentCols+` FROM payment_transactions WHERE gateway_ref = $1`, gatewayRef)
	return scanPayment(row)
}

func (r *PaymentRepository) MarkPaid(ctx context.Context, id int64, gatewayRef string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET status = 'paid', gateway_ref = COALESCE($1, gateway_ref), updated_at = NOW() WHERE id = $2`,
		nullIfEmpty(gatewayRef), id)
	return err
}

func (r *PaymentRepository) MarkExpired(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_transactions SET status = 'expired', updated_at = NOW() WHERE id = $1`, id)
	return err
}

func scanPayment(row pgx.Row) (*domain.PaymentTransaction, error) {
	var p domain.PaymentTransaction
	var paymentURL, paymentNo, qrString, qrImage, expiredAt, sessionID, gatewayRef, transactionID *string
	err := row.Scan(&p.ID, &p.MerchantID, &p.OrderID, &p.Gateway, &p.Status, &p.Amount,
		&paymentURL, &paymentNo, &qrString, &qrImage, &expiredAt, &sessionID, &gatewayRef, &transactionID,
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
	p.GatewayRef = deref(gatewayRef)
	p.TransactionID = deref(transactionID)
	return &p, nil
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
