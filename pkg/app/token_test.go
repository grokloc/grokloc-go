package app

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func (s *SessionSuite) TestNewToken() {
	// get a token
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+"/token", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(s.srv.ST.RootUser+s.srv.ST.RootUserAPISecret))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	contentType := resp.Header.Get("content-type")
	require.Equal(s.T(), "application/json", contentType)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)
	now := time.Now().Unix()
	require.GreaterOrEqual(s.T(), tok.Expires, now)
	require.NotEmpty(s.T(), tok.Bearer)

	// now try using the token
	req, err = http.NewRequest(http.MethodGet, s.ts.URL+"/verify", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, tok.Bearer)
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *SessionSuite) TestNewTokenMissingHeader() {
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+"/token", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *SessionSuite) TestNewTokenMalformedHeader() {
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+"/token", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(uuid.NewString()))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

func (s *SessionSuite) TestOtherUsersToken() {
	_, u, err := util.NewOrgOwner(s.ctx, s.srv.ST.Master, s.srv.ST.Key)
	require.Nil(s.T(), err)
	claims, err := jwt.New(*u)
	require.Nil(s.T(), err)
	token := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(u.ID + string(s.srv.ST.SigningKey)))
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/verify", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, signedToken)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}
