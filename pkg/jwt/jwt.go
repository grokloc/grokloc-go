// Package jwt manage authorization claims for the app
package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/grokloc/grokloc-go/pkg/models/user"
)

// JWT related constants
const (
	Authorization = "Authorization"
	TokenType     = "Bearer"
	Expiration    = 86400
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
			ExpiresAt: now + int64(Expiration),
			Id:        u.ID,
			Issuer:    "grokLOC.com",
			IssuedAt:  now,
		}}
	return claims, nil
}

// ToHeaderVal prepends the JWTTokenType
func ToHeaderVal(token string) string {
	return fmt.Sprintf("%s %s", TokenType, token)
}

// FromHeaderVal will remove the JWTTokenType if it prepends the string s,
// but is also safe to use if s is just the token
func FromHeaderVal(s string) string {
	return strings.TrimPrefix(s, fmt.Sprintf("%s ", TokenType))
}

// Decode returns the claims from a signed string jwt
func Decode(id, token string, signingKey []byte) (*Claims, error) {
	// token may come in as '$TokenType $val' if from a web context

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
