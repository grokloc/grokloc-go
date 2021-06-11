package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
	srv *Instance
	ctx context.Context
	ts  *httptest.Server
}

func (suite *ClientSuite) SetupTest() {
	var err error
	suite.srv, err = New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}
	suite.ctx = context.Background()
	suite.ts = httptest.NewServer(suite.srv.Router())
}

func (suite *ClientSuite) TestOk() {
	c, err := NewClient(suite.ts.URL, suite.srv.ST.RootUser, suite.srv.ST.RootUserAPISecret)
	require.Nil(suite.T(), err)
	resp, _, err := c.Ok()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
