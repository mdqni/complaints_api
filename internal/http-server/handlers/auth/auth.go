package auth

import (
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/lib/jwt"
	"complaint_server/internal/lib/logger/sl"
	service "complaint_server/internal/service/admin"
	"encoding/json"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// RequestBody структура для парсинга тела запроса
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
func New(log *slog.Logger, adminService *service.AdminService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.New"
		log := log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()))

		var creds RequestBody
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Invalid request body",
				StatusCode: http.StatusBadRequest,
				Data:       nil,
			})
			return
		}

		admin, err := adminService.Login(r.Context(), creds.Username, creds.Password)
		if err != nil {
			log.Error("invalid credentials", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Invalid credentials",
				StatusCode: http.StatusUnauthorized,
				Data:       nil,
			})
			return
		}

		// Генерация JWT
		token, err := jwt.GenerateJWT(admin)
		if err != nil {
			log.Error("failed to generate JWT", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Failed to generate token",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
