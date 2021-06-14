package app

import (
	"context"
	"log"
	"net/http/httptest"
	"testing"

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
	s.ctx = context.Background()
	s.ts = httptest.NewServer(s.srv.Router())
}

func (s *SessionSuite) TestFoundAndActive() {
	require.True(s.T(), true)
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionSuite))
}
