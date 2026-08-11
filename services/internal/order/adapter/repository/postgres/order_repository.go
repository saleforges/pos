package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.OrderRepository = (*OrderRepository)(nil)

const orderCols = `id, merchant_id, branch_id, created_by, customer_id, client_order_id, status, subtotal, discount, tax, total, paid_amount, due_date, note, created_at, updated_at`
const orderItemCols = `id, order_id, product_item_id, item_name, unit_price, quantity, line_total`
const paymentCols = `id, order_id, amount, method, created_by, paid_at, created_at`

type OrderRepository struct {
	pool *otel.TracedPool
}

func NewOrderRepository(pool *otel.TracedPool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (merchant_id, branch_id, created_by, customer_id, client_order_id, status, subtotal, discount, tax, total, paid_amount, due_date, note, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`,
		order.MerchantID, order.BranchID, order.CreatedBy, order.CustomerID, nullIfEmpty(order.ClientOrderID), order.Status,
		order.Subtotal, order.Discount, order.Tax, order.Total, order.PaidAmount,
		order.DueDate, nullIfEmpty(order.Note), order.CreatedAt, order.UpdatedAt,
	).Scan(&order.ID)
	if err != nil {
		return err
	}

	for i := range order.Items {
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_item_id, item_name, unit_price, quantity, line_total)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			order.ID, order.Items[i].ProductItemID, order.Items[i].ItemName,
			order.Items[i].UnitPrice, order.Items[i].Quantity, order.Items[i].LineTotal,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return err
		}
		order.Items[i].OrderID = order.ID
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, order); err != nil {
		return nil, err
	}
	if err := r.loadPayments(ctx, order); err != nil {
		return nil, err
	}
	order.PaymentStatus = order.ComputePaymentStatus()
	return order, nil
}

func (r *OrderRepository) GetByClientOrderID(ctx context.Context, clientOrderID string, merchantID int64) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+orderCols+` FROM orders WHERE client_order_id = $1 AND merchant_id = $2`, clientOrderID, merchantID)
	order, err := scanOrder(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	if err := r.loadItems(ctx, order); err != nil {
		return nil, err
	}
	if err := r.loadPayments(ctx, order); err != nil {
		return nil, err
	}
	order.PaymentStatus = order.ComputePaymentStatus()
	return order, nil
}

func (r *OrderRepository) List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
	query := `SELECT ` + orderCols + ` FROM orders WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if branchID != nil {
		args = append(args, *branchID)
		query += ` AND branch_id = $` + itoa(len(args))
	}
	if status != nil {
		args = append(args, string(*status))
		query += ` AND status = $` + itoa(len(args))
	}
	if paymentStatus != nil {
		switch *paymentStatus {
		case domain.PaymentStatusUnpaid:
			query += ` AND paid_amount <= 0`
		case domain.PaymentStatusPaid:
			query += ` AND paid_amount >= total`
		case domain.PaymentStatusPartial:
			query += ` AND paid_amount > 0 AND paid_amount < total`
		}
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		if err := r.loadItems(ctx, &orders[i]); err != nil {
			return nil, err
		}
		orders[i].PaymentStatus = orders[i].ComputePaymentStatus()
	}
	return orders, nil
}

func (r *OrderRepository) ListChangedSince(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Order, error) {
	query := `SELECT ` + orderCols + ` FROM orders WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if branchID > 0 {
		query += ` AND branch_id = $2`
		args = append(args, branchID)
	}
	if since != nil {
		query += ` AND updated_at > $` + itoa(len(args)+1)
		args = append(args, *since)
	}
	query += ` ORDER BY id`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := scanOrders(rows)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		if err := r.loadItems(ctx, &orders[i]); err != nil {
			return nil, err
		}
		orders[i].PaymentStatus = orders[i].ComputePaymentStatus()
	}
	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET due_date = $1, note = $2, updated_at = $3 WHERE id = $4 AND merchant_id = $5`,
		order.DueDate, nullIfEmpty(order.Note), order.UpdatedAt, order.ID, order.MerchantID)
	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error) {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2 AND merchant_id = $3`,
		status, id, merchantID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, merchantID)
}

func (r *OrderRepository) AddPayment(ctx context.Context, orderID int64, merchantID int64, payment *domain.PaymentRecord) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO payment_records (order_id, amount, method, created_by, paid_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		orderID, payment.Amount, payment.Method, payment.CreatedBy, payment.PaidAt, payment.CreatedAt,
	).Scan(&payment.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders SET paid_amount = paid_amount + $1, updated_at = NOW() WHERE id = $2 AND merchant_id = $3`,
		payment.Amount, orderID, merchantID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) loadItems(ctx context.Context, order *domain.Order) error {
	rows, err := r.pool.Query(ctx,
		`SELECT `+orderItemCols+` FROM order_items WHERE order_id = $1 ORDER BY id`, order.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	items, err := scanOrderItems(rows)
	if err != nil {
		return err
	}
	order.Items = items
	return nil
}

func (r *OrderRepository) loadPayments(ctx context.Context, order *domain.Order) error {
	rows, err := r.pool.Query(ctx,
		`SELECT `+paymentCols+` FROM payment_records WHERE order_id = $1 ORDER BY paid_at`, order.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	payments, err := scanPayments(rows)
	if err != nil {
		return err
	}
	order.Payments = payments
	return nil
}

func salesFilterClause(alias string, merchantID, branchID int64, from, to *time.Time) (string, []interface{}) {
	col := func(name string) string {
		if alias == "" {
			return name
		}
		return alias + "." + name
	}
	clause := col("merchant_id") + ` = $1 AND ` + col("status") + ` = 'completed'`
	args := []interface{}{merchantID}
	if branchID > 0 {
		args = append(args, branchID)
		clause += ` AND ` + col("branch_id") + ` = $` + fmt.Sprint(len(args))
	}
	if from != nil {
		args = append(args, *from)
		clause += ` AND ` + col("created_at") + ` >= $` + fmt.Sprint(len(args))
	}
	if to != nil {
		args = append(args, *to)
		clause += ` AND ` + col("created_at") + ` <= $` + fmt.Sprint(len(args))
	}
	return clause, args
}

func (r *OrderRepository) SalesReport(ctx context.Context, merchantID, branchID int64, from, to *time.Time) (*domain.SalesReport, error) {
	where, args := salesFilterClause("", merchantID, branchID, from, to)
	query := `SELECT
		COALESCE(SUM(paid_amount), 0),
		COUNT(*),
		COUNT(*) FILTER (WHERE paid_amount >= total),
		COUNT(*) FILTER (WHERE paid_amount < total),
		COALESCE(SUM(total - paid_amount) FILTER (WHERE paid_amount < total), 0)
	FROM orders WHERE ` + where

	var report domain.SalesReport
	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&report.TotalRevenue, &report.TotalOrders, &report.PaidOrders, &report.DebtOrders, &report.Outstanding,
	); err != nil {
		return nil, err
	}

	topProducts, err := r.topProducts(ctx, merchantID, branchID, from, to)
	if err != nil {
		return nil, err
	}
	report.TopProducts = topProducts

	breakdown, err := r.paymentBreakdown(ctx, merchantID, branchID, from, to)
	if err != nil {
		return nil, err
	}
	report.PaymentBreakdown = breakdown

	return &report, nil
}

func (r *OrderRepository) topProducts(ctx context.Context, merchantID, branchID int64, from, to *time.Time) ([]domain.ProductSales, error) {
	where, args := salesFilterClause("o", merchantID, branchID, from, to)
	query := `SELECT oi.product_item_id, oi.item_name, SUM(oi.quantity), SUM(oi.line_total)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE ` + where + `
		GROUP BY oi.product_item_id, oi.item_name
		ORDER BY SUM(oi.quantity) DESC
		LIMIT 5`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.ProductSales{}
	for rows.Next() {
		var p domain.ProductSales
		if err := rows.Scan(&p.ProductItemID, &p.Name, &p.Quantity, &p.Revenue); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *OrderRepository) paymentBreakdown(ctx context.Context, merchantID, branchID int64, from, to *time.Time) ([]domain.PaymentMethodTotal, error) {
	where, args := salesFilterClause("o", merchantID, branchID, from, to)
	query := `SELECT pr.method, COALESCE(SUM(pr.amount), 0), COUNT(*)
		FROM payment_records pr
		JOIN orders o ON o.id = pr.order_id
		WHERE ` + where + `
		GROUP BY pr.method
		ORDER BY SUM(pr.amount) DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.PaymentMethodTotal{}
	for rows.Next() {
		var p domain.PaymentMethodTotal
		if err := rows.Scan(&p.Method, &p.Amount, &p.Count); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func scanOrder(row pgx.Row) (*domain.Order, error) {
	var o domain.Order
	var note, clientOrderID *string
	err := row.Scan(&o.ID, &o.MerchantID, &o.BranchID, &o.CreatedBy, &o.CustomerID, &clientOrderID, &o.Status,
		&o.Subtotal, &o.Discount, &o.Tax, &o.Total, &o.PaidAmount, &o.DueDate, &note, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	o.Note = deref(note)
	o.ClientOrderID = deref(clientOrderID)
	return &o, nil
}

func scanOrders(rows pgx.Rows) ([]domain.Order, error) {
	var result []domain.Order
	for rows.Next() {
		var o domain.Order
		var note, clientOrderID *string
		if err := rows.Scan(&o.ID, &o.MerchantID, &o.BranchID, &o.CreatedBy, &o.CustomerID, &clientOrderID, &o.Status,
			&o.Subtotal, &o.Discount, &o.Tax, &o.Total, &o.PaidAmount, &o.DueDate, &note, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.Note = deref(note)
		o.ClientOrderID = deref(clientOrderID)
		result = append(result, o)
	}
	return result, rows.Err()
}

func scanOrderItems(rows pgx.Rows) ([]domain.OrderItem, error) {
	var result []domain.OrderItem
	for rows.Next() {
		var i domain.OrderItem
		if err := rows.Scan(&i.ID, &i.OrderID, &i.ProductItemID, &i.ItemName, &i.UnitPrice, &i.Quantity, &i.LineTotal); err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, rows.Err()
}

func scanPayments(rows pgx.Rows) ([]domain.PaymentRecord, error) {
	var result []domain.PaymentRecord
	for rows.Next() {
		var p domain.PaymentRecord
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.CreatedBy, &p.PaidAt, &p.CreatedAt); err != nil {
			return nil, err
		}
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
