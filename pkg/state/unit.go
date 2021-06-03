// Package state manages all external state conns
package state

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" //

	"github.com/grokloc/grokloc-go/pkg/schemas"
)

func unitInstance() Instance {
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
	return Instance{}
}
