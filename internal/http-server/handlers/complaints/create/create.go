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
)

type Request struct {
	Message    string `json:"message" validate:"required"`
	CategoryID int    `json:"category_id" validate:"required"`
	Barcode    string `json:"barcode"`
}

//type ComplaintResponse struct {
//	Status int `json:"status"`
//	Data   struct {
//		Answer string `json:"answer"`
//	} `json:"data"`
//}

// New CreateComplaint godoc
// @Summary Create a new complaint
// @Description Create a new complaint for a specific user and category. Only one complaint can be submitted per hour for the same user.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param Request body Request true "Complaint details" // Request body with message, category_id, and barcode
// @Success 200 {object} response.Response "Success response with complaint ID and answer"
// @Failure 400 {object} response.Response "Invalid request, bad input or validation error"
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
		categoryID := req.CategoryID //Фронт должен преобразовать
		barcode := req.Barcode
		complaintID, answer, err := service.CreateComplaint(r.Context(), barcode, categoryID, message)
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
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.Response{
			Message:    answer,
			StatusCode: http.StatusOK,
			Data:       complaintID,
		})

	}
}
