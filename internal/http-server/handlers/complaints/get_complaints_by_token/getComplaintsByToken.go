package get_complaints_by_token

import (
	"complaint_server/internal/lib/fetchStudentProfile"
	service "complaint_server/internal/service/complaint"
	"encoding/json"
	"log/slog"
	"net/http"
)

// New creates a handler that returns the complaints of a user by their token.
// @Summary Get user complaints by token
// @Description Retrieves all complaints associated with a user based on the provided token.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param token query string true "User token"
// @Success 200 {array} response.Response "List of complaints"
// @Failure 400 {object}  response.Response "Token required"
// @Failure 401 {object} response.Response "Invalid token or failed to fetch profile"
// @Failure 500 {object} response.Response "Failed to serialize complaints"
// @Router /complaints/by-token [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			log.Error("token required", http.StatusBadRequest)
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}

		profile, err := fetchStudentProfile.FetchStudentProfile(token)
		log.Info("Profile: ", profile)
		if err != nil {
			log.Error("invalid token or failed to fetch profile", http.StatusUnauthorized)
			http.Error(w, "invalid token or failed to fetch profile", http.StatusUnauthorized)
			return
		}

		complaints, err := service.GetComplaintsByBarcode(r.Context(), profile.Barcode)
		if err != nil {
			http.Error(w, "failed to get complaints", http.StatusInternalServerError)
			return
		}

		responseData, err := json.Marshal(complaints)
		if err != nil {
			http.Error(w, "failed to serialize complaints", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}
}
