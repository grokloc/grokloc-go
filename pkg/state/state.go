// Package state manages all external state conns
package state

import (
	"database/sql"
	"errors"

	"github.com/grokloc/grokloc-go/pkg/env"
)

// Instance is a single set of conns
type Instance struct {
	Master                               *sql.DB
	Replicas                             []*sql.DB
	Key                                  string
	RootOrg, RootUser, RootUserAPISecret string
}

// NewInstanceFromLevel creates a new instance for the given level
func NewInstanceFromLevel(level env.Level) (Instance, error) {
	if level == env.None {
		return Instance{}, errors.New("no instance for None")
	}
	if level == env.Unit {
		return unitInstance(), nil
	}
	return Instance{}, errors.New("no instance available")
}
