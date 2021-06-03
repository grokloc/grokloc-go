// Package org models an organization
package org

import (
	"github.com/grokloc/grokloc-go/pkg/models"
)

// exported org symbols
const (
	SchemaVersion = 0
)

// Instance is an organization model
type Instance struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}
