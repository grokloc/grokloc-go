package client

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grokloc/grokloc-go/pkg/app"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
	srv *app.Instance
	ctx context.Context
	ts  *httptest.Server
}

func (s *ClientSuite) SetupTest() {
	var err error
	s.srv, err = app.New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.ctx = context.Background()
	s.ts = httptest.NewServer(s.srv.Router())
}

func (s *ClientSuite) TestOk() {
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	resp, _, err := c.Ok()
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *ClientSuite) TestStatus() {
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	resp, _, err := c.Status()
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
