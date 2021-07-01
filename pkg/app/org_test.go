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
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/models/user"
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
	bs, err := json.Marshal(CreateOrgMsg{Name: uuid.NewString()})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
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
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
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

	bs, err := json.Marshal(CreateOrgMsg{Name: uuid.NewString()})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
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
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
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
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
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
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	require.Nil(s.T(), err)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *OrgSuite) TestReadOrgNotFound() {
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+OrgRoute+"/"+uuid.NewString(), nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *OrgSuite) TestUpdateOrg() {
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

	bs, err := json.Marshal(UpdateOrgOwnerMsg{Owner: rUser.ID})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+OrgRoute+"/"+o.ID, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
	oRead, err := org.Read(s.ctx, s.srv.ST.RandomReplica(), o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), rUser.ID, oRead.Owner)

	bs, err = json.Marshal(UpdateStatusMsg{Status: models.StatusInactive})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+OrgRoute+"/"+o.ID, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
	oRead, err = org.Read(s.ctx, s.srv.ST.RandomReplica(), o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, oRead.Meta.Status)
}

func (s *OrgSuite) TestUpdateOrgForbidden() {
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
	bs, err := json.Marshal(UpdateStatusMsg{Status: models.StatusInactive})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+OrgRoute+"/"+o.ID, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(jwt.Authorization, tok.Bearer)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *OrgSuite) TestUpdateOrgBadUpdate() {
	o, _, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	bs, err := json.Marshal(map[string]interface{}{"x": 1})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+OrgRoute+"/"+o.ID, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *OrgSuite) TestUpdateOrgNotFound() {
	bs, err := json.Marshal(UpdateOrgOwnerMsg{Owner: uuid.NewString()})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+OrgRoute+"/"+uuid.NewString(), bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
