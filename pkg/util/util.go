// Package util has labor-saving devices
package util

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/matthewhartstonge/argon2"
)

// NewOrgOwner returns a new, inserted org and a new, inserted user as owner
func NewOrgOwner(ctx context.Context, db *sql.DB, key []byte) (*org.Instance, *user.Instance, error) {
	o, err := org.New(uuid.NewString())
	if err != nil {
		return nil, nil, err
	}

	o.Meta.Status = models.StatusActive
	err = o.Insert(context.Background(), db)
	if err != nil {
		return nil, nil, err
	}

	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	if err != nil {
		return nil, nil, err
	}

	u, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, password)
	if err != nil {
		return nil, nil, err
	}

	u.Meta.Status = models.StatusActive
	err = u.Insert(context.Background(), db, key)
	if err != nil {
		return nil, nil, err
	}

	err = o.UpdateOwner(context.Background(), db, u.ID)
	if err != nil {
		return nil, nil, err
	}
	return o, u, nil
}
