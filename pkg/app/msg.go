package app

import (
	"encoding/json"
	"errors"

	"github.com/grokloc/grokloc-go/pkg/models"
)

// UpdateStatusMsg is what a client should marshal to send as a json body to
// status update endpoints
type UpdateStatusMsg struct {
	Status models.Status `json:"status"`
}

// UnmarshalJSON is a custom unmarshal for UpdateStatusMsg
func (m *UpdateStatusMsg) UnmarshalJSON(bs []byte) error {
	var t map[string]int
	err := json.Unmarshal(bs, &t)
	if err != nil {
		return err
	}
	v, ok := t["status"]
	if !ok {
		return errors.New("no status field found")
	}
	s, err := models.NewStatus(v)
	if err != nil {
		return err
	}
	m.Status = s
	return nil
}
