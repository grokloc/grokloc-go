module github.com/grokloc/grokloc-go

go 1.16

replace (
	github.com/grokloc/grokloc-go/pkg/env => ./pkg/env
	github.com/grokloc/grokloc-go/pkg/jwt => ./pkg/jwt
	github.com/grokloc/grokloc-go/pkg/models => ./pkg/models
	github.com/grokloc/grokloc-go/pkg/models/org => ./pkg/models/org
	github.com/grokloc/grokloc-go/pkg/models/user => ./pkg/models/user
	github.com/grokloc/grokloc-go/pkg/schemas => ./pkg/schemas
	github.com/grokloc/grokloc-go/pkg/security => ./pkg/security
	github.com/grokloc/grokloc-go/pkg/state => ./pkg/state
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.2.0
	github.com/matthewhartstonge/argon2 v0.1.4
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
)
