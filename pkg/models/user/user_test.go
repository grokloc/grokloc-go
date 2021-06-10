// Package user models an user
package user

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-go/pkg/env"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/security"
	"github.com/grokloc/grokloc-go/pkg/state"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	suite.Suite
	ST  *state.Instance
	Org *org.Instance
}

func (suite *UserSuite) SetupTest() {
	var err error
	suite.ST, err = state.New(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
	suite.Org, err = org.New(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	suite.Org.Meta.Status = models.StatusActive
	err = suite.Org.Insert(context.Background(), suite.ST.Master)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *UserSuite) TestInsertUser() {
	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)

	// org not there
	u, err := New(uuid.NewString(), uuid.NewString(), uuid.NewString(), password)
	require.Nil(suite.T(), err)
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Error(suite.T(), err)

	// org there but not active
	o, err := org.New(uuid.NewString())
	require.Nil(suite.T(), err)
	o.Meta.Status = models.StatusInactive
	err = o.Insert(context.Background(), suite.ST.Master)
	require.Nil(suite.T(), err)

	u, err = New(uuid.NewString(), uuid.NewString(), o.ID, password)
	require.Nil(suite.T(), err)
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Error(suite.T(), err)

	// org there and active
	u, err = New(uuid.NewString(), uuid.NewString(), suite.Org.ID, password)
	require.Nil(suite.T(), err)
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	// duplicate
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrConflict, err)
}

func (suite *UserSuite) TestReadUser() {
	// not found
	_, err := Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), suite.Org.ID, password)
	require.Nil(suite.T(), err)
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	uRead, err := Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), u.ID, uRead.ID)
	require.Equal(suite.T(), u.APISecret, uRead.APISecret)
	require.Equal(suite.T(), u.APISecretDigest, uRead.APISecretDigest)
	require.Equal(suite.T(), u.DisplayName, uRead.DisplayName)
	require.Equal(suite.T(), u.DisplayNameDigest, uRead.DisplayNameDigest)
	require.Equal(suite.T(), u.Email, uRead.Email)
	require.Equal(suite.T(), u.EmailDigest, uRead.EmailDigest)
	require.Equal(suite.T(), u.Org, uRead.Org)
	require.Equal(suite.T(), u.Password, uRead.Password)
	require.NotEqual(suite.T(), u.Meta.Ctime, uRead.Meta.Ctime)
	require.NotEqual(suite.T(), u.Meta.Mtime, uRead.Meta.Mtime)
}

func (suite *UserSuite) TestUpdateUserDisplayName() {
	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), suite.Org.ID, password)
	require.Nil(suite.T(), err)

	// not yet inserted
	err = u.UpdateDisplayName(context.Background(), suite.ST.Master, suite.ST.Key, uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	// read in current state
	uRead, err := Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), u.DisplayName, uRead.DisplayName)
	require.Equal(suite.T(), u.DisplayNameDigest, uRead.DisplayNameDigest)

	// update again
	displayName := uuid.NewString()
	err = u.UpdateDisplayName(context.Background(), suite.ST.Master, suite.ST.Key, displayName)
	require.Nil(suite.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), displayName, uRead.DisplayName)
	require.Equal(suite.T(), security.EncodedSHA256(displayName), uRead.DisplayNameDigest)
}

func (suite *UserSuite) TestUpdateUserPassword() {
	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), suite.Org.ID, password)
	require.Nil(suite.T(), err)

	// not yet inserted
	err = u.UpdatePassword(context.Background(), suite.ST.Master, uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	// read in current state
	uRead, err := Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), u.Password, uRead.Password)

	// update again
	password, err = security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	err = u.UpdatePassword(context.Background(), suite.ST.Master, password)
	require.Nil(suite.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), password, uRead.Password)
}

func (suite *UserSuite) TestUpdateUserStatus() {
	password, err := security.DerivePassword(uuid.NewString(), suite.ST.Argon2Cfg)
	require.Nil(suite.T(), err)
	u, err := New(uuid.NewString(), uuid.NewString(), suite.Org.ID, password)
	require.Nil(suite.T(), err)

	// not yet inserted
	err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	// fix that
	err = u.Insert(context.Background(), suite.ST.Master, suite.ST.Key)
	require.Nil(suite.T(), err)

	// demonstrate that the status is not active
	uRead, err := Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusUnconfirmed, uRead.Meta.Status)

	// update again
	err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	require.Nil(suite.T(), err)

	// re-read to be sure, check changed status
	uRead, err = Read(context.Background(), suite.ST.RandomReplica(), suite.ST.Key, u.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusActive, uRead.Meta.Status)

	// None not allowed
	err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusNone)
	require.Error(suite.T(), err)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
