package categoriesGetAll

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/categoryService"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// New @Summary      Получить все категории
// @Description  Возвращает список всех категорий жалоб.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Category
// @Failure      500  {object}  response.Response  "Внутренняя ошибка сервера"
// @Router       /categories [get]
func New(log *slog.Logger, service *categoryService.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.get_all.New"
		ctx := r.Context()
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		result, err := service.GetCategories(ctx)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("Categories found")
		render.JSON(w, r, result)
	}
}
