package store_repository

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/siderustler/go-ecommerce/store"
	store_domain "github.com/siderustler/go-ecommerce/store/domain"
)

type repository struct {
	db *sql.DB
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var _ store.Repository = repository{}

func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS baskets (
			id UUID PRIMARY KEY,
			customer_id TEXT NOT NULL,
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
			customer_id TEXT NOT NULL REFERENCES customers(customer_id),
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
			product_id UUID PRIMARY KEY NOT NULL REFERENCES products(id),
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
	_, err = db.ExecContext(ctx,
		`INSERT INTO stock (product_id, available_amount, reserved_amount) SELECT id, 9999,0 FROM products ON CONFLICT (product_id) DO NOTHING`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating stock_reservations table: %w", err)
	}
	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY, 
			checkout_id UUID REFERENCES checkouts(id),
			created_at TIMESTAMP NOT NULL,
			status TEXT CHECK (status IN ('FINALIZED', 'PAID', 'SHIPPING'))
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating orders table: %w", err)
	}
	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS order_products (
		id SERIAL PRIMARY KEY,
		order_id UUID REFERENCES orders(id),
		name TEXT NOT NULL,
		price REAL NOT NULL,
		count INT NOT NULL
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating order_products table: %w", err)
	}
	return &repository{db: db}, nil
}

