package can_submit

import (
	"complaint_server/internal/lib/api/response"
	service "complaint_server/internal/service/complaint"
	"encoding/json"
	"log/slog"
	"net/http"
)

// New godoc
// @Summary      Проверка возможности отправки жалобы
// @Description  Проверяет, может ли пользователь отправить новую жалобу (прошел ли час с последнего запроса)
// @Tags         Complaints
// @Accept       json
// @Produce      json
// @Param        barcode  query     string  true  "Barcode пользователя"
// @Success      200    {object}  response.Response  "true/false"
// @Failure      400    {object}  response.Response  "missing barcode"
// @Failure      500    {object}  response.Response  "internal error"
// @Router       /complaints/can-submit [get]
func New(log *slog.Logger, service *service.ComplaintService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("barcode")
		if userID == "" {
			log.Error("missing barcode", http.StatusBadRequest)
			http.Error(w, "missing barcode", http.StatusBadRequest)
			return
		}

		canSubmit, _ := service.CanSubmitByBarcode(r.Context(), userID)
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusOK,
			Data:       map[string]interface{}{"canSubmit": canSubmit},
		})
		_, _ = w.Write(responseData)
		return
	}
}
