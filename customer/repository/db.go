package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/siderustler/go-ecommerce/customer"
)

type repository struct {
	db *sql.DB
}

func NewRepository(ctx context.Context, db *sql.DB) (*repository, error) {
	_, err := db.ExecContext(ctx,
		`CREATE TABLE billings (
			id UUID PRIMARY KEY,
			nip_code TEXT,
			company TEXT,
			city TEXT NOT NULL,
			address TEXT NOT NULL,
			postal_code TEXT NOT NULL,
			local TEXT
		) ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating baskets table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE shippings (
			id UUID PRIMARY KEY,
			city TEXT NOT NULL,
			address TEXT NOT NULL,
			postal_code TEXT NOT NULL,
			local TEXT
		) ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating baskets table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE users (
			user_id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			phone TEXT NOT NULL,
			billing UUID REFERENCES billings(id),
			shipping UUID REFERENCES shippings(id)
		) ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating users table: %w", err)
	}
	return &repository{db: db}, nil
}

func (r repository) CustomerByID(ctx context.Context, userID string) (customer.Customer, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT 
		c.user_id, c.name, c.email, c.phone, 
		s.id, s.city, s.address, s.local, s.postal_code,
		b.id, b.city, b.address, b.local, b.postal_code, b.nip_code, b.company
		FROM users AS c 
		LEFT JOIN billings AS b ON c.billing = b.id 
		LEFT JOIN shippings AS s ON c.shipping = s.id
		WHERE c.user_id = $1`,
		userID,
	)
	var cust customer.Customer
	err := row.Scan(
		&cust.ID,
		&cust.Credentials.Name,
		&cust.Credentials.Email,
		&cust.Credentials.Phone,
		&cust.Shipping.ID,
		&cust.Shipping.Address.City,
		&cust.Shipping.Address.Address,
		&cust.Shipping.Address.Local,
		&cust.Shipping.Address.PostalCode,
		&cust.Billing.ID,
		&cust.Billing.Address.City,
		&cust.Billing.Address.Address,
		&cust.Billing.Address.Local,
		&cust.Billing.Address.PostalCode,
		&cust.Billing.NIPCode,
		&cust.Billing.Company,
	)

	if err != nil {
		return customer.Customer{}, fmt.Errorf("scanning customer row: %w", err)
	}
	return cust, nil
}

func (r repository) UpsertCredentials(ctx context.Context, customer customer.Customer) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (user_id, name, email, phone) VALUES ($1, $2, $3, $4) ON CONFLICT DO UPDATE SET name = $2, email = $3, phone = $4",
		customer.ID, customer.Credentials.Name, customer.Credentials.Email, customer.Credentials.Phone,
	)
	if err != nil {
		return fmt.Errorf("creating credentials: %w", err)
	}
	return nil
}

func (r repository) UpsertBillingAddress(ctx context.Context, userID string, billing customer.Billing) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO billings (id, nip_code, company, city, address, postal_code, local) VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT DO UPDATE SET id = $7, nip_code = $1, company = $2, city = $3, address = $4, postal_code = $5, local = $6`,
			billing.NIPCode,
			billing.Company,
			billing.Address.City,
			billing.Address.Address,
			billing.Address.PostalCode,
			billing.Address.Local,
			billing.ID,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
		}

		_, err = r.db.ExecContext(
			ctx,
			`UPDATE users SET billing = $1 WHERE user_id = $2`,
			billing.ID, userID,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
		}
		return nil
	})
}

func (r repository) UpsertShippingAddress(ctx context.Context, userID string, shipping customer.ShippingAddress) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO shippings (id, city, address, postal_code, local) VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT DO UPDATE SET id = $1, city = $2, address = $3, postal_code = $4, local = $5`,
			shipping.ID,
			shipping.Address.City,
			shipping.Address.Address,
			shipping.Address.PostalCode,
			shipping.Address.Local,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
		}

		_, err = r.db.ExecContext(
			ctx,
			`UPDATE users SET shipping = $1 WHERE user_id = $2`,
			shipping.ID, userID,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
		}
		return nil
	})
}
func (r repository) CreateCredentials(ctx context.Context, customer customer.Customer) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (user_id, name, email, phone) VALUES ($1, $2, $3, $4)",
		customer.ID, customer.Credentials.Name, customer.Credentials.Email, customer.Credentials.Phone,
	)
	if err != nil {
		return fmt.Errorf("creating credentials: %w", err)
	}
	return nil
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
