package updateCategory

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
	Id          int    `json:"id"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Answer      string `json:"answer" validate:"required"`
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
// @Router /category [put]
func New(ctx context.Context, log *slog.Logger, service *service.CategoryService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.update.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err)) //Пишем в лог
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request", http.StatusBadRequest)) //Возвращаем ошибку
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			errors.As(err, &validationErrors)
			log.Error("failed to validate request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validationErrors))
			return
		}
		log.Info("Req: ", req)
		description := req.Description
		categoryName := req.Title
		answer := req.Answer
		id := req.Id
		_, err = service.UpdateCategory(r.Context(), domain.Category{ID: id, Title: categoryName, Description: description, Answer: answer})
		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to update category", http.StatusInternalServerError))
		}
		client.Del(ctx, "cache:/categories")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.Response{
			StatusCode: http.StatusOK,
			Data:       map[string]interface{}{"id": id},
		})
	}
}
