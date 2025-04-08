package deleteCategoryById

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	categoryService "complaint_server/internal/service/category"
	"complaint_server/internal/storage"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
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
// @Success      200  {object}  response.Response "Категория успешно удалена"
// @Failure      400  {object}  response.Response "Неверный запрос (не указан ID) или некорректный ID"
// @Failure      404  {object}  response.Response "Категория не найдена"
// @Failure      500  {object}  response.Response "Ошибка сервера"
// @Router       /categories/{id} [delete]
func New(ctx context.Context, log *slog.Logger, service *categoryService.CategoryService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id can not be empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "id can not be empty"})
			return
		}
		atoi, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error()})
			return
		}
		err = service.DeleteCategoryById(r.Context(), atoi)

		if errors.Is(err, storage.ErrCategoryNotFound) {
			log.Error("category not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, response.Response{StatusCode: http.StatusNotFound, Message: err.Error()})
			return
		}
		if errors.Is(err, storage.ErrHasRelatedRows) {
			log.Error("there are related rows", sl.Err(err))
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, response.Response{StatusCode: http.StatusConflict, Message: err.Error()})
			return
		}
		if err != nil {
			log.Info("failed to delete category", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error(), Data: nil})
			return
		}
		log.Info("category deleted")
		client.Del(ctx, "cache:/categories")
		render.JSON(w, r, response.Response{StatusCode: http.StatusOK})
	}
}
