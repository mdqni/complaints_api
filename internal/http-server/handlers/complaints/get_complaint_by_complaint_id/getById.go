package get_complaint_by_complaint_id

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaintService"
	"complaint_server/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

// New GetComplaintById godoc
// @Summary Get a complaint by ID
// @Description Retrieve a complaint using its unique identifier
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Success 200 {object} domain.Complaint "Complaint details"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /{id} [get]
func New(log *slog.Logger, service *complaintService.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.get_by_complaint_id.New"
		ctx := r.Context()
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		result, err := service.GetComplaintById(ctx, id)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaint not found", sl.Err(err))
			render.JSON(w, r, response.Error("complaint with this id not found", http.StatusBadRequest))
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("complaints found")
		w.WriteHeader(200)
		render.JSON(w, r, result)
	}
}
