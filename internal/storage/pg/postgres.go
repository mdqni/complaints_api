package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

// New открывает соединение с БД и создаёт таблицы, если их нет.
func New(ctx context.Context, connString string) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	// Создание таблиц
	tables := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL,
			answer TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			uuid TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS complaints (
    id SERIAL PRIMARY KEY,
    user_uuid TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answer TEXT,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    FOREIGN KEY (user_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);`,
	}

	for _, table := range tables {
		if _, err := pool.Exec(ctx, table); err != nil {
			pool.Close() // ✅ Закрываем соединение при ошибке
			return nil, fmt.Errorf("%s: failed to create tables: %w", op, err)
		}
	}

	return &Storage{db: pool}, nil
}
