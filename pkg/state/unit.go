// Package state manages all external state conns
package state

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/matthewhartstonge/argon2"
	_ "github.com/mattn/go-sqlite3" //
	"go.uber.org/zap"

	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/schemas"
	"github.com/grokloc/grokloc-go/pkg/security"
)

// unitInstance builds an instance for the Unit environment
func unitInstance() *Instance {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	// avoid concurrency bug with the sqlite library
	db.SetMaxOpenConns(1)
	_, err = db.Exec(schemas.AppCreate)
	if err != nil {
		log.Fatal(err)
	}
	key, err := security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	signingKey, err := security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	return &Instance{
		Level:      env.Unit,
		Master:     db,
		Replicas:   []*sql.DB{db},
		Key:        key,
		SigningKey: signingKey,
		Argon2Cfg:  argon2.DefaultConfig(),
		L:          logger,
	}
}
