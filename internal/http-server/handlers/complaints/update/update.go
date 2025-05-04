package update

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

type Request struct {
	Complaint domain.Complaint `json:"data"`
}

// New @Summary Update a complaint
// @Description Updates an existing complaint based on the provided complaint ID and new data.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID"
// @Param request body Request true "Complaint resolution details"
// @Success 200 {object} Request "Complaint updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/complaints/{id} [put]
func New(context context.Context, log *slog.Logger, service *serviceComplaint.ComplaintService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.update.New"

		ctx := r.Context()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(ctx)),
		)

		// Get complaint ID from URL parameters
		complaintID := chi.URLParam(r, "id")
		if complaintID == "" {
			log.Error("Missing complaint_id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
			return
		}
		id, err := uuid.Parse(complaintID)
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

		log.Info("complaint updated", slog.Any("id", id))
		err = client.Del(ctx, "cache:/complaints").Err()
		if err != nil {
			log.Error("failed to deleteByAdmin cache", sl.Err(err))
		}
		client.Del(context, fmt.Sprintf("cache:/complaints/%d", req.Complaint.ID))
		// Send success response
		render.JSON(w, r,
			response.Response{Message: complaintID, StatusCode: http.StatusOK, Data: complaint},
		)
	}
}
