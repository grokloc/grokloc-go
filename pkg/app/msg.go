package app

import (
	"encoding/json"

	"github.com/grokloc/grokloc-go/pkg/models"
)

// UpdateStatusMsg is what a client should marshal to send as a json body to
// status update endpoints
type UpdateStatusMsg struct {
	Status models.Status `json:"status"`
}

// UnmarshalJSON is a custom unmarshal for UpdateStatusMsg
func (m *UpdateStatusMsg) UnmarshalJSON(bs []byte) error {
	type T struct {
		Status int `json:"status"`
	}
	var t T
	err := json.Unmarshal(bs, &t)
	if err != nil {
		return err
	}
	s, err := models.NewStatus(t.Status)
	if err != nil {
		return err
	}
	m.Status = s
	return nil
}
