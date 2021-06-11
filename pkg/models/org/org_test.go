package org

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/schemas"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/matthewhartstonge/argon2"
	_ "github.com/mattn/go-sqlite3" //
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OrgSuite cannot user a state unit instance as it will create
// an import cycle, so the relevant fields are just instantiated
// directly
type OrgSuite struct {
	suite.Suite
	DB  *sql.DB
	Key []byte
}

func (suite *OrgSuite) SetupTest() {
	var err error
	suite.DB, err = sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	// avoid concurrency bug with the sqlite library
	suite.DB.SetMaxOpenConns(1)
	_, err = suite.DB.Exec(schemas.AppCreate)
	if err != nil {
		log.Fatal(err)
	}
	suite.Key, err = security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *OrgSuite) TestInsertOrg() {
	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	err = o.Insert(context.Background(), suite.DB)
	require.Nil(suite.T(), err)

	// duplicate
	err = o.Insert(context.Background(), suite.DB)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrConflict, err)
}

func (suite *OrgSuite) TestReadOrg() {
	// not found
	_, err := Read(context.Background(), suite.DB, uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	err = o.Insert(context.Background(), suite.DB)
	require.Nil(suite.T(), err)

	oRead, err := Read(context.Background(), suite.DB, o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), o.ID, oRead.ID)
	require.Equal(suite.T(), o.Name, oRead.Name)
	require.Equal(suite.T(), o.Owner, oRead.Owner)
	require.NotEqual(suite.T(), o.Meta.Ctime, oRead.Meta.Ctime)
	require.NotEqual(suite.T(), o.Meta.Mtime, oRead.Meta.Mtime)
}

func (suite *OrgSuite) TestUpdateOrgOwner() {
	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	o.Meta.Status = models.StatusActive
	err = o.Insert(context.Background(), suite.DB)
	require.Nil(suite.T(), err)

	// try setting owner to an id not in the db
	err = o.UpdateOwner(context.Background(), suite.DB, uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrRelatedUser, err)

	// new owner in db but not active
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(suite.T(), err)
	u, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(suite.T(), err)
	u.Meta.Status = models.StatusInactive
	err = u.Insert(context.Background(), suite.DB, suite.Key)
	require.Nil(suite.T(), err)
	err = o.UpdateOwner(context.Background(), suite.DB, u.ID)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrRelatedUser, err)

	// user active but in different org
	oOther, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	oOther.Meta.Status = models.StatusActive
	err = oOther.Insert(context.Background(), suite.DB)
	require.Nil(suite.T(), err)
	password, err = security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(suite.T(), err)
	uOther, err := user.New(uuid.NewString(), uuid.NewString(), oOther.ID, password)
	require.Nil(suite.T(), err)
	uOther.Meta.Status = models.StatusActive
	err = uOther.Insert(context.Background(), suite.DB, suite.Key)
	require.Nil(suite.T(), err)
	err = o.UpdateOwner(context.Background(), suite.DB, uOther.ID)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrRelatedUser, err)
}

func (suite *OrgSuite) TestUpdateOrgStatus() {
	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)

	// not yet inserted
	err = o.UpdateStatus(context.Background(), suite.DB, models.StatusActive)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	// fix that
	err = o.Insert(context.Background(), suite.DB)
	require.Nil(suite.T(), err)

	// demonstrate that the status is not active
	oRead, err := Read(context.Background(), suite.DB, o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusUnconfirmed, oRead.Meta.Status)

	// update again
	err = o.UpdateStatus(context.Background(), suite.DB, models.StatusActive)
	require.Nil(suite.T(), err)

	// re-read to be sure, check changed status
	oRead, err = Read(context.Background(), suite.DB, o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusActive, oRead.Meta.Status)

	// None not allowed
	err = o.UpdateStatus(context.Background(), suite.DB, models.StatusNone)
	require.Error(suite.T(), err)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
