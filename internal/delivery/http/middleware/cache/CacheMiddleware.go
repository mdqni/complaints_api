package cache

import (
	"bytes"
	"complaint_server/internal/shared/logger/sl"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"time"
)

func CacheMiddleware(redis *redis.Client, ttl time.Duration, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			ctx := r.Context()
			key := fmt.Sprintf("cache:%s", r.URL.Path)

			cached, err := redis.Get(ctx, key).Bytes()
			if err == nil {
				log.Info("Cache hit", slog.String("key", key))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(cached)
				return
			}

			recorder := &responseRecorder{
				ResponseWriter: w,
				body:           new(bytes.Buffer),
			}
			next.ServeHTTP(recorder, r)

			if recorder.status == http.StatusOK {
				err := redis.Set(ctx, key, recorder.body.Bytes(), ttl).Err()
				if err != nil {
					log.Error("Failed to set cache", sl.Err(err))
				} else {
					log.Info("Cache set", slog.String("key", key))
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

func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseRecorder) Write(p []byte) (int, error) {
	rw.body.Write(p)
	return rw.ResponseWriter.Write(p)
}
