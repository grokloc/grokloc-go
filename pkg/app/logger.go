package app

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger is a middleware that adds logging to each request
func (srv Instance) RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer srv.ST.L.Sync() // nolint
		sugar := srv.ST.L.Sugar()
		sugar.Infow("request",
			"reqid", middleware.GetReqID(ctx),
			"method", r.Method,
			"path", r.URL,
			"remote", r.RemoteAddr,
			"headers", r.Header,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
