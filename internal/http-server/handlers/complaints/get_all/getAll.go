package get_all_complaint

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
)

// New GetComplaintById godoc
// @Summary Get a complaint by ID
// @Description Retrieve a complaint using its unique identifier. The ID must be an integer that corresponds to a valid complaint in the database.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID (unique identifier of the complaint)"
// @Success 200 {object} domain.Complaint "Complaint details"
// @Failure 400 {object} response.Response "Invalid request, incorrect ID format"
// @Failure 404 {object} response.Response "Complaint with the given ID not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaint.get_by_complaint_id.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
		)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("incorrect id on params", sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				StatusCode: http.StatusBadRequest,
				Data:       nil,
				Message:    "incorrect id on params",
			})
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseData)
			return
		}

		result, err := service.GetComplaintById(r.Context(), id)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaint not found", sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				StatusCode: http.StatusNotFound,
				Data:       nil,
				Message:    fmt.Sprintf("complaint with this id %d not found", id),
			})
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(responseData)
			return
		}
		if err != nil {
			log.Error(op, sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
				Message:    "internal error",
			})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(responseData)
			return
		}

		log.Info(fmt.Sprintf("complaint found with id: %d", id))
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusOK,
			Data:       result,
		})
		_, _ = w.Write(responseData)
	}
}
