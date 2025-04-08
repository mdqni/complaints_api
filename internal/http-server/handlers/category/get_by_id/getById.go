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
		const op = "handlers.category.get_all.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		ctx := r.Context()
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("incorrect id on params", sl.Err(err))
			render.JSON(w, r, response.Error("incorrect id on params", http.StatusBadRequest))
			return
		}
		result, err := service.GetCategoryById(ctx, id)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("Categories found")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.Response{StatusCode: http.StatusOK, Data: result})
	}
}
