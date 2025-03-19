package main

import (
	"complaint_server/internal/config"
	categoriesCreate "complaint_server/internal/http-server/handlers/category/create"
	deleteCategoryById "complaint_server/internal/http-server/handlers/category/delete"
	categoriesGetAll "complaint_server/internal/http-server/handlers/category/get_all"
	"complaint_server/internal/http-server/handlers/complaints/create"
	deleteComplaint "complaint_server/internal/http-server/handlers/complaints/delete"
	"complaint_server/internal/http-server/handlers/complaints/get_all"
	"complaint_server/internal/http-server/handlers/complaints/get_complaint_by_complaint_id"
	"complaint_server/internal/http-server/handlers/complaints/get_complaints_by_category_id"
	"complaint_server/internal/http-server/handlers/complaints/update_complaint_status"
	"complaint_server/internal/http-server/middleware/admin_only"
	mwLogger "complaint_server/internal/http-server/middleware/logger"
	"complaint_server/internal/lib/logger/handlers/slogpretty"
	"complaint_server/internal/lib/logger/sl"
	"complaint_server/internal/service"
	"complaint_server/internal/storage/pg"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
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
// @host			localhost:8082
// @BasePath		/
func main() {
	//Init Logger
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info(
		"starting complaints server",
		slog.String("env", cfg.Env),
	)
	log.Debug("Debug message are enabled")

	storage := setupStorage(cfg.ConnString, log)
	router := setupRouter(log, cfg, storage)

	startServer(cfg, router, log)
}
func setupStorage(connString string, log *slog.Logger) *pg.Storage {
	storage, err := pg.New(connString)
	if err != nil {
		log.Error("error creating db", sl.Err(err))
		os.Exit(1)
	}
	return storage
}

func setupRouter(log *slog.Logger, cfg *config.Config, storage *pg.Storage) chi.Router {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, httprate.Limit(50, 1*time.Minute))
	router.Use(mwLogger.New(log))

	// Routes
	setupRoutes(cfg, router, log, storage)
	return router
}

func setupRoutes(cfg *config.Config, router chi.Router, log *slog.Logger, storage *pg.Storage) {
	complaintService := service.New(storage)
	router.Route("/complaint", func(r chi.Router) {
		r.Post("/", create.New(log, complaintService))                           //Создать компл
		r.Get("/", get_all.New(log, complaintService))                           //Получить все компл
		r.Get("/{id}", get_complaint_by_complaint_id.New(log, complaintService)) //Получить компл по айди
	})
	router.Route("/category", func(r chi.Router) {
		r.Get("/", categoriesGetAll.New(log, storage))                           //Удаление категории
		r.Get("/{id}", get_complaints_by_category_id.New(log, complaintService)) //Получить компл по категории айди
	})
	router.Route("/docs", func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Get("/*", httpSwagger.WrapHandler)
	})

	router.Route("/admin", func(r chi.Router) {
		r.Use(admin_only.AdminOnlyMiddleware)
		//Complaint
		r.Put("/complaint/{id}/status", resolveComplaint.New(log, complaintService))
		r.Delete("/complaint/{id}", deleteComplaint.New(log, complaintService)) //Удалить компл

		//Category
		r.Post("/category", categoriesCreate.New(log, storage))          //Создание категории
		r.Delete("/category/{id}", deleteCategoryById.New(log, storage)) //Удалить категории по АЙДИ

	})
}

func startServer(cfg *config.Config, router chi.Router, log *slog.Logger) {
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", sl.Err(err))
		}
	}()

	log.Info("server started", slog.String("address", cfg.Address))
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
