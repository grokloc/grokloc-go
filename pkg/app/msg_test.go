package app

import (
	"encoding/json"
	"testing"

	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MsgSuite is responsible fo testing app message functionality
type MsgSuite struct {
	suite.Suite
}

func (s *MsgSuite) TestUnmarshalUpdateMsg() {
	type T struct {
		Status int `json:"status"`
	}

	t := T{Status: int(models.StatusNone) - 1}
	bs, err := json.Marshal(t)
	require.Nil(s.T(), err)
	var m UpdateStatusMsg
	err = m.UnmarshalJSON(bs)
	require.Error(s.T(), err)

	t = T{Status: int(models.StatusNone)}
	bs, err = json.Marshal(t)
	require.Nil(s.T(), err)
	err = m.UnmarshalJSON(bs)
	require.Error(s.T(), err)

	t = T{Status: int(models.StatusInactive) + 1}
	bs, err = json.Marshal(t)
	require.Nil(s.T(), err)
	err = m.UnmarshalJSON(bs)
	require.Error(s.T(), err)

	for _, v := range []models.Status{
		models.StatusUnconfirmed,
		models.StatusActive,
		models.StatusInactive,
	} {
		t = T{Status: int(v)}
		bs, err = json.Marshal(t)
		require.Nil(s.T(), err)
		err = m.UnmarshalJSON(bs)
		require.Nil(s.T(), err)
	}
}

func TestMsgSuite(t *testing.T) {
	suite.Run(t, new(MsgSuite))
}
