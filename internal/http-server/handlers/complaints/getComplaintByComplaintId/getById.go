package getComplaintByComplaintId

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

// New GetComplaintById godoc
// @Summary Get a complaint by ID
// @Description Retrieve a complaint using its unique identifier. The UUID must be a string that corresponds to a valid complaint in the database.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID (unique identifier of the complaint)"
// @Success 200 {object} domain.Complaint "Complaint details"
// @Failure 400 {object} response.Response "Invalid request, incorrect ID format"
// @Failure 404 {object} response.Response "Complaint with the given ID not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.get_by_complaint_id.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Error("Missing complaint_id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
			return
		}
		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Error("Invalid complaint_id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: err.Error()})
		}
		result, err := service.GetComplaintByUUID(r.Context(), uuid)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaint not found", sl.Err(err))
			render.JSON(w, r, response.Error("complaint with this id not found", http.StatusNotFound))
			return
		}
		if errors.Is(err, storage.ErrScanFailure) {
			log.Error("complaint scan failure", sl.Err(err))
			render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error()})
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusOK,
			Data:       result,
		})
		_, _ = w.Write(responseData)
	}
}
