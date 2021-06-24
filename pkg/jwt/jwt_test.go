package jwt

import (
	"context"
	"log"
	"testing"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/state"
	"github.com/grokloc/grokloc-go/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JWTSuite struct {
	suite.Suite
	ST *state.Instance
}

func (s *JWTSuite) SetupTest() {
	var err error
	s.ST, err = state.New(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *JWTSuite) TestJWT() {
	// make a new org and user as owner
	o, u, err := util.NewOrgOwner(context.Background(), s.ST.Master, s.ST.Key)
	require.Nil(s.T(), err)

	claims, err := New(*u)
	require.Nil(s.T(), err)
	token := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(u.ID + string(s.ST.SigningKey)))
	require.Nil(s.T(), err)
	claimsDecoded, err := Decode(u.ID, signedToken, s.ST.SigningKey)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.ID, claimsDecoded.Id)
	require.Equal(s.T(), u.Org, claimsDecoded.Org)

	// wrong user
	password, err := security.DerivePassword(uuid.NewString(), s.ST.Argon2Cfg)
	require.Nil(s.T(), err)
	uOther, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(s.T(), err)
	u.Meta.Status = models.StatusActive
	err = uOther.Insert(context.Background(), s.ST.Master, s.ST.Key)
	require.Nil(s.T(), err)
	_, err = Decode(uOther.ID, signedToken, s.ST.SigningKey)
	require.Error(s.T(), err)

	// bad JWT
	_, err = Decode(u.ID, uuid.NewString(), s.ST.SigningKey)
	require.Error(s.T(), err)

	// bad signing key
	otherSigningKey, err := security.MakeKey(uuid.NewString())
	require.Nil(s.T(), err)
	_, err = Decode(u.ID, signedToken, otherSigningKey)
	require.Error(s.T(), err)
}

func (s *JWTSuite) TestHeaderVal() {
	token := uuid.NewString() // it just needs to be some string
	require.Equal(s.T(), token, FromHeaderVal(ToHeaderVal(token)))
	require.Equal(s.T(), token, FromHeaderVal(token))
}

func TestJWTSuite(t *testing.T) {
	suite.Run(t, new(JWTSuite))
}
