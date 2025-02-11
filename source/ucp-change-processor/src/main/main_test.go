// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	customerprofiles "tah/upt/source/storage"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	eventbridge "tah/upt/source/tah-core/eventbridge"
	firehose "tah/upt/source/tah-core/firehose"
	glue "tah/upt/source/tah-core/glue"
	iam "tah/upt/source/tah-core/iam"
	s3 "tah/upt/source/tah-core/s3"
	sqs "tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	awsfirehose "github.com/aws/aws-sdk-go/service/firehose"
	s3p "github.com/xitongsys/parquet-go-source/s3"
	"github.com/xitongsys/parquet-go/reader"
)

var UCP_REGION = getRegion()
var GLUE_SCHEMA_PATH = os.Getenv("GLUE_SCHEMA_PATH")
var TRAVELLER_EVENT_BUS_TEST = "test_bus"

var expectedEvents = []TravelerEvent{
	{
		TravelerID: "ae5e36d15d9641e8978ecce97c542cca",
		Events: []AccpEvent{{
			EventType:  "CREATED",
			ObjectType: "_profile",
		}},
		Domain: "ucp-test-sprint-9-4",
	},
	{
		TravelerID: "6831e372b0d149bf9a0e2f6772c9c0a1",
		Events: []AccpEvent{{
			EventType:  "UPDATED",
			ObjectType: "air_booking",
		}},
		Domain: "ucp-test-sprint-9-4",
	},
	{
		TravelerID: "111",
		Events: []AccpEvent{{
			EventType:  "MERGED",
			ObjectType: "_profile",
			MergeEventDetails: MergeEvent{
				SourceConnectID:  "111",
				TargetConnectIDs: []string{"222", "333"},
				SourceProfileID:  "tid_111",
				TargetProfileIDs: []string{"tid_222", "tid_333"}}}},
		Domain: "IrvinDomain",
	},
}

func TestLogErrorIfNotNil(t *testing.T) {

	svc := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[testMain] Could not create test queue")
	}
	r := Runner{
		Sqs: svc,
	}
	r.Tx = core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	r.LogErrorIfNotNil(nil, "log line", map[string]string{"key": "value"})
	r.LogErrorIfNotNil(errors.New("testError"), "log line", map[string]string{"key": "value"})

	content, err1 := svc.Get(sqs.GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestMain] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) != 1 {
		t.Errorf("[TestMain] exactly 1 error should has been logged to SQS: %+v", content.Peek)
	}
	log.Printf("content: %v", content.Peek[0])
	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestMain]error deleting queue by name: %v", err)
	}
}

func TestRetrieveTravellerWithRetries(t *testing.T) {
	svc := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[testMain] Could not create test queue")
	}
	r := Runner{
		Tx: core.NewTransaction(t.Name(), "", core.LogLevelDebug),
		Accp: &customerprofiles.MockCustomerProfileLowCostConfig{
			Profiles:    &[]profilemodel.Profile{{ProfileId: "123456"}},
			Profile:     &profilemodel.Profile{ProfileId: "123456"},
			MaxFailures: 2,
			Mappings: &[]customerprofiles.ObjectMapping{
				{Name: "air_booking"},
				{Name: "air_loyalty"},
				{Name: "hotel_loyalty"},
				{Name: "hotel_booking"},
				{Name: "hotel_stay_revenue_items"},
				{Name: "customer_service_interaction"},
				{Name: "ancillary_service"},
				{Name: "alternate_profile_id"},
				{Name: "guest_profile"},
				{Name: "pax_profile"},
				{Name: "email_history"},
				{Name: "phone_history"},
				{Name: "loyalty_transaction"},
				{Name: "clickstream"},
			},
		},
		Sqs: svc,
	}
	tvl, err := r.retrieveTravellerWithRetries("123456")
	if err != nil {
		t.Errorf("[TestRetrieveTravellerWithRetries] error retrieving traveller: %v. should have succeeded after retrying twice", err)
	}
	log.Printf("Retrieved traveller: %+v", tvl)
	if tvl.ConnectID != "123456" {
		t.Errorf("[TestRetrieveTravellerWithRetries] Should return traveller with connectId %v after successful retry", "123456")
	}
	(r.Accp.(*customerprofiles.MockCustomerProfileLowCostConfig)).MaxFailures = 3
	(r.Accp.(*customerprofiles.MockCustomerProfileLowCostConfig)).Failures = 0
	_, err = r.retrieveTravellerWithRetries("123456")
	if err == nil {
		t.Errorf("[TestRetrieveTravellerWithRetries] Should have returned an error after retrying twice")
	}
	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestChangeProcessing]error deleting queue by name: %v", err)
	}
}

