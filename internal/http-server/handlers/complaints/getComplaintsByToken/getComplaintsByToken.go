package getComplaintsByToken

import (
	"complaint_server/internal/config"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/jwt"
	service "complaint_server/internal/service/complaint"
	"encoding/json"
	"github.com/go-chi/render"
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
// @Success 200 {object} response.Response "List of complaints"
// @Failure 400 {object}  response.Response "Token required"
// @Failure 401 {object} response.Response "Invalid token or failed to fetch profile"
// @Failure 500 {object} response.Response "Failed to serialize complaints"
// @Router /complaints/by-token [get]
func New(cfg *config.Config, log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			log.Error("token required", http.StatusBadRequest)
			render.JSON(w, r, response.Response{
				Message:    "token required",
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		log.Info("token", token)
		profile, err := jwt.EncodeJWT(cfg.JwtSecret, token)
		log.Info("Profile: ", profile)
		if err != nil {
			log.Error("invalid token or failed to fetch profile", "Err", response.Response{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			render.JSON(w, r, response.Response{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		complaints, err := service.GetComplaintsByBarcode(r.Context(), profile.Barcode)
		log.Error("scan error", "err", err)
		if err != nil {
			log.Error("failed to get complaints", "err", err)
			render.JSON(w, r, response.Response{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		responseData, err := json.Marshal(response.Response{Data: complaints, StatusCode: http.StatusOK})
		if err != nil {
			render.JSON(w, r, response.Response{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}
}
