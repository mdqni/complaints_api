package get_complaints_by_category_id

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service/complaint"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

// New GetComplaintsByCategoryId godoc
// @Summary Get complaints by category ID
// @Description Retrieve all complaints that belong to a specific category
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {array} domain.Complaint "List of complaints"
// @Failure 400 {object} response.Response "Invalid category ID"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/category/{id} [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.complaints.getByCategoryId.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		categoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid category id", http.StatusBadRequest))
		}
		result, err := service.GetComplaintsByCategoryId(categoryId)

		if err != nil {
			log.Error(op, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
			return
		}
		log.Info("Categories found")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, result)
	}
}
