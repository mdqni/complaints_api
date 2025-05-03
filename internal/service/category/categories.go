package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage/pg"
	"context"
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
	answer, err := s.storage.CreateCategory(ctx, category)

	return answer, err
}

// UpdateCategory обновляет категорию
func (s *CategoryService) UpdateCategory(ctx context.Context, uuid_ uuid.UUID, category domain.Category) (uuid.UUID, error) {
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
	return categories, err
}

// DeleteCategoryById  удаляет жалобу
func (s *CategoryService) DeleteCategoryById(ctx context.Context, complaintID uuid.UUID) error {
	return s.storage.DeleteCategoryById(ctx, complaintID)
}
