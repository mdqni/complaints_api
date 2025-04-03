package update

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaintService"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	complaint domain.Complaint
}

// New UpdateComplaint godoc
// @Summary Update a complaint
// @Description Update a complaint
// @Tags Complaints
// @Accept json
// @Produce json
// @Param request body Request true "Complaint resolution details"
// @Success 200 {object} response.Response "Complaint updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [put]
func New(log *slog.Logger, service *complaintService.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.update.New"

		ctx := r.Context()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(ctx)),
		)

		// Get complaint ID from URL parameters
		complaintID := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(complaintID, 10, 64)
		if err != nil {
			log.Error("invalid complaint ID", sl.Err(err))
			render.JSON(w, r, response.Error("invalid complaint ID", http.StatusBadRequest))
			return
		}

		// Decode the request body into a Request struct
		var req Request
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request", http.StatusBadRequest))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// Update the complaint
		err = service.UpdateComplaint(ctx, id, req.complaint)
		if err != nil {
			log.Error("failed to update complaint status", sl.Err(err))
			render.JSON(w, r, response.Error("failed to update complaint status", http.StatusInternalServerError))
			return
		}

		log.Info("complaint updated", slog.Int64("id", id))

		// Send success response
		render.JSON(w, r, response.Response{
			Status: http.StatusOK,
		})
		return
	}
}