func products(ctx context.Context, exec executor, ids ...string) ([]store_domain.Product, error) {
	row, err := exec.QueryContext(ctx, `SELECT id, name, price, price_before FROM products WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("retrieving products: %w", err)
	}
	products := make([]store_domain.Product, 0, len(ids))
	for row.Next() {
		var product store_domain.Product
		err = row.Scan(&product.ID, &product.Name, &product.ActualPrice, &product.DiscountPrice)
		if err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		products = append(products, product)
	}
	return products, nil
}

// CreateOrder implements store.Repository.
func (r repository) CreateOrder(
	ctx context.Context,
	checkoutID string,
	createFn func(
		cart *store_domain.Cart,
		checkout *store_domain.Checkout,
		stock *store_domain.Stock,
		products []store_domain.Product,
	) (store_domain.Order, error),
) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		checkout, err := checkoutByID(ctx, tx, checkoutID)
		if err != nil {
			return fmt.Errorf("retrieving checkout by id: %w", err)
		}
		stock, err := stockForUpdate(ctx, tx, slices.Collect(maps.Keys(checkout.Items))...)
		if err != nil {
			return fmt.Errorf("retrieving stock: %w", err)
		}
		cart, err := cart(ctx, tx, checkout.UserID)
		if err != nil {
			return fmt.Errorf("retrieving cart: %w", err)
		}
		products, err := products(ctx, tx, slices.Collect(maps.Keys(cart.Products))...)
		if err != nil {
			return fmt.Errorf("retrieving products: %w", err)
		}
		order, err := createFn(&cart, &checkout, &stock, products)
		if err != nil {
			return fmt.Errorf("domain: %w", err)
		}
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO orders (id, checkout_id, status, created_at) 
			VALUES ($1, $2, $3, $4)`,
			order.ID, order.CheckoutID, order.Status, order.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("inserting order: %w", err)
		}
		for _, orderProduct := range order.Products {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO order_products (order_id, name, price, count) 
				VALUES ($1, $2, $3, $4, $5)`,
				order.ID, orderProduct.Name, orderProduct.ItemPrice, orderProduct.Count,
			)
			if err != nil {
				return fmt.Errorf("inserting order product: %w", err)
			}
		}
		_, err = tx.ExecContext(
			ctx,
			`UPDATE checkouts SET status = $1 WHERE id = $2`,
			checkout.Status, checkout.ID,
		)
		if err != nil {
			return fmt.Errorf("updating checkout: %w", err)
		}
		_, err = tx.ExecContext(ctx, "UPDATE baskets SET status = $1 WHERE id = $2", cart.Status, cart.ID)
		if err != nil {
			return fmt.Errorf("updating cart: %w", err)
		}
		for itemID, stockItem := range stock.Items {
			_, err = tx.ExecContext(
				ctx,
				`UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3`,
				stockItem.AvailableAmount, stockItem.ReservedAmount, itemID,
			)
			if err != nil {
				return fmt.Errorf("updating stock: %w", err)
			}
		}
		return nil
	})
}

// CheckoutByUserID implements store.Repository.
func (r repository) CheckoutByUserID(ctx context.Context, userID string) (store_domain.Checkout, error) {
	return checkoutByUserID(ctx, r.db, userID)
}

// MergeUserCarts implements store.Repository.
func (r repository) MergeUserCarts(
	ctx context.Context,
	fromUserID string,
	toUserID string,
	mergeFn func(
		fromCart store_domain.Cart,
		toCart *store_domain.Cart,
		fromCheckout *store_domain.Checkout,
		toCheckout *store_domain.Checkout,
		stock *store_domain.Stock,
	) error,
) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		fromCart, err := cart(ctx, tx, fromUserID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving from cart: %w", err)
		}
		toCart, err := cart(ctx, tx, toUserID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving to cart: %w", err)
		}
		var fromCheckout store_domain.Checkout
		if !fromCart.IsZero() {
			fromCheckout, err = checkoutByBasketID(ctx, tx, fromCart.ID)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("retrieving from checkout: %w", err)
			}
		}
		var toCheckout store_domain.Checkout
		if !toCart.IsZero() {
			toCheckout, err = checkoutByBasketID(ctx, tx, toCart.ID)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("retrieving to checkout: %w", err)
			}
		}
		stock, err := stockForUpdate(ctx, tx, slices.AppendSeq(slices.Collect(maps.Keys(toCart.Products)), maps.Keys(fromCart.Products))...)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving stock for update: %w", err)
		}
		err = mergeFn(fromCart, &toCart, &fromCheckout, &toCheckout, &stock)
		if err != nil {
			return fmt.Errorf("domain merge: %w", err)
		}
		if !fromCart.IsZero() {
			_, err = tx.ExecContext(
				ctx,
				"UPDATE baskets SET status = $1 WHERE id = $2",
				fromCart.Status,
				fromCart.ID,
			)
			if err != nil {
				return fmt.Errorf("updating from cart: %w", err)
			}
		}
		if !toCart.IsZero() {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO baskets (id, customer_id, status, last_modified_at) VALUES ($1, $2, $3, $4) 
				ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, last_modified_at = EXCLUDED.last_modified_at`,
				toCart.ID,
				toCart.CustomerID,
				toCart.Status,
				toCart.LastModifiedAt,
			)
			if err != nil {
				return fmt.Errorf("updating to cart: %w", err)
			}
		}
		for _, product := range toCart.Products {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO basket_products (id, product_id, count) VALUES ($1, $2, $3)
				 ON CONFLICT (id,product_id) DO UPDATE SET count = EXCLUDED.count`,
				toCart.ID, product.ProductID, product.Count,
			)
			if err != nil {
				return fmt.Errorf("upserting cart products: %w", err)
			}
		}
		if !fromCheckout.IsZero() {
			_, err = tx.ExecContext(ctx, "UPDATE checkouts SET status = $1 WHERE id = $2", fromCheckout.Status, fromCheckout.ID)
			if err != nil {
				return fmt.Errorf("updating from checkout: %w", err)
			}
		}
		if !toCheckout.IsZero() {
			//FIXME -- merge to single clause with upper
			_, err = tx.ExecContext(ctx, "UPDATE checkouts SET status = $1 WHERE id = $2", toCheckout.Status, toCheckout.ID)
			if err != nil {
				return fmt.Errorf("updating to checkout: %w", err)
			}
		}
		for _, stockItem := range stock.Items {
			_, err = tx.ExecContext(
				ctx,
				"UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3",
				stockItem.AvailableAmount, stockItem.ReservedAmount, stockItem.ProductID,
			)
			if err != nil {
				return fmt.Errorf("updating stock item: %w", err)
			}
		}
		return nil
	})
}

func cart(ctx context.Context, exec executor, userID string) (store_domain.Cart, error) {
	rows, err := exec.QueryContext(
		ctx,
		`SELECT b.id, b.customer_id, b.last_modified_at, b.status, bp.product_id, bp.count FROM baskets AS b 
		JOIN basket_products AS bp ON bp.id = b.id WHERE b.customer_id = $1 AND b.status = $2 AND bp.count > 0`,
		userID, store_domain.CartActive,
	)
	if err != nil {
		return store_domain.Cart{}, fmt.Errorf("retrieving cart: %w", err)
	}
	defer rows.Close()
	cart := store_domain.Cart{Products: make(map[string]store_domain.CartProduct)}
	for rows.Next() {
		var cartProduct store_domain.CartProduct
		err := rows.Scan(&cart.ID, &cart.CustomerID, &cart.LastModifiedAt, &cart.Status, &cartProduct.ProductID, &cartProduct.Count)
		if err != nil {
			return store_domain.Cart{}, fmt.Errorf("scanning cart: %w", err)
		}
		cart.Products[cartProduct.ProductID] = cartProduct
	}
	return cart, nil
}

// Cart implements store.Repository.
func (r repository) Cart(ctx context.Context, userID string) (store_domain.Cart, error) {
	return cart(ctx, r.db, userID)
}

// CartCount implements store.Repository.
func (r repository) CartCount(ctx context.Context, userID string) (int, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(bp.id) FROM basket_products AS bp JOIN baskets AS b ON bp.id = b.id WHERE b.customer_id = $1 AND bp.count > 0`,
		userID,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("retrieving basket count: %w", err)
	}
	return count, nil
}

