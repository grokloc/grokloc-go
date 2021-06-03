// Package state manages all external state conns
package state

import (
	"testing"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StateSuite struct {
	suite.Suite
}

func (suite *StateSuite) TestUnit() {
	_, err := NewInstanceFromLevel(env.Unit)
	require.Nil(suite.T(), err)
}

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}
