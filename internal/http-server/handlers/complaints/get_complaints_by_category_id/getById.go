package get_complaints_by_category_id

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
)

// New GetComplaintsByCategoryId godoc
// @Summary Get complaints by category ID
// @Description Retrieve all complaints that belong to a specific category based on its unique identifier (Category ID).
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Category ID (unique identifier of the category)"
// @Success 200 {array} domain.Complaint "List of complaints associated with the given category"
// @Failure 400 {object} response.Response "Invalid category ID format"
// @Failure 404 {object} response.Response "No complaints found for the given category"
// @Failure 500 {object} response.Response "Internal server error while fetching complaints"
// @Router /complaints/category/{id} [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.getByCategoryId.New"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
		)

		categoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("incorrect category id format", sl.Err(err))
			responseData, _ := json.Marshal(response.Response{
				Message:    "invalid category id format",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseData)
			return
		}

		result, err := service.GetComplaintsByCategoryId(r.Context(), categoryId)
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

		log.Info("Complaints found for category", slog.Int("category_id", categoryId))
		responseData, _ := json.Marshal(response.Response{
			Message:    "Complaints fetched successfully",
			StatusCode: http.StatusOK,
			Data:       result,
		})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseData)
	}
}
