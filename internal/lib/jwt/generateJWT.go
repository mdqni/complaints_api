package jwt

import (
	"complaint_server/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GenerateJWT(admin *domain.Admin) (string, error) {
	claims := jwt.MapClaims{
		"id":       admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Токен на 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("a-string-secret-at-least-256-bits-long"))
}
