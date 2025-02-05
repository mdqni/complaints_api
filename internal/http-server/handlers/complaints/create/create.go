package create

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service"
	"complaint_server/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	_ "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
)

type Request struct {
	Message    string `json:"message" validate:"required"`
	CategoryID int    `json:"categoryId" validate:"required"`
	UserUUID   string `json:"user_uuid"`
}

// New CreateComplaint godoc
// @Summary Create a new complaint
// @Description Create a new complaint for a specific user and category
// @Tags Complaints
// @Accept json
// @Produce json
// @Param Request body Request true "Complaint details"
// @Success 200 {object} map[string]interface{} "Response with complaint ID"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 429 {object} response.Response "Limit of one complaint per hour exceeded"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaint [post]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.create.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))       //Пишем в лог
			render.JSON(w, r, response.Error("failed to decode request")) //Возвращаем ошибку
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			errors.As(err, &validationErrors)
			log.Error("failed to validate request", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validationErrors))
			return
		}
		message := req.Message
		categoryID := req.CategoryID
		userUUID := req.UserUUID
		answer, err := service.CreateComplaint(userUUID, categoryID, message)
		if errors.Is(err, storage.ErrLimitOneComplaintInOneHour) {
			log.Error("failed to create complaints", sl.Err(err))
			render.JSON(w, r, response.Error("You can only submit one complaint per hour. Please try again later."))
			return
		}
		if err != nil {
			log.Error("failed to create complaints", sl.Err(err))
			render.JSON(w, r, response.Error("failed to save complaints"))
		}
		render.JSON(w, r, map[string]interface{}{
			"status": http.StatusOK,
			"data": map[string]interface{}{
				"answer": answer,
			},
		})

	}
}
