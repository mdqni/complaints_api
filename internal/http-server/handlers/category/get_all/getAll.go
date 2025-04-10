package categoriesGetAll

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	service "complaint_server/internal/service/category"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// New @Summary Получить все категории
// @Description Возвращает список всех категорий жалоб, доступных в системе.
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {array} domain.Category "List of all categories"
// @Failure 500 {object} response.Response "Internal server error while fetching categories"
// @Router /categories [get]
func New(log *slog.Logger, service *service.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.get_all.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		ctx := r.Context()

		result, err := service.GetCategories(ctx)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Response{
				Message:    "internal error",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		log.Info("Categories found")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		render.JSON(w, r, response.Response{
			Message:    "Categories fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		})
		log.Info("result: ", result[0])
	}
}
