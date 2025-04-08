package update

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strconv"
)

type Request struct {
	Complaint domain.Complaint `json:"data"`
}

// New @Summary Update a complaint
// @Description Updates an existing complaint based on the provided complaint ID and new data.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body Request true "Complaint resolution details"
// @Success 200 {object} Request "Complaint updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [put]
func New(log *slog.Logger, service *service.ComplaintService, client *redis.Client) http.HandlerFunc {
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
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "invalid complaint ID"})
			return
		}

		// Decode the request body into a Request struct
		var req Request
		err = render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Response{Message: "failed to decode request", StatusCode: http.StatusBadRequest})
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// Update the complaint
		log.Info("Complaints:", req.Complaint)
		complaint, err := service.UpdateComplaint(ctx, id, req.Complaint)
		if err != nil {
			log.Error("failed to update complaint", sl.Err(err))
			render.JSON(w, r, response.Response{Message: "failed to update complaint", StatusCode: http.StatusInternalServerError})
			return
		}

		log.Info("complaint updated", slog.Int64("id", id))
		err = client.Del(ctx, "cache:/complaints").Err()
		if err != nil {
			log.Error("failed to delete cache", sl.Err(err))
		}
		w.WriteHeader(http.StatusOK)
		// Send success response
		render.JSON(w, r,
			response.Response{Message: complaintID, StatusCode: http.StatusOK, Data: complaint},
		)
	}
}
