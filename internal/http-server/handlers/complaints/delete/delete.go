package deleteComplaint

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

// New DeleteComplaint godoc
// @Summary Delete a complaint
// @Description Delete a complaint by its ID
// @Tags Complaints
// @Param id path int true "Complaint ID"
// @Success 200 {object} response.Response "Complaint successfully deleted"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaint/{id} [delete]
func New(log *slog.Logger, service *complaintService.ComplaintService) http.HandlerFunc {
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
			return
		}
		err = service.DeleteComplaint(ctx, atoi)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaint not found", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("complaint not found", http.StatusBadRequest))
			return
		}
		if err != nil {
			log.Info("failed to delete complaint", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error", http.StatusBadRequest))
			return
		}
		log.Info("complaint deleted")
		render.JSON(w, r, response.OK())

	}
}
