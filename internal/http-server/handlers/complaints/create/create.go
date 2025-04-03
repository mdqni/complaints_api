package create

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	_ "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	Message    string `json:"message" validate:"required"`
	CategoryID string `json:"categoryId" validate:"required"`
	UserUUID   string `json:"user_uuid"`
}
type ComplaintResponse struct {
	Status int `json:"status"`
	Data   struct {
		Answer string `json:"answer"`
	} `json:"data"`
}

// New CreateComplaint godoc
// @Summary Create a new complaint
// @Description Create a new complaint for a specific user and category
// @Tags Complaints
// @Accept json
// @Produce json
// @Param Request body Request true "Complaint details"
// @Success 200 {object} ComplaintResponse "Success response"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 429 {object} response.Response "Limit of one complaint per hour exceeded"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints [post]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.register.New"
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
		message := req.Message
		categoryID, _ := strconv.Atoi(req.CategoryID) //Фронт должен преобразовать
		userUUID := req.UserUUID
		answer, err := service.CreateComplaint(userUUID, categoryID, message)
		if errors.Is(err, storage.ErrLimitOneComplaintInOneHour) {
			log.Error("failed to register complaints", sl.Err(err))
			w.WriteHeader(http.StatusTooManyRequests)
			render.JSON(w, r, response.Error("You can only submit one complaint per hour. Please try again later.", http.StatusTooManyRequests))
			return
		}
		if err != nil {
			log.Error("failed to register complaints", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to save complaints", http.StatusInternalServerError))
		}
		render.JSON(w, r, map[string]interface{}{
			"status": http.StatusOK,
			"data": map[string]interface{}{
				"answer": answer,
			},
		})

	}
}
