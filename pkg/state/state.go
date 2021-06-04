// Package state manages all external state conns
package state

import (
	"database/sql"
	"errors"
	"math/rand"

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
func NewInstanceFromLevel(level env.Level) (*Instance, error) {
	if level == env.None {
		return nil, errors.New("no instance for None")
	}
	if level == env.Unit {
		return unitInstance(), nil
	}
	return nil, errors.New("no instance available")
}

// RandomReplica selects a random replica
func (s *Instance) RandomReplica() *sql.DB {
	l := len(s.Replicas)
	if l == 0 {
		panic("there are no replicas")
	}
	return s.Replicas[rand.Intn(l)]
}
