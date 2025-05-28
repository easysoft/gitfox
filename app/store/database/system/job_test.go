// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package system_test

import (
	"context"
	"testing"
	"time"

	"github.com/easysoft/gitfox/app/store/database"
	"github.com/easysoft/gitfox/app/store/database/system"
	"github.com/easysoft/gitfox/app/store/database/testsuite"
	"github.com/easysoft/gitfox/job"
	"github.com/easysoft/gitfox/store"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTableJob = "jobs"
)

type JobSuite struct {
	testsuite.BaseSuite

	ormStore  *system.JobStore
	sqlxStore *database.JobStore
}

func TestJobSuite(t *testing.T) {
	ctx := context.Background()

	st := &JobSuite{
		BaseSuite: testsuite.BaseSuite{
			Ctx:  ctx,
			Name: "jobs",
		},
	}

	st.BaseSuite.Constructor = func(ts *testsuite.TestStore) {
		st.ormStore = system.NewJobOrmStore(st.Gdb)
		st.sqlxStore = database.NewJobStore(st.Sdb)

		// add init data
		testsuite.AddUser(st.Ctx, t, ts.Principal, 1, true)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 2, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 3, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 4, false)
		testsuite.AddUser(st.Ctx, t, ts.Principal, 5, true)

		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 1, 0)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 2, 1)
		testsuite.AddSpace(st.Ctx, t, ts.Space, ts.SpacePath, 1, 3, 1)
	}

	suite.Run(t, st)
}

func (suite *JobSuite) SetupTest() {
	suite.addData()
}

func (suite *JobSuite) TearDownTest() {
	suite.Gdb.WithContext(suite.Ctx).Table(testTableJob).Where("1 = 1").Delete(nil)
}

const (
	type1 = "type1"
	type2 = "type2"
	type3 = "type3"
)

var testAddJobs = []job.Job{
	{UID: "job_1", GroupID: "g1", Type: type1},
	{UID: "job_2", GroupID: "g1", Type: type3},
	{UID: "job_3", GroupID: "g2", Type: type1},
	{UID: "job_4", GroupID: "g3", Type: type2},
	{UID: "job_5", GroupID: "g3", Type: type1},
}

func (suite *JobSuite) addData() {
	now := time.Now().UnixMilli()
	for id, item := range testAddJobs {
		item.Created = now
		item.Updated = now
		err := suite.ormStore.Create(suite.Ctx, &item)
		require.NoError(suite.T(), err, testsuite.InvalidLoopMsgF, id)
	}
}

type updateJob struct {
	obj   *job.Job
	field string
}

func (suite *JobSuite) TestUpsert() {
	uid := "job_100"
	//now := time.Now().UnixMilli() - 1

	_, err := suite.ormStore.Find(suite.Ctx, uid)
	require.ErrorIs(suite.T(), err, store.ErrResourceNotFound)

	// create
	newObj := &job.Job{UID: uid, GroupID: "g3", Type: type2}
	err = suite.ormStore.Upsert(suite.Ctx, newObj)
	require.NoError(suite.T(), err)

	currObj, err := suite.ormStore.Find(suite.Ctx, uid)
	require.NoError(suite.T(), err)

	// nothing updates
	noUp := newNoChangeObj(currObj)
	err = suite.sqlxStore.Upsert(suite.Ctx, noUp)
	require.NoError(suite.T(), err)
	suite.equalObj(uid, currObj)

	err = suite.ormStore.Upsert(suite.Ctx, noUp)
	require.NoError(suite.T(), err)
	suite.equalObj(uid, currObj)

	// require update
	reqUpdates := make([]updateJob, 0)
	for range make([]int, 7) {
		n := *currObj
		reqUpdates = append(reqUpdates, updateJob{obj: &n})
	}
	reqUpdates[0].obj.Type = type1
	reqUpdates[0].field = "Type"
	reqUpdates[1].obj.Priority = 99
	reqUpdates[1].field = "Priority"
	reqUpdates[2].obj.Data = "data11"
	reqUpdates[2].field = "Data"
	reqUpdates[3].obj.MaxDurationSeconds = 1801
	reqUpdates[3].field = "MaxDurationSeconds"
	reqUpdates[4].obj.MaxRetries = 9
	reqUpdates[4].field = "MaxRetries"
	reqUpdates[5].obj.IsRecurring = true
	reqUpdates[5].field = "IsRecurring"
	reqUpdates[6].obj.RecurringCron = "*/1"
	reqUpdates[6].field = "RecurringCron"

	lastObj := currObj
	// do updates
	for id, up := range reqUpdates {
		var s job.Store = suite.ormStore
		if id%2 == 0 {
			s = suite.sqlxStore
		}

		err = s.Upsert(suite.Ctx, up.obj)
		require.NoError(suite.T(), err)

		obj, err := suite.ormStore.Find(suite.Ctx, uid)
		require.NoError(suite.T(), err)

		testsuite.NotEqualStructFieldValue(suite.T(), lastObj, obj, up.field, testsuite.InvalidLoopMsgF, id)
		lastObj = obj
	}
}

func (suite *JobSuite) equalObj(uid string, target *job.Job) {
	obj, err := suite.ormStore.Find(suite.Ctx, uid)
	require.NoError(suite.T(), err)
	require.EqualExportedValues(suite.T(), *obj, *target)
}

func newNoChangeObj(j *job.Job) *job.Job {
	now := time.Now().UnixMilli()

	n := *j
	n.Created = now
	n.Updated = now
	n.State = job.JobStateFailed
	n.Scheduled = now
	n.TotalExecutions = 10
	n.RunBy = "u1"
	n.RunDeadline = now + 3600*1000
	n.RunProgress = 90
	n.LastExecuted = now - 60*1000
	n.ConsecutiveFailures = 3
	n.LastFailureError = "err1"
	n.GroupID = "g110"
	return &n
}