func TestChangeProcessing(t *testing.T) {
	svc := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[E2EDataStream] Could not create test queue")
	}
	s3c, err := s3.InitWithRandBucket("test-change-proc", "", UCP_REGION, "", "")
	if err != nil {
		t.Errorf("[testMain] Could not create test bucket %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	r := &Runner{
		Tx:   tx,
		Sqs:  svc,
		S3:   s3c,
		Accp: &customerprofiles.MockCustomerProfileLowCostConfig{},
	}
	//test profile record
	err = r.parseKinesisRecord(events.KinesisEventRecord{
		Kinesis: events.KinesisRecord{
			Data: []byte(PROFILE_REC_EXAMPLE),
		},
	})
	state := r.State
	log.Printf("State without Object: %+v", state)
	if err != nil {
		t.Errorf("[TestChangeProcessing] error parsing Kinesis record %v", err)
	}
	if state.PartialTraveler[0].ConnectID != "ae5e36d15d9641e8978ecce97c542cca" {
		t.Errorf("[TestChangeProcessing] invalid ConnectID in traveller record %v", state.PartialTraveler)
	}
	if state.Domain != "ucp-test-sprint-9-4" {
		t.Errorf("[TestChangeProcessing] invalid domain %v", state.Domain)
	}
	if state.EventType != "CREATED" {
		t.Errorf("[TestChangeProcessing] invalid event type %v", state.EventType)
	}
	if state.RawData != escapeRawData(PROFILE_REC_EXAMPLE) {
		t.Errorf("[TestChangeProcessing] invalid raw data %v", state.RawData)
	}

	//test object record
	err = r.parseKinesisRecord(events.KinesisEventRecord{
		Kinesis: events.KinesisRecord{
			Data: []byte(PROFILE_OBJECT_REC_EXAMPLE),
		},
	})
	state = r.State
	log.Printf("State with object: %+v", state)
	if err != nil {
		t.Errorf("[TestChangeProcessing] error parsing Kinesis record %v", err)
	}
	//traveller object should be correct
	if state.PartialTraveler[0].ConnectID != "6831e372b0d149bf9a0e2f6772c9c0a1" {
		t.Errorf("[TestChangeProcessing] invalid ConnectID in traveller record %v", state.PartialTraveler)
	}
	if len(state.PartialTraveler[0].AirBookingRecords) != 1 {
		t.Errorf("[TestChangeProcessing] traveller should have 1 Air Booking record %v", state.PartialTraveler)
	}
	//domain should be correct
	if state.Domain != "ucp-test-sprint-9-4" {
		t.Errorf("[TestChangeProcessing] invalid domain %v", state.Domain)
	}
	if state.EventType != "UPDATED" {
		t.Errorf("[TestChangeProcessing] invalid eventType %v", state.EventType)
	}
	if state.RawData != escapeRawData(PROFILE_OBJECT_REC_EXAMPLE) {
		t.Errorf("[TestChangeProcessing] invalid raw data %v", state.RawData)
	}

	//test merge record
	err = r.parseKinesisRecord(events.KinesisEventRecord{
		Kinesis: events.KinesisRecord{
			Data: []byte(MERGE_EVENT_EXAMPLE),
		},
	})
	state = r.State
	log.Printf("State with object: %+v", state)
	if err != nil {
		t.Errorf("[TestChangeProcessing] error parsing Kinesis record %v", err)
	}
	//domain should be correct
	if state.Domain != "IrvinDomain" {
		t.Errorf("[TestChangeProcessing] invalid domain %v", state.Domain)
	}
	if state.EventType != "MERGED" {
		t.Errorf("[TestChangeProcessing] invalid eventType %v", state.EventType)
	}
	if state.MergeEvent.SourceConnectID != "111" {
		t.Errorf("[TestChangeProcessing] Merge Event sourceConnectId should be 111 and not %s", state.MergeEvent.SourceConnectID)
	}
	if len(state.MergeEvent.TargetConnectIDs) != 2 {
		t.Errorf("[TestChangeProcessing] Merge Event should hev 2 targetConnectIds and not %v", len(state.MergeEvent.TargetConnectIDs))
	}
	if state.MergeEvent.TargetConnectIDs[0] != "222" || state.MergeEvent.TargetConnectIDs[1] != "333" {
		t.Errorf("[TestChangeProcessing] invalid raw target connect id array: %v (should be [222,333])", state.MergeEvent.TargetConnectIDs)
	}
	if state.MergeEvent.SourceProfileID != "tid_111" {
		t.Errorf("[TestChangeProcessing] Merge Event sourceProfileId should be tid_111 and not %s", state.MergeEvent.SourceProfileID)
	}
	if len(state.MergeEvent.TargetProfileIDs) != 2 {
		t.Errorf("[TestChangeProcessing] Merge Event should hev 2 targetProfileIds and not %v", len(state.MergeEvent.TargetProfileIDs))
	}
	if state.MergeEvent.TargetProfileIDs[0] != "tid_222" || state.MergeEvent.TargetProfileIDs[1] != "tid_333" {
		t.Errorf("[TestChangeProcessing] invalid raw target profile id array: %v (should be [tid_222,tid_333])", state.MergeEvent.TargetProfileIDs)
	}

	//queue should have no errors
	content, err1 := svc.Get(sqs.GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestChangeProcessing] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) > 0 {
		t.Errorf("[TestChangeProcessing] Error during ingestion: %v", content.Peek)
	}
	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestChangeProcessing]error deleting queue by name: %v", err)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestMain]error deleting s3 by name: %v", err)
	}
}

