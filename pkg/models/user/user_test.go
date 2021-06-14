package user

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/schemas"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/matthewhartstonge/argon2"
	_ "github.com/mattn/go-sqlite3" //
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserSuite cannot user a state unit instance as it will create
// an import cycle, so the relevant fields are just instantiated
// directly
type UserSuite struct {
	suite.Suite
	DB  *sql.DB
	Key []byte
	Org *org.Instance
}

func (s *UserSuite) SetupTest() {
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
	s.Org, err = org.New(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	s.Org.Meta.Status = models.StatusActive
	err = s.Org.Insert(context.Background(), s.DB)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserSuite) TestInsertUser() {
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)

	// org not there
	u, err := New(uuid.NewString(), uuid.NewString(), uuid.NewString(), password)
	require.Nil(s.T(), err)
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Error(s.T(), err)

	// org there but not active
	o, err := org.New(uuid.NewString())
	require.Nil(s.T(), err)
	o.Meta.Status = models.StatusInactive
	err = o.Insert(context.Background(), s.DB)
	require.Nil(s.T(), err)

	u, err = New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(s.T(), err)
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Error(s.T(), err)

	// org there and active
	u, err = New(uuid.NewString(), uuid.NewString(), s.Org.ID, password)
	require.Nil(s.T(), err)
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)

	// duplicate
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Error(s.T(), err)
	require.Equal(s.T(), models.ErrConflict, err)
}

func (s *UserSuite) TestReadUser() {
	// not found
	_, err := Read(context.Background(), s.DB, s.Key, uuid.NewString())
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), s.Org.ID, password)
	require.Nil(s.T(), err)
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)

	uRead, err := Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.ID, uRead.ID)
	require.Equal(s.T(), u.APISecret, uRead.APISecret)
	require.Equal(s.T(), u.APISecretDigest, uRead.APISecretDigest)
	require.Equal(s.T(), u.DisplayName, uRead.DisplayName)
	require.Equal(s.T(), u.DisplayNameDigest, uRead.DisplayNameDigest)
	require.Equal(s.T(), u.Email, uRead.Email)
	require.Equal(s.T(), u.EmailDigest, uRead.EmailDigest)
	require.Equal(s.T(), u.Org, uRead.Org)
	require.Equal(s.T(), u.Password, uRead.Password)
	require.NotEqual(s.T(), u.Meta.Ctime, uRead.Meta.Ctime)
	require.NotEqual(s.T(), u.Meta.Mtime, uRead.Meta.Mtime)
}

func (s *UserSuite) TestUpdateUserDisplayName() {
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), s.Org.ID, password)
	require.Nil(s.T(), err)

	// not yet inserted
	err = u.UpdateDisplayName(context.Background(), s.DB, s.Key, uuid.NewString())
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)

	// read in current state
	uRead, err := Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.DisplayName, uRead.DisplayName)
	require.Equal(s.T(), u.DisplayNameDigest, uRead.DisplayNameDigest)

	// update again
	displayName := uuid.NewString()
	err = u.UpdateDisplayName(context.Background(), s.DB, s.Key, displayName)
	require.Nil(s.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), displayName, uRead.DisplayName)
	require.Equal(s.T(), security.EncodedSHA256(displayName), uRead.DisplayNameDigest)
}

func (s *UserSuite) TestUpdateUserPassword() {
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), s.Org.ID, password)
	require.Nil(s.T(), err)

	// not yet inserted
	err = u.UpdatePassword(context.Background(), s.DB, uuid.NewString())
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)

	// read in current state
	uRead, err := Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.Password, uRead.Password)

	// update again
	password, err = security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	err = u.UpdatePassword(context.Background(), s.DB, password)
	require.Nil(s.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), password, uRead.Password)
}

func (s *UserSuite) TestUpdateUserStatus() {
	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), s.Org.ID, password)
	require.Nil(s.T(), err)

	// not yet inserted
	err = u.UpdateStatus(context.Background(), s.DB, models.StatusActive)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), s.DB, s.Key)
	require.Nil(s.T(), err)

	// demonstrate that the status is not active
	uRead, err := Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusUnconfirmed, uRead.Meta.Status)

	// update again
	err = u.UpdateStatus(context.Background(), s.DB, models.StatusActive)
	require.Nil(s.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), s.DB, s.Key, u.ID)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusActive, uRead.Meta.Status)

	// None not allowed
	err = u.UpdateStatus(context.Background(), s.DB, models.StatusNone)
	require.Error(s.T(), err)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
