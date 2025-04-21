package deleteComplaint

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

// New @Summary Delete a complaint
// @Description Delete a complaint by its ID. If the complaint is not found, an error is returned.
// @Tags Complaints
// @Param id path string true "Complaint ID"
// @Success 200 {object} response.Response "Complaint successfully deleted"
// @Failure 400 {object} response.Response "Invalid request or complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/complaints/{id} [delete]
func New(context context.Context, log *slog.Logger, service *service.ComplaintService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.delete.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id can not be empty")
			render.JSON(w, r, response.Response{
				Message:    "Complaint ID is required",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}

		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Error("invalid complaint ID format", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Invalid complaint ID format",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}

		err = service.DeleteComplaintById(r.Context(), uuid)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaint not found", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Complaint not found",
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			return
		}

		if err != nil {
			log.Error("failed to delete complaint", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Internal error while deleting complaint",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		client.Del(r.Context(), "cache:/complaints")
		log.Info("complaint successfully deleted")
		client.Del(context, fmt.Sprintf("cache:/complaints"))
		// Успешный ответ
		render.JSON(w, r, response.Response{
			Message:    "Complaint successfully deleted",
			StatusCode: http.StatusOK,
			Data:       nil,
		})
	}
}
