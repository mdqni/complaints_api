package storage

import (
	"complaint_server/internal/config"
	"complaint_server/internal/shared/logger/sl"
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

func NewClient(ctx context.Context, cfg *config.Config, log *slog.Logger) (*redis.Client, error) {
	dialTimeout, err := time.ParseDuration(cfg.RedisClient.DialTimeout)
	if err != nil {
		log.Error("invalid REDIS_DIAL_TIMEOUT: %v", err)
	}

	timeout, err := time.ParseDuration(cfg.RedisClient.Timeout)
	if err != nil {
		log.Error("invalid REDIS_TIMEOUT: %v", err)
	}

	db := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisClient.Addr,
		Username:     cfg.RedisClient.User,
		DB:           0,
		Password:     cfg.RedisClient.Password,
		MaxRetries:   cfg.RedisClient.MaxRetries,
		DialTimeout:  dialTimeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
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

type Cache interface {
	Delete(ctx context.Context, key string) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
