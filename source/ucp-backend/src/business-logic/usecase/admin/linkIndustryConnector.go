// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"errors"
	"fmt"

	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/datasync"
	"tah/upt/source/tah-core/db"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	services "tah/upt/source/ucp-common/src/services/admin"

	"github.com/aws/aws-lambda-go/events"
)

type LinkIndustryConnector struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

type TransferObject struct {
	DomainName              string
	ObjectName              string
	SourceBucketPrefix      string
	DestinationBucketName   string
	DestinationBucketPrefix string
}

func NewLinkIndustryConnector() *LinkIndustryConnector {
	return &LinkIndustryConnector{name: "LinkIndustryConnector"}
}

func (u *LinkIndustryConnector) Name() string {
	return u.name
}
func (u *LinkIndustryConnector) Tx() core.Transaction {
	return *u.tx
}
func (u *LinkIndustryConnector) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *LinkIndustryConnector) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *LinkIndustryConnector) Registry() *registry.Registry {
	return u.reg
}

func (u *LinkIndustryConnector) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *LinkIndustryConnector) AccessPermission() constants.AppPermission {
	return constants.IndustryConnectorPermission
}

func (u *LinkIndustryConnector) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[%v] Validating request", u.Name())

	if rq.LinkIndustryConnectorRq.DomainName == "" {
		return errors.New("domain name cannot be empty")
	}
	//https://docs.aws.amazon.com/customerprofiles/latest/APIReference/API_CreateDomain.html
	if len(rq.LinkIndustryConnectorRq.DomainName) > customerprofileslcs.MAX_DOMAIN_NAME_LENGTH || len(rq.LinkIndustryConnectorRq.DomainName) < customerprofileslcs.MIN_DOMAIN_NAME_LENGTH {
		return fmt.Errorf("domain name length should be between %v and %v chars", customerprofileslcs.MIN_DOMAIN_NAME_LENGTH, customerprofileslcs.MAX_DOMAIN_NAME_LENGTH)
	}
	return nil
}

func (u *LinkIndustryConnector) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	var err error
	if req.LinkIndustryConnectorRq.RemoveLink {
		u.tx.Info("[LinkIndustryConnector] Removing link to domain %v", req.LinkIndustryConnectorRq.DomainName)
		err = removeLink(*u.reg.ConfigDB, req.LinkIndustryConnectorRq.DomainName)
	} else {
		u.tx.Info("[LinkIndustryConnector] Linking domain %v", req.LinkIndustryConnectorRq.DomainName)
		err = link(*u.reg.ConfigDB, req.LinkIndustryConnectorRq.DomainName)
		if err != nil {
			u.tx.Error("[LinkIndustryConnector] Error %v", err)
			return model.ResponseWrapper{}, err
		}
		err = beginDataTransfer(u, u.reg.DataSync, req.LinkIndustryConnectorRq.DomainName)
	}

	if err != nil {
		u.tx.Error("[LinkIndustryConnector] Error %v", err)
	}

	return model.ResponseWrapper{}, err
}

func (u *LinkIndustryConnector) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

// Link the domain to Industry Connector by adding the domain to the list of domains
// subscribed to Industry Connector PutObject events. All future objects processed
// by the Industry Connector will be processed via real time data stream.
func link(configDbClient db.DBConfig, domainName string) error {
	err := services.AddLinkedDomain(configDbClient, domainName)
	return err
}

// Stop the event subscription for this domain. Future records will no longer be
// processed by the real time data stream.
func removeLink(configDbClient db.DBConfig, domainName string) error {
	err := services.RemoveLinkedDomain(configDbClient, domainName)
	return err
}

// Start a one-time data transfer. This sends all existing data from the Industry Connector
// to be added to business object buckets and processed by the batch processing Glue jobs.
func beginDataTransfer(u *LinkIndustryConnector, dsc datasync.IConfig, domainName string) error {
	connectorBucketName := u.reg.Env["INDUSTRY_CONNECTOR_BUCKET_NAME"]
	dataSyncRoleArn := u.reg.Env["DATASYNC_ROLE_ARN"]
	dataSyncLogGroupArn := u.reg.Env["DATASYNC_LOG_GROUP_ARN"]

	transferObjects := []TransferObject{
		{
			ObjectName:              constants.BIZ_OBJECT_AIR_BOOKING,
			SourceBucketPrefix:      "bookings",
			DestinationBucketName:   u.reg.Env["S3_AIR_BOOKING"],
			DestinationBucketPrefix: constants.BIZ_OBJECT_AIR_BOOKING,
		},
		{
			ObjectName:              constants.BIZ_OBJECT_CLICKSTREAM,
			SourceBucketPrefix:      "clickstream",
			DestinationBucketName:   u.reg.Env["S3_CLICKSTREAM"],
			DestinationBucketPrefix: constants.BIZ_OBJECT_CLICKSTREAM,
		},
		{
			ObjectName:              constants.BIZ_OBJECT_GUEST_PROFILE,
			SourceBucketPrefix:      "profiles",
			DestinationBucketName:   u.reg.Env["S3_GUEST_PROFILE"],
			DestinationBucketPrefix: constants.BIZ_OBJECT_GUEST_PROFILE,
		},
		{
			ObjectName:              constants.BIZ_OBJECT_STAY_REVENUE,
			SourceBucketPrefix:      "stays",
			DestinationBucketName:   u.reg.Env["S3_STAY_REVENUE"],
			DestinationBucketPrefix: constants.BIZ_OBJECT_STAY_REVENUE,
		},
	}

	// Create tasks for each business object
	haveError := false
	for _, obj := range transferObjects {
		// Create source location
		sourceLocation, err := dsc.CreateS3Location(connectorBucketName, obj.SourceBucketPrefix, dataSyncRoleArn)
		if err != nil {
			u.tx.Error("LinkIndustryConnector][BeginDataTransfer] Error creating source location: %v", err)
			haveError = true
		}
		// Create destination location
		destinationLocation, err := dsc.CreateS3Location(obj.DestinationBucketName, obj.DestinationBucketPrefix, dataSyncRoleArn)
		if err != nil {
			u.tx.Error("LinkIndustryConnector][BeginDataTransfer] Error creating source location: %v", err)
			haveError = true
		}
		// Create task
		taskArn, err := dsc.CreateTask("ucp-datasync-"+domainName+"-"+obj.ObjectName, sourceLocation, destinationLocation, &dataSyncLogGroupArn)
		if err != nil {
			u.tx.Error("LinkIndustryConnector][BeginDataTransfer] Error creating task: %v", err)
			haveError = true
		}
		// Start task
		executionArn, err := dsc.StartTaskExecution(taskArn)
		if err != nil {
			u.tx.Error("LinkIndustryConnector][BeginDataTransfer] Error starting task execution: %v", err)
		} else {
			u.tx.Info("LinkIndustryConnector][BeginDataTransfer] Task started: %v", executionArn)
		}
	}

	if haveError {
		return errors.New("error creating data transfer tasks, please try linking the Industry Connector again")
	}

	return nil
}
