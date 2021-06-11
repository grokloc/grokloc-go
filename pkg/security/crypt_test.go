package security

import (
	"testing"

	"github.com/google/uuid"
	"github.com/matthewhartstonge/argon2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CryptSuite struct {
	suite.Suite
	Argon2Cfg argon2.Config
}

func (suite *CryptSuite) SetupTest() {
	suite.Argon2Cfg = argon2.DefaultConfig()
}

func (suite *CryptSuite) TestEncrypt() {
	key, err := MakeKey(uuid.NewString())
	require.Nil(suite.T(), err)
	s := uuid.NewString()
	e, err := Encrypt(s, key)
	require.Nil(suite.T(), err)
	d, err := Decrypt(e, key)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), s, d)
	notKey, err := MakeKey(uuid.NewString())
	require.Nil(suite.T(), err)
	_, err = Decrypt(e, notKey)
	require.Error(suite.T(), err)
}

func (suite *CryptSuite) TestDerivePassword() {
	password := uuid.NewString()
	derived, err := DerivePassword(password, suite.Argon2Cfg)
	require.Nil(suite.T(), err)
	good, err := VerifyPassword(password, derived)
	require.Nil(suite.T(), err)
	require.True(suite.T(), good)
	bad, err := VerifyPassword(uuid.NewString(), derived)
	require.Nil(suite.T(), err)
	require.False(suite.T(), bad)
}

func TestCryptSuite(t *testing.T) {
	suite.Run(t, new(CryptSuite))
}
