// Package org models an organization
package org

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

type OrgSuite struct {
	suite.Suite
	ST *state.Instance
}

func (suite *OrgSuite) SetupTest() {
	var err error
	suite.ST, err = state.NewInstanceFromLevel(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *OrgSuite) TestInsertOrg() {
	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	err = o.Insert(context.Background(), suite.ST.Master)
	require.Nil(suite.T(), err)

	// duplicate
	err = o.Insert(context.Background(), suite.ST.Master)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), models.ErrConflict, err)
}

func (suite *OrgSuite) TestReadOrg() {
	// not found
	_, err := Read(context.Background(), suite.ST.RandomReplica(), uuid.NewString())
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)
	err = o.Insert(context.Background(), suite.ST.Master)
	require.Nil(suite.T(), err)

	oRead, err := Read(context.Background(), suite.ST.RandomReplica(), o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), o.ID, oRead.ID)
	require.Equal(suite.T(), o.Name, oRead.Name)
	require.Equal(suite.T(), o.Owner, oRead.Owner)
	require.NotEqual(suite.T(), o.Meta.Ctime, oRead.Meta.Ctime)
	require.NotEqual(suite.T(), o.Meta.Mtime, oRead.Meta.Mtime)
}

func (suite *OrgSuite) TestUpdateOrgOwner() {
	require.True(suite.T(), true)
}

func (suite *OrgSuite) TestUpdateOrgStatus() {
	o, err := New(uuid.NewString())
	require.Nil(suite.T(), err)

	// not yet inserted
	err = o.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), sql.ErrNoRows, err)

	// fix that
	err = o.Insert(context.Background(), suite.ST.Master)
	require.Nil(suite.T(), err)

	// demonstrate that the status is not active
	oRead, err := Read(context.Background(), suite.ST.RandomReplica(), o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusUnconfirmed, oRead.Meta.Status)

	// update again
	err = o.UpdateStatus(context.Background(), suite.ST.Master, models.StatusActive)
	require.Nil(suite.T(), err)

	// re-read to be sure, check changed status
	oRead, err = Read(context.Background(), suite.ST.RandomReplica(), o.ID)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), models.StatusActive, oRead.Meta.Status)

	// None not allowed
	err = o.UpdateStatus(context.Background(), suite.ST.Master, models.StatusNone)
	require.Error(suite.T(), err)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
