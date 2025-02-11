// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"fmt"

	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	common "tah/upt/source/ucp-common/src/constant/admin"
	constant "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

type StartJobs struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewStartJobs() *StartJobs {
	return &StartJobs{name: "StartJobs"}
}

func (u *StartJobs) Name() string {
	return u.name
}
func (u *StartJobs) Tx() core.Transaction {
	return *u.tx
}
func (u *StartJobs) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *StartJobs) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *StartJobs) Registry() *registry.Registry {
	return u.reg
}

func (u *StartJobs) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *StartJobs) AccessPermission() constant.AppPermission {
	return constant.RunGlueJobsPermission
}

func (u *StartJobs) ValidateRequest(req model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	jobName := req.StartJobRq.JobName
	return u.validateJobName(jobName)
}

func (u *StartJobs) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("Manually Starting glue jobs")
	cwe, err := u.buildCloudwatchEvent(req)
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	_, err = u.reg.SyncLambda.InvokeAsync(cwe)
	return model.ResponseWrapper{}, err
}

func (u *StartJobs) buildCloudwatchEvent(req model.RequestWrapper) (events.CloudWatchEvent, error) {
	domainName := u.reg.Env["ACCP_DOMAIN_NAME"]
	if domainName == "" {
		return events.CloudWatchEvent{}, errors.New("No Domain found in context. Cannot start jobs")
	}
	if req.StartJobRq.DomainName != "" {
		domainName = req.StartJobRq.DomainName
		u.tx.Debug("DomainName found in request, using domain from request: %s", domainName)
	} else {
		u.tx.Debug("DomainName not found in request, using domain from environment: %s", domainName)
	}
	cwe := events.CloudWatchEvent{
		DetailType: common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER,
		Resources:  []string{"domain:" + domainName},
	}
	jobName := req.StartJobRq.JobName
	// Usecase logic under ucp-sync/src/business-logic/usecase/maintainGluePartitions
	// cwe.Resources must have no job field to run all jobs
	if jobName != u.reg.Env["RUN_ALL_JOBS"] {
		u.tx.Info("JobName found in request, using jobName from request: %s", jobName)
		cwe.Resources = append(cwe.Resources, "job:"+jobName)
	} else {
		u.tx.Info("JobName run-all-jobs requested")
	}
	return cwe, nil
}

func (u *StartJobs) validateJobName(jobName string) error {
	validJobNames := []string{
		u.reg.Env["HOTEL_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["AIR_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["GUEST_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["PAX_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["CLICKSTREAM_JOB_NAME_CUSTOMER"],
		u.reg.Env["HOTEL_STAY_JOB_NAME_CUSTOMER"],
		u.reg.Env["CSI_JOB_NAME_CUSTOMER"],
		u.reg.Env["RUN_ALL_JOBS"],
	}
	for _, v := range validJobNames {
		if jobName == v {
			return nil
		}
	}
	return fmt.Errorf("invalid job name: %s, should be within %v", jobName, validJobNames)
}

func (u *StartJobs) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
