package get_all

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	_ "database/sql"
	"errors"
	"github.com/go-chi/render"
	_ "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
)

// New @Summary Get all complaints
// @Summary Get all complaints
// @Description Retrieve all complaints from the database
// @Tags Complaints
// @Produce json
// @Success 200 {array} domain.Complaint "List of complaints"
// @Failure 404 {object} response.Response "No complaints found"
// @Failure 500 {object} response.Response "Internal error"
// @Router /complaints [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.get_all.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))
		result, err := service.GetAllComplaints()
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
		render.JSON(w, r, result)
	}
}
