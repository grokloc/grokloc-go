package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SessionSuite is responsible fo testing interior methods used in auth
type SessionSuite struct {
	suite.Suite
	srv *Instance
	ctx context.Context
	ts  *httptest.Server
	c   *http.Client
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
	s.c = &http.Client{}
}

func (s *SessionSuite) TestFoundAndActive() {
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *SessionSuite) TestUserNotFound() {
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, uuid.NewString())
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *SessionSuite) TestUserInactive() {
	_, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	err = u.UpdateStatus(s.ctx, s.srv.ST.Master, models.StatusInactive)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *SessionSuite) TestOrgInactive() {
	o, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	err = o.UpdateStatus(s.ctx, s.srv.ST.Master, models.StatusInactive)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionSuite))
}
