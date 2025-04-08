package storage

import (
	"complaint_server/internal/config"
	"complaint_server/internal/lib/logger/sl"
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

func NewClient(ctx context.Context, cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         "34.89.141.175:1006",
		Username:     "default",
		DB:           0,
		Password:     "Aitusa2025!",
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.RedisClient.Timeout,
		WriteTimeout: cfg.RedisClient.Timeout,
	})
	log.Info(db.String())

	if err := db.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to redis server", sl.Err(err))
		return nil, err
	} else {
		log.Info("connected to Redis successfully")
	}

	return db, nil
}
