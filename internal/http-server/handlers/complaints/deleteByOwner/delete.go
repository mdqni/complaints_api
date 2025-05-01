package deleteComplaintByOwner

import (
	"complaint_server/internal/config"
	"complaint_server/internal/lib/api/response"
	"complaint_server/internal/service/complaint"
	"complaint_server/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strings"
)

// New @Summary Delete a complaint
// @Description Delete a complaint by its ID by owner. If the complaint is not found or user is not owner, an error is returned.
// @Tags Complaints
// @Param id path string true "Complaint ID"
// @Success 200 {object} response.Response "Complaint successfully deleted"
// @Failure 400 {object} response.Response "Invalid request or complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [delete]
func New(context context.Context, log *slog.Logger, service *service.ComplaintService, cfg *config.Config, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("No Authorization header found")
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Error("Invalid token format")
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			log.Info("Token found", slog.String("token", tokenString))
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Error("Invalid token", slog.String("token", tokenString))
			fmt.Println(err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Error("Invalid token claims")
			render.JSON(w, r, response.Response{
				Message:    "Invalid token claims",
				StatusCode: http.StatusInternalServerError,
				Data:       nil,
			})
			return
		}

		barcodeFloat, ok := claims["barcode"].(float64)
		if !ok {
			log.Error("Invalid token barcode type")
			render.JSON(w, r, response.Response{
				Message:    "Invalid token barcode",
				StatusCode: http.StatusForbidden,
				Data:       nil,
			})
			return
		}
		barcode := int(barcodeFloat)

		id := chi.URLParam(r, "id")
		complaintID, err := uuid.Parse(id)
		if err != nil {
			log.Error("Invalid complaint id", slog.String("complaintID", id))
			render.JSON(w, r, response.Response{
				Message:    "Invalid complaint id",
				StatusCode: http.StatusForbidden,
				Data:       nil,
			})
		}
		can, err := service.CanUserDeleteComplaintById(ctx, complaintID, barcode)
		if errors.Is(err, storage.ErrComplaintNotFound) {
			log.Error("complaints not found", slog.String("barcode", claims["barcode"].(string)))
			render.JSON(w, r, response.Response{
				Message:    "complaints not found",
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			return
		}
		if err != nil {
			log.Error("Error on finding owner", "err", err)
			render.JSON(w, r, response.Response{
				Message:    "Error on finding owner",
				StatusCode: http.StatusNotFound,
				Data:       nil,
			})
			return
		}
		if !can {
			log.Error("err", response.Response{
				Message:    "User is not owner of complaint",
				StatusCode: http.StatusForbidden,
				Data:       nil,
			})
			render.JSON(w, r, response.Response{
				Message:    "User is not owner of complaint",
				StatusCode: http.StatusForbidden,
				Data:       nil,
			})
			return
		}
		err = service.DeleteComplaintById(ctx, complaintID)
		if err != nil {
			log.Error("Error on deleting complaint", "err", err)
			render.JSON(w, r, response.Response{Message: "Error on deleting complaint", Data: nil, StatusCode: http.StatusInternalServerError})
			return
		}
		log.Info("complaint successfully deleted")

		client.Del(context, fmt.Sprintf("cache:/complaints/%s", id))
		client.Del(context, fmt.Sprintf("cache:/complaints"))
		render.JSON(w, r, response.Response{
			Message:    "Complaint successfully deleted",
			StatusCode: http.StatusOK,
			Data:       nil,
		})
		return
	}
}
