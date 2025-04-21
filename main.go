package main

import (
	_ "complaint_server/docs"
	"complaint_server/internal/config"
	categoriesCreate "complaint_server/internal/http-server/handlers/category/create"
	deleteCategoryById "complaint_server/internal/http-server/handlers/category/delete"
	categoriesGetAll "complaint_server/internal/http-server/handlers/category/get_all"
	"complaint_server/internal/http-server/handlers/category/get_by_id"
	updateCategory "complaint_server/internal/http-server/handlers/category/update"
	"complaint_server/internal/http-server/handlers/complaints/can_submit"
	"complaint_server/internal/http-server/handlers/complaints/create"
	deleteComplaint "complaint_server/internal/http-server/handlers/complaints/delete"
	getallcomplaint "complaint_server/internal/http-server/handlers/complaints/get_all"
	"complaint_server/internal/http-server/handlers/complaints/get_complaint_by_complaint_id"
	"complaint_server/internal/http-server/handlers/complaints/get_complaints_by_category_id"
	"complaint_server/internal/http-server/handlers/complaints/get_complaints_by_token"
	"complaint_server/internal/http-server/handlers/complaints/update"
	"complaint_server/internal/http-server/middleware/admin_only"
	"complaint_server/internal/http-server/middleware/cache"
	mwLogger "complaint_server/internal/http-server/middleware/logger"
	"complaint_server/internal/lib/logger/handlers/slogpretty"
	"complaint_server/internal/lib/logger/sl"
	authService "complaint_server/internal/service/admin"
	categoryService "complaint_server/internal/service/category"
	complaintService "complaint_server/internal/service/complaint"
	strg "complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

// @title			Complaint Server API
// @version		1.0
// @description	This is a server for managing complaints and categories.
// @contact.name	API Support
// @contact.email	quanaimadi@.gmail.com
// @license.name	MIT
// @license.url	https://opensource.org/licenses/MIT
// @host			complaints-api.yeunikey.dev
// @BasePath		/
// @schemes https
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Init Logger
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	//log.Info(password) Admin
	log.Info(
		"starting complaints server",
		slog.String("env", cfg.Env),
	)
	log.Debug("Debug message are enabled")
	rdb := setupRedis(ctx, cfg, log)
	storage := setupStorage(cfg.ConnString, log)
	router := setupRouter(ctx, log, cfg, storage, rdb)

	startServer(cfg, router, log)
}

func setupRedis(ctx context.Context, cfg *config.Config, log *slog.Logger) *redis.Client {
	client, err := strg.NewClient(ctx, cfg, log)
	if err != nil {
		panic(err)
	}
	return client
}

func setupStorage(connString string, log *slog.Logger) *pg.Storage {
	storage, err := pg.New(connString)
	log.Info("setup/creating storage")
	if err != nil {
		log.Error("error creating db", sl.Err(err))
		os.Exit(1)
	}
	return storage
}

func setupRouter(ctx context.Context, log *slog.Logger, cfg *config.Config, storage *pg.Storage, client *redis.Client) chi.Router {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, httprate.Limit(50, 1*time.Minute))
	router.Use(mwLogger.New(log))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	// Routes
	setupRoutes(ctx, cfg, router, log, storage, client)
	return router
}

func setupRoutes(ctx context.Context, cfg *config.Config, router chi.Router, log *slog.Logger, storage *pg.Storage, client *redis.Client) {
	_complaintService := complaintService.NewComplaintsService(storage)
	_categoryService := categoryService.NewCategoriesService(storage)
	_adminService := authService.NewAdminService(storage)
	router.Route("/complaints", func(r chi.Router) {
		r.With(cache.CacheMiddleware(client, 1*time.Minute, log)).
			Get("/", getallcomplaint.New(log, _complaintService)) // Получить все компл
		r.Post("/", create.New(log, _complaintService))                           //Создать компл
		r.Get("/{id}", get_complaint_by_complaint_id.New(log, _complaintService)) // Получить компл по айди
		r.Get("/can-submit", can_submit.New(log, _complaintService))
		r.Get("/by-token", get_complaints_by_token.New(cfg, log, _complaintService))
	})
	router.Route("/categories", func(r chi.Router) {
		r.Use(cache.CacheMiddleware(client, time.Minute*1, log))
		r.Get("/", categoriesGetAll.New(log, _categoryService))
		r.Get("/{id}", categories_get_by_id.New(log, _categoryService))
		r.Get("/{id}/complaints", get_complaints_by_category_id.New(log, _complaintService)) //Получить компл по категории айди
	})
	router.Get("/docs/*", httpSwagger.WrapHandler)

	router.Route("/admin", func(r chi.Router) {

		r.Use(admin_only.AdminOnlyMiddleware(log, cfg, _adminService))
		//Complaint
		r.Put("/complaints/{id}", update.New(ctx, log, _complaintService, client))
		r.Delete("/complaints/{id}", deleteComplaint.New(ctx, log, _complaintService, client)) //Удалить компл

		//Category
		r.Post("/categories", categoriesCreate.New(ctx, log, _categoryService, client)) //Создание категории
		r.Put("/categories/{id}", updateCategory.New(ctx, log, _categoryService, client))
		r.Delete("/categories/{id}", deleteCategoryById.New(ctx, log, _categoryService, client)) //Удалить категории по ID
	})
}

func startServer(cfg *config.Config, router chi.Router, log *slog.Logger) {
	timeout, err := time.ParseDuration(cfg.HTTPServer.Timeout)
	if err != nil {
		log.Error("invalid HTTP_TIMEOUT: %v", err)
	}
	idleTimeout, err := time.ParseDuration(cfg.HTTPServer.IdleTimeout)
	if err != nil {
		log.Error("invalid HTTP_TIMEOUT: %v", err)
	}
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  idleTimeout,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", sl.Err(err))
		}
	}()

	log.Info("server started", slog.String("address", cfg.HTTPServer.Address))
	<-done
	log.Info("stopping server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("error during server shutdown", sl.Err(err))
	} else {
		log.Info("server stopped successfully")
	}
}

func setupLogger(env string) *slog.Logger {

	defaultLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	switch env {
	case envLocal:
		return setupPrettySlog()
	case envDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return defaultLogger
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
