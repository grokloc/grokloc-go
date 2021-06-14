package env

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EnvSuite struct {
	suite.Suite
}

func (s *EnvSuite) TestEnv() {
	var err error
	var level Level
	_, err = NewLevel("")
	require.Error(s.T(), err)
	level, err = NewLevel("UNIT")
	require.Nil(s.T(), err)
	require.Equal(s.T(), Unit, level)
}

func TestEnvSuite(t *testing.T) {
	suite.Run(t, new(EnvSuite))
}
