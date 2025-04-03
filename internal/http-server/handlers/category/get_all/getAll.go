package categoriesGetAll

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type CategoriesGetter interface {
	GetCategories(context.Context) ([]domain.Category, error)
}

// New @Summary      Получить все категории
// @Description  Возвращает список всех категорий жалоб.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Category
// @Failure      500  {object}  response.Response  "Внутренняя ошибка сервера"
// @Router       /categories [get]
func New(log *slog.Logger, getCategories CategoriesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.category.get_all.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		result, err := getCategories.GetCategories(nil)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("Categories found")
		render.JSON(w, r, result)
	}
}
