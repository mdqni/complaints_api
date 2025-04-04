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
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Answer      string `json:"answer" validate:"required"`
}

// New @Summary      Создать категорию
// @Description  Создает новую категорию жалоб.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        request  body  Request  true  "Данные категории"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response  "Ошибка валидации или декодирования"
// @Failure      500  {object}  response.Response  "Ошибка сервера"
// @Router       /category [post]
func New(ctx context.Context, log *slog.Logger, service *service.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.register.New"
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
		description := req.Description
		categoryName := req.Title
		answer := req.Answer

		categoryID, err := service.CreateCategory(ctx, domain.Category{Title: categoryName, Description: description, Answer: answer})
		if err != nil {
			log.Error("failed to save category", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to save complaints", http.StatusInternalServerError))
		}
		log.Info("category saved on "+strconv.Itoa(int(categoryID))+" ID", slog.Int64("id", categoryID))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.Response{Status: http.StatusOK})
	}
}
