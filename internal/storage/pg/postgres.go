package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(connString string) (*Storage, error) {
	const op = "storage.postgres.New"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,

		`SET search_path TO public;`,

		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'complaint_status') THEN
				CREATE TYPE complaint_status AS ENUM ('pending', 'approved', 'rejected');
			END IF;
		END
		$$;`,

		`CREATE TABLE IF NOT EXISTS categories (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL,
			answer TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS complaints (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			barcode INTEGER NOT NULL,
			category_id UUID NOT NULL,
			message TEXT NOT NULL,
			status complaint_status NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			answer TEXT,
			FOREIGN KEY (category_id) REFERENCES categories(uuid)
		);`,

		`CREATE TABLE IF NOT EXISTS admins (
			id SERIAL PRIMARY KEY,
			barcode INTEGER UNIQUE NOT NULL
		);`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return nil, fmt.Errorf("%s: failed to execute statement: %w\nSQL: %s", op, err, stmt)
		}
	}

	return &Storage{db: pool}, nil
}
