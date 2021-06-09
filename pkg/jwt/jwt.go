// Package jwt manage authorization claims for the app
package jwt

import (
	"errors"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/grokloc/grokloc-go/pkg/models/user"
)

// Claims are the JWT claims for the app
type Claims struct {
	Scope string `json:"scope"`
	Org   string `json:"org"`
	jwt_go.StandardClaims
}

// New returns a new Claims instance
func New(u user.Instance) (*Claims, error) {
	now := time.Now().Unix()
	claims := &Claims{
		"app",
		u.Org,
		jwt_go.StandardClaims{
			Audience:  u.EmailDigest,
			ExpiresAt: now + (int64(30) * int64(86400)),
			Id:        u.ID,
			Issuer:    "grokLOC.com",
			IssuedAt:  now,
		}}
	return claims, nil
}

// Decode returns the claims from a signed string jwt
func Decode(id, token string, signingKey []byte) (*Claims, error) {
	f := func(token *jwt_go.Token) (interface{}, error) {
		return []byte(id + string(signingKey)), nil
	}
	parsed, err := jwt_go.ParseWithClaims(token, &Claims{}, f)
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if ok {
		return claims, nil
	}
	return nil, errors.New("Token claims")
}
