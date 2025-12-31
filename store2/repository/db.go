package store_repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	store "github.com/siderustler/go-ecommerce/store2"
	store_domain "github.com/siderustler/go-ecommerce/store2/domain"
)

type repository struct {
	db *sql.DB
}

var _ store.Repository = repository{}

func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS baskets (
			id UUID PRIMARY KEY,
			customer_id UUID NOT NULL REFERENCES customers(customer_id),
			last_modified_at TIMESTAMP NOT NULL,
			status TEXT CHECK (status IN ('ACTIVE', 'INACTIVE'))
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating baskets table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS checkouts (
			id UUID PRIMARY KEY,
			basket_id UUID NOT NULL REFERENCES baskets(id),
			created_at TIMESTAMP NOT NULL,
			status TEXT CHECK (status IN ('INVALIDATED', 'PENDING', 'FINALIZED'))
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating checkouts table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS basket_products (
			id UUID NOT NULL REFERENCES baskets(id),
			product_id UUID NOT NULL REFERENCES products(id),
			count INT NOT NULL,
			PRIMARY KEY(id, product_id)
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating basket_products table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS stock (
			product_id UUID NOT NULL REFERENCES products(id),
			available_amount INT DEFAULT 0,
			reserved_amount INT DEFAULT 0
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating stock table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS stock_reservations (
			checkout_id UUID PRIMARY KEY REFERENCES checkouts(id),
			product_id UUID REFERENCES products(id),
			amount INT DEFAULT 0,
			reserved_at TIMESTAMP NOT NULL
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating stock_reservations table: %w", err)
	}
	return &repository{db: db}, nil
}

// CartCount implements store.Repository.
func (r repository) CartCount(ctx context.Context, userID string) (int, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(bp.id) FROM basket_products AS bp JOIN baskets AS b ON bp.id = b.id WHERE b.customer_id = $1`,
		userID,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("retrieving basket count: %w", err)
	}
	return count, nil
}

// CreateCheckout implements store.Repository.
func (r repository) CreateCheckout(
	ctx context.Context,
	userID string,
	insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(
			ctx,
			`SELECT b.id, b.customer_id, b.last_modified_at, b.status, bp.product_id, bp.count
			FROM baskets AS b
			JOIN basket_products AS bp ON b.id = bp.id
			WHERE b.status = $1 AND b.customer_id = $2 FOR UPDATE OF b`,
			store_domain.CartActive, userID,
		)
		if err != nil {
			return fmt.Errorf("retrieving user basket: %w", err)
		}
		defer rows.Close()
		cart := store_domain.Cart{Products: make(map[string]store_domain.CartProduct)}
		var cartProductIds []string
		for rows.Next() {
			var cartProduct store_domain.CartProduct
			err := rows.Scan(&cart.ID, &cart.CustomerID, &cart.LastModifiedAt, &cart.Status, &cartProduct.ProductID, &cartProduct.Count)
			if err != nil {
				return fmt.Errorf("scanning baskets: %w", err)
			}
			cart.Products[cartProduct.ProductID] = cartProduct
			cartProductIds = append(cartProductIds, cartProduct.ProductID)
		}
		stock := store_domain.Stock{Items: make(map[string]store_domain.StockItem, len(cart.Products))}
		if !cart.IsZero() {
			rows, err = tx.QueryContext(
				ctx,
				`SELECT product_id, available_amount, reserved_amount 
				FROM stock WHERE product_id ANY ($1) FOR UPDATE`,
				cartProductIds,
			)
			if err != nil {
				return fmt.Errorf("retrieving stock: %w", err)
			}
			defer rows.Close()
			var stockItem store_domain.StockItem
			err := rows.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
			if err != nil {
				return fmt.Errorf("scanning stock item: %w", err)
			}
			stock.Items[stockItem.ProductID] = stockItem
		}
		checkout, err := insertFn(&cart, &stock)
		if err != nil {
			return fmt.Errorf("insertfn domain checkout: %w", err)
		}
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO checkouts (id, basket_id, created_at, status) VALUES ($1,$2,$3,$4)`,
			checkout.ID, cart.ID, checkout.CreatedAt, checkout.Status,
		)
		if err != nil {
			return fmt.Errorf("inserting checkout to repository: %w", err)
		}
		_, err = tx.ExecContext(
			ctx,
			`UPDATE baskets SET last_modified_at = $2, status = $3 
			WHERE id = $1 AND status = $4`,
			cart.ID, cart.LastModifiedAt, cart.Status, store_domain.CartActive,
		)
		if err != nil {
			return fmt.Errorf("updating baskets: %w", err)
		}
		reservationTime := time.Now().UTC().Format(time.RFC3339)
		//FIXME -- single statement
		for productID, product := range checkout.Items {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO stock_reservations (checkout_id, product_id, amount, reserved_at) 
				VALUES ($1,$2,$3,$4)`,
				checkout.ID, productID, product.Count, reservationTime,
			)
			if err != nil {
				return fmt.Errorf("inserting reservation: %w", err)
			}
			stockItem := stock.Items[productID]
			_, err = tx.ExecContext(
				ctx,
				`UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3`,
				stockItem.AvailableAmount, stockItem.ReservedAmount, productID,
			)
			if err != nil {
				return fmt.Errorf("updating stock: %w", err)
			}
		}

		return nil
	})
}

// InsertStockItem implements store.Repository.
func (r repository) InsertStockItem(ctx context.Context, stockItem store_domain.StockItem, product store_domain.Product) error {
	panic("unimplemented")
}

// UpdateCheckout implements store.Repository.
func (r repository) UpdateCheckout(ctx context.Context, checkoutID string, updateFn func(checkout *store_domain.Checkout, stock *store_domain.Stock) error) error {
	panic("unimplemented")
}

// UpdateStockItem implements store.Repository.
func (r repository) UpdateStockItem(ctx context.Context, stockItem store_domain.StockItem, updateFn func(stockItem *store_domain.StockItem) error) {
	panic("unimplemented")
}

// UpsertCart implements store.Repository.
func (r repository) UpsertCart(ctx context.Context, userID string, item store_domain.CartProduct, upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(
			ctx,
			`SELECT b.id, b.customer_id, b.last_modified_at, b.status, bp.product_id, bp.count
			FROM baskets AS b
			JOIN basket_products AS bp ON b.id = bp.id
			WHERE b.status = $1 AND b.customer_id = $2`,
			store_domain.CartActive, userID,
		)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving user basket: %w", err)
		}
		defer rows.Close()
		cart := store_domain.Cart{Products: make(map[string]store_domain.CartProduct)}
		for rows.Next() {
			var cartProduct store_domain.CartProduct
			err := rows.Scan(&cart.ID, &cart.CustomerID, &cart.LastModifiedAt, &cart.Status, &cartProduct.ProductID, &cartProduct.Count)
			if err != nil {
				return fmt.Errorf("scanning baskets: %w", err)
			}
			cart.Products[cartProduct.ProductID] = cartProduct
		}
		row := tx.QueryRowContext(
			ctx,
			`SELECT product_id, available_amount, reserved_amount 
			 FROM stock WHERE product_id = $1`,
			item.ProductID,
		)

		var stockItem store_domain.StockItem
		err = row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("scanning stock item: %w", err)
		}

		var checkout store_domain.Checkout
		var stock store_domain.Stock
		var checkoutProductIds []string
		if !cart.IsZero() {
			checkout = store_domain.Checkout{Items: make(map[string]store_domain.CartProduct, len(cart.Products))}
			checkout.UserID = userID
			rows, err = tx.QueryContext(
				ctx,
				`SELECT c.id, c.created_at, c.status, sr.product_id, sr.count FROM checkouts AS c 
				JOIN stock_reservations AS sr ON c.id = sr.checkout_id
				WHERE c.status = $1 AND c.basket_id = $2 GROUP BY c.id, sr.product_id`,
				store_domain.CheckoutPending, cart.ID,
			)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("retrieving checkout: %w", err)
			}
			defer rows.Close()
			checkoutProductIds = make([]string, 0, len(cart.Products))
			for rows.Next() {
				var cartProduct store_domain.CartProduct
				err = rows.Scan(
					&checkout.ID,
					&checkout.CreatedAt,
					&checkout.Status,
					&cartProduct.ProductID,
					&cartProduct.Count,
				)
				if err != nil {
					return fmt.Errorf("scanning checkout and stock: %w", err)
				}
				checkout.Items[cartProduct.ProductID] = cartProduct
				checkoutProductIds = append(checkoutProductIds, cartProduct.ProductID)
			}
		}
		if !checkout.IsZero() {
			rows, err = tx.QueryContext(
				ctx,
				`SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id ANY($1) FOR UPDATE`,
				checkoutProductIds,
			)
			if err != nil {
				return fmt.Errorf("retrieving stock and resrevations: %w", err)
			}
			defer rows.Close()
			stock = store_domain.Stock{Items: make(map[string]store_domain.StockItem, len(checkoutProductIds))}
			for rows.Next() {
				var stockItem store_domain.StockItem

				err := rows.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
				if err != nil {
					return fmt.Errorf("scanning stock: %w", err)
				}
				stock.Items[stockItem.ProductID] = stockItem
			}
		}
		err = upsertFn(&cart, &checkout, &stock, stockItem)
		if err != nil {
			return fmt.Errorf("upserting domain fn: %w", err)
		}
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO baskets (id, customer_id, last_modified_at, status) VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET last_modified_at = $3, status = $4`,
			cart.ID, cart.CustomerID, cart.LastModifiedAt, cart.Status,
		)
		if err != nil {
			return fmt.Errorf("upserting onto baskets: %w", err)
		}
		productToUpdate := cart.Products[item.ProductID]
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO basket_products (id, product_id, count) VALUES ($1,$2,$3) 
			ON CONFLICT (id,product_id) DO UPDATE SET count = $3`,
			cart.ID, productToUpdate.ProductID, productToUpdate.Count,
		)
		if err != nil {
			return fmt.Errorf("updating basket products: %w", err)
		}
		if !checkout.IsZero() {
			_, err = tx.ExecContext(
				ctx,
				`UPDATE checkouts SET status = $1 WHERE id = $2`,
				checkout.Status, checkout.ID,
			)
			if err != nil {
				return fmt.Errorf("updating checkout %s: %w", checkout.ID, err)
			}
			for productID, stockItem := range stock.Items {
				_, err = tx.ExecContext(
					ctx,
					`UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3`,
					stockItem.AvailableAmount, stockItem.ReservedAmount, productID,
				)
				if err != nil {
					return fmt.Errorf("updating stock item %s: %w", productID, err)
				}
			}
		}
		return nil
	})
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
