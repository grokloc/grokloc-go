// Package user models a user
package user

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

// exported user symbols
const (
	SchemaVersion = 0
)

// Instance is a user model
type Instance struct {
	models.Base
	APISecret         string `json:"api_secret"`
	APISecretDigest   string `json:"api_secret_digest"`
	DisplayName       string `json:"display_name"`
	DisplayNameDigest string `json:"display_name_digest"`
	Email             string `json:"email"`
	EmailDigest       string `json:"email_digest"`
	Org               string `json:"org"`
	Password          string `json:"password"`
}

// New creates a new user that hasn't been created before
func New(displayName, email, org, password string) (*Instance, error) {
	for _, v := range []string{displayName, email, org, password} {
		if !security.SafeStr(v) {
			return nil, errors.New("malformed user arg")
		}
	}
	u := &Instance{Org: org, Password: password}
	u.ID = uuid.NewString()
	u.Meta.SchemaVersion = SchemaVersion
	u.Meta.Status = models.StatusUnconfirmed

	u.APISecret = uuid.NewString()
	u.APISecretDigest = security.EncodedSHA256(u.APISecret)
	u.DisplayName = displayName
	u.DisplayNameDigest = security.EncodedSHA256(u.DisplayName)
	u.Email = email
	u.EmailDigest = security.EncodedSHA256(u.Email)

	return u, nil
}

// Insert a new row.
// TODO - make sure org exists and is active
func (u *Instance) Insert(ctx context.Context, db *sql.DB, key []byte) error {
	encryptedAPISecret, err := security.Encrypt(u.APISecret, key)
	if err != nil {
		return err
	}
	encryptedDisplayName, err := security.Encrypt(u.DisplayName, key)
	if err != nil {
		return err
	}
	encryptedEmail, err := security.Encrypt(u.Email, key)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("insert into %s (id,api_secret,api_secret_digest,display_name,display_name_digest,email,email_digest,org,password,status,schema_version) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		schemas.UsersTableName)
	result, err := db.ExecContext(ctx,
		q,
		u.ID,
		encryptedAPISecret,
		u.APISecretDigest,
		encryptedDisplayName,
		u.DisplayNameDigest,
		encryptedEmail,
		u.EmailDigest,
		u.Org,
		u.Password,
		u.Meta.Status,
		SchemaVersion)
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
func Read(ctx context.Context, db *sql.DB, key []byte, id string) (*Instance, error) {
	q := fmt.Sprintf("select api_secret,api_secret_digest,display_name,display_name_digest,email,email_digest,org,password,ctime,mtime,status,schema_version from %s where id = $1",
		schemas.UsersTableName)
	var statusRaw int
	u := &Instance{}
	u.ID = id
	var encryptedAPISecret, encryptedDisplayName, encryptedEmail string
	err := db.QueryRowContext(ctx, q, id).Scan(
		&encryptedAPISecret,
		&u.APISecretDigest,
		&encryptedDisplayName,
		&u.DisplayNameDigest,
		&encryptedEmail,
		&u.EmailDigest,
		&u.Org,
		&u.Password,
		&u.Meta.Ctime,
		&u.Meta.Mtime,
		&statusRaw,
		&u.Meta.SchemaVersion)
	if err != nil {
		return nil, err
	}
	u.APISecret, err = security.Decrypt(encryptedAPISecret, key)
	if err != nil {
		return nil, err
	}
	u.DisplayName, err = security.Decrypt(encryptedDisplayName, key)
	if err != nil {
		return nil, err
	}
	u.Email, err = security.Decrypt(encryptedEmail, key)
	if err != nil {
		return nil, err
	}
	u.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}
	if u.Meta.SchemaVersion != SchemaVersion {
		// handle migrating different versions, or err
		return nil, models.ErrModelMigrate
	}
	return u, nil
}

// UpdateStatus sets the user status
func (u *Instance) UpdateStatus(ctx context.Context, db *sql.DB, status models.Status) error {
	if status == models.StatusNone {
		return errors.New("cannot use None as a stored status")
	}
	return models.Update(ctx, db, schemas.UsersTableName, u.ID, "status", status)
}
