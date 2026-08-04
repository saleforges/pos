package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

var _ repository.ProductComponentRepository = (*ProductComponentRepository)(nil)

const pcCols = `id, merchant_id, product_item_id, created_at, updated_at`
const pciCols = `id, product_component_id, component_product_item_id, quantity, unit_id, created_at`

type ProductComponentRepository struct {
	pool *otel.TracedPool
}

func NewProductComponentRepository(pool *otel.TracedPool) *ProductComponentRepository {
	return &ProductComponentRepository{pool: pool}
}

func (r *ProductComponentRepository) Create(ctx context.Context, component *domain.ProductComponent) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO product_components (merchant_id, product_item_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		component.MerchantID, component.ProductItemID, component.CreatedAt, component.UpdatedAt,
	).Scan(&component.ID)
	if err != nil {
		return err
	}

	for i := range component.Items {
		err = tx.QueryRow(ctx,
			`INSERT INTO product_component_items (product_component_id, component_product_item_id, quantity, unit_id, created_at)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			component.ID, component.Items[i].ComponentProductItemID, component.Items[i].Quantity, component.Items[i].UnitID, component.Items[i].CreatedAt,
		).Scan(&component.Items[i].ID)
		if err != nil {
			return err
		}
		component.Items[i].ProductComponentID = component.ID
	}

	return tx.Commit(ctx)
}

func (r *ProductComponentRepository) GetByProductItem(ctx context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+pcCols+` FROM product_components WHERE product_item_id = $1 AND merchant_id = $2`,
		productItemID, merchantID)
	component, err := scanProductComponent(row)
	if err != nil {
		return nil, err
	}

	items, err := r.listItems(ctx, component.ID)
	if err != nil {
		return nil, err
	}
	component.Items = items
	return component, nil
}

func (r *ProductComponentRepository) List(ctx context.Context, merchantID int64) ([]domain.ProductComponent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+pcCols+` FROM product_components WHERE merchant_id = $1 ORDER BY id`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	components, err := scanProductComponents(rows)
	if err != nil {
		return nil, err
	}

	for i := range components {
		items, err := r.listItems(ctx, components[i].ID)
		if err != nil {
			return nil, err
		}
		components[i].Items = items
	}
	return components, nil
}

func (r *ProductComponentRepository) Update(ctx context.Context, component *domain.ProductComponent) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE product_components SET updated_at=$1 WHERE id=$2 AND merchant_id=$3`,
		component.UpdatedAt, component.ID, component.MerchantID)
	if err != nil {
		return err
	}

	// Delete existing items and re-insert
	_, err = tx.Exec(ctx, `DELETE FROM product_component_items WHERE product_component_id = $1`, component.ID)
	if err != nil {
		return err
	}

	for i := range component.Items {
		err = tx.QueryRow(ctx,
			`INSERT INTO product_component_items (product_component_id, component_product_item_id, quantity, unit_id, created_at)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			component.ID, component.Items[i].ComponentProductItemID, component.Items[i].Quantity, component.Items[i].UnitID, component.Items[i].CreatedAt,
		).Scan(&component.Items[i].ID)
		if err != nil {
			return err
		}
		component.Items[i].ProductComponentID = component.ID
	}

	return tx.Commit(ctx)
}

func (r *ProductComponentRepository) Delete(ctx context.Context, id int64, merchantID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM product_components WHERE id = $1 AND merchant_id = $2`, id, merchantID)
	return err
}

func (r *ProductComponentRepository) listItems(ctx context.Context, componentID int64) ([]domain.ProductComponentItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+pciCols+` FROM product_component_items WHERE product_component_id = $1 ORDER BY id`, componentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProductComponentItems(rows)
}

func scanProductComponent(row pgx.Row) (*domain.ProductComponent, error) {
	var c domain.ProductComponent
	err := row.Scan(&c.ID, &c.MerchantID, &c.ProductItemID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrProductComponentNotFound
		}
		return nil, err
	}
	return &c, nil
}

func scanProductComponents(rows pgx.Rows) ([]domain.ProductComponent, error) {
	var result []domain.ProductComponent
	for rows.Next() {
		var c domain.ProductComponent
		if err := rows.Scan(&c.ID, &c.MerchantID, &c.ProductItemID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func scanProductComponentItems(rows pgx.Rows) ([]domain.ProductComponentItem, error) {
	var result []domain.ProductComponentItem
	for rows.Next() {
		var i domain.ProductComponentItem
		if err := rows.Scan(&i.ID, &i.ProductComponentID, &i.ComponentProductItemID, &i.Quantity, &i.UnitID, &i.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, rows.Err()
}
