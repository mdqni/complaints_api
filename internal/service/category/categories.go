package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage/pg"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type CategoryService struct {
	storage *pg.Storage
}

func NewCategoriesService(strg *pg.Storage) *CategoryService {
	return &CategoryService{storage: strg}
}

// CreateCategory создаёт категорию
func (s *CategoryService) CreateCategory(ctx context.Context, category domain.Category) (uuid.UUID, error) {
	// Сохраняем жалобу
	answer, err := s.storage.CreateCategory(ctx, category)
	if err != nil {
		return answer, fmt.Errorf("failed to save category: %w", err)
	}

	return answer, nil
}

// UpdateCategory обновляет категорию
func (s *CategoryService) UpdateCategory(ctx context.Context, uuid_ uuid.UUID, category domain.Category) (uuid.UUID, error) {
	const op = "storage.category.UpdateCategory"
	id, err := s.storage.UpdateCategory(ctx, uuid_, category)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

// GetCategoryById получает жалобу по ID
func (s *CategoryService) GetCategoryById(ctx context.Context, categoryId uuid.UUID) (domain.Category, error) {
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

// DeleteCategoryById  удаляет жалобу
func (s *CategoryService) DeleteCategoryById(ctx context.Context, complaintID uuid.UUID) error {
	return s.storage.DeleteCategoryById(ctx, complaintID)
}
