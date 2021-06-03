// Package models provides shared model definitions
package models

import "errors"

// Status is an int when stored
type Status int

// exported status values
const (
	None        = Status(-1)
	Unconfirmed = Status(0)
	Active      = Status(1)
	Inactive    = Status(2)
)

// NewStatus creates a Status from an int
func NewStatus(status int) (Status, error) {
	switch status {
	case 0:
		return Unconfirmed, nil
	case 1:
		return Active, nil
	case 2:
		return Inactive, nil
	default:
		return None, errors.New("unknown status")
	}
}

// Meta models metadata common to all models
type Meta struct {
	Ctime         int64  `json:"ctime"`
	Mtime         int64  `json:"mtime"`
	SchemaVersion int    `json:"schema_version"`
	Status        Status `json:"status"`
}

// Base models core attributes common to all models
type Base struct {
	ID   string `json:"id"`
	Meta Meta   `json:"meta"`
}
