package pg

import "C"
import (
	"context"
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

func (s *Storage) GetCategoryById(ctx context.Context, categoryId int) (domain.Category, error) {
	const op = "storage.category.GetCategoryById"
	query := `
		SELECT *
		FROM categories
		WHERE id = $1`
	var category domain.Category
	err := s.db.QueryRow(ctx, query, categoryId).Scan(
		&category.ID,
		&category.Title,
		&category.Description,
		&category.Answer,
	)

	if err != nil {
		return domain.Category{}, fmt.Errorf("%s: %w", op, err)
	}

	return category, nil
}

func (s *Storage) CreateCategory(ctx context.Context, category domain.Category) (int64, error) {
	const op = "storage.category.CreateCategory"
	query := `INSERT INTO categories (title, description, answer) VALUES ($1, $2, $3) RETURNING id`
	var categoryID int64
	err := s.db.QueryRow(ctx, query, category.Title, category.Description, category.Answer).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to save category: %w", op, err)
	}

	return categoryID, nil
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
