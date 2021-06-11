package env

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EnvSuite struct {
	suite.Suite
}

func (suite *EnvSuite) TestEnv() {
	var err error
	var level Level
	_, err = NewLevel("")
	require.Error(suite.T(), err)
	level, err = NewLevel("UNIT")
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), Unit, level)
}

func TestEnvSuite(t *testing.T) {
	suite.Run(t, new(EnvSuite))
}
