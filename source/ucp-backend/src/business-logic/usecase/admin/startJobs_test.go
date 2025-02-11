// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	common "tah/upt/source/ucp-common/src/constant/admin"

	"github.com/aws/aws-lambda-go/events"
)

func TestStartJobs(t *testing.T) {
	t.Parallel()
	log.Printf("TestStartJobs: initialize test resources")
	lambdaCfg := lambda.LambdaMock{}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{SyncLambda: &lambdaCfg, Env: map[string]string{"ACCP_DOMAIN_NAME": "test_domain"}}
	uc := NewStartJobs()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}
	uc.reg.SetAppAccessPermission(common.RunGlueJobsPermission)
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "StartJobs", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "StartJobs", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "StartJobs", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	log.Printf("TestStartJobs: deleting test resources")
}

func TestValidateRequestAndBuildCloudwatchEvent(t *testing.T) {
	log.Printf("TestBuildCloudwatchEvent: initialize test resources")
	lambdaCfg := lambda.LambdaMock{}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{SyncLambda: &lambdaCfg, Env: map[string]string{
		"ACCP_DOMAIN_NAME":                "test_domain_in_context",
		"HOTEL_BOOKING_JOB_NAME_CUSTOMER": "validHotelBookingJobName",
		"AIR_BOOKING_JOB_NAME_CUSTOMER":   "airBookingJobName",
		"GUEST_PROFILE_JOB_NAME_CUSTOMER": "guestProfileJobName",
		"PAX_PROFILE_JOB_NAME_CUSTOMER":   "paxJobName",
		"CLICKSTREAM_JOB_NAME_CUSTOMER":   "clickstreamJobName",
		"HOTEL_STAY_JOB_NAME_CUSTOMER":    "hotelStayJobName",
		"CSI_JOB_NAME_CUSTOMER":           "csiJobName",
		"RUN_ALL_JOBS":                    "run-all-jobs",
	}}
	regNoDomain := registry.Registry{SyncLambda: &lambdaCfg, Env: map[string]string{}}
	uc := NewStartJobs()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&regNoDomain)

	// Initialize Requests
	rq := events.APIGatewayProxyRequest{
		Body: `{"startJobRq":{}}`,
	}
	rqWithValidJob := events.APIGatewayProxyRequest{
		Body: `{"startJobRq":{"jobName": "validHotelBookingJobName"}}`,
	}
	rqWithDomainValidJob := events.APIGatewayProxyRequest{
		Body: `{"startJobRq":{"domainName": "testDomain_in_rq", "jobName": "validHotelBookingJobName"}}`,
	}
	rqWithInvalidJob := events.APIGatewayProxyRequest{
		Body: `{"startJobRq":{"jobName": "InvalidJobName"}}`,
	}
	rqWithValidRunAllJobs := events.APIGatewayProxyRequest{
		Body: `{"startJobRq":{"jobName": "run-all-jobs"}}`,
	}

	// Wrap all Requests
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}
	wrapperWithValidJob, err0 := uc.CreateRequest(rqWithValidJob)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}
	wrapperWithDomainAndValidJob, err0 := uc.CreateRequest(rqWithDomainValidJob)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}
	wrapperWithInvalidJob, err0 := uc.CreateRequest(rqWithInvalidJob)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}
	wrapperWithValidRunAllJobs, err0 := uc.CreateRequest(rqWithValidRunAllJobs)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartJobs", err0)
	}

	// Test build cloudwatch event with no domain in context
	cwe, err := uc.buildCloudwatchEvent(wrapperWithValidJob)
	if err == nil {
		t.Errorf("Building cloudwatch event with no domain in context should return and error")
	}

	// Set domain in context
	log.Printf("Setting domain in context")
	uc.SetRegistry(&reg)

	// Validate All Requests
	uc.reg.SetAppAccessPermission(common.RunGlueJobsPermission)
	err = uc.ValidateRequest(wrapper)
	if err == nil {
		t.Errorf("Request with no job name should error")
	}
	err = uc.ValidateRequest(wrapperWithValidJob)
	if err != nil {
		t.Errorf("Request with valid job name should succeed: %v", err)
	}
	err = uc.ValidateRequest(wrapperWithDomainAndValidJob)
	if err != nil {
		t.Errorf("Request with valid job name should succeed: %v", err)
	}
	err = uc.ValidateRequest(wrapperWithInvalidJob)
	if err == nil {
		t.Errorf("Request with invalid job should error")
	}
	err = uc.ValidateRequest(wrapperWithValidRunAllJobs)
	if err != nil {
		t.Errorf("Request with valid run-all-jobs job name should succeed: %v", err)
	}

	// Test build cloudwatch events
	cwe, err = uc.buildCloudwatchEvent(wrapperWithValidJob)
	if err != nil {
		t.Errorf("Building cloudwatch event returned unexpected error %v", err)
	}
	if cwe.DetailType != common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
		t.Errorf("Cloudwatch event has unexpected detail type %v. Should be %v", cwe.DetailType, common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER)
	}
	if len(cwe.Resources) != 2 {
		t.Errorf("Cloudwatch event has unexpected number of resources %v. Should be 2", len(cwe.Resources))
		return
	}
	if cwe.Resources[0] != "domain:test_domain_in_context" {
		t.Errorf("Cloudwatch event has unexpected resource %v. Should be domain:testDomain_in_context", cwe.Resources[0])
	}
	if cwe.Resources[1] != "job:validHotelBookingJobName" {
		t.Errorf("Cloudwatch event has unexpected resource %v. Should be job:validHotelBookingJobName", cwe.Resources[1])
	}

	cwe, err = uc.buildCloudwatchEvent(wrapperWithDomainAndValidJob)
	if err != nil {
		t.Errorf("Building cloudwatch event returned unexpected error %v", err)
	}
	if len(cwe.Resources) != 2 {
		t.Errorf("Cloudwatch event has unexpected number of resources %v. Should be 2", len(cwe.Resources))
		return
	}
	if cwe.Resources[0] != "domain:testDomain_in_rq" {
		t.Errorf("Cloudwatch event has unexpected resource %v. Should be domain:testDomain_in_rq", cwe.Resources[0])
	}
	if cwe.Resources[1] != "job:validHotelBookingJobName" {
		t.Errorf("Cloudwatch event has unexpected resource %v. Should be job:validHotelBookingJobName", cwe.Resources[1])
	}

	cwe, err = uc.buildCloudwatchEvent(wrapperWithValidRunAllJobs)
	if err != nil {
		t.Errorf("Building cloudwatch event returned unexpected error %v", err)
	}
	if cwe.DetailType != common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
		t.Errorf("Cloudwatch event has unexpected detail type %v. Should be %v", cwe.DetailType, common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER)
	}
	if len(cwe.Resources) != 1 {
		t.Errorf("Cloudwatch event has unexpected number of resources %v. Should be 1", len(cwe.Resources))
		return
	}
	if cwe.Resources[0] != "domain:test_domain_in_context" {
		t.Errorf("Cloudwatch event has unexpected resource %v. Should be domain:testDomain_in_context", cwe.Resources[0])
	}
}
