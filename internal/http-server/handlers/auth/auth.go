package auth

import (
	"complaint_server/internal/lib/jwt"
	service "complaint_server/internal/service/admin"
	"encoding/json"
	"log/slog"
	"net/http"
)

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
		if err != nil {
			log.Error(err.Error())
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Отправляем токен
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
