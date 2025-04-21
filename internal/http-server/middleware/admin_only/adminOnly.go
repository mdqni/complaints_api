package admin_only

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"log/slog"
)

func AdminOnlyMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
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
				return []byte("a-string-secret-at-least-256-bits-long"), nil
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

			role, ok := claims["role"].(string)
			if !ok || role != "admin" {
				logger.Warn("Forbidden access", slog.String("role", role))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			logger.Info("Admin passed")
			ctx := context.WithValue(r.Context(), "barcode", claims["barcode"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
