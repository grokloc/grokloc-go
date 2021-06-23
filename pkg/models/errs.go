package models

import (
	"errors"
	"strings"
)

// ErrConflict describes a duplicate row insertion
var ErrConflict error = errors.New("row insertion conflict")

// ErrRowsAffected describes an incorrect number of rows changed from a db mutation
var ErrRowsAffected error = errors.New("db RowsAffected was not correct")

// ErrRelatedOrg signals that an org is missing or not Active
var ErrRelatedOrg error = errors.New("related org is missing or not Active")

// ErrRelatedUser signals that a user is missing, not Active, or is in a different org
var ErrRelatedUser error = errors.New("related user is missing, not Active, or is in a different org")

// ErrModelMigrate signals a model could not be migrated to a different version
var ErrModelMigrate error = errors.New("schema version error; cannot migrate model")

// ErrDisallowedValue signals a value of the right type, just not allowed
var ErrDisallowedValue error = errors.New("value disallowed in this context")

// UniqueConstraint will try to match the db unique constraint violation
func UniqueConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
