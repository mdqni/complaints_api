package updateCategory

import (
	"complaint_server/internal/domain"
	service "complaint_server/internal/service/category"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"complaint_server/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

type Request struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Answer      string    `json:"answer" validate:"required"`
}

// New @Summary Обновить категорию
// @Description Обновляет информацию о категории жалоб. Требуется предоставить ID категории и новые данные (название, описание и ответ).
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body Request true "Данные категории"
// @Success 200 {object} response.Response "Категория успешно обновлена"
// @Failure 400 {object} response.Response "Ошибка валидации или декодирования данных"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /admin/categories/{id} [put]
func New(ctx context.Context, log *slog.Logger, service *service.CategoryService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.update.New"
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
		categoryId := chi.URLParam(r, "id")
		if categoryId == "" {
			log.Error("Missing category id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing category id"})
			return
		}
		id, err := uuid.Parse(categoryId)
		_, err = service.UpdateCategory(r.Context(), id, domain.Category{
			ID:          req.Id,
			Title:       req.Title,
			Description: req.Description,
			Answer:      req.Answer,
		})
		if errors.Is(err, storage.ErrCategoryNotFound) {
			log.Error(op, sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				Message:    fmt.Sprintf("category with %d not found", id),
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(responseData)
			return
		}
		if err != nil {
			log.Error("failed to update category", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Failed to update category",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		client.Del(ctx, "cache:/categories")

		render.JSON(w, r, response.Response{
			Message:    "Category updated successfully",
			StatusCode: http.StatusOK,
			Data:       map[string]interface{}{"id": req.Id},
		})
	}
}
