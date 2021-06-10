// Package state manages all external state conns
package state

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/matthewhartstonge/argon2"
	"go.uber.org/zap"
)

// Instance is a single set of conns
type Instance struct {
	Level                                env.Level
	Master                               *sql.DB
	Replicas                             []*sql.DB
	Key                                  []byte
	SigningKey                           []byte
	Argon2Cfg                            argon2.Config
	RootOrg, RootUser, RootUserAPISecret string
	L                                    *zap.Logger
}

// New creates a new instance for the given level
func New(level env.Level) (*Instance, error) {
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
		log.Fatal("there are no replicas")
	}
	return s.Replicas[rand.Intn(l)]
}

// Close should be deferred in the main context
func (s *Instance) Close() error {
	if s.Level == env.Unit {
		err := s.Master.Close()
		if err != nil {
			return err
		}
	} else {
		log.Fatal("env unsupported")
	}
	return nil
}
