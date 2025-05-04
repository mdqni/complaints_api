package getComplaintsByCategoryId

import (
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

// New GetComplaintsByCategoryId godoc
// @Summary Get complaints by category UUID
// @Description Retrieve all complaints that belong to a specific category based on its unique identifier (Category ID).
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Category UUID (unique identifier of the category)"
// @Success 200 {array} domain.Complaint "List of complaints associated with the given category"
// @Failure 400 {object} response.Response "Invalid category ID format"
// @Failure 404 {object} response.Response "No complaints found for the given category"
// @Failure 500 {object} response.Response "Internal server error while fetching complaints"
// @Router /categories/{id}/complaints [get]
func New(log *slog.Logger, service *serviceComplaint.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.getByCategoryId.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
		)

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Error("Missing category id")
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
			return
		}

		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Error(op, sl.Err(err))
			render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: err.Error()})
			return
		}

		result, err := service.GetComplaintsByCategoryId(r.Context(), uuid)
		if err != nil {
			log.Error("failed to get complaints", sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				Message:    "no complaints found for the given category",
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(responseData)
			return
		}

		log.Info("Complaints found for category", slog.Any("category_id", uuid))
		responseData, _ := json.Marshal(response.Response{
			Message:    "Complaints fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseData)
	}
}
