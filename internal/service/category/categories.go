package serviceCategory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"complaint_server/internal/storage"

	"github.com/google/uuid"
)

const cacheKey = "cache:/categories"

type CategoryService struct {
	repo  repository.CategoryRepository
	cache storage.Cache
}

func NewCategoriesService(repo repository.CategoryRepository, cache storage.Cache) *CategoryService {
	return &CategoryService{repo: repo, cache: cache}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category domain.Category) (uuid.UUID, error) {
	id, err := s.repo.Create(ctx, category)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return id, err
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, category domain.Category) (uuid.UUID, error) {
	updatedID, err := s.repo.Update(ctx, id, category)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return updatedID, err
}

func (s *CategoryService) GetCategoryById(ctx context.Context, categoryId uuid.UUID) (domain.Category, error) {
	return s.repo.GetByID(ctx, categoryId)
}

func (s *CategoryService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && len(cached) > 0 {
		var categories []domain.Category
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			return categories, nil
		}
	}

	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrCategoryNotFound) {
			return nil, err
		}
		return nil, err
	}

	data, err := json.Marshal(categories)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return categories, nil
}

func (s *CategoryService) DeleteCategoryById(ctx context.Context, categoryID uuid.UUID) error {
	err := s.repo.Delete(ctx, categoryID)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return err
}