func stockForUpdate(ctx context.Context, exec executor, stockItemIds ...string) (store_domain.Stock, error) {
	stock := store_domain.Stock{Items: make(map[string]store_domain.StockItem, len(stockItemIds))}
	rows, err := exec.QueryContext(
		ctx,
		`SELECT product_id, available_amount, reserved_amount 
				FROM stock WHERE product_id = ANY ($1) FOR UPDATE`,
		stockItemIds,
	)
	if err != nil {
		return store_domain.Stock{}, fmt.Errorf("retrieving stock: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var stockItem store_domain.StockItem
		err := rows.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
		if err != nil {
			return store_domain.Stock{}, fmt.Errorf("scanning stock item: %w", err)
		}
		stock.Items[stockItem.ProductID] = stockItem
	}
	return stock, nil
}

// CreateCheckout implements store.Repository.
func (r repository) CreateCheckout(
	ctx context.Context,
	userID string,
	insertFn func(cart *store_domain.Cart, stock *store_domain.Stock) (store_domain.Checkout, error),
) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		var exists int
		row := tx.QueryRowContext(ctx, `SELECT 1 FROM checkouts WHERE status = $1 AND customer_id = $2`, store_domain.CheckoutPending, userID)
		_ = row.Scan(&exists)
		if exists == 1 {
			return nil
		}
		cart, err := cart(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf("retrieving cart: %w", err)
		}
		cartProductIds := slices.Collect(maps.Keys(cart.Products))
		var stock store_domain.Stock
		if !cart.IsZero() {
			stock, err = stockForUpdate(ctx, tx, cartProductIds...)
			if err != nil {
				return fmt.Errorf("stock for update: %w", err)
			}
		}
		checkout, err := insertFn(&cart, &stock)
		if err != nil {
			return fmt.Errorf("insertfn domain checkout: %w", err)
		}
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO checkouts (id, customer_id, basket_id, created_at, status) VALUES ($1,$2,$3,$4,$5)`,
			checkout.ID, checkout.UserID, cart.ID, checkout.CreatedAt, checkout.Status,
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

func stockItem(ctx context.Context, exec executor, itemID string) (store_domain.StockItem, error) {
	var stockItem store_domain.StockItem
	row := exec.QueryRowContext(
		ctx,
		`SELECT product_id, available_amount, reserved_amount 
			 FROM stock WHERE product_id = $1`,
		itemID,
	)
	err := row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
	if err != nil {
		return store_domain.StockItem{}, fmt.Errorf("scanning stock item: %w", err)
	}
	return stockItem, nil
}

func checkoutByID(ctx context.Context, exec executor, id string) (store_domain.Checkout, error) {
	checkout := store_domain.Checkout{Items: make(map[string]store_domain.CartProduct)}
	rows, err := exec.QueryContext(
		ctx,
		`SELECT c.id, c.customer_id, c.created_at, c.status, sr.product_id, sr.amount FROM checkouts AS c 
				JOIN stock_reservations AS sr ON c.id = sr.checkout_id
				WHERE c.status = $1 AND c.id = $2 GROUP BY c.id, sr.product_id, sr.amount`,
		store_domain.CheckoutPending, id,
	)
	if err != nil {
		return store_domain.Checkout{}, fmt.Errorf("retrieving checkout: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var checkoutProduct store_domain.CartProduct
		err = rows.Scan(
			&checkout.ID,
			&checkout.UserID,
			&checkout.CreatedAt,
			&checkout.Status,
			&checkoutProduct.ProductID,
			&checkoutProduct.Count,
		)
		if err != nil {
			return store_domain.Checkout{}, fmt.Errorf("scanning checkout and stock: %w", err)
		}
		checkout.Items[checkoutProduct.ProductID] = checkoutProduct
	}
	return checkout, nil
}

func checkoutByBasketID(ctx context.Context, exec executor, basketID string) (store_domain.Checkout, error) {
	checkout := store_domain.Checkout{Items: make(map[string]store_domain.CartProduct)}
	rows, err := exec.QueryContext(
		ctx,
		`SELECT c.id, c.customer_id, c.created_at, c.status, sr.product_id, sr.amount FROM checkouts AS c 
				JOIN stock_reservations AS sr ON c.id = sr.checkout_id
				WHERE c.status = $1 AND c.basket_id = $2 GROUP BY c.id, sr.product_id, sr.amount`,
		store_domain.CheckoutPending, basketID,
	)
	if err != nil {
		return store_domain.Checkout{}, fmt.Errorf("retrieving checkout: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var checkoutProduct store_domain.CartProduct
		err = rows.Scan(
			&checkout.ID,
			&checkout.UserID,
			&checkout.CreatedAt,
			&checkout.Status,
			&checkoutProduct.ProductID,
			&checkoutProduct.Count,
		)
		if err != nil {
			return store_domain.Checkout{}, fmt.Errorf("scanning checkout and stock: %w", err)
		}
		checkout.Items[checkoutProduct.ProductID] = checkoutProduct
	}
	return checkout, nil
}

func checkoutByUserID(ctx context.Context, exec executor, userID string) (store_domain.Checkout, error) {
	checkout := store_domain.Checkout{Items: make(map[string]store_domain.CartProduct)}
	rows, err := exec.QueryContext(
		ctx,
		`SELECT c.id, c.customer_id, c.created_at, c.status, sr.product_id, sr.amount FROM checkouts AS c 
				JOIN stock_reservations AS sr ON c.id = sr.checkout_id
				WHERE c.status = $1 AND c.customer_id = $2 GROUP BY c.id, sr.product_id, sr.amount`,
		store_domain.CheckoutPending, userID,
	)
	if err != nil {
		return store_domain.Checkout{}, fmt.Errorf("retrieving checkout: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var checkoutProduct store_domain.CartProduct
		err = rows.Scan(
			&checkout.ID,
			&checkout.UserID,
			&checkout.CreatedAt,
			&checkout.Status,
			&checkoutProduct.ProductID,
			&checkoutProduct.Count,
		)
		if err != nil {
			return store_domain.Checkout{}, fmt.Errorf("scanning checkout and stock: %w", err)
		}
		checkout.Items[checkoutProduct.ProductID] = checkoutProduct
	}
	return checkout, nil
}

// UpsertCart implements store.Repository.
func (r repository) UpsertCart(ctx context.Context, userID string, item store_domain.CartProduct, upsertFn func(cart *store_domain.Cart, checkout *store_domain.Checkout, stock *store_domain.Stock, stockItem store_domain.StockItem) error) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		cart, err := cart(ctx, tx, userID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving user cart: %w", err)
		}
		stockItem, err := stockItem(ctx, tx, item.ProductID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("retrieving stock item: %w", err)
		}
		var domainCheckout store_domain.Checkout
		var domainStock store_domain.Stock
		if !cart.IsZero() {
			domainCheckout, err = checkoutByBasketID(ctx, tx, cart.ID)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("retrieving checkout: %w", err)
			}
		}
		checkoutProductIds := slices.Collect(maps.Keys(domainCheckout.Items))

		if !domainCheckout.IsZero() {
			domainStock, err = stockForUpdate(ctx, tx, checkoutProductIds...)
			if err != nil {
				return fmt.Errorf("retrieving stock for update: %w", err)
			}
		}
		err = upsertFn(&cart, &domainCheckout, &domainStock, stockItem)
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
		if !domainCheckout.IsZero() {
			_, err = tx.ExecContext(
				ctx,
				`UPDATE checkouts SET status = $1 WHERE id = $2`,
				domainCheckout.Status, domainCheckout.ID,
			)
			if err != nil {
				return fmt.Errorf("updating checkout %s: %w", domainCheckout.ID, err)
			}
			for productID, stockItem := range domainStock.Items {
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
