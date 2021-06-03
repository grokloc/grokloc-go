// Package org models an organization
package org

import (
	"errors"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/security"
)

// exported org symbols
const (
	SchemaVersion = 0
	OwnerNone     = "OWNER.NONE"
)

// Instance is an organization model
type Instance struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// New creates a new org that hasn't been created before
func New(name string) (*Instance, error) {
	if !security.SafeStr(name) {
		return nil, errors.New("malformed name")
	}
	o := &Instance{Name: name, Owner: OwnerNone}
	o.ID = uuid.NewString()
	o.Meta.SchemaVersion = SchemaVersion
	o.Meta.Status = models.Unconfirmed
	return o, nil
}
