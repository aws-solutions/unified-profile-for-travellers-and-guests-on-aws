// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"sync"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	admmodel "tah/upt/source/ucp-common/src/model/admin"

	"github.com/aws/aws-lambda-go/events"
)

type GetJobsStatus struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetJobsStatus() *GetJobsStatus {
	return &GetJobsStatus{name: "GetJobsStatus"}
}

func (u *GetJobsStatus) Name() string {
	return u.name
}
func (u *GetJobsStatus) Tx() core.Transaction {
	return *u.tx
}
func (u *GetJobsStatus) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetJobsStatus) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetJobsStatus) Registry() *registry.Registry {
	return u.reg
}

func (u *GetJobsStatus) AccessPermission() constant.AppPermission {
	return constant.PublicAccessPermission
}

func (u *GetJobsStatus) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *GetJobsStatus) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())
	return nil
}

func (u *GetJobsStatus) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	res := model.ResponseWrapper{}
	jobNames := []string{
		u.reg.Env["HOTEL_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["AIR_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["GUEST_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["PAX_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["CLICKSTREAM_JOB_NAME_CUSTOMER"],
		u.reg.Env["HOTEL_STAY_JOB_NAME_CUSTOMER"],
		u.reg.Env["CSI_JOB_NAME_CUSTOMER"],
	}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	jobs := []admmodel.JobSummary{}
	wg.Add(len(jobNames))
	var jobStatusErr error
	for _, jobName := range jobNames {
		go func(name string) {
			lastRun, status, err := u.reg.Glue.GetJobRunStatusWithFilter(name, map[string]string{"--ACCP_DOMAIN": u.reg.Env["ACCP_DOMAIN_NAME"]})
			if err != nil {
				u.tx.Error("Error getting job run status %v => %v", name, err)
				jobStatusErr = err
			}

			mu.Lock()
			jobs = append(jobs, admmodel.JobSummary{
				JobName:     name,
				LastRunTime: lastRun,
				Status:      status,
			})
			mu.Unlock()
			wg.Done()
		}(jobName)
	}
	wg.Wait()

	res.AwsResources = model.AwsResources{
		Jobs: jobs,
	}

	return res, jobStatusErr
}

func (u *GetJobsStatus) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
