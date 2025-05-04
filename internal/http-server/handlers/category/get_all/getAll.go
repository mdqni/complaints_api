package categoriesGetAll

import (
	service "complaint_server/internal/service/category"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"encoding/json"
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8") // Убедись, что установил utf-8
		const op = "handlers.category.get_all.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		ctx := r.Context()

		// Получаем категории
		result, err := service.GetCategories(ctx)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.Header().Set("Content-Type", "application/json; charset=utf-8") // Устанавливаем правильный контент
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Response{
				Message:    "internal error",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}
		log.Info("Categories found")

		// Ручная сериализация JSON с ensure_ascii=false
		responseData := response.Response{
			Message:    "Categories fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		}

		encodedResponse, err := json.MarshalIndent(responseData, "", "  ")
		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Response{
				Message:    "Failed to serialize response",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(encodedResponse)
		if err != nil {
			log.Error(op, sl.Err(err))
		}
		log.Info("result: ", result)
	}
}
