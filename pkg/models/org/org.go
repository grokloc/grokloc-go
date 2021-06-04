// Package org models an organization
package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/schemas"
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
	o.Meta.Status = models.StatusUnconfirmed
	return o, nil
}

// Insert a new row.
// TODO - make sure owner is in the db and active prior to insertion
func (o *Instance) Insert(ctx context.Context, db *sql.DB) error {
	q := fmt.Sprintf("insert into %s (id,name,owner,status,schema_version) values ($1,$2,$3,$4,$5)",
		schemas.OrgsTableName)
	result, err := db.ExecContext(ctx, q, o.ID, o.Name, o.Owner, o.Meta.Status, SchemaVersion)
	if err != nil {
		if models.UniqueConstraint(err) {
			return models.ErrConflict
		}
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}
	if inserted != 1 {
		return models.ErrRowsAffected
	}
	return nil
}

// Read initializes an Instance based on a database row
func Read(ctx context.Context, db *sql.DB, id string) (*Instance, error) {
	q := fmt.Sprintf("select name,owner,ctime,mtime,status,schema_version from %s where id = $1",
		schemas.OrgsTableName)
	var statusRaw int
	o := &Instance{}
	o.ID = id
	err := db.QueryRowContext(ctx, q, id).Scan(
		&o.Name,
		&o.Owner,
		&o.Meta.Ctime,
		&o.Meta.Mtime,
		&statusRaw,
		&o.Meta.SchemaVersion)
	if err != nil {
		return nil, err
	}
	o.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}
	if o.Meta.SchemaVersion != SchemaVersion {
		// handle migrating different versions, or err
		return nil, models.ErrModelMigrate
	}
	return o, nil
}

// UpdateStatus sets the org status
func (o *Instance) UpdateStatus(ctx context.Context, db *sql.DB, status models.Status) error {
	if status == models.StatusNone {
		return errors.New("cannot use None as a stored status")
	}
	return models.Update(ctx, db, schemas.OrgsTableName, o.ID, "status", status)
}
