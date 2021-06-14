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

func (s *OrgSuite) SetupTest() {
	var err error
	s.DB, err = sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	// avoid concurrency bug with the sqlite library
	s.DB.SetMaxOpenConns(1)
	_, err = s.DB.Exec(schemas.AppCreate)
	if err != nil {
		log.Fatal(err)
	}
	s.Key, err = security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
}

func (s *OrgSuite) TestInsertOrg() {
	o, err := New(uuid.NewString())
	require.Nil(s.T(), err)
	err = o.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)

	// duplicate
	err = o.Insert(context.Background(), s.DB)
	require.Error(s.T(), err)
	require.Equal(s.T(), models.ErrConflict, err)
}

func (s *OrgSuite) TestReadOrg() {
	// not found
	_, err := Read(context.Background(), s.DB, uuid.NewString())
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	o, err := New(uuid.NewString())
	require.Nil(s.T(), err)
	err = o.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)

	oRead, err := Read(context.Background(), s.DB, o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.ID, oRead.ID)
	require.Equal(s.T(), o.Name, oRead.Name)
	require.Equal(s.T(), o.Owner, oRead.Owner)
	require.NotEqual(s.T(), o.Meta.Ctime, oRead.Meta.Ctime)
	require.NotEqual(s.T(), o.Meta.Mtime, oRead.Meta.Mtime)
}

func (s *OrgSuite) TestUpdateOrgOwner() {
	o, err := New(uuid.NewString())
	require.Nil(s.T(), err)
	o.Meta.Status = models.StatusActive
	err = o.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)

	// try setting owner to an id not in the db
	err = o.UpdateOwner(context.Background(), s.DB, uuid.NewString())
	require.Error(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)

	// new owner in db but not active
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := user.New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(s.T(), err)
	u.Meta.Status = models.StatusInactive
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)
	err = o.UpdateOwner(context.Background(), s.DB, u.ID)
	require.Error(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)

	// user active but in different org
	oOther, err := New(uuid.NewString())
	require.Nil(s.T(), err)
	oOther.Meta.Status = models.StatusActive
	err = oOther.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)
	password, err = security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	uOther, err := user.New(uuid.NewString(), uuid.NewString(), oOther.ID, password)
	require.Nil(s.T(), err)
	uOther.Meta.Status = models.StatusActive
	err = uOther.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)
	err = o.UpdateOwner(context.Background(), s.DB, uOther.ID)
	require.Error(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)
}

func (s *OrgSuite) TestUpdateOrgStatus() {
	o, err := New(uuid.NewString())
	require.Nil(s.T(), err)

	// not yet inserted
	err = o.UpdateStatus(context.Background(), s.DB, models.StatusActive)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	// fix that
	err = o.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)

	// demonstrate that the status is not active
	oRead, err := Read(context.Background(), s.DB, o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusUnconfirmed, oRead.Meta.Status)

	// update again
	err = o.UpdateStatus(context.Background(), s.DB, models.StatusActive)
	require.Nil(s.T(), err)

	// re-read to be sure, check changed status
	oRead, err = Read(context.Background(), s.DB, o.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusActive, oRead.Meta.Status)

	// None not allowed
	err = o.UpdateStatus(context.Background(), s.DB, models.StatusNone)
	require.Error(s.T(), err)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
