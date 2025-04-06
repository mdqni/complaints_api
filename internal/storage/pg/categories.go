package pg

import (
	"complaint_server/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"complaint_server/internal/domain"
)

func (s *Storage) GetCategories(ctx context.Context) ([]domain.Category, error) {
	const op = "storage.category.GetCategories"
	rows, err := s.db.Query(ctx, `
		SELECT id, title, description, answer
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
func (s *Storage) GetCategoryById(ctx context.Context, categoryID int) (domain.Category, error) {
	const op = "storage.category.GetCategoriesByID"
	query := `
		SELECT id, title, description, answer
		FROM categories WHERE id = $1`
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
func (s *Storage) CreateCategory(ctx context.Context, category domain.Category) (int64, error) {
	const op = "storage.category.CreateCategory"
	query := `INSERT INTO categories (title, description, answer) VALUES ($1, $2, $3) RETURNING id`
	var categoryID int64
	err := s.db.QueryRow(context.Background(), query, category.Title, category.Description, category.Answer).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to save category: %w", op, err)
	}

	return categoryID, nil
}

func (s *Storage) CategoryExists(name string) (bool, error) {
	const op = "storage.category.CategoryExists"
	query := `SELECT COUNT(1) FROM categories WHERE title = $1`
	var count int
	err := s.db.QueryRow(context.Background(), query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check category existence: %w", op, err)
	}
	return count > 0, nil
}

func (s *Storage) DeleteCategoryById(ctx context.Context, index int) error {
	const op = "storage.category.DeleteCategoryById"
	query := `DELETE FROM categories WHERE id = $1`
	_, err := s.db.Exec(ctx, query, index)
	if err != nil {
		return fmt.Errorf("%s: failed to delete category: %w", op, err)
	}
	return nil
}
func (s *Storage) UpdateCategory(ctx context.Context, category domain.Category) (int, error) {
	const op = "storage.category.UpdateCategory"

	query := `
		UPDATE categories
		SET title = $1,
			description = $2,
			answer = $3
		WHERE id = $4
	`

	cmdTag, err := s.db.Exec(ctx, query, category.Title, category.Description, category.Answer, category.ID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to update category: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return 0, storage.ErrCategoryNotFound
	}

	return category.ID, nil
}
