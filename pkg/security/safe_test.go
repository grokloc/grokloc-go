// Package security provides crypto and hashing support
package security

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SafeSuite struct {
	suite.Suite
}

func (suite *SafeSuite) TestSafeStr() {
	require.False(suite.T(), SafeStr(""))
	require.False(suite.T(), SafeStr("hello'"))
	require.False(suite.T(), SafeStr("hello`"))
	require.True(suite.T(), SafeStr("hello"))
}

func TestSafeSuite(t *testing.T) {
	suite.Run(t, new(SafeSuite))
}