func TestSqsErrors(t *testing.T) {
	svc := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2" + core.GenerateUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[E2EDataStream] Could not create test queue")
	}
	s3c, err := s3.InitWithRandBucket("test-change-proc-sqs-errors", "", UCP_REGION, "", "")
	if err != nil {
		t.Errorf("[testMain] Could not create test bucket %v", err)
	}
	req := events.KinesisEvent{Records: []events.KinesisEventRecord{
		{
			Kinesis: events.KinesisRecord{
				Data: []byte("illFormatted record"),
			},
		},
	}}
	r := Runner{
		Sqs:      svc,
		Accp:     &customerprofiles.MockCustomerProfileLowCostConfig{},
		S3:       s3c,
		EvBridge: &eventbridge.MockConfig{},
	}
	r.HandleRequestWithServices(context.Background(), req)

	content, err1 := svc.Get(sqs.GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestSQS] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "illFormatted record" {
		t.Errorf("[TestSQS] message in queue should be %v", "illFormatted record")
	}

	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue by name: %v", err)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestMain]error deleting s3 by name: %v", err)
	}
}

func TestMain(t *testing.T) {
	buildRunnerFromGlobals()
	testPostFix := strings.ToLower(core.GenerateUniqueId())
	svc := sqs.Init(UCP_REGION, "", "")
	qName := "Test-Queue-2-" + testPostFix
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[testMain] Could not create test queue")
	}
	s3c, err := s3.InitWithRandBucket("test-change-proc-main", "", UCP_REGION, "", "")
	if err != nil {
		t.Errorf("[testMain] Could not create test bucket %v", err)
	}
	name := s3c.Bucket
	bucketArn := "arn:aws:s3:::" + name
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:PutObject"}
	principal := map[string][]string{"Service": {"firehose.amazonaws.com"}}
	err = s3c.AddPolicy(name, resources, actions, principal)
	if err != nil {
		t.Errorf("[TestCreateDeleteBucket] error adding bucket policy %+v", err)
	}

	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	glueDbName := "test_database-" + testPostFix
	glueTableName := "test_table-" + testPostFix
	glueRoleName := "test_glue_role-" + testPostFix
	glueCfg := glue.Init(UCP_REGION, glueDbName, "", "")
	filePath := GLUE_SCHEMA_PATH + "/" + "tah-common-glue-schemas/traveller.glue.json"
	schemaBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("[TestSchema] error reading file %v", err)
	}
	glueSchema, err := glue.ParseSchema(string(schemaBytes))
	if err != nil {
		t.Errorf("[TestSchema] error parsing schema %v", err)
	}

	err = glueCfg.CreateDatabase(glueDbName)
	if err != nil {
		t.Errorf("[TestMain] error creating database %v", err)
	}
	err = glueCfg.CreateParquetTable(glueTableName, name, nil, glueSchema)
	if err != nil {
		t.Errorf("[TestMain] error creating table %v", err)
	}
	glueRoleArn, gluePolicyArn, err := iamCfg.CreateRoleWithActionsResource(glueRoleName, "firehose.amazonaws.com", []string{
		"glue:GetTable",
		"glue:GetTableVersion",
		"glue:GetTableVersions",
	}, []string{
		"arn:aws:glue:" + UCP_REGION + ":*:catalog",
		"arn:aws:glue:" + UCP_REGION + ":*:database/" + glueDbName,
		"arn:aws:glue:" + UCP_REGION + ":*:table/" + glueDbName + "/" + glueTableName,
	})
	if err != nil {
		t.Errorf("[TestMain] error creating role %v", err)
	}

	t.Cleanup(func() {
		err = iamCfg.DetachRolePolicy(glueRoleName, gluePolicyArn)
		if err != nil {
			t.Errorf("[%v] error detaching policy from role: %v", t.Name(), err)
		}
		err = iamCfg.DeletePolicy(gluePolicyArn)
		if err != nil {
			t.Errorf("[%v] error deleting glue policy: %v", t.Name(), err)
		}
		err = iamCfg.DeleteRole(glueRoleName)
		if err != nil {
			t.Errorf("[%v] error deleting glue role: %v", t.Name(), err)
		}
	})

	roleName := "firehose-unit-test-role-" + testPostFix
	roleArn, policyArn, err := iamCfg.CreateRoleWithActionsResource(roleName, "firehose.amazonaws.com", []string{"s3:PutObject",
		"s3:PutObjectAcl",
		"s3:ListBucket"}, resources)
	if err != nil {
		t.Errorf("[TestIam] error %v", err)
	}

	t.Cleanup(func() {
		err = iamCfg.DetachRolePolicy(roleName, policyArn)
		if err != nil {
			t.Errorf("[%v] error detaching policy from role: %v", t.Name(), err)
		}
		err := iamCfg.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[%v] error deleting firehose policy: %v", t.Name(), err)
		}
		err = iamCfg.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[%v] error deleting firehose role: %v", t.Name(), err)
		}
	})
	log.Printf("Waiting 10 seconds role to propagate")
	time.Sleep(10 * time.Second)
	streamName := "test_stream_change_proc-" + testPostFix
	log.Printf("0-Creating stream %v", streamName)
	firehoseCfg, err := firehose.InitAndCreateWithPartitioningAndDataFormatConversion(
		streamName,
		bucketArn,
		roleArn,
		[]firehose.PartitionKeyVal{{Key: "domainname", Val: ".domain"}},
		UCP_REGION,
		"solutionId",
		"solutionVersion",
		&awsfirehose.DataFormatConversionConfiguration{
			SchemaConfiguration: &awsfirehose.SchemaConfiguration{
				Region:       &UCP_REGION,
				DatabaseName: &glueCfg.DbName,
				TableName:    &glueTableName,
				VersionId:    aws.String("LATEST"),
				RoleARN:      &glueRoleArn,
			},
			Enabled: aws.Bool(true),
			InputFormatConfiguration: &awsfirehose.InputFormatConfiguration{
				Deserializer: &awsfirehose.Deserializer{
					HiveJsonSerDe: &awsfirehose.HiveJsonSerDe{},
				},
			},
			OutputFormatConfiguration: &awsfirehose.OutputFormatConfiguration{
				Serializer: &awsfirehose.Serializer{
					ParquetSerDe: &awsfirehose.ParquetSerDe{
						Compression: aws.String("GZIP"),
					},
				},
			},
		})
	if err != nil {
		t.Errorf("[TestMain] error creating firehose stream %v", err)
	}
	_, err = firehoseCfg.WaitForStreamCreation(streamName, 300)
	if err != nil {
		t.Errorf("Timeout while waiting for stream creation")
	}

	req := events.KinesisEvent{Records: []events.KinesisEventRecord{
		{
			Kinesis: events.KinesisRecord{
				Data: []byte(PROFILE_REC_EXAMPLE),
			},
		},
		{
			Kinesis: events.KinesisRecord{
				Data: []byte(PROFILE_OBJECT_REC_EXAMPLE),
			},
		},
		{
			Kinesis: events.KinesisRecord{
				Data: []byte(MERGE_EVENT_EXAMPLE),
			},
		},
	}}
	r := Runner{
		Sqs: svc,
		Accp: &customerprofiles.MockCustomerProfileLowCostConfig{
			ProfilesByID: map[string]profilemodel.Profile{
				"111": {
					Domain:    "IrvinDomain",
					ProfileId: "111",
				},
				"222": {
					Domain:    "IrvinDomain",
					ProfileId: "222",
				},
				"333": {
					Domain:    "IrvinDomain",
					ProfileId: "333",
				},
				"ae5e36d15d9641e8978ecce97c542cca": {
					Domain:    "ucp-test-sprint-9-4",
					ProfileId: "ae5e36d15d9641e8978ecce97c542cca",
				},
				"6831e372b0d149bf9a0e2f6772c9c0a1": {
					Domain:    "ucp-test-sprint-9-4",
					ProfileId: "6831e372b0d149bf9a0e2f6772c9c0a1",
				},
			},
			Mappings: &[]customerprofiles.ObjectMapping{
				{Name: "air_booking"},
				{Name: "air_loyalty"},
				{Name: "hotel_loyalty"},
				{Name: "hotel_booking"},
				{Name: "hotel_stay_revenue_items"},
				{Name: "customer_service_interaction"},
				{Name: "ancillary_service"},
				{Name: "alternate_profile_id"},
				{Name: "guest_profile"},
				{Name: "pax_profile"},
				{Name: "email_history"},
				{Name: "phone_history"},
				{Name: "loyalty_transaction"},
				{Name: "clickstream"},
			},
		},
		EvBridge: &eventbridge.MockConfig{},
		S3:       s3c,
		Firehose: *firehoseCfg,
		Env: map[string]string{
			"EVENTBRIDGE_ENABLED":  "true",
			"FIREHOSE_STREAM_NAME": streamName,
		},
	}

	r.HandleRequestWithServices(context.Background(), req)

	content, err1 := svc.Get(sqs.GetMessageOptions{})
	if err1 != nil {
		t.Errorf("[TestMain] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) > 0 {
		t.Errorf("[TestMain] An error has been logged to SQS: %v", content.Peek)
	}

	ebCfg := r.EvBridge.(*eventbridge.MockConfig)
	if len(ebCfg.Events) != 3 {
		t.Errorf("[TestMain] Expected 3 events to be sent to EventBridge, got %v", len(ebCfg.Events))
	} else {
		matched := map[string]bool{}
		for i, event := range ebCfg.Events {
			tEvent := event.(TravelerEvent)
			for _, expected := range expectedEvents {
				if tEvent.TravelerID == expected.TravelerID {
					matched[tEvent.TravelerID] = true
					if tEvent.Domain != expected.Domain {
						t.Errorf("[mock] invalid domain for event %v should be %s, got %v ", i, expected.Domain, tEvent.Domain)
					}
					if tEvent.Events[0].EventType != expected.Events[0].EventType {
						t.Errorf("[mock] invalid event type for event %v should be %s, got %v ", i, expected.Events[0].EventType, tEvent.Events[0].EventType)
					}
					if tEvent.Events[0].EventType == "MERGED" {
						accpEvent := tEvent.Events[0]
						if accpEvent.MergeEventDetails.SourceProfileID != expectedEvents[2].Events[0].MergeEventDetails.SourceProfileID {
							t.Errorf("[mock] Both source and target event traveler ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
						if accpEvent.MergeEventDetails.TargetProfileIDs[0] != expectedEvents[2].Events[0].MergeEventDetails.TargetProfileIDs[0] {
							t.Errorf("[mock] Both source and target event traveler ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
						if accpEvent.MergeEventDetails.TargetProfileIDs[1] != expectedEvents[2].Events[0].MergeEventDetails.TargetProfileIDs[1] {
							t.Errorf("[mock] Both source and target event traveler ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
						if accpEvent.MergeEventDetails.SourceConnectID != expectedEvents[2].Events[0].MergeEventDetails.SourceConnectID {
							t.Errorf("[mock] Both source and target event connect ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
						if accpEvent.MergeEventDetails.TargetConnectIDs[0] != expectedEvents[2].Events[0].MergeEventDetails.TargetConnectIDs[0] {
							t.Errorf("[mock] Both source and target event connect ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
						if accpEvent.MergeEventDetails.TargetConnectIDs[1] != expectedEvents[2].Events[0].MergeEventDetails.TargetConnectIDs[1] {
							t.Errorf("[mock] Both source and target event connect ids should be provided in Merge event, got %v ", accpEvent.MergeEventDetails)
						}
					}
				}
			}
		}
		for _, expected := range expectedEvents {
			if !matched[expected.TravelerID] {
				t.Errorf("[mock] expected event %v not found in events sent to EventBridge", expected)
			}
		}
	}

	err = firehose.WaitForObject(s3c, "data/domainname=IrvinDomain", 600)
	if err != nil {
		t.Errorf("[TestMain] error waiting for data in firehose: %v", err)
	}
	err = firehose.WaitForObject(s3c, "data/domainname=ucp-test-sprint-9-4", 300)
	if err != nil {
		t.Errorf("[TestMain] error waiting for data in firehose: %v", err)
	}
	ctx := context.Background()
	hasMergedProfile1 := false
	hasMergedProfile2 := false
	nObjectsPerDomain := map[string]int{}
	for _, expected := range []string{"data/domainname=IrvinDomain", "data/domainname=ucp-test-sprint-9-4"} {
		res, err := s3c.Search(expected, 100)
		if err != nil || len(res) == 0 {
			t.Errorf("object with prefix %v should be uploaded in S3: got error: %v or no data", expected, err)
		} else {
			for _, r := range res {
				log.Printf("Checking domain %s content", expected)
				fr, err := s3p.NewS3FileReader(ctx, s3c.Bucket, r, &s3c.Svc.Config)
				if err != nil {
					t.Errorf("[TestMain] error reading file in bucket: %v", err)
					panic("Error reading file")
				}

				pr, err := reader.NewParquetReader(fr, nil, 4)
				if err != nil {
					t.Errorf("[TestMain] error creating parquet reader: %v", err)
				}

				num := int(pr.GetNumRows())
				nObjectsPerDomain[expected] += num
				res, err := pr.ReadByNumber(num)
				if err != nil {
					t.Errorf("[TestMain] error reading parquet: %v", err)
				}
				log.Printf("Parquet reader response: %+v", res)
				jsonBytes, err := json.Marshal(res)
				if err != nil {
					t.Errorf("[TestMain] error marshalling json: %v", err)
				}
				log.Printf("Parquet reader response (Json string): %+v", string(jsonBytes))

				recs := make([]model.Traveller, num)

				if err = json.Unmarshal(jsonBytes, &recs); err != nil {
					log.Printf("[TestMain] error unmarshalling json: %v", err)
				}
				for _, rec := range recs {
					if rec.ConnectID == "222" && rec.MergedIn == "111" && expected == "data/domainname=IrvinDomain" {
						log.Printf("[TestMain] Found merged profile 222")
						hasMergedProfile1 = true
					}
					if rec.ConnectID == "333" && rec.MergedIn == "111" && expected == "data/domainname=IrvinDomain" {
						log.Printf("[TestMain] Found merged profile 333")
						hasMergedProfile2 = true
					}
					tZero := time.Time{}
					if rec.LastUpdated == tZero {
						t.Errorf("[TestMain] Last updated timestamp should be greater than %v on S3 traveller", tZero)
					}
				}

				pr.ReadStop()
				fr.Close()
			}
		}
	}
	if !hasMergedProfile1 {
		t.Errorf("[TestMain] Bucket should contain merged traveller with ID 222")
	}
	if !hasMergedProfile2 {
		t.Errorf("[TestMain] Bucket should contain merged traveller with ID 333")
	}
	if nObjectsPerDomain["data/domainname=IrvinDomain"] != 3 {
		t.Errorf("[TestMain] Bucket should contain 3 objects with prefix %v not %v", "data/domainname=IrvinDomain", nObjectsPerDomain["data/domainname=IrvinDomain"])
	}
	if nObjectsPerDomain["data/domainname=ucp-test-sprint-9-4"] != 2 {
		t.Errorf("[TestMain] Bucket should contain 2 objects with prefix %v not %v", "data/domainname=ucp-test-sprint-9-4", nObjectsPerDomain["data/domainname=ucp-test-sprint-9-4"])
	}

	err = glueCfg.DeleteTable(glueTableName)
	if err != nil {
		t.Errorf("[TestMain] error deleting table: %v", err)
	}

	err = glueCfg.DeleteDatabase(glueCfg.DbName)
	if err != nil {
		t.Errorf("[TestMain] error deleting database: %v", err)
	}

	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestMain]error deleting queue by name: %v", err)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestMain]error deleting s3 by name: %v", err)
	}
	log.Printf("4-Deleting stream %v", streamName)
	err = firehoseCfg.Delete(streamName)
	if err != nil {
		t.Errorf("[TestKinesis] error deleting stream: %v", err)
	}
}

func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}

func escapeRawData(data string) string {
	edata := strings.Replace(data, "\n", "", -1)
	edata = strings.Replace(edata, "\t", "", -1)
	return edata
}

var PROFILE_REC_EXAMPLE = `{
	"SchemaVersion": 0,
	"EventId": "f2a97405-4ed6-4c09-90a7-b37eb9966d22",
	"EventTimestamp": "2023-04-28 14:42:18.761 +0000 UTC",
	"EventType": "CREATED",
	"DomainName": "ucp-test-sprint-9-4",
	"ObjectTypeName": "_profile",
	"Object": {
		"Address":{},
		"Attributes":{
			"company":"CrowdANALYTIX",
			"email_business":"joeyaufderhar@example.com",
			"honorific":"Ms",
			"job_title":"Architect",
			"payment_type":"voucher",
			"phone_business":"2459719737",
			"profile_id":"4443810144"
		},
		"BillingAddress":{},
		"BirthDate":"2022-03-15",
		"FirstName":"Giovanny",
		"Gender":"other",
		"GenderString":"other",
		"LastName":"Walsh",
		"MailingAddress":{
			"Address1":"86331 Manors port",
			"Address2":"more content line 2",
			"Address3":"more content line 3",
			"Address4":"more content line 4",
			"City":"St. Petersburg",
			"Country":"IQ",
			"PostalCode":"78847"
			},
		"MiddleName":"Marietta",
		"ProfileId":"ae5e36d15d9641e8978ecce97c542cca"
	},
	"IsMessageRealTime": true
  }`

var PROFILE_OBJECT_REC_EXAMPLE = `{
	"SchemaVersion": 0,
	"EventId": "f2a97405-4ed6-4c09-90a7-b37eb9966d22",
	"EventTimestamp": "2023-04-28 14:42:18.761 +0000 UTC",
	"EventType": "UPDATED",
	"DomainName": "ucp-test-sprint-9-4",
	"ObjectTypeName": "air_booking",
	"AssociatedProfileId":"6831e372b0d149bf9a0e2f6772c9c0a1",
	"ProfileObjectUniqueKey":"w7rb3HISZsELqNtJuEC/oh7flFAT2YJclNS1joYceJY=",
	"Object": {
		"address_billing_city":"",
		"address_billing_country":"",
		"address_billing_line1":"",
		"address_billing_line2":"",
		"address_billing_line3":"",
		"address_billing_line4":"",
		"address_billing_postal_code":"",
		"address_billing_state_province":"",
		"address_business_city":"",
		"address_business_country":"",
		"address_business_line1":"",
		"address_business_line2":"",
		"address_business_line3":"",
		"address_business_line4":"",
		"address_business_postal_code":"",
		"address_business_state_province":"",
		"address_city":"Henderson",
		"address_country":"CX",
		"address_line1":"65845 Ridge borough",
		"address_line2":"more content line 2",
		"address_line3":"more content line 3",
		"address_line4":"more content line 4",
		"address_mailing_city":"",
		"address_mailing_country":"",
		"address_mailing_line1":"",
		"address_mailing_line2":"",
		"address_mailing_line3":"",
		"address_mailing_line4":"",
		"address_mailing_postal_code":"",
		"address_mailing_state_province":"",
		"address_postal_code":"32941",
		"address_state_province":"",
		"address_type":"",
		"arrival_date":"2021-11-03",
		"arrival_time":"16:21",
		"booking_id":"YQ3P9B",
		"cc_cvv":"",
		"cc_exp":"",
		"cc_name":"",
		"cc_token":"",
		"cc_type":"",
		"channel":"mobile",
		"company":"Healthgrades",
		"date_of_birth":"1919-06-01",
		"departure_date":"2021-11-03",
		"departure_time":"12:12",
		"email":"",
		"email_business":"montyswaniawski@example.com",
		"email_type":"business", "first_name":"Lacy",
		"flight_number":"UA3833",
		"from":"SFO",
		"gds_id":"",
		"gender":"male",
		"honorific":"Ms",
		"job_title":"Analyst",
		"language_code":"",
		"language_name":"",
		"last_name":"Kessler",
		"last_updated":"2021-12-31T025607.912662687Z",
		"last_updated_by":"Martina Upton",
		"middle_name":"Earnest",
		"model_version":"1.0",
		"nationality_code":"",
		"nationality_name":"",
		"object_type":"air_booking",
		"payment_type":"bank_account",
		"phone":"",
		"phone_business":"351.804.6565",
		"phone_home":"",
		"phone_mobile":"",
		"phone_type":"business",
		"price":"0.0",
		"pronoun":"he",
		"pss_id":"",
		"segment_id":"YQ3P9B-SFO-LAS",
		"status":"canceled",
		"to":"LAS",
		"traveller_id":"2661022692"
	},
	"IsMessageRealTime": true
}`

var MERGE_EVENT_EXAMPLE = `
{
	"SchemaVersion": 0,
	"EventId": "b88d5339-69a3-4ef0-ac0e-208be1823a55",
	"EventTimestamp": "2023-08-24T02:43:20.000Z",
	"EventType": "MERGED",
	"DomainName": "IrvinDomain",
	"ObjectTypeName": "_profile",
	"Object": {
	  "FirstName": "FirstName",
	  "LastName": "LastName",
	  "AdditionalInformation": "Main",
	  "ProfileId": "111",
	  "Attributes": {
		"profile_id": "tid_111"
	}
	},
	"MergeId": "045d69ae-42c2-11ee-be56-0242ac120002",
	"DuplicateProfiles": [
	  {
		"FirstName": "FirstName",
		"LastName": "LastName",
		"AdditionalInformation": "Duplicate1",
		"ProfileId": "222",
		"Attributes": {
			"profile_id": "tid_222"
		}
	  },
	  {
		"FirstName": "FirstName",
		"LastName": "LastName",
		"AdditionalInformation": "Duplicate2",
		"ProfileId": "333",
		"Attributes": {
			"profile_id": "tid_333"
		}
	  }
	],
	"IsMessageRealTime": true
}`
