package models

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatusSuite struct {
	suite.Suite
}

func (s *StatusSuite) TestStatus() {
	var err error
	var status Status
	_, err = NewStatus(-1)
	require.Error(s.T(), err)
	_, err = NewStatus(100)
	require.Error(s.T(), err)
	status, err = NewStatus(0)
	require.Nil(s.T(), err)
	require.Equal(s.T(), StatusUnconfirmed, status)
	status, err = NewStatus(1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), StatusActive, status)
	status, err = NewStatus(2)
	require.Nil(s.T(), err)
	require.Equal(s.T(), StatusInactive, status)
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(StatusSuite))
}
