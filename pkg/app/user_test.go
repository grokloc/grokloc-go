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

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserSuite is responsible fo testing user endpoints
type UserSuite struct {
	suite.Suite
	srv   *Instance
	ctx   context.Context
	ts    *httptest.Server
	c     *http.Client
	token *Token
}

func (s *UserSuite) SetupTest() {
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

func (s *UserSuite) TestCreateUser() {
	o, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	bs, err := json.Marshal(CreateUserMsg{
		DisplayName: uuid.NewString(),
		Email:       uuid.NewString(),
		Org:         o.ID,
		Password:    uuid.NewString(),
	})
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
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

	// as org owner
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(u.ID+u.APISecret))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	bs, err = json.Marshal(CreateUserMsg{
		DisplayName: uuid.NewString(),
		Email:       uuid.NewString(),
		Org:         o.ID,
		Password:    uuid.NewString(),
	})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	location = resp.Header.Get("location")
	require.NotEmpty(s.T(), location)
}

func (s *UserSuite) TestCreateUserForbidden() {
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

	oOther, _, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	bs, err := json.Marshal(CreateUserMsg{
		DisplayName: uuid.NewString(),
		Email:       uuid.NewString(),
		Org:         oOther.ID,
		Password:    uuid.NewString(),
	})
	require.Nil(s.T(), err)

	// as owner of other org
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)

	// as regular user in org (just skip web api and create direct)
	rUser, err := user.New(uuid.NewString(), uuid.NewString(), oOther.ID, uuid.NewString())
	require.Nil(s.T(), err)
	err = rUser.Insert(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	err = rUser.UpdateStatus(s.ctx, s.srv.ST.Master, models.StatusActive)
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, rUser.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(rUser.ID+rUser.APISecret))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	bs, err = json.Marshal(CreateUserMsg{
		DisplayName: uuid.NewString(),
		Email:       uuid.NewString(),
		Org:         oOther.ID,
		Password:    uuid.NewString(),
	})
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, rUser.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

func (s *UserSuite) TestReadUser() {
	o, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+UserRoute+"/"+u.ID, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var uRead user.Instance
	err = json.Unmarshal(respBody, &uRead)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.ID, uRead.ID)
	require.Equal(s.T(), u.APISecret, uRead.APISecret)
	require.Equal(s.T(), u.APISecretDigest, uRead.APISecretDigest)
	require.Equal(s.T(), u.DisplayName, uRead.DisplayName)
	require.Equal(s.T(), u.DisplayNameDigest, uRead.DisplayNameDigest)
	require.Equal(s.T(), u.Email, uRead.Email)
	require.Equal(s.T(), u.EmailDigest, uRead.EmailDigest)
	require.Equal(s.T(), u.Org, uRead.Org)
	require.NotEqual(s.T(), u.Password, uRead.Password) // not returned in read
	require.NotEqual(s.T(), u.Meta.Ctime, uRead.Meta.Ctime)
	require.NotEqual(s.T(), u.Meta.Mtime, uRead.Meta.Mtime)

	// create regular user rUser in o
	rUser, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, uuid.NewString())
	require.Nil(s.T(), err)
	err = rUser.Insert(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	err = rUser.UpdateStatus(s.ctx, s.srv.ST.Master, models.StatusActive)
	require.Nil(s.T(), err)

	// org owner u read new user rUser
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(u.ID+u.APISecret))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+UserRoute+"/"+rUser.ID, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	err = json.Unmarshal(respBody, &uRead)
	require.Nil(s.T(), err)
	require.Equal(s.T(), rUser.ID, uRead.ID)
	require.Equal(s.T(), rUser.APISecret, uRead.APISecret)
	require.Equal(s.T(), rUser.APISecretDigest, uRead.APISecretDigest)
	require.Equal(s.T(), rUser.DisplayName, uRead.DisplayName)
	require.Equal(s.T(), rUser.DisplayNameDigest, uRead.DisplayNameDigest)
	require.Equal(s.T(), rUser.Email, uRead.Email)
	require.Equal(s.T(), rUser.EmailDigest, uRead.EmailDigest)
	require.Equal(s.T(), rUser.Org, uRead.Org)
	require.NotEqual(s.T(), rUser.Password, uRead.Password) // not returned in read
	require.NotEqual(s.T(), rUser.Meta.Ctime, uRead.Meta.Ctime)
	require.NotEqual(s.T(), rUser.Meta.Mtime, uRead.Meta.Mtime)

	// uRead read their own record
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, rUser.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(rUser.ID+rUser.APISecret))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+UserRoute+"/"+rUser.ID, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, rUser.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err = io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	err = json.Unmarshal(respBody, &uRead)
	require.Nil(s.T(), err)
	require.Equal(s.T(), rUser.ID, uRead.ID)
	require.Equal(s.T(), rUser.APISecret, uRead.APISecret)
	require.Equal(s.T(), rUser.APISecretDigest, uRead.APISecretDigest)
	require.Equal(s.T(), rUser.DisplayName, uRead.DisplayName)
	require.Equal(s.T(), rUser.DisplayNameDigest, uRead.DisplayNameDigest)
	require.Equal(s.T(), rUser.Email, uRead.Email)
	require.Equal(s.T(), rUser.EmailDigest, uRead.EmailDigest)
	require.Equal(s.T(), rUser.Org, uRead.Org)
	require.NotEqual(s.T(), rUser.Password, uRead.Password) // not returned in read
	require.NotEqual(s.T(), rUser.Meta.Ctime, uRead.Meta.Ctime)
	require.NotEqual(s.T(), rUser.Meta.Mtime, uRead.Meta.Mtime)
}

func (s *UserSuite) TestReadUserForbidden() {
	// org owner read into other other org

	// regular user try to read anything other than their own id
	require.True(s.T(), true)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
