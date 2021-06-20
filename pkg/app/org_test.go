package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OrgSuite is responsible fo testing org endpoints
type OrgSuite struct {
	suite.Suite
	srv   *Instance
	ctx   context.Context
	ts    *httptest.Server
	c     *http.Client
	token *Token
}

func (s *OrgSuite) SetupTest() {
	var err error
	s.srv, err = New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}

	s.ctx = context.Background()
	s.ts = httptest.NewServer(s.srv.Router())
	s.c = &http.Client{}

	// for making authenticated requests, get a token
	// (these steps are already run through real tests in token_test)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(s.srv.ST.RootUser+s.srv.ST.RootUserAPISecret))
	resp, err := s.c.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.token = &tok
}

func (s *OrgSuite) TestCreateOrg() {
	bs, err := json.Marshal(CreateMsg{Name: uuid.NewString()})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenHeader, s.token.Bearer)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	location := resp.Header.Get("location")
	require.NotEmpty(s.T(), location)

	// duplicate
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)
}

func (s *OrgSuite) TestCreateOrgBadPayload() {
	req, err := http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer([]byte(uuid.NewString())))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenHeader, s.token.Bearer)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *OrgSuite) TestCreateOrgNotRoot() {
	_, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(u.ID+u.APISecret))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)

	bs, err := json.Marshal(CreateMsg{Name: uuid.NewString()})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenHeader, tok.Bearer)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *OrgSuite) TestReadOrg() {
	now := time.Now()
	o, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(u.ID+u.APISecret))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+OrgRoute+"/"+o.ID, nil)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenHeader, tok.Bearer)
	require.Nil(s.T(), err)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var oRead org.Instance
	err = json.Unmarshal(respBody, &oRead)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.ID, oRead.ID)
	require.Equal(s.T(), o.Name, oRead.Name)
	require.GreaterOrEqual(s.T(), oRead.Meta.Ctime, now.Unix())
	require.GreaterOrEqual(s.T(), oRead.Meta.Mtime, now.Unix())

	// root can read it too
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+OrgRoute+"/"+o.ID, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenHeader, s.token.Bearer)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var oReadRoot org.Instance
	err = json.Unmarshal(respBody, &oReadRoot)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.ID, oReadRoot.ID)
}

func (s *OrgSuite) TestReadOtherOrg() {
	_, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(u.ID+u.APISecret))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	// root org cannot be read by u
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+OrgRoute+"/"+s.srv.ST.RootOrg, nil)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenHeader, tok.Bearer)
	require.Nil(s.T(), err)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
