package models

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatusSuite struct {
	suite.Suite
}

func (suite *StatusSuite) TestStatus() {
	var err error
	var status Status
	_, err = NewStatus(-1)
	require.Error(suite.T(), err)
	_, err = NewStatus(100)
	require.Error(suite.T(), err)
	status, err = NewStatus(0)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), StatusUnconfirmed, status)
	status, err = NewStatus(1)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), StatusActive, status)
	status, err = NewStatus(2)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), StatusInactive, status)
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(StatusSuite))
}
