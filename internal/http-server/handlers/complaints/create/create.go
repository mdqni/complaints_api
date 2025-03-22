package create

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaintService"
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
	CategoryID int    `json:"category_id" validate:"required"`
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
// @Router /complaint [post]
func New(log *slog.Logger, service *complaintService.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.create.New"

		ctx := r.Context()
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		defer r.Body.Close()

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Response{Message: "Invalid request body"})
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Валидация запроса
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			log.Error("failed to validate request", sl.Err(err))

			// Преобразуем ошибки валидации в JSON-ответ
			validationErrors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors[err.Field()] = err.Tag()
			}

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Response{
				Message: "Validation failed",
				Errors:  validationErrors,
			})
			return
		}

		// Обрабатываем жалобу
		answer, err := service.CreateComplaint(ctx, req.UserUUID, req.CategoryID, req.Message)
		if errors.Is(err, storage.ErrLimitOneComplaintInOneHour) {
			log.Warn("complaint limit exceeded", sl.Err(err))
			render.Status(r, http.StatusTooManyRequests)
			render.JSON(w, r, response.Response{Message: "You can only submit one complaint per hour. Please try again later."})
			return
		}
		if err != nil {
			log.Error("failed to create complaint", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Response{Message: "Failed to save complaint"})
			return
		}

		// Успешный ответ
		render.Status(r, http.StatusOK)
		render.JSON(w, r, ComplaintResponse{
			Status: http.StatusOK,
			Data: struct {
				Answer string `json:"answer"`
			}{Answer: answer},
		})
	}
}
