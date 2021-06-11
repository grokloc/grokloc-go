package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// path/route constants
const (
	APIPath    = "/api"
	APIRoute   = APIPath + "/" + Version
	TokenRoute = APIRoute + "/token"

	OkPath      = "/ok"
	OkRoute     = APIRoute + OkPath
	StatusPath  = "/status"
	StatusRoute = APIRoute + StatusPath // like ping, but requires auth
)

// URL parameter names
const (
	IDParam = "id"
)

// Router provides API route handlers
func (srv Instance) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(srv.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Second))

	r.Get(OkRoute, Ok)

	return r
}

// Ok is just a ping-acknowledgement
func Ok(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		panic(err.Error())
	}
}
