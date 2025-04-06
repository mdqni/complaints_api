package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage/pg"
	"context"
	"fmt"
)

type CategoryService struct {
	storage *pg.Storage
}

func NewCategoriesService(strg *pg.Storage) *CategoryService {
	return &CategoryService{storage: strg}
}

// CreateCategory создаёт категорию
func (s *CategoryService) CreateCategory(ctx context.Context, category domain.Category) (int64, error) {
	// Сохраняем жалобу
	answer, err := s.storage.CreateCategory(ctx, category)
	if err != nil {
		return answer, fmt.Errorf("failed to save category: %w", err)
	}

	return answer, nil
}

// UpdateCategory обновляет категорию
func (s *CategoryService) UpdateCategory(ctx context.Context, category domain.Category) (int, error) {
	const op = "storage.category.UpdateCategory"
	id, err := s.storage.UpdateCategory(ctx, category)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetCategoryById получает жалобу по ID
func (s *CategoryService) GetCategoryById(ctx context.Context, categoryId int) (domain.Category, error) {
	category, err := s.storage.GetCategoryById(ctx, categoryId)
	if err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

// GetCategories получает все жалобы
func (s *CategoryService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	categories, err := s.storage.GetCategories(ctx)
	if err != nil {
		return []domain.Category{}, err
	}
	return categories, nil
}

// DeleteComplaint удаляет жалобу
func (s *CategoryService) DeleteCategoryById(ctx context.Context, complaintID int) error {
	return s.storage.DeleteCategoryById(ctx, complaintID)
}
