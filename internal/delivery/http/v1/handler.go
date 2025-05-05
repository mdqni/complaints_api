package httpserver

import (
	"complaint_server/internal/config"
	mwLogger "complaint_server/internal/delivery/http/middleware/logger"
	"complaint_server/internal/delivery/http/v1/categories"
	"complaint_server/internal/delivery/http/v1/complaints"

	serviceAdmin "complaint_server/internal/service/admin"
	serviceCategory "complaint_server/internal/service/category"
	serviceComplaint "complaint_server/internal/service/complaint"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

const requestLimitTimeout = 1 * time.Minute

func RegisterRoutes(
	ctx context.Context,
	cfg *config.Config,
	router chi.Router,
	log *slog.Logger,
	client *redis.Client,
	complaintsService *serviceComplaint.ComplaintService,
	categoriesService *serviceCategory.CategoryService,
	adminService *serviceAdmin.AdminService,
) {

	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger,
		middleware.Recoverer, httprate.Limit(50, requestLimitTimeout))
	router.Use(mwLogger.New(log))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	router.Route("/categories", func(r chi.Router) {
		categoryHandler := categories.NewHandler(ctx, complaintsService, adminService, categoriesService, log, client, cfg)
		categories.RegisterRoutes(r, categoryHandler)
	})
	router.Route("/complaints", func(r chi.Router) {
		complaintHandler := complaints.NewHandler(ctx, complaintsService, adminService, categoriesService, log, client, cfg)
		complaints.RegisterRoutes(r, complaintHandler)
	})
}
