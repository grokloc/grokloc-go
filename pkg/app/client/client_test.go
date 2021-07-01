package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/app"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientSuite tests will act as implicit tests for all public app endpoints
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

func (s *ClientSuite) TestCreateOrg() {
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	resp, _, err := c.CreateOrg(uuid.NewString())
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
}

func (s *ClientSuite) TestReadOrg() {
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	name := uuid.NewString()
	resp, _, err := c.CreateOrg(name)
	require.Nil(s.T(), err)
	location := resp.Header.Get("location")
	require.NotEmpty(s.T(), location)
	pathElts := strings.Split(location, "/")
	require.True(s.T(), len(pathElts) != 0)
	var respBody []byte
	resp, respBody, err = c.ReadOrg(pathElts[len(pathElts)-1])
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var o org.Instance
	err = json.Unmarshal(respBody, &o)
	require.Nil(s.T(), err)
	require.Equal(s.T(), name, o.Name)
	require.Equal(s.T(), pathElts[len(pathElts)-1], o.ID)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
