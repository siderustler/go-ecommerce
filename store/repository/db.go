package repository

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	"github.com/siderustler/go-ecommerce/store"
// )

// type repository struct {
// 	db *sql.DB
// }

// func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
// 	_, err := db.ExecContext(ctx,
// 		`CREATE TABLE IF NOT EXISTS baskets (
// 			id UUID PRIMARY KEY,
// 			customer_id UUID NOT NULL REFERENCES customers(customer_id),
// 			last_modified_at TIMESTAMP NOT NULL,
// 		)`,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating baskets table: %w", err)
// 	}
// 	_, err = db.ExecContext(ctx,
// 		`CREATE TABLE IF NOT EXISTS basket_products (
// 			id UUID NOT NULL REFERENCES baskets(id),
// 			product_id UUID NOT NULL REFERENCES products(id),
// 			count INT NOT NULL,
// 			PRIMARY KEY(id, product_id)
// 		)`,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating basket_products table: %w", err)
// 	}
// 	_, err = db.ExecContext(ctx,
// 		`CREATE TABLE IF NOT EXISTS stock (
// 			product_id UUID NOT NULL REFERENCES products(id),
// 			available_amount INT DEFAULT 0,
// 			reserved_amount INT DEFAULT 0
// 		)`,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating stock table: %w", err)
// 	}
// 	_, err = db.ExecContext(ctx,
// 		`CREATE TABLE IF NOT EXISTS stock_reservations (
// 			checkout_id UUID PRIMARY KEY REFERENCES checkouts(id),
// 			product_id UUID REFERENCES products(id),
// 			amount INT DEFAULT 0,
// 			reserved_at TIMESTAMP NOT NULL
// 		)`,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating stock_reservations table: %w", err)
// 	}
// 	_, err = db.ExecContext(ctx,
// 		`CREATE TABLE IF NOT EXISTS checkouts (
// 			id UUID PRIMARY KEY,
// 			basket_id UUID NOT NULL REFERENCES baskets(id),
// 			created_at TIMESTAMP NOT NULL,
// 			status TEXT CHECK (status IN ('INVALIDATED', 'PENDING', 'FINALIZED'))
// 		)`,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating checkouts table: %w", err)
// 	}
// 	return &repository{db: db}, nil
// }

// func (r repository) InsertStockItem(ctx context.Context, stockItem store.StockItem) error {
// 	_, err := r.db.ExecContext(ctx,
// 		`INSERT INTO stock (product_id, available_amount, reserved_amount) VALUES ($1,$2,$3)`,
// 		stockItem.ProductID, stockItem.AvailableAmount, stockItem.ReservedAmount,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("updating stock: %w", err)
// 	}
// 	return nil
// }

// func (r repository) StockItem(ctx context.Context, itemID string) (store.StockItem, error) {
// 	row := r.db.QueryRowContext(ctx, "SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1", itemID)
// 	var stockItem store.StockItem

// 	err := row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 	if err != nil {
// 		return store.StockItem{}, fmt.Errorf("scanning row: %w", err)
// 	}
// 	return stockItem, nil
// }

