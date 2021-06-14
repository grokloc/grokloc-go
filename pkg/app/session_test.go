package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionSuite struct {
	suite.Suite
	srv *Instance
	ctx context.Context
	ts  *httptest.Server
}

func (s *SessionSuite) SetupTest() {
	var err error
	s.srv, err = New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(s.srv.GetUserAndOrg)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			_, err := w.Write([]byte("OK"))
			if err != nil {
				panic(err.Error())
			}
		})
	})

	s.ctx = context.Background()
	s.ts = httptest.NewServer(r)
}

func (s *SessionSuite) TestFoundAndActive() {
	require.True(s.T(), true)
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionSuite))
}
