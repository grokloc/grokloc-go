package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// path/route constants
const (
	APIPath    = "/api/" + Version
	TokenRoute = APIPath + "/token"

	OkPath      = "/ok"
	OkRoute     = APIPath + OkPath
	OrgPath     = "/org"
	OrgRoute    = APIPath + OrgPath
	StatusPath  = "/status"
	StatusRoute = APIPath + StatusPath // auth + Ok
)

// URL parameter names
const (
	IDParam = "id"
)

// Router provides API route handlers
func (srv *Instance) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(srv.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Second))

	r.Get(OkRoute, Ok)

	r.Route(TokenRoute, func(r chi.Router) {
		r.Use(srv.GetUserAndOrg)
		r.Put("/", srv.NewToken)
	})

	r.Route(APIPath, func(r chi.Router) {
		r.Use(srv.GetUserAndOrg)
		r.Use(srv.VerifyToken)
		r.Get(StatusPath, Ok)
	})

	r.Route(OrgRoute, func(r chi.Router) {
		r.Use(srv.GetUserAndOrg)
		r.Use(srv.VerifyToken)
		r.Post("/", srv.CreateOrg)
	})

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