// func (r repository) CreateCheckout(
// 	ctx context.Context,
// 	checkout store.Checkout,
// 	onUpdateFn func(stock store.Stock) (updatedStock store.Stock, err error),
// ) error {
// 	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
// 		_, err := tx.ExecContext(ctx,
// 			`INSERT INTO checkouts (id, basket_id, created_at, status) VALUES ($1, $2, $3, $4)`,
// 			checkout.ID, checkout.BasketID, checkout.CreatedAt, checkout.Status,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("inserting to checkouts: %w", err)
// 		}
// 		stock := make(store.Stock, len(checkout.BasketProducts))
// 		for _, basketProduct := range checkout.BasketProducts {
// 			_, err = tx.ExecContext(ctx,
// 				`INSERT INTO stock_reservations (checkout_id, product_id, amount, reserved_at) VALUES ($1, $2, $3, $4)`,
// 				checkout.ID, basketProduct.ProductID, basketProduct.ProductID, reservation.ReservedAt,
// 			)
// 			if err != nil {
// 				return fmt.Errorf("inserting stock reservations: %w", err)
// 			}
// 			row := tx.QueryRowContext(ctx,
// 				`SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1 FOR UPDATE`,
// 				reservation.ProductID,
// 			)
// 			var stockItem store.StockItem
// 			err = row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 			if err != nil {
// 				return fmt.Errorf("scanning stock item: %w", err)
// 			}
// 			stock[store.ProductID(stockItem.ProductID)] = stockItem
// 		}
// 		updatedStock, err := onUpdateFn(stock)
// 		if err != nil {
// 			return fmt.Errorf("on stock update: %w", err)
// 		}
// 		for _, stockItem := range updatedStock {
// 			_, err = tx.ExecContext(ctx,
// 				`UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3`,
// 				stockItem.AvailableAmount, stockItem.ReservedAmount, stockItem.ProductID,
// 			)
// 			if err != nil {
// 				return fmt.Errorf("updating stock: %w", err)
// 			}
// 		}
// 		return nil
// 	})
// }

// func (r repository) UpdateCheckout(
// 	ctx context.Context,
// 	checkout store.Checkout,
// 	onUpdateFn func(stock store.Stock) (updatedStock store.Stock, err error),
// ) error {
// 	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
// 		_, err := tx.ExecContext(ctx,
// 			`UPDATE checkouts SET status = $1 WHERE checkout_id = $2`,
// 			checkout.Status, checkout.ID,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("updating checkouts status: %w", err)
// 		}
// 		stock := make(store.Stock, len(checkout.Reservations))
// 		for _, reservation := range checkout.Reservations {
// 			row := tx.QueryRowContext(ctx,
// 				`SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1 FOR UPDATE`,
// 				reservation.ProductID,
// 			)
// 			var stockItem store.StockItem
// 			err = row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 			if err != nil {
// 				return fmt.Errorf("scanning stock item: %w", err)
// 			}
// 		}
// 		updatedStock, err := onUpdateFn(stock)
// 		if err != nil {
// 			return fmt.Errorf("on stock invalidate: %w", err)
// 		}
// 		for _, stockItem := range updatedStock {
// 			_, err = tx.ExecContext(ctx,
// 				`UPDATE stock SET available_amount = $1, reserved_amount = $2 WHERE product_id = $3`,
// 				stockItem.AvailableAmount, stockItem.ReservedAmount, stockItem.ProductID,
// 			)
// 			if err != nil {
// 				return fmt.Errorf("updating stock: %w", err)
// 			}
// 		}
// 		return nil
// 	})
// }

// func upsertReservation(
// 	ctx context.Context,
// 	exec *sql.Tx,
// 	basketID string,
// 	productID string,
// 	upsertFn func(
// 		stockItem store.StockItem,
// 		actualReservation store.Reservation,
// 	) (updatedReservation store.Reservation, updatedStockItem store.StockItem, err error),
// ) error {
// 	row := exec.QueryRowContext(ctx, "SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1 FOR UPDATE", productID)
// 	var stockItem store.StockItem

// 	err := row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 	if err != nil {
// 		return fmt.Errorf("scanning stock row: %w", err)
// 	}

