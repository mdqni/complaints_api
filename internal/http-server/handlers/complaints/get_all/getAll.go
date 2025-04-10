package get_all

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

// New @Summary Get all complaints
// @Description Retrieve all complaints from the database. Caches the result for 5 minutes.
// @Tags Complaints
// @Produce json
// @Success 200 {array} response.Response "List of all complaints"
// @Failure 404 {object} response.Response "No complaints found in the database"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints [get]
func New(ctx context.Context, log *slog.Logger, service *service.ComplaintService, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.get_all.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		result, err := service.GetAllComplaints(r.Context())
		log.Info("url", r.URL.String())

		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusOK)
			responseData, _ := json.Marshal(response.Response{
				StatusCode: http.StatusOK,
				Data:       nil,
				Message:    storage.ErrComplaintNotFound.Error(),
			})
			_, _ = w.Write(responseData)
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			_, _ = w.Write(responseData)
			return
		}

		log.Info("complaints found")
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusOK,
			Data:       result,
		})
		_, _ = w.Write(responseData)
	}
}
