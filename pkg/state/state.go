// Package state manages all external state conns
package state

import (
	"database/sql"
	"errors"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/matthewhartstonge/argon2"
)

// Instance is a single set of conns
type Instance struct {
	Master                               *sql.DB
	Replicas                             []*sql.DB
	Key                                  []byte
	Argon2Cfg                            argon2.Config
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
