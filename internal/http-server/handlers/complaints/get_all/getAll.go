package get_all

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"context"
	_ "database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	_ "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"time"
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

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		result, err := service.GetAllComplaints(r.Context())
		log.Info("url", r.URL.String())

		client.Set(ctx, fmt.Sprintf("cache:%s", r.URL.Path), result, 5*time.Minute)

		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusOK)
			render.JSON(w, r, nil)
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("complaints found")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response.Response{StatusCode: http.StatusOK, Data: result})
	}
}
