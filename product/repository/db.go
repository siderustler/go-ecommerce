package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/siderustler/go-ecommerce/product"
)

type repository struct {
	db *sql.DB
}

// FIXME
func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS product_categories (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating product_categories table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			main_image TEXT NOT NULL,
			price INT NOT NULL,
			category_id INT NOT NULL REFERENCES product_categories(id),
			price_before INT DEFAULT 0
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating products table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS product_images (
			id UUID REFERENCES products(id),
			image_url TEXT NOT NULL,
			PRIMARY KEY (id, image_url)
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating product_images table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS product_details (
			id SERIAL PRIMARY KEY,
			product_id UUID REFERENCES products(id),
			product_info TEXT NOT NULL,
			technical_parameters TEXT NOT NULL
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating product_details table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO product_categories (name) VALUES ('GARDENING'), ('MACHINES')`,
	)
	if err != nil {
		return nil, fmt.Errorf("INSERT product_categories table: %w", err)
	}
	productUUID := uuid.NewString()
	_, err = db.ExecContext(ctx,
		`INSERT INTO products (id, name, main_image, price, category_id, price_before) VALUES ($1, $2, $3,$4,$5,$6)`,
		productUUID, "DLUGIE NAME BARDZO DLUGIE OJOJO JOJ ", "/public/products/essa/1.webp", 1900, 1, 0,
	)
	if err != nil {
		return nil, fmt.Errorf("INSERT products table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO product_images (id, image_url) VALUES ($1,$2), ($3,$4)`,
		productUUID, "/public/products/essa/2.webp",
		productUUID, "/public/products/essa/3.webp",
	)
	if err != nil {
		return nil, fmt.Errorf("INSERT product_images table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO product_details (product_id, product_info, technical_parameters) VALUES ($1,$2,$3)`,
		productUUID, "SOME SOME SOME SOME SOME", "TEHCNICAL TECHNICLA TEHCNIALCA",
	)
	if err != nil {
		return nil, fmt.Errorf("INSERT product_details table: %w", err)
	}
	return &repository{db: db}, nil
}

func mapDomainFilterToSQLQueryFilter(filter product.Filter, limit, offset int) (clause string, args []any) {
	var where []string
	if filter.IncludeElectro {
		args = append(args, "1")
	}
	if filter.IncludeElectroMachines {
		args = append(args, "2")
	}
	if filter.IncludeGardening {
		args = append(args, "3")
	}
	if filter.IncludeMachines {
		args = append(args, "4")
	}
	if filter.IncludeParts {
		args = append(args, "5")
	}
	placeholders := make([]string, 0, len(args))
	if len(args) > 0 {
		for i := 1; i < len(args)+1; i++ {
			placeholders = append(placeholders, "$"+strconv.Itoa(i))
		}
		where = append(where, fmt.Sprintf(`category_id IN (%s)`, strings.Join(placeholders, ",")))
	}
	if filter.Search != "" {
		placeholder := "$" + strconv.Itoa(len(placeholders)+1)
		placeholders = append(placeholders, placeholder)
		where = append(where, "name ILIKE "+placeholder)
		args = append(args, "%"+filter.Search+"%")
	}
	var sort string
	if filter.Sort != "" {
		switch filter.Sort {
		case product.NameDesc:
			sort = " ORDER BY name DESC"
		case product.PriceAsc:
			sort = " ORDER BY price ASC"
		case product.PriceDesc:
			sort = " ORDER BY price DESC"
		default:
			sort = " ORDER BY name ASC"
		}
	}
	if filter.PriceFrom != 0 {
		where = append(where, "price >= "+strconv.Itoa(filter.PriceFrom))
	}
	if filter.PriceTo != 0 {
		where = append(where, "price <= "+strconv.Itoa(filter.PriceTo))
	}

	shouldIncludeWhere := len(where) > 0
	if shouldIncludeWhere {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	clause += sort
	clause += " LIMIT $" + strconv.Itoa(len(placeholders)+1) + " OFFSET $" + strconv.Itoa(len(placeholders)+2)
	args = append(args, limit, offset)
	return clause, args
}

func (r repository) Products(ctx context.Context, offset int, limit int, filter product.Filter) ([]product.Product, error) {
	filterClause, args := mapDomainFilterToSQLQueryFilter(filter, limit, offset)
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, main_image, price, price_before,category_id FROM products"+filterClause, args...)
	if err != nil {
		return nil, fmt.Errorf("retrieving products: %w", err)
	}
	var products []product.Product
	for rows.Next() {
		var product product.Product
		var xd string
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.PriceBefore, &xd)
		if err != nil {
			return nil, fmt.Errorf("scannig product: %w", err)
		}
		products = append(products, product)
	}
	return products, nil
}

func (r repository) ProductsByIDs(ctx context.Context, ids []string) (map[string]product.Product, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, main_image, price, price_before FROM products WHERE id = ANY ($1)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("retrieving products: %w", err)
	}
	products := make(map[string]product.Product, len(ids))
	for rows.Next() {
		var product product.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.PriceBefore)
		if err != nil {
			return nil, fmt.Errorf("scannig product: %w", err)
		}
		products[product.ID] = product
	}
	return products, nil
}

func (r repository) Promotions(ctx context.Context, offset, limit int) (promos []product.Product, promoCount int, err error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, main_image, price, price_before FROM products WHERE price_before > 0 LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("retrieving promotions: %w", err)
	}
	for rows.Next() {
		var product product.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.PriceBefore)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning product: %w", err)
		}
		promos = append(promos, product)
	}
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM products")
	err = row.Scan(&promoCount)
	if err != nil {
		return nil, 0, fmt.Errorf("scanning promo count: %w", err)
	}
	return promos, promoCount, nil
}

func RunInTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	return txFunc(tx)
}
