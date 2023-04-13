package main

import (
	"context"
	"os"
	"sync"
	athena "tah/core/athena"
	core "tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	db "tah/core/db"
	glue "tah/core/glue"
	common "tah/ucp-common/src/constant/admin"
	model "tah/ucp-sync/src/business-logic/model"
	maintainGluePartitions "tah/ucp-sync/src/business-logic/usecase/maintainGluePartitions"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Resources
var LAMBDA_ENV = os.Getenv("LAMBDA_ENV")
var LAMBDA_ACCOUNT_ID = os.Getenv("LAMBDA_ACCOUNT_ID")
var LAMBDA_REGION = os.Getenv("AWS_REGION")

var ATHENA_WORKGROUP = os.Getenv("ATHENA_WORKGROUP")
var ATHENA_TABLE = os.Getenv("ATHENA_TABLE")
var ATHENA_DB = os.Getenv("ATHENA_DB")

// Athena Table Names
var HOTEL_BOOKING_JOB_NAME = os.Getenv("HOTEL_BOOKING_JOB_NAME_CUSTOMER")
var AIR_BOOKING_JOB_NAME = os.Getenv("AIR_BOOKING_JOB_NAME_CUSTOMER")
var GUEST_PROFILE_JOB_NAME = os.Getenv("GUEST_PROFILE_JOB_NAME_CUSTOMER")
var PAX_PROFILE_JOB_NAME = os.Getenv("PAX_PROFILE_JOB_NAME_CUSTOMER")
var CLICKSTREAM_JOB_NAME = os.Getenv("CLICKSTREAM_JOB_NAME_CUSTOMER")
var HOTEL_STAY_JOB_NAME = os.Getenv("HOTEL_STAY_JOB_NAME_CUSTOMER")

var S3_HOTEL_BOOKING = os.Getenv("S3_HOTEL_BOOKING")
var S3_AIR_BOOKING = os.Getenv("S3_AIR_BOOKING")
var S3_GUEST_PROFILE = os.Getenv("S3_GUEST_PROFILE")
var S3_PAX_PROFILE = os.Getenv("S3_PAX_PROFILE")
var S3_CLICKSTREAM = os.Getenv("S3_CLICKSTREAM")
var S3_STAY_REVENUE = os.Getenv("S3_STAY_REVENUE")

var DYNAMO_TABLE = os.Getenv("DYNAMO_TABLE")
var DYNAMO_PK = os.Getenv("DYNAMO_PK")
var DYNAMO_SK = os.Getenv("DYNAMO_SK")

var ORIGIN_DATE = os.Getenv("PARTITION_START_DATE")

var configDb = db.Init(DYNAMO_TABLE, DYNAMO_PK, DYNAMO_SK)

var athenaCfg = athena.Init(ATHENA_DB, ATHENA_TABLE, ATHENA_WORKGROUP)
var glueCfg = glue.Init(LAMBDA_REGION, ATHENA_DB)
var accpCfg = customerprofiles.Init(LAMBDA_REGION)

var buckets = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: S3_HOTEL_BOOKING,
	common.BIZ_OBJECT_AIR_BOOKING:   S3_AIR_BOOKING,
	common.BIZ_OBJECT_GUEST_PROFILE: S3_GUEST_PROFILE,
	common.BIZ_OBJECT_PAX_PROFILE:   S3_PAX_PROFILE,
	common.BIZ_OBJECT_CLICKSTREAM:   S3_CLICKSTREAM,
	common.BIZ_OBJECT_STAY_REVENUE:  S3_STAY_REVENUE,
}
var jobs = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: HOTEL_BOOKING_JOB_NAME,
	common.BIZ_OBJECT_AIR_BOOKING:   AIR_BOOKING_JOB_NAME,
	common.BIZ_OBJECT_GUEST_PROFILE: GUEST_PROFILE_JOB_NAME,
	common.BIZ_OBJECT_PAX_PROFILE:   PAX_PROFILE_JOB_NAME,
	common.BIZ_OBJECT_CLICKSTREAM:   CLICKSTREAM_JOB_NAME,
	common.BIZ_OBJECT_STAY_REVENUE:  HOTEL_STAY_JOB_NAME,
}

func HandleRequest(ctx context.Context, req events.CloudWatchEvent) (model.ResponseWrapper, error) {
	tx := core.NewTransaction("ucp-sync", "")
	tx.Log("Received Request %+v with context %+v", req, ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		tx.Log("Starting use case %+v", "maintainGluePartitions")
		maintainGluePartitions.Run(tx, glueCfg, configDb, buckets, accpCfg, LAMBDA_ENV, ORIGIN_DATE, jobs)
		wg.Done()
	}()
	wg.Wait()
	return model.ResponseWrapper{}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
