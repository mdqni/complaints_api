package serviceCategory

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"context"
	"github.com/google/uuid"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoriesService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category domain.Category) (uuid.UUID, error) {
	answer, err := s.repo.Create(ctx, category)
	return answer, err
}

func (s *CategoryService) UpdateCategory(ctx context.Context, uuid_ uuid.UUID, category domain.Category) (uuid.UUID, error) {
	id, err := s.repo.Update(ctx, uuid_, category)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func (s *CategoryService) GetCategoryById(ctx context.Context, categoryId uuid.UUID) (domain.Category, error) {
	category, err := s.repo.GetByID(ctx, categoryId)
	if err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (s *CategoryService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	categories, err := s.repo.GetAll(ctx)
	return categories, err
}

func (s *CategoryService) DeleteCategoryById(ctx context.Context, complaintID uuid.UUID) error {
	return s.repo.Delete(ctx, complaintID)
}
