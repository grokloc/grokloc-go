// Package app provides support for the ReST API
package app

import (
	"time"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/state"
)

// Version is the current API version
const Version = "v0"

// API headers
// TokenRequest is formatted as hex-encode(id+api-secret)
const (
	IDHeader           = "X-GrokLOC-ID"
	TokenHeader        = "X-GrokLOC-Token"
	TokenRequestHeader = "X-GrokLOC-TokenRequest"
)

// Auth levels to be found in ctx with key authLevelCtxKey
const (
	AuthUser = iota
	AuthOrg
	AuthRoot
)

// contextKey is used to dismbiguate keys for vars put into request contexts
type contextKey struct {
	name string
}

// Context key instances for inserting and reading context vars
var (
	sessionCtxKey   = &contextKey{"session"}   // nolint
	authLevelCtxKey = &contextKey{"authlevel"} // nolint
)

// Instance is a single app server
type Instance struct {
	ST      *state.Instance
	Started time.Time
}

// New creates a new app server Instance
func New(level env.Level) (*Instance, error) {
	st, err := state.New(level)
	if err != nil {
		return nil, err
	}
	return &Instance{ST: st, Started: time.Now()}, nil
}
