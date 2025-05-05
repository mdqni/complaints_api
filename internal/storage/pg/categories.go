package pg

import (
	"complaint_server/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"complaint_server/internal/domain"
)

func (s *Storage) GetCategories(ctx context.Context) ([]domain.Category, error) {
	const op = "storage.categories.GetCategories"
	rows, err := s.db.Query(ctx, `
		SELECT uuid, title, description, answer
		FROM categories`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(&category.ID, &category.Title, &category.Description, &category.Answer)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return categories, nil
}
func (s *Storage) GetCategoryById(ctx context.Context, categoryID uuid.UUID) (domain.Category, error) {
	const op = "storage.categories.GetCategoriesByID"
	query := `
		SELECT uuid, title, description, answer
		FROM categories WHERE uuid = $1`
	var category domain.Category
	err := s.db.QueryRow(ctx, query, categoryID).Scan(
		&category.ID,
		&category.Title,
		&category.Description,
		&category.Answer,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Category{}, storage.ErrCategoryNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("%s: %w", op, err)
	}

	return category, nil
}
func (s *Storage) CreateCategory(ctx context.Context, category domain.Category) (uuid.UUID, error) {
	const op = "storage.categories.CreateCategory"
	query := `INSERT INTO categories (title, description, answer) VALUES ($1, $2, $3) RETURNING uuid`
	var categoryID uuid.UUID
	err := s.db.QueryRow(ctx, query, category.Title, category.Description, category.Answer).Scan(&categoryID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, storage.ErrCategoryNotFound
	}
	if errors.Is(err, sql.ErrConnDone) {
		return uuid.Nil, storage.ErrDBConnection
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: failed to save categories: %w", op, err)
	}

	return categoryID, nil
}

func (s *Storage) CategoryExists(ctx context.Context, name string) (bool, error) {
	const op = "storage.categories.CategoryExists"
	query := `SELECT COUNT(1) FROM categories WHERE title = $1`
	var count int
	err := s.db.QueryRow(ctx, query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check categories existence: %w", op, err)
	}
	return count > 0, nil
}

func (s *Storage) DeleteCategoryById(ctx context.Context, id uuid.UUID) error {
	const op = "storage.categories.DeleteCategoryById"
	query := `DELETE FROM categories WHERE uuid = $1`
	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return storage.ErrHasRelatedRows
		}
		return fmt.Errorf("не удалось удалить категорию: %w", err)
	}
	return nil
}
func (s *Storage) UpdateCategory(ctx context.Context, id uuid.UUID, category domain.Category) (uuid.UUID, error) {
	const op = "storage.categories.UpdateCategory"

	query := `
		UPDATE categories
		SET uuid = $1,
		    title = $2,
			description = $3,
			answer = $4
		WHERE uuid = $5
	`

	cmdTag, err := s.db.Exec(ctx, query, category.ID, category.Title, category.Description, category.Answer, id)
	if cmdTag.RowsAffected() == 0 {
		return uuid.UUID{}, storage.ErrCategoryNotFound
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: failed to update categories: %w", op, err)
	}

	return category.ID, nil
}
