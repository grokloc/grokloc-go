// Package jwt manage authorization claims for the app
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

func (suite *JWTSuite) SetupTest() {
	var err error
	suite.ST, err = state.New(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *JWTSuite) TestJWT() {
	// make a new org and user as owner
	o, u, err := util.NewOrgOwner(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	claims, err := New(*u)
	require.Nil(suite.T(), err)
	token := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(u.ID + string(suite.ST.SigningKey)))
	require.Nil(suite.T(), err)
	claimsDecoded, err := Decode(u.ID, signedToken, suite.ST.SigningKey)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), u.ID, claimsDecoded.Id)
	require.Equal(suite.T(), u.Org, claimsDecoded.Org)

	// wrong user
	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	uOther, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(suite.T(), err)
	u.Meta.Status = models.StatusActive
	err = uOther.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)
	_, err = Decode(uOther.ID, signedToken, suite.ST.SigningKey)
	require.Error(suite.T(), err)

	// bad JWT
	_, err = Decode(u.ID, uuid.NewString(), suite.ST.SigningKey)
	require.Error(suite.T(), err)

	// bad signing key
	otherSigningKey, err := security.MakeKey(uuid.NewString())
	require.Nil(suite.T(), err)
	_, err = Decode(u.ID, signedToken, otherSigningKey)
	require.Error(suite.T(), err)
}

func TestJWTSuite(t *testing.T) {
	suite.Run(t, new(JWTSuite))
}
