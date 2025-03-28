package deleteCategoryById

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/categoryService"
	"complaint_server/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

// New @Summary      Удалить категорию по ID
// @Description  Удаляет категорию жалоб по переданному идентификатору.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID категории"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response  "Неверный запрос (не указан ID)"
// @Failure      404  {object}  response.Response  "Категория не найдена"
// @Failure      500  {object}  response.Response  "Внутренняя ошибка сервера"
// @Router       /categories/{id} [delete]
func New(log *slog.Logger, service *categoryService.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.delete.New"
		ctx := r.Context()
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id can not be empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request", http.StatusBadRequest))
			return
		}
		atoi, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("Internal Server Message", http.StatusInternalServerError))
			return
		}
		err = service.DeleteCategoryById(ctx, atoi)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("category not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, response.Error("category not found", http.StatusNotFound))
			return
		}
		if err != nil {
			log.Info("failed to delete category", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("category deleted")
		render.JSON(w, r, response.OK())
	}
}
