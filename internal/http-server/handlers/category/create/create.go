package categoriesCreate

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	service "complaint_server/internal/service/category"
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

type Request struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Answer      string `json:"answer" validate:"required"`
}

// New @Summary Создать категорию
// @Description Создает новую категорию жалоб с необходимыми данными: название, описание и ответ.
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body Request true "Данные категории"
// @Success 200 {object} response.Response "Категория успешно создана"
// @Failure 400 {object} response.Response "Ошибка валидации или декодирования данных"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /admin/categories [post]
func New(ctx context.Context, log *slog.Logger, service *service.CategoryService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.create.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Failed to decode request body",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// Валидация данных
		if err := validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				log.Error("validation failed", sl.Err(err))
				render.JSON(w, r, response.Response{
					Message:    "Validation failed",
					StatusCode: http.StatusBadRequest,
					Data:       validationErrors,
				})
				return
			}
			log.Error("unknown validation error", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Unknown validation error",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}

		// Создание категории
		categoryID, err := service.CreateCategory(r.Context(), domain.Category{
			Title:       req.Title,
			Description: req.Description,
			Answer:      req.Answer,
		})
		if err != nil {
			log.Error("failed to save category", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Failed to save category",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		log.Info("category saved", slog.Any("id", categoryID))
		client.Del(ctx, "cache:/categories")

		render.JSON(w, r, response.Response{
			Message:    "Category created successfully",
			StatusCode: http.StatusOK,
			Data:       categoryID,
		})
	}
}