// 	row = exec.QueryRowContext(ctx, "SELECT amount, reserved_at FROM stock_reservations WHERE basket_id = $1 AND product_id = $2")
// 	actualReservation := store.Reservation{ProductID: productID}
// 	err = row.Scan(&actualReservation.Amount, &actualReservation.ReservedAt)
// 	if err != nil {
// 		return fmt.Errorf("scanning stock_reservations row: %w", err)
// 	}
// 	reservation, updatedStockItem, err := upsertFn(stockItem, actualReservation)
// 	if err != nil {
// 		return fmt.Errorf("upserting domain reservation: %w", err)
// 	}
// 	_, err = exec.ExecContext(
// 		ctx,
// 		`INSERT INTO stock_reservations (basket_id, product_id, amount, reserved_at) VALUES ($1, $2, $3, $4)
// 			 ON CONFLICT DO UPDATE SET amount = $3, reserved_at = $4`,
// 		basketID, reservation.ProductID, reservation.Amount, reservation.ReservedAt,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("upserting repo reservation: %w", err)
// 	}
// 	_, err = exec.ExecContext(ctx,
// 		"UPDATE stock SET available_amount = $2, reserved_amount = $3 WHERE product_id = $1",
// 		productID,
// 		updatedStockItem.AvailableAmount,
// 		updatedStockItem.ReservedAmount,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("update repo stock item: %w", err)
// 	}
// 	return nil
// }

// func (r repository) Checkout(ctx context.Context, checkoutID string) (store.Checkout, error) {
// 	row := r.db.QueryRowContext(ctx, "SELECT checkout_id, created_at FROM checkouts WHERE checkout_id = $1", checkoutID)

// 	var checkout store.Checkout
// 	err := row.Scan(&checkout.ID, &checkout.CreatedAt)
// 	if err != nil {
// 		return store.Checkout{}, fmt.Errorf("scanning checkout row: %w", err)
// 	}
// 	return checkout, nil
// }

// func (r repository) BasketByUserID(ctx context.Context, userID string) (store.Basket, error) {
// 	row := r.db.QueryRowContext(ctx, "SELECT basket_id FROM baskets WHERE customer_id = $1", userID)

// 	var basketID string
// 	err := row.Scan(&basketID)
// 	if err != nil {
// 		return store.Basket{}, fmt.Errorf("scanning basketID: %w", err)
// 	}
// 	rows, err := r.db.QueryContext(ctx, `
// 		SELECT b.id, b.customer_id, b.last_modified_at, bp.count, bp.product_id FROM baskets AS b
// 		JOIN basket_products AS bp ON bp.id = b.id WHERE b.id = $1`,
// 		basketID,
// 	)
// 	if err != nil {
// 		return store.Basket{}, fmt.Errorf("retrieving basket: %w", err)
// 	}
// 	defer rows.Close()

// 	var basket store.Basket
// 	for rows.Next() {
// 		var product store.BasketProduct
// 		err := rows.Scan(&basket.ID, &basket.CustomerID, &basket.LastModifiedAt, &product.Count, &product.ProductID)
// 		if err != nil {
// 			return store.Basket{}, fmt.Errorf("scanning basket: %w", err)
// 		}

// 		basket.Products[store.ProductID(product.ProductID)] = product
// 	}

// 	return basket, nil
// }

// func (r repository) BasketModifyTime(ctx context.Context, basketID string) (string, error) {
// 	row := r.db.QueryRowContext(ctx, "SELECT last_modified_at FROM baskets WHERE id = $1", basketID)

// 	var time string
// 	err := row.Scan(&time)
// 	if err != nil {
// 		return "", fmt.Errorf("scanning basket modify time: %w", err)
// 	}
// 	return time, nil
// }

// func (r repository) UpsertReservations(
// 	ctx context.Context,
// 	basketID string,
// 	productIDs []string,
// 	reservationTime string,
// 	upsertFn func(
// 		stockItem store.StockItem,
// 		actualReservation store.Reservation,
// 	) (updatedReservation store.Reservation, updatedStockItem store.StockItem, err error),
// ) error {
// 	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
// 		for _, productID := range productIDs {
// 			err := upsertReservation(ctx, tx, basketID, productID, upsertFn)
// 			if err != nil {
// 				return fmt.Errorf("updating reservation for productID %s: %w", productID, err)
// 			}
// 		}
// 		_, err := tx.ExecContext(ctx,
// 			`INSERT INTO checkouts (checkout_id, created_at) VALUES ($1, $2)
// 			 ON CONFLICT DO UPDATE SET created_at = $2 WHERE checkout_id = $1`,
// 			basketID, reservationTime,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("upserting checkouts: %w", err)
// 		}
// 		return nil
// 	})
// }

