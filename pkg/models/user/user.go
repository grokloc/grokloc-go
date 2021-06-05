// Package user models a user
package user

import (
	"github.com/grokloc/grokloc-go/pkg/models"
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
