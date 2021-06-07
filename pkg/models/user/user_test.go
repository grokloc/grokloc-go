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
	"github.com/grokloc/grokloc-go/pkg/state"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	suite.Suite
	ST *state.Instance
}

func (suite *UserSuite) SetupTest() {
	var err error
	suite.ST, err = state.NewInstanceFromLevel(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *UserSuite) TestInsertUser() {
	u, err := New(uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString())
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

	u, err := New(uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString())
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

func (suite *UserSuite) TestUpdateUserStatus() {
	// o, err := New(uuid.NewString())
	// require.Nil(suite.T(), err)

	// // not yet inserted
	// err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	// require.Error(suite.T(), err)
	// require.Equal(suite.T(), sql.ErrNoRows, err)

	// // fix that
	// err = u.Insert(context.Background(), suite.ST.Master)
	// require.Nil(suite.T(), err)

	// // demonstrate that the status is not active
	// oRead, err := Read(context.Background(), suite.ST.RandomReplica(), o.ID)
	// require.Nil(suite.T(), err)
	// require.Equal(suite.T(), models.StatusUnconfirmed, oRead.Meta.Status)

	// // update again
	// err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	// require.Nil(suite.T(), err)

	// // re-read to be sure, check changed status
	// oRead, err = Read(context.Background(), suite.ST.RandomReplica(), o.ID)
	// require.Nil(suite.T(), err)
	// require.Equal(suite.T(), models.StatusActive, oRead.Meta.Status)

	// // None not allowed
	// err = u.UpdateStatus(context.Background(), suite.ST.Master, models.StatusNone)
	// require.Error(suite.T(), err)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