// func (r repository) UpdateStockItem(
// 	ctx context.Context,
// 	itemID string,
// 	updateFn func(item store.StockItem) (updatedItem store.StockItem, err error),
// ) error {
// 	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
// 		row := tx.QueryRowContext(ctx, "SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1 FOR UPDATE", itemID)
// 		var stockItem store.StockItem

// 		err := row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 		if err != nil {
// 			return fmt.Errorf("scanning row: %w", err)
// 		}
// 		updatedItem, err := updateFn(stockItem)
// 		if err != nil {
// 			return fmt.Errorf("update domain stock item: %w", err)
// 		}
// 		_, err = tx.ExecContext(ctx,
// 			"UPDATE stock SET available_amount = $2, reserved_amount = $3 WHERE product_id = $1",
// 			itemID,
// 			updatedItem.AvailableAmount,
// 			updatedItem.ReservedAmount,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("update repo stock item: %w", err)
// 		}

// 		return nil
// 	})
// }

// func (r repository) UpdateBasket(
// 	ctx context.Context,
// 	customerID string,
// 	basketProduct store.BasketProduct,
// 	onUpdateFn func(stockItem store.StockItem) error,
// ) error {
// 	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
// 		//FIXME abstract it with function and interface which will allow db and tx operations
// 		row := tx.QueryRowContext(ctx, `SELECT id, customer_id FROM baskets WHERE customer_id = $1`)

// 		var basketID, customerID string
// 		err := row.Scan(basketID, customerID)
// 		if err != nil {
// 			return fmt.Errorf("retrieving basket: %w", err)
// 		}

// 		//FIXME abstract it with function and interface which will allow db and tx operations
// 		row = r.db.QueryRowContext(ctx, "SELECT product_id, available_amount, reserved_amount FROM stock WHERE product_id = $1", basketProduct.ProductID)
// 		var stockItem store.StockItem

// 		err = row.Scan(&stockItem.ProductID, &stockItem.AvailableAmount, &stockItem.ReservedAmount)
// 		if err != nil {
// 			return fmt.Errorf("retrieving stock item: %w", err)
// 		}
// 		err = onUpdateFn(stockItem)
// 		if err != nil {
// 			return fmt.Errorf("on domain update stock item: %w", err)
// 		}
// 		_, err = tx.ExecContext(ctx,
// 			`INSERT INTO baskets (id, customer_id, last_modified_at) VALUES ($1, $2, $3)
// 			ON CONFLICT DO NOTHING`,
// 			basketID, customerID, time.Now().UTC().Format(time.RFC3339),
// 		)
// 		if err != nil {
// 			return fmt.Errorf("updating basket: %w", err)
// 		}
// 		_, err = tx.ExecContext(ctx,
// 			`INSERT INTO basket_products
// 				(id, product_id, count)
// 				VALUES ($1, $2, $3)
// 				ON CONFLICT DO
// 				UPDATE SET product_id = $2, count = $3 WHERE id = $1`,
// 			basketID, basketProduct.ProductID, basketProduct.Count,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("updating basket products: %w", err)
// 		}

// 		return nil
// 	})
// }

// func RunInTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, txFunc func(tx *sql.Tx) error) (err error) {
// 	tx, err := db.BeginTx(ctx, nil)
// 	defer func() {
// 		if err != nil {
// 			_ = tx.Rollback()
// 		} else {
// 			err = tx.Commit()
// 		}
// 	}()
// 	if err != nil {
// 		return fmt.Errorf("starting transaction: %w", err)
// 	}

// 	return txFunc(tx)
// }
