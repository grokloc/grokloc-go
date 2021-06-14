package security

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SafeSuite struct {
	suite.Suite
}

func (s *SafeSuite) TestSafeStr() {
	require.False(s.T(), SafeStr(""))
	require.False(s.T(), SafeStr("hello'"))
	require.False(s.T(), SafeStr("hello`"))
	require.True(s.T(), SafeStr("hello"))
}

func TestSafeSuite(t *testing.T) {
	suite.Run(t, new(SafeSuite))
}
