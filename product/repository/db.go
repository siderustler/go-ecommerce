package repository

import (
	"context"
	"database/sql"
	"fmt"

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
			price REAL NOT NULL,
			category_id INT NOT NULL REFERENCES product_categories(id),
			price_before REAL
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
		productUUID, "DLUGIE NAME BARDZO DLUGIE OJOJO JOJ ", "/public/products/essa/1.webp", 1.90, 1, 0,
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

func (r repository) Products(ctx context.Context, filter product.Filter) ([]product.Product, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, main_image, price, price_before,category_id FROM products")
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

func (r repository) ProductsByIDs(ctx context.Context, ids []string) ([]product.Product, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, main_image, price, price_before FROM products WHERE id = ANY ($1)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("retrieving products: %w", err)
	}
	products := make([]product.Product, len(ids))
	for rows.Next() {
		var product product.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.PriceBefore)
		if err != nil {
			return nil, fmt.Errorf("scannig product: %w", err)
		}
		products = append(products, product)
	}
	return products, nil
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
