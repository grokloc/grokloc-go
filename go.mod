module github.com/grokloc/grokloc-go

go 1.16

replace (
	github.com/grokloc/grokloc-go/pkg/env => ./pkg/env
	github.com/grokloc/grokloc-go/pkg/models => ./pkg/models
	github.com/grokloc/grokloc-go/pkg/models/org => ./pkg/models/org
	github.com/grokloc/grokloc-go/pkg/schemas => ./pkg/schemas
	github.com/grokloc/grokloc-go/pkg/security => ./pkg/security
)

require (
	github.com/google/uuid v1.2.0
	github.com/matthewhartstonge/argon2 v0.1.4
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/stretchr/testify v1.7.0
)
