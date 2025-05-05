package app

import (
	"complaint_server/internal/config"
	httpserver "complaint_server/internal/delivery/http/v1"
	serviceAdmin "complaint_server/internal/service/admin"
	serviceCategory "complaint_server/internal/service/category"
	serviceComplaint "complaint_server/internal/service/complaint"
	redisClient "complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"context"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	server *http.Server
}

func NewApp(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	router := chi.NewRouter()

	db, err := pg.New(cfg.ConnString)
	if err != nil {
		return nil, err
	}
	client, err := redisClient.NewClient(ctx, cfg, log)
	if err != nil {
		return nil, err
	}
	complaintsRepo := pg.NewComplaintRepo(db)
	categoryRepo := pg.NewCategoryRepo(db, client)

	complaintsService := serviceComplaint.NewComplaintsService(complaintsRepo)
	categoriesService := serviceCategory.NewCategoriesService(db)
	adminService := serviceAdmin.NewAdminService(db)

	httpserver.RegisterRoutes(ctx, cfg, router, log, client, complaintsService, categoriesService, adminService)

	readTimeout, err := time.ParseDuration(cfg.HTTPServer.Timeout)
	if err != nil {
		log.Error("invalid HTTP_TIMEOUT", slog.String("error", err.Error()))
	}

	writeTimeout, err := time.ParseDuration(cfg.HTTPServer.IdleTimeout)
	if err != nil {
		log.Error("invalid HTTP_IDLE_TIMEOUT", slog.String("error", err.Error()))
	}

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return &App{
		server: server,
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
