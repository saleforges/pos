package postgres

import (
	"context"
	"strings"

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

func scanCustomer(row pgx.Row) (*domain.Customer, error) {
	var c domain.Customer
	err := row.Scan(&c.ID, &c.MerchantID, &c.Name, &c.Phone, &c.Address, &c.Note, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanCustomers(rows pgx.Rows) ([]domain.Customer, error) {
	var result []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(&c.ID, &c.MerchantID, &c.Name, &c.Phone, &c.Address, &c.Note, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
