package resolveComplaint

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	Status string `json:"status" validate:"required"`
	Answer string `json:"answer" validate:"required"`
}

// New UpdateComplaintStatus godoc
// @Summary Resolve a complaint !!NOT ENABLE RIGHT NOW
// @Description Update the status of a complaint ("approved" or "rejected") with an answer
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body Request true "Complaint resolution details"
// @Success 200 {object} response.Response "Complaint status updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id}/status [put]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.update.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
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

		var status domain.ComplaintStatus
		switch strings.ToLower(req.Status) {
		case "approved":
			status = domain.StatusApproved
		case "rejected":
			status = domain.StatusRejected
		default:
			log.Error("invalid status", slog.String("status", req.Status))
			render.JSON(w, r, response.Error("invalid status", http.StatusBadRequest))
			return
		}
		answer := req.Answer
		// Update the complaint status
		err = service.UpdateComplaintStatus(r.Context(), id, status, answer)
		if err != nil {
			log.Error("failed to update complaint status", sl.Err(err))
			render.JSON(w, r, response.Error("failed to update complaint status", http.StatusInternalServerError))
			return
		}

		log.Info("complaint status updated", slog.Int64("id", id), slog.String("status", string(status)))

		// Send success response
		render.JSON(w, r, response.Response{
			Status: http.StatusOK,
		})
	}
}
