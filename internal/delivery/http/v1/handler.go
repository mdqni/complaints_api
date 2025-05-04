package httpserver

import (
	"complaint_server/internal/config"
	categoriesCreate "complaint_server/internal/http-server/handlers/category/create"
	deleteCategoryById "complaint_server/internal/http-server/handlers/category/delete"
	categoriesGetAll "complaint_server/internal/http-server/handlers/category/get_all"
	categoriesGetById "complaint_server/internal/http-server/handlers/category/get_by_id"
	updateCategory "complaint_server/internal/http-server/handlers/category/update"
	"complaint_server/internal/http-server/handlers/complaints/canSubmit"
	"complaint_server/internal/http-server/handlers/complaints/create"
	deleteComplaint "complaint_server/internal/http-server/handlers/complaints/deleteByAdmin"
	deleteComplaintByOwner "complaint_server/internal/http-server/handlers/complaints/deleteByOwner"
	"complaint_server/internal/http-server/handlers/complaints/getComplaintByComplaintId"
	"complaint_server/internal/http-server/handlers/complaints/getComplaintsByCategoryId"
	"complaint_server/internal/http-server/handlers/complaints/getComplaintsByToken"
	getAllComplaint "complaint_server/internal/http-server/handlers/complaints/get_all"
	"complaint_server/internal/http-server/handlers/complaints/update"
	"complaint_server/internal/http-server/middleware/admin_only"
	"complaint_server/internal/http-server/middleware/cache"
	mwLogger "complaint_server/internal/http-server/middleware/logger"
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

const CacheDuration = time.Minute

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
		middleware.Recoverer, httprate.Limit(50, 1*time.Minute))
	router.Use(mwLogger.New(log))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	router.Route("/complaints", func(r chi.Router) {
		r.With(cache.CacheMiddleware(client, CacheDuration, log)).
			Get("/", getAllComplaint.New(log, complaintsService))

		r.Post("/", create.New(log, complaintsService))
		r.Get("/{id}", getComplaintByComplaintId.New(log, complaintsService))
		r.Get("/can-submit", canSubmit.New(log, complaintsService))
		r.Get("/by-token", getComplaintsByToken.New(cfg, log, complaintsService))
		r.Delete("/{id}", deleteComplaintByOwner.New(ctx, log, complaintsService, cfg, client))
	})

	router.Route("/categories", func(r chi.Router) {
		r.Use(cache.CacheMiddleware(client, CacheDuration, log))
		r.Get("/", categoriesGetAll.New(log, categoriesService))
		r.Get("/{id}", categoriesGetById.New(log, categoriesService))
		r.Get("/{id}/complaints", getComplaintsByCategoryId.New(log, complaintsService))
	})

	router.Route("/admin", func(r chi.Router) {
		r.Use(admin_only.AdminOnlyMiddleware(log, cfg, adminService))

		r.Put("/complaints/{id}", update.New(ctx, log, complaintsService, client))
		r.Delete("/complaints/{id}", deleteComplaint.New(ctx, log, complaintsService, client))

		r.Post("/categories", categoriesCreate.New(ctx, log, categoriesService, client))
		r.Put("/categories/{id}", updateCategory.New(ctx, log, categoriesService, client))
		r.Delete("/categories/{id}", deleteCategoryById.New(ctx, log, categoriesService, client))
	})
}
