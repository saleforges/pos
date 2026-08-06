package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.CustomerRepository = (*CustomerRepository)(nil)

const customerCols = `id, merchant_id, name, phone, address, note, created_at, updated_at`

type CustomerRepository struct {
	pool *otel.TracedPool
}

func NewCustomerRepository(pool *otel.TracedPool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO customers (merchant_id, name, phone, address, note, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		c.MerchantID, c.Name, nullIfEmpty(c.Phone), nullIfEmpty(c.Address), nullIfEmpty(c.Note), c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
}

func (r *CustomerRepository) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+customerCols+` FROM customers WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return scanCustomer(row)
}

func (r *CustomerRepository) List(ctx context.Context, merchantID int64, search string) ([]domain.Customer, error) {
	query := `SELECT ` + customerCols + ` FROM customers WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		query += ` AND (LOWER(name) LIKE $2 OR LOWER(COALESCE(phone,'')) LIKE $2)`
	}
	query += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomers(rows)
}

// ListChangedSince returns customers updated after the given time — the
// incremental payload for mobile customer sync.
func (r *CustomerRepository) ListChangedSince(ctx context.Context, merchantID int64, since *time.Time) ([]domain.Customer, error) {
	query := `SELECT ` + customerCols + ` FROM customers WHERE merchant_id = $1`
	args := []interface{}{merchantID}
	if since != nil {
		query += ` AND updated_at > $2`
		args = append(args, *since)
	}
	query += ` ORDER BY id`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomers(rows)
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE customers SET name=$1, phone=$2, address=$3, note=$4, updated_at=$5 WHERE id=$6 AND merchant_id=$7`,
		c.Name, nullIfEmpty(c.Phone), nullIfEmpty(c.Address), nullIfEmpty(c.Note), c.UpdatedAt, c.ID, c.MerchantID)
	return err
}

func (r *CustomerRepository) Delete(ctx context.Context, id int64, merchantID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM customers WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return err
}

const customerPriceCols = `id, merchant_id, customer_id, product_item_id, price, currency, created_at, updated_at`

func (r *CustomerRepository) UpsertPrices(ctx context.Context, merchantID, customerID int64, prices []domain.CustomerPrice) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM customer_prices WHERE customer_id = $1 AND merchant_id = $2`, customerID, merchantID); err != nil {
		return err
	}
	for _, p := range prices {
		if _, err := tx.Exec(ctx,
			`INSERT INTO customer_prices (merchant_id, customer_id, product_item_id, price, currency, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			merchantID, customerID, p.ProductItemID, p.Price, p.Currency, p.CreatedAt, p.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CustomerRepository) ListPrices(ctx context.Context, merchantID, customerID int64) ([]domain.CustomerPrice, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+customerPriceCols+` FROM customer_prices WHERE customer_id = $1 AND merchant_id = $2 ORDER BY product_item_id`,
		customerID, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomerPrices(rows)
}

func (r *CustomerRepository) ListAllPrices(ctx context.Context, merchantID int64) ([]domain.CustomerPrice, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+customerPriceCols+` FROM customer_prices WHERE merchant_id = $1 ORDER BY customer_id, product_item_id`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomerPrices(rows)
}

func (r *CustomerRepository) GetPriceMap(ctx context.Context, merchantID, customerID int64, productItemIDs []int64) (map[int64]float64, error) {
	if len(productItemIDs) == 0 {
		return map[int64]float64{}, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT product_item_id, price FROM customer_prices
		 WHERE customer_id = $1 AND merchant_id = $2 AND product_item_id = ANY($3)`,
		customerID, merchantID, productItemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[int64]float64)
	for rows.Next() {
		var id int64
		var price float64
		if err := rows.Scan(&id, &price); err != nil {
			return nil, err
		}
		prices[id] = price
	}
	return prices, rows.Err()
}

func scanCustomerPrices(rows pgx.Rows) ([]domain.CustomerPrice, error) {
	var result []domain.CustomerPrice
	for rows.Next() {
		var p domain.CustomerPrice
		if err := rows.Scan(&p.ID, &p.MerchantID, &p.CustomerID, &p.ProductItemID, &p.Price, &p.Currency, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func scanCustomer(row pgx.Row) (*domain.Customer, error) {
	var c domain.Customer
	var phone, address, note *string
	err := row.Scan(&c.ID, &c.MerchantID, &c.Name, &phone, &address, &note, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	c.Phone = deref(phone)
	c.Address = deref(address)
	c.Note = deref(note)
	return &c, nil
}

func scanCustomers(rows pgx.Rows) ([]domain.Customer, error) {
	var result []domain.Customer
	for rows.Next() {
		var c domain.Customer
		var phone, address, note *string
		if err := rows.Scan(&c.ID, &c.MerchantID, &c.Name, &phone, &address, &note, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Phone = deref(phone)
		c.Address = deref(address)
		c.Note = deref(note)
		result = append(result, c)
	}
	return result, rows.Err()
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
