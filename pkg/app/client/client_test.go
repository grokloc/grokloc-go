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
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/util"
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
	orgID := pathElts[len(pathElts)-1]
	var respBody []byte
	resp, respBody, err = c.ReadOrg(orgID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var o org.Instance
	err = json.Unmarshal(respBody, &o)
	require.Nil(s.T(), err)
	require.Equal(s.T(), name, o.Name)
	require.Equal(s.T(), orgID, o.ID)
}

func (s *ClientSuite) TestUpdateOrgOwner() {
	o, _, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	derived, err := security.DerivePassword(uuid.NewString(), s.srv.ST.Argon2Cfg)
	require.Nil(s.T(), err)
	rUser, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, derived)
	require.Nil(s.T(), err)
	err = rUser.Insert(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	err = rUser.UpdateStatus(s.ctx, s.srv.ST.Master, models.StatusActive)
	require.Nil(s.T(), err)
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	resp, _, err := c.UpdateOrgOwner(o.ID, rUser.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
	var respBody []byte
	resp, respBody, err = c.ReadOrg(o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	var oRead org.Instance
	err = json.Unmarshal(respBody, &oRead)
	require.Nil(s.T(), err)
	require.Equal(s.T(), rUser.ID, oRead.Owner)
}

func (s *ClientSuite) TestUpdateOrgStatus() {
	o, _, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	c, err := NewClient(s.ts.URL, s.srv.ST.RootUser, s.srv.ST.RootUserAPISecret)
	require.Nil(s.T(), err)
	resp, _, err := c.UpdateOrgStatus(o.ID, models.StatusInactive)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
	oRead, err := org.Read(s.ctx, s.srv.ST.RandomReplica(), o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, oRead.Meta.Status)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
