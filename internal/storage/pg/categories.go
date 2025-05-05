package pg

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"complaint_server/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type categoryRepo struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewCategoryRepo(db *Storage, redis *redis.Client) repository.CategoryRepository {
	return &categoryRepo{db: db.db, redis: redis}
}

const (
	allCategoriesKey = "cache:/categories"
	categoryKey      = "cache:category:"
)

func (c *categoryRepo) Create(ctx context.Context, category domain.Category) (uuid.UUID, error) {
	const op = "storage.categories.CreateCategory"
	query := `INSERT INTO categories (title, description, answer) VALUES ($1, $2, $3) RETURNING uuid`
	var categoryID uuid.UUID
	err := c.db.QueryRow(ctx, query, category.Title, category.Description, category.Answer).Scan(&categoryID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	c.redis.Del(ctx, allCategoriesKey)

	return categoryID, nil
}

func (c *categoryRepo) GetAll(ctx context.Context) ([]domain.Category, error) {
	const op = "storage.categories.GetAll"

	cached, err := c.redis.Get(ctx, allCategoriesKey).Result()
	if err == nil {
		var categories []domain.Category
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			return categories, nil
		}
	}

	rows, err := c.db.Query(ctx, `SELECT uuid, title, description, answer FROM categories`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.Title, &cat.Description, &cat.Answer); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, cat)
	}

	data, _ := json.Marshal(categories)
	c.redis.Set(ctx, allCategoriesKey, data, 0)

	return categories, nil
}

func (c *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	const op = "storage.categories.GetByID"

	key := categoryKey + id.String()
	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cat domain.Category
		if err := json.Unmarshal([]byte(cached), &cat); err == nil {
			return cat, nil
		}
	}

	var category domain.Category
	err = c.db.QueryRow(ctx, `
		SELECT uuid, title, description, answer
		FROM categories WHERE uuid = $1`,
		id,
	).Scan(&category.ID, &category.Title, &category.Description, &category.Answer)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Category{}, storage.ErrCategoryNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("%s: %w", op, err)
	}

	data, _ := json.Marshal(category)
	c.redis.Set(ctx, key, data, 0)

	return category, nil
}

func (c *categoryRepo) Update(ctx context.Context, id uuid.UUID, category domain.Category) (uuid.UUID, error) {
	const op = "storage.categories.Update"

	query := `
		UPDATE categories
		SET title = $1, description = $2, answer = $3
		WHERE uuid = $4`
	res, err := c.db.Exec(ctx, query, category.Title, category.Description, category.Answer, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		return uuid.Nil, storage.ErrCategoryNotFound
	}

	c.redis.Del(ctx, allCategoriesKey, categoryKey+id.String())

	return id, nil
}

func (c *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "storage.categories.Delete"

	_, err := c.db.Exec(ctx, `DELETE FROM categories WHERE uuid = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c.redis.Del(ctx, allCategoriesKey, categoryKey+id.String())

	return nil
}
