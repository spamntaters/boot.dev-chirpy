package api

import (
	"net/http"
	"sync/atomic"

	"github.com/spamntaters/boot.dev-chirpy/internal/database"
)

type Config struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	Secret         string
}

func (cfg *Config) MiddlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
