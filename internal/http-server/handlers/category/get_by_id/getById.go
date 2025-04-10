package categories_get_by_id

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	service "complaint_server/internal/service/category"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

// New @Summary Получить категорию по ID
// @Description Возвращает категорию по уникальному идентификатору (ID).
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID (unique identifier of the category)"
// @Success 200 {object} domain.Category "Category details"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 500 {object} response.Response "Internal server error while fetching the category"
// @Router /categories/{id} [get]
func New(log *slog.Logger, service *service.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.get_by_id.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		ctx := r.Context()

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("incorrect id in params", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "incorrect id in params",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}

		result, err := service.GetCategoryById(ctx, id)
		if err != nil {
			log.Error(op, sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "internal error",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		log.Info("Category found", slog.Int("category_id", id))
		render.JSON(w, r, response.Response{
			Message:    "Category fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		})
	}
}
