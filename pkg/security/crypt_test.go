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

func (s *CryptSuite) SetupTest() {
	s.Argon2Cfg = argon2.DefaultConfig()
}

func (s *CryptSuite) TestEncrypt() {
	key, err := MakeKey(uuid.NewString())
	require.Nil(s.T(), err)
	str := uuid.NewString()
	e, err := Encrypt(str, key)
	require.Nil(s.T(), err)
	d, err := Decrypt(e, key)
	require.Nil(s.T(), err)
	require.Equal(s.T(), str, d)
	notKey, err := MakeKey(uuid.NewString())
	require.Nil(s.T(), err)
	_, err = Decrypt(e, notKey)
	require.Error(s.T(), err)
}

func (s *CryptSuite) TestDerivePassword() {
	password := uuid.NewString()
	derived, err := DerivePassword(password, s.Argon2Cfg)
	require.Nil(s.T(), err)
	good, err := VerifyPassword(password, derived)
	require.Nil(s.T(), err)
	require.True(s.T(), good)
	bad, err := VerifyPassword(uuid.NewString(), derived)
	require.Nil(s.T(), err)
	require.False(s.T(), bad)
}

func TestCryptSuite(t *testing.T) {
	suite.Run(t, new(CryptSuite))
}
