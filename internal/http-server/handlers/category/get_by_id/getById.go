package categoriesGetById

import (
	service "complaint_server/internal/service/category"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"complaint_server/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

// New @Summary Получить категорию по ID
// @Description Возвращает категорию по уникальному идентификатору (ID).
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID (unique identifier of the category)"
// @Success 200 {object} domain.Category "Category details"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 500 {object} response.Response "Internal server error while fetching the category"
// @Router /categories/{id} [get]
func New(log *slog.Logger, service *service.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.get_by_id.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
		)
		ctx := r.Context()

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Error("Missing complaint_id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
			return
		}
		uuid, err := uuid.Parse(id)
		result, err := service.GetCategoryById(ctx, uuid)
		if errors.Is(err, storage.ErrCategoryNotFound) {
			log.Error(op, sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				Message:    fmt.Sprintf("category with %d not found", uuid),
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(responseData)
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				Message:    "internal error",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(responseData)
			return
		}

		log.Info("Category found", slog.Any("category_id", uuid))
		responseData, _ := json.Marshal(response.Response{
			Message:    "Category fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseData)
	}
}
