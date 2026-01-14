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
		`CREATE TABLE IF NOT EXISTS billings (
			id UUID PRIMARY KEY,
			nip_code TEXT,
			company TEXT,
			city TEXT,
			address TEXT,
			postal_code TEXT,
			local TEXT
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating billings table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS shippings (
			id UUID PRIMARY KEY,
			city TEXT,
			address TEXT,
			postal_code TEXT,
			local TEXT
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating shippings table: %w", err)
	}
	_, err = db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS customers (
			customer_id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			phone TEXT,
			billing UUID REFERENCES billings(id),
			shipping UUID REFERENCES shippings(id)
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("creating customers table: %w", err)
	}
	return &repository{db: db}, nil
}

func (r repository) CustomerByID(ctx context.Context, userID string) (customer.Customer, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT 
		c.customer_id, c.name, c.email, c.phone, 
		s.id, s.city, s.address, s.local, s.postal_code,
		b.id, b.city, b.address, b.local, b.postal_code, b.nip_code, b.company
		FROM customers AS c 
		LEFT JOIN billings AS b ON c.billing = b.id 
		LEFT JOIN shippings AS s ON c.shipping = s.id
		WHERE c.customer_id = $1`,
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

func (r repository) CreateCustomer(ctx context.Context, customer customer.Customer) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		if !customer.Billing.IsZero() {
			_, err := tx.ExecContext(
				ctx,
				`INSERT INTO billings (id, nip_code, company, city, address, postal_code, local) VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (id) DO UPDATE SET nip_code = $2, company = $3, city = $4, address = $5, postal_code = $6, local = $7`,
				customer.Billing.ID,
				customer.Billing.NIPCode,
				customer.Billing.Company,
				customer.Billing.Address.City,
				customer.Billing.Address.Address,
				customer.Billing.Address.PostalCode,
				customer.Billing.Address.Local,
			)
			if err != nil {
				return fmt.Errorf("creating repo billings: %w", err)
			}
		}
		if !customer.Shipping.IsZero() {
			_, err := tx.ExecContext(
				ctx,
				`INSERT INTO shippings (id, city, address, postal_code, local) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE SET city = $2, address = $3, postal_code = $4, local = $5`,
				customer.Shipping.ID,
				customer.Shipping.Address.City,
				customer.Shipping.Address.Address,
				customer.Shipping.Address.PostalCode,
				customer.Shipping.Address.Local,
			)
			if err != nil {
				return fmt.Errorf("creating repo shippings: %w", err)
			}
		}
		//FIXME -- multiple functions (update shipping /billing ) with shallow reference to customer
		var err error
		if !customer.Shipping.IsZero() && !customer.Billing.IsZero() {
			statement := `INSERT INTO customers (customer_id, name, email, phone, billing, shipping) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			ON CONFLICT (customer_id) DO UPDATE SET name = $2, email = $3, phone = $4, billing = $5, shipping = $6`
			_, err = tx.ExecContext(ctx,
				statement,
				customer.ID,
				customer.Credentials.Name,
				customer.Credentials.Email,
				customer.Credentials.Phone,
				customer.Billing.ID,
				customer.Shipping.ID,
			)
		} else if !customer.Billing.IsZero() {
			statement := `INSERT INTO customers (customer_id, name, email, phone, billing) 
			VALUES ($1, $2, $3, $4, $5) 
			ON CONFLICT (customer_id) DO UPDATE SET name = $2, email = $3, phone = $4, billing = $5`
			_, err = tx.ExecContext(ctx,
				statement,
				customer.ID,
				customer.Credentials.Name,
				customer.Credentials.Email,
				customer.Credentials.Phone,
				customer.Billing.ID,
			)
		} else if !customer.Shipping.IsZero() {
			statement := `INSERT INTO customers (customer_id, name, email, phone, shipping) 
			VALUES ($1, $2, $3, $4, $5) 
			ON CONFLICT (customer_id) DO UPDATE SET name = $2, email = $3, phone = $4, shipping = $5`
			_, err = tx.ExecContext(ctx,
				statement,
				customer.ID,
				customer.Credentials.Name,
				customer.Credentials.Email,
				customer.Credentials.Phone,
				customer.Shipping.ID,
			)
		} else {
			statement := `INSERT INTO customers (customer_id, name, email, phone)
			 VALUES ($1,$2,$3,$4)
			 ON CONFLICT (customer_id) DO UPDATE SET name = $2, email = $3, phone = $4`
			_, err = tx.ExecContext(ctx,
				statement,
				customer.ID,
				customer.Credentials.Name,
				customer.Credentials.Email,
				customer.Credentials.Phone,
			)
		}
		if err != nil {
			return fmt.Errorf("creating repo customer: %w", err)
		}
		return nil
	})
}

func (r repository) UpdateShippingAddress(ctx context.Context, userID string, shipping customer.ShippingAddress) error {
	return RunInTx(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO shippings (id, city, address, postal_code, local) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE SET city = $2, address = $3, postal_code = $4, local = $5`,
			shipping.ID,
			shipping.Address.City,
			shipping.Address.Address,
			shipping.Address.PostalCode,
			shipping.Address.Local,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE customers SET shipping = $1 WHERE customer_id = $2`,
			shipping.ID, userID,
		)
		if err != nil {
			return fmt.Errorf("creating credentials: %w", err)
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
