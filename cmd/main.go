package main

import (
	_ "complaint_server/docs"
	"complaint_server/internal/app"
	"complaint_server/internal/config"
	"complaint_server/internal/shared/logger/handlers/slogpretty"
	"complaint_server/internal/shared/logger/sl"
	"context"
	"errors"
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
const contextTimeout = 5 * time.Second

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
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	application, err := app.NewApp(ctx, cfg, log)
	if err != nil {
		log.Error("failed to create app", sl.Err(err))
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", sl.Err(err))
		}
	}()
	log.Info("server started", slog.String("address", cfg.HTTPServer.Address))
	<-done
	log.Info("stopping server gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), contextTimeout)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
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
