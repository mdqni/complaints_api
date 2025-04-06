package pg

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/hash"
	"context"
	"fmt"
)

func (s *Storage) CreateAdmin(ctx context.Context, admin *domain.Admin) error {
	const op = "storage.postgres.CreateAdmin"

	// Хэшируем пароль
	hashedPassword, err := hash.HashPassword(admin.PasswordHash)
	if err != nil {
		return fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO admins (barcode, password_hash, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (barcode) DO NOTHING;
	`, admin.Barcode, string(hashedPassword))

	if err != nil {
		return fmt.Errorf("%s: failed to insert auth: %w", op, err)
	}

	return nil
}

func (s *Storage) GetAdminByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	query := "SELECT id, barcode, password_hash, created_at FROM admins WHERE barcode = $1"
	row := s.db.QueryRow(ctx, query, username)

	var admin domain.Admin
	err := row.Scan(&admin.ID, &admin.Barcode, &admin.PasswordHash, &admin.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth: %w", err)
	}

	return &admin, nil
}
