package auth

import (
	"complaint_server/internal/lib/jwt"
	service "complaint_server/internal/service/admin"
	"encoding/json"
	"log/slog"
	"net/http"
)

type RequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// New @Summary Authenticate an admin user
// @Description Authenticates an admin user with their username and password, and generates a JWT token if the credentials are correct.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body RequestBody true "Admin credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Failure 500 {object} response.Response "Failed to generate token"
// @Router /login [post]
func New(log *slog.Logger, admin *service.AdminService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			log.Error(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		admin, err := admin.Login(r.Context(), creds.Username, creds.Password)
		if err != nil {
			log.Error("Can not login", err.Error())
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Генерация JWT
		token, err := jwt.GenerateJWT(admin)
		log.Debug(token)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Отправляем токен
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
