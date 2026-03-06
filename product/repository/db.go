package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/siderustler/go-ecommerce/product"
	"strconv"
	"strings"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func mapDomainFilterToSQLQueryFilter(filter product.Filter) (clause string, args []any) {
	var where []string
	if filter.IncludeElectro {
		args = append(args, "ELECTRO")
	}
	if filter.IncludeElectroMachines {
		args = append(args, "ELECTROMACHINES")
	}
	if filter.IncludeGardening {
		args = append(args, "GARDENING")
	}
	if filter.IncludeMachines {
		args = append(args, "MACHINES")
	}
	if filter.IncludeParts {
		args = append(args, "PARTS")
	}
	if len(args) > 0 {
		placeholders := make([]string, 0, len(args))
		for i := 1; i < len(args)+1; i++ {
			placeholders = append(placeholders, "$"+strconv.Itoa(i))
		}
		where = append(where, fmt.Sprintf(`category IN (%s)`, strings.Join(placeholders, ",")))
	}
	if filter.Search != "" {
		placeholder := "$" + strconv.Itoa(len(args)+1)
		where = append(where, "name ILIKE "+placeholder)
		args = append(args, "%"+filter.Search+"%")
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
	return clause, args
}

func (r repository) Products(ctx context.Context, offset int, limit int, filter product.Filter) ([]product.Product, int, error) {
	filterClause, args := mapDomainFilterToSQLQueryFilter(filter)
	args = append(args, limit, offset)
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
	stmt := `SELECT id, 
		name, 
		main_image, 
		price, 
		price_before, 
		category
		FROM products` + filterClause + sort + " LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))
	rows, err := r.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("retrieving products: %w", err)
	}
	var products []product.Product
	for rows.Next() {
		var product product.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.DiscountPrice, &product.Category)
		if err != nil {
			return nil, 0, fmt.Errorf("scannig product: %w", err)
		}
		products = append(products, product)
	}
	var allFilteredProductsCount int
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM products"+filterClause, args[:len(args)-2]...)
	err = row.Scan(&allFilteredProductsCount)
	if err != nil {
		return nil, 0, fmt.Errorf("scanning product count: %w", err)
	}
	return products, allFilteredProductsCount, nil
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
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.DiscountPrice)
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
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Price, &product.DiscountPrice)
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
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	return txFunc(tx)
}
