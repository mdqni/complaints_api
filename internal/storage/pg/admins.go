package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) AdminExistsByBarcode(ctx context.Context, barcode int) (bool, error) {
	query := "SELECT 1 FROM admins WHERE barcode = $1"
	row := s.db.QueryRow(ctx, query, barcode)

	var exists int
	err := row.Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check admin existence: %w", err)
	}

	return true, nil
}
