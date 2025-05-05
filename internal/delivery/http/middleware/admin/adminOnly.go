package admin

import (
	"complaint_server/internal/config"
	authService "complaint_server/internal/service/admin"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"log/slog"
)

func AdminOnlyMiddleware(logger *slog.Logger, cfg *config.Config, service *authService.AdminService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("No Authorization header found")
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Error("Invalid token format")
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				logger.Info("Token found", slog.String("token", tokenString))
				return []byte(cfg.JwtSecret), nil
			})

			if err != nil || !token.Valid {
				logger.Error("Invalid token", slog.String("token", tokenString))
				fmt.Println(err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Error("Invalid token claims")
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			bcode, err := strconv.Atoi(strconv.Itoa(int(claims["barcode"].(float64))))
			if err != nil {
				logger.Error("Invalid token barcode", slog.String("barcode", claims["barcode"].(string)))
				http.Error(w, "Invalid token barcode", http.StatusUnauthorized)
				return
			}
			isAdmin, err := service.IsAdmin(r.Context(), bcode)
			if err != nil {
				logger.Error("Invalid token barcode", slog.String("barcode", claims["barcode"].(string)))
				http.Error(w, "Invalid token barcode", http.StatusUnauthorized)
				return
			}
			if !isAdmin {
				logger.Error("User is not admin", slog.String("barcode", claims["barcode"].(string)))
				http.Error(w, "Access denied", http.StatusMethodNotAllowed)
				return
			}
			ctx := context.WithValue(r.Context(), "barcode", claims["barcode"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
