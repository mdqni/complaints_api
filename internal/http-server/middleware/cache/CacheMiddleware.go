package cache

import (
	"bytes"
	"complaint_server/internal/lib/logger/sl"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"time"
)

func CacheMiddleware(redis *redis.Client, ttl time.Duration, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := fmt.Sprintf("cache:%s", r.URL.Path) //Можно юзать юрл как ключ
			log.Info(key)
			ctx := r.Context()
			// Проверяем кеш
			cached, err := redis.Get(ctx, key).Result()
			if err == nil {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(cached))
				return
			}
			recorder := &responseRecorder{ResponseWriter: w, body: new(bytes.Buffer)}
			next.ServeHTTP(recorder, r)

			if recorder.status == http.StatusOK {
				err := redis.Set(ctx, key, recorder.body.String(), ttl).Err()
				if err != nil {
					log.Error("Failed to set cache", sl.Err(err))
				} else {
					log.Info("Cache set for key", slog.String("key", key))
				}
			}

		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rw *responseRecorder) Write(p []byte) (int, error) {
	rw.body.Write(p)                  // Сохраняем ответ в буфер
	return rw.ResponseWriter.Write(p) // Пишем ответ клиенту
}

func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
