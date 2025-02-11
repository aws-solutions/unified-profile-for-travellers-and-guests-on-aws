// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maintainGluePartitions

import (
	"log"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/kms"
	"tah/upt/source/tah-core/sqs"
	common "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/config"
	testutils "tah/upt/source/ucp-common/src/utils/test"
	"tah/upt/source/ucp-sync/src/business-logic/model"
	"testing"
	"time"
	//model "cloudrack-lambda-core/config/model"
)

var buckets = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: "s3_hotel_booking",
	common.BIZ_OBJECT_AIR_BOOKING:   "s3_air_booking",
	common.BIZ_OBJECT_GUEST_PROFILE: "s3_guest_profile",
	common.BIZ_OBJECT_PAX_PROFILE:   "s3_pax_profile",
	common.BIZ_OBJECT_CLICKSTREAM:   "s3_clickstream",
	common.BIZ_OBJECT_STAY_REVENUE:  "s3_stay_revenue",
	common.BIZ_OBJECT_CSI:           "s3_csi",
}
var CONNECT_PROFILE_SOURCE_BUCKET string = "s3_connect_profile"
var jobs = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: "GLUE_JOB_NAME_HOTEL_BOOKINGS",
	common.BIZ_OBJECT_AIR_BOOKING:   "GLUE_JOB_NAME_AIR_BOOKING",
	common.BIZ_OBJECT_GUEST_PROFILE: "GLUE_JOB_NAME_GUEST_PROFILES",
	common.BIZ_OBJECT_PAX_PROFILE:   "GLUE_JOB_NAME_PAX_PROFILES",
	common.BIZ_OBJECT_CLICKSTREAM:   "GLUE_JOB_NAME_CLICKSTREAM",
	common.BIZ_OBJECT_STAY_REVENUE:  "GLUE_JOB_NAME_STAY_REVENUE",
	common.BIZ_OBJECT_CSI:           "GLUE_JOB_NAME_CSI",
}

func TestBuildTableConfig(t *testing.T) {
	domains := []commonModel.Domain{{Name: "test_dom_1"}, {Name: "test_dom_2"}}
	env := "test_env"
	configs := buildTableConfig(domains, env, buckets, jobs, "")
	nTableExpected := len(domains) * len(common.BUSINESS_OBJECTS)
	if len(configs) != nTableExpected {
		t.Errorf("[TestBuildTableConfig] %v tables should be created", nTableExpected)
	}
	for i, domain := range domains {
		for j, bo := range common.BUSINESS_OBJECTS {
			expected := "ucp_" + env + "_" + bo.Name + "_" + domain.Name
			if configs[i*len(common.BUSINESS_OBJECTS)+j].TableName != expected {
				t.Errorf("[TestBuildTableConfig] table config at pos (%v,%v) should be %v", i, j, expected)
			}
			if configs[i*len(common.BUSINESS_OBJECTS)+j].JobName != jobs[bo.Name] {
				t.Errorf("[TestBuildTableConfig] job name in config at pos (%v,%v) should be %v", i, j, jobs[bo.Name])
			}
		}
	}

	configs = buildTableConfig(domains, env, buckets, jobs, "GLUE_JOB_NAME_GUEST_PROFILES")
	nTableExpected = len(domains)
	if len(configs) != nTableExpected {
		t.Errorf("[TestBuildTableConfig] %v tables should be created", nTableExpected)
	}
	for i, domain := range domains {
		expected := "ucp_" + env + "_" + common.BIZ_OBJECT_GUEST_PROFILE + "_" + domain.Name
		if configs[i].TableName != expected {
			t.Errorf("[TestBuildTableConfig] table config at pos (%v) should be %v", i, expected)
		}
		if configs[i].JobName != "GLUE_JOB_NAME_GUEST_PROFILES" {
			t.Errorf("[TestBuildTableConfig] job name in config at pos (%v) should be %v", i, "GLUE_JOB_NAME_GUEST_PROFILES")
		}
	}
}

func TestResetJobBookmark(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	cfg, err := db.InitWithNewTable("ucp_sync_unit_tests_job_bookmark-"+core.GenerateUniqueId(), "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("error creating dynamo table: %v", err)
	}
	t.Cleanup(func() { cfg.DeleteTable(cfg.TableName) })
	cfg.WaitForTableCreation()
	bizObject := common.BIZ_OBJECT_HOTEL_BOOKING
	domainName := "test_dom_1_rest_bookmark"
	pk := "glue_job_bookmark"
	sk := bizObject + "_" + domainName
	_, err = cfg.Save(model.JobBookmark{Pk: pk, Sk: sk})
	if err != nil {
		t.Fatalf("Error saving test bookmark %v", err)
	}
	bookmark := model.JobBookmark{}
	err = cfg.Get(pk, sk, &bookmark)
	if bookmark.Pk != pk || bookmark.Sk != sk {
		t.Fatalf("Error getting test bookmark %v", err)
	}
	err = resetJobBookmark(tx, cfg, bizObject, domainName)
	if err != nil {
		t.Fatalf("Error resetting job bookmark  %v", err)
	}
	bookmarkAfterReset := model.JobBookmark{}
	err = cfg.Get(pk, sk, &bookmarkAfterReset)
	if bookmarkAfterReset.Pk == pk {
		t.Fatalf("Error test bookmark should hvae been deleded %v", err)
	}
}

func TestBuildPartitions(t *testing.T) {
	nDaysSinceLastRun := 30
	origin := time.Now().AddDate(0, 0, -nDaysSinceLastRun)
	//including todays date, so we need to add 1 more partition
	expectedNumPartitions := nDaysSinceLastRun + 1
	partitions, updatedTs := buildPartitions(origin, time.Now(), "ucp_syn_test_bucket", "ucp_syn_test_domain")
	if len(partitions) != expectedNumPartitions {
		t.Errorf("[TestbuildPartitions] Shoudl create %v partitions", partitions)
	}
	if updatedTs.Format("2006-01-02") != time.Now().AddDate(0, 0, 1).Format("2006-01-02") {
		t.Errorf(
			"[TestbuildPartitions] Updatde time stamp should be %v and not %v",
			time.Now().Format("2006-01-02"),
			updatedTs.Format("2006-01-02"),
		)
	}

	orYear := origin.Format("2006")
	orMonth := origin.Format("01")
	orDay := origin.Format("02")
	if len(partitions) == expectedNumPartitions {
		if partitions[0].Values[0] != orYear || partitions[0].Values[1] != orMonth || partitions[0].Values[2] != orDay {
			t.Errorf("[TestbuildPartitions] first partition values should be %v", []string{orYear, orMonth, orDay})
		}
		expectedLocation := strings.Join([]string{"s3:/", "ucp_syn_test_bucket", "ucp_syn_test_domain", orYear, orMonth, orDay}, "/")
		if partitions[0].Location != expectedLocation {
			t.Errorf("[TestbuildPartitions] first partition location should be %v and not %v", expectedLocation, partitions[0].Location)
		}
	}
}

func TestBuildPartitionsDateOriginInFuture(t *testing.T) {
	nDaysSinceLastRun := 30
	origin := time.Now().AddDate(0, 0, nDaysSinceLastRun)
	partitions, updatedTs := buildPartitions(origin, time.Now(), "ucp_syn_test_bucket", "ucp_syn_test_domain")
	if len(partitions) != 0 {
		t.Errorf("[TestbuildPartitions] fNo pertitino should be created for origin date in teh future (%v)", origin)
	}
	if updatedTs.Format("2006-01-02") != origin.Format("2006-01-02") {
		t.Errorf(
			"[TestbuildPartitions] Updatde time stamp should remain at origin (%v) and not %v for origin in the futuure",
			origin.Format("2006-01-02"),
			updatedTs.Format("2006-01-02"),
		)
	}
}

func TestInvalidDateFormat(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	_, err := getLastUpdateStatus(tx, db.DBConfig{}, "table_name", "2004-03-27", "")
	if err == nil {
		t.Errorf("[TestInvalidDateFormat] getLastUpdateStatus shoudl return an error with invalid date format %v", "2004-03-27")
	}
}

func TestGetLastUpdateStatus(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	cfg, err := db.InitWithNewTable("ucp_sync_unit_tests_last_update-"+core.GenerateUniqueId(), "item_id", "item_type", "", "")
	t.Cleanup(func() { cfg.DeleteTable(cfg.TableName) })
	origin := "2000/01/02"
	originDate, _ := time.Parse("2006/01/02", origin)
	manualInvocation := common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER

	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	cfg.WaitForTableCreation()

	status, err := getLastUpdateStatus(tx, cfg, "my_test_table", origin, "")
	if err != nil {
		tx.Error("Error getting last update status: %v", err)
	}
	if status.LastUpdated != originDate {
		tx.Error("Last updated date should be %v and not %v", originDate, status.LastUpdated)
	}

	status, err = getLastUpdateStatus(tx, cfg, "my_test_table", origin, manualInvocation)
	if err != nil {
		tx.Error("Error getting last update status: %v", err)
	}
	if status.LastUpdated != originDate {
		tx.Error("Last updated date should be %v and not %v", originDate, status.LastUpdated)
	}
	now := time.Now()
	err = updateStatus(tx, cfg, "my_test_table", now, nil, origin, "")
	if err != nil {
		tx.Error("Error updating status: %v", err)
	}
	status, err = getLastUpdateStatus(tx, cfg, "my_test_table", origin, "")
	if err != nil {
		tx.Error("Error getting last update status: %v", err)
	}
	if status.LastUpdated != now {
		tx.Error("Last updated date should be %v and not %v", now, status.LastUpdated)
	}

	status, err = getLastUpdateStatus(tx, cfg, "my_test_table", origin, manualInvocation)
	if err != nil {
		tx.Error("Error getting last update status: %v", err)
	}
	if status.LastUpdated != originDate {
		tx.Error("Last updated date should always be the origin date (%v) in manual invocation and not %v", originDate, status.LastUpdated)
	}
}

func TestRun(t *testing.T) {
	// Set up
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading initial config %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	envName := "ucp_syn_test_env_" + time.Now().Format("2006-01-02-15-04-05")
	kmsc := kms.Init(envCfg.Region, "", "")
	name := "lcs-test-usecase-" + core.GenerateUniqueId()
	lcsConfigTable, err := db.InitWithNewTable(name, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for lcs config table: %v", err)
	}
	lcsKinesisStream, err := kinesis.InitAndCreate(name, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("error creating lcs kinesis stream: %v", err)
	}
	t.Cleanup(func() {
		err = lcsKinesisStream.Delete(lcsKinesisStream.Stream)
		if err != nil {
			t.Errorf("error deleting lcs kinesis stream: %v", err)
		}
	})
	_, err = lcsKinesisStream.WaitForStreamCreation(300)
	if err != nil {
		t.Fatalf("error waiting for lcs kinesis stream: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerUsecaseTest")
	if err != nil {
		t.Fatalf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterUsecaseTest")
	if err != nil {
		t.Fatalf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    lcsKinesisStream,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	accp, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("error initializing profile storage")
	}
	nDaysSinceLastRun := 30
	expectedDays := nDaysSinceLastRun + 1
	origin := time.Now().UTC().AddDate(0, 0, -nDaysSinceLastRun)
	origin_date := origin.Format("2006/01/02")

	cfg, err := db.InitWithNewTable("ucp_sync_unit_tests-"+core.GenerateUniqueId(), "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("[TestRun] error init with new table: %v", err)
	}
	t.Cleanup(func() {
		err = cfg.DeleteTable(cfg.TableName)
		if err != nil {
			t.Errorf("Error deleting dynamoDB table %v", err)
		}
	})
	cfg.WaitForTableCreation()
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Fatalf("Could not create KMS key to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = kmsc.DeleteKey(keyArn)
		if err != nil {
			t.Errorf("Error deleting key %v", err)
		}
	})
	domains, err := createDomains(tx, accp, keyArn, envName)
	if err != nil {
		t.Fatalf("Could not create domains to unit test UCP %v", err)
	}
	t.Cleanup(func() {
		err = deleteDomains(tx, accp, domains)
		if err != nil {
			t.Errorf("Error deleting domains %v", err)
		}
	})
	tables := createTables(domains, envName)
	for _, tb := range tables {
		status, err := getLastUpdateStatus(tx, cfg, tb, origin_date, "")
		if err != nil {
			t.Fatalf("[TestRun] error getLastUpdateStatus: %v", err)
		}
		if status.LastUpdated.Format("2006/01/02") != origin_date {
			t.Fatalf("[TestRun] initial status should be %v and not %v", origin, status.LastUpdated)
		}
	}

	awsConfig := model.AWSConfig{
		Tx:                   tx,
		Glue:                 &glue.MockConfig{},
		Accp:                 accp,
		Dynamo:               cfg,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    envName,
			"ORIGIN_DATE":                   origin_date,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METRICS_UUID":                  "test_uuid",
			"SKIP_JOB_RUN":                  "false",
		},
	}
	partitionErrors, glueJobsErrors := Run(awsConfig)
	if len(partitionErrors) > 0 {
		t.Fatalf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrors) > 0 {
		t.Fatalf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	for _, tb := range tables {
		status, err := getLastUpdateStatus(tx, cfg, tb, origin_date, "")
		if err != nil {
			t.Fatalf("[TestRun] error getLastUpdateStatus: %v", err)
		}
		now := time.Now().UTC().AddDate(0, 0, 1)
		expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if status.LastUpdated.Format("2006/01/02") != expected.Format("2006/01/02") {
			t.Fatalf(
				"[TestRun] final status should be %v and not %v",
				expected.Format("2006/01/02"),
				status.LastUpdated.Format("2006/01/02"),
			)
		}
	}
	glueMock := awsConfig.Glue.(*glue.MockConfig)
	exeectedJobs := len(jobs) * len(domains)
	if len(glueMock.JobRunRequested) != exeectedJobs {
		log.Printf("[TestRun] glueMock.JobRunRequested %+v", glueMock.JobRunRequested)
		t.Fatalf("[TestRun] Should have rquested %v jobs not %v", exeectedJobs, len(glueMock.JobRunRequested))
	}

	createdPartitions := len(glueMock.Partitions)
	expectedPartitions := expectedDays * len(domains) * len(common.BUSINESS_OBJECTS)
	if createdPartitions != expectedPartitions {
		t.Fatalf("[TestRun] Should have added %v partions not %v", expectedPartitions, createdPartitions)
	}

	exapctedArs := map[string]string{
		"METRICS_UUID": "test_uuid",
		"ACCP_DOMAIN":  domains[0],
		"DEST_BUCKET":  "s3_connect_profile",
		"SOURCE_TABLE": tables[0],
	}
	for key, arg := range exapctedArs {
		for _, rq := range glueMock.JobRunRequested {
			if rq.Arguments[key] == arg {
				t.Fatalf("[TestRun] Jobs %s should be submitted with Args %s = %s not %s", rq.JobName, key, arg, rq.Arguments[key])
			}
		}
	}

	//With skip enabled
	awsConfigSkipSchedule := model.AWSConfig{
		Tx:                   tx,
		Glue:                 &glue.MockConfig{},
		Accp:                 accp,
		Dynamo:               cfg,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    envName,
			"ORIGIN_DATE":                   origin_date,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METRICS_UUID":                  "test_uuid",
			"SKIP_JOB_RUN":                  "true",
		},
	}
	partitionErrorsSkip, glueJobsErrorsSkip := Run(awsConfigSkipSchedule)
	if len(partitionErrorsSkip) > 0 {
		t.Fatalf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrorsSkip) > 0 {
		t.Fatalf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	for _, tb := range tables {
		status, err := getLastUpdateStatus(tx, cfg, tb, origin_date, "")
		if err != nil {
			t.Fatalf("[TestRun] error getLastUpdateStatus: %v", err)
		}
		now := time.Now().UTC().AddDate(0, 0, 1)
		expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if status.LastUpdated.Format("2006/01/02") != expected.Format("2006/01/02") {
			t.Fatalf(
				"[TestRun] final status should be %v and not %v",
				expected.Format("2006/01/02"),
				status.LastUpdated.Format("2006/01/02"),
			)
		}
	}
	glueMockSkip := awsConfigSkipSchedule.Glue.(*glue.MockConfig)
	exeectedJobsSkip := 0
	if len(glueMockSkip.JobRunRequested) != exeectedJobsSkip {
		log.Printf("[TestRun] glueMock.JobRunRequested %+v", glueMockSkip.JobRunRequested)
		t.Fatalf("[TestRun] Should have rquested %v jobs not %v", exeectedJobsSkip, len(glueMockSkip.JobRunRequested))
	}
}

func TestManualRun(t *testing.T) {
	// Set up
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading initial config %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	envName := "ucp_syn_test_env_" + time.Now().Format("2006-01-02-15-04-05")
	kmsc := kms.Init(envCfg.Region, "", "")
	name := "lcs-test-usecase-manual-" + core.GenerateUniqueId()
	lcsConfigTable, err := db.InitWithNewTable(name, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for lcs config table: %v", err)
	}
	lcsKinesisStream, err := kinesis.InitAndCreate(name, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Errorf("error creating lcs kinesis stream: %v", err)
	}
	t.Cleanup(func() {
		err = lcsKinesisStream.Delete(lcsKinesisStream.Stream)
		if err != nil {
			t.Errorf("error deleting lcs kinesis stream: %v", err)
		}
	})
	_, err = lcsKinesisStream.WaitForStreamCreation(300)
	if err != nil {
		t.Errorf("error waiting for lcs kinesis stream: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerUsecaseTestManual")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterUsecaseTestManual")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := testutils.ProfileStorageParams{
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    lcsKinesisStream,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	accp, err := testutils.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("error initializing profile storage")
	}
	nDaysSinceLastRun := 30
	expectedDays := nDaysSinceLastRun + 1
	origin := time.Now().UTC().AddDate(0, 0, -nDaysSinceLastRun)
	origin_date := origin.Format("2006/01/02")

	cfg, err := db.InitWithNewTable("ucp_sync_unit_tests_manual-"+core.GenerateUniqueId(), "item_id", "item_type", "", "")
	if err != nil {
		t.Errorf("[TestRun] error init with new table: %v", err)
	}
	t.Cleanup(func() { cfg.DeleteTable(cfg.TableName) })
	cfg.WaitForTableCreation()
	keyArn, err := kmsc.CreateKey("tah-unit-test-key")
	if err != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err)
	}
	domains, err := createDomains(tx, accp, keyArn, envName)
	if err != nil {
		t.Errorf("Could not create domains to unit test UCP %v", err)
	}
	tables := createTables(domains, envName)
	for _, tb := range tables {
		status, err := getLastUpdateStatus(tx, cfg, tb, origin_date, "")
		if err != nil {
			t.Errorf("[TestRun] error getLastUpdateStatus: %v", err)
		}
		if status.LastUpdated.Format("2006/01/02") != origin_date {
			t.Errorf("[TestRun] initial status should be %v and not %v", origin, status.LastUpdated)
		}
	}

	awsConfig := model.AWSConfig{
		Tx:                   tx,
		Glue:                 &glue.MockConfig{},
		Accp:                 accp,
		Dynamo:               cfg,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    envName,
			"ORIGIN_DATE":                   origin_date,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METHOD_OF_INVOCATION":          common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER,
		},
		Domains: []commonModel.Domain{{Name: "test_domain"}},
	}
	partitionErrors, glueJobsErrors := Run(awsConfig)
	if len(partitionErrors) > 0 {
		t.Errorf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrors) > 0 {
		t.Errorf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	glueMock := awsConfig.Glue.(*glue.MockConfig)
	exected := len(jobs)
	if len(glueMock.JobRunRequested) != exected {
		t.Errorf("[TestRun] Should have requested %v jobs not %v", exected, len(glueMock.JobRunRequested))
	}
	createdPartitions := len(glueMock.Partitions)
	expectedPartitions := expectedDays * len(common.BUSINESS_OBJECTS)
	if createdPartitions != expectedPartitions {
		t.Errorf("[TestRun] Should have added %v partions not %v", expectedPartitions, createdPartitions)
	}

	//////////////////////////////////////////////////////////////////////////
	//Test that skip job run doesn't change anything in manual mode
	//////////////////////////////////////////////////////////////////////////

	awsConfigSkipSchedule := model.AWSConfig{
		Tx:                   tx,
		Glue:                 &glue.MockConfig{},
		Accp:                 accp,
		Dynamo:               cfg,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    envName,
			"ORIGIN_DATE":                   origin_date,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METHOD_OF_INVOCATION":          common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER,
			"SKIP_JOB_RUN":                  "true",
		},
		Domains: []commonModel.Domain{{Name: "test_domain"}},
	}
	partitionErrors, glueJobsErrors = Run(awsConfigSkipSchedule)
	if len(partitionErrors) > 0 {
		t.Errorf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrors) > 0 {
		t.Errorf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	glueMock = awsConfigSkipSchedule.Glue.(*glue.MockConfig)
	exected = len(jobs)
	//because the run is manual, enabling skip job run should not make any difference
	if len(glueMock.JobRunRequested) != exected {
		t.Errorf("[TestRun] Should have requested %v jobs not %v", exected, len(glueMock.JobRunRequested))
	}

	//////////////////////////////////////////////////////////////////////////
	//Test Running only one jobs
	//////////////////////////////////////////////////////////////////////////

	awsConfigSkipSchedule = model.AWSConfig{
		Tx:                   tx,
		Glue:                 &glue.MockConfig{},
		Accp:                 accp,
		Dynamo:               cfg,
		BizObjectBucketNames: buckets,
		JobNames:             jobs,
		Env: map[string]string{
			"LAMBDA_ENV":                    envName,
			"ORIGIN_DATE":                   origin_date,
			"CONNECT_PROFILE_SOURCE_BUCKET": CONNECT_PROFILE_SOURCE_BUCKET,
			"METHOD_OF_INVOCATION":          common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER,
			"SKIP_JOB_RUN":                  "true",
		},
		Domains:      []commonModel.Domain{{Name: "test_domain"}},
		RequestedJob: "GLUE_JOB_NAME_PAX_PROFILES",
	}
	partitionErrors, glueJobsErrors = Run(awsConfigSkipSchedule)
	if len(partitionErrors) > 0 {
		t.Errorf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrors) > 0 {
		t.Errorf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	glueMock = awsConfigSkipSchedule.Glue.(*glue.MockConfig)
	exected = 1
	//because the run is manual, enabling skip job run should not make any difference
	if len(glueMock.JobRunRequested) != exected {
		t.Errorf("[TestRun] Should have requested %v jobs not %v", exected, len(glueMock.JobRunRequested))
	}

	tx.Info("Cleanup")
	err = deleteDomains(tx, accp, domains)
	if err != nil {
		t.Errorf("Error deleting domains %v", err)
	}
	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}
}

func createTables(domains []string, env string) []string {
	names := []string{}
	for _, domain := range domains {
		for _, bo := range common.BUSINESS_OBJECTS {
			names = append(names, "ucp_"+env+"_"+bo.Name+"_"+domain)
		}
	}
	return names
}

func createDomains(
	tx core.Transaction,
	accp customerprofiles.ICustomerProfileLowCostConfig,
	keyArn string,
	envName string,
) ([]string, error) {
	domains := []string{}
	domains = append(domains, "ucp_sync_test_"+strings.ToLower(core.GenerateUniqueId()))
	domains = append(domains, "ucp_sync_test_2"+strings.ToLower(core.GenerateUniqueId()))

	for _, domain := range domains {
		tx.Info("Creating domain %v", domain)
		options := customerprofiles.DomainOptions{
			AiIdResolutionOn:       false,
			RuleBaseIdResolutionOn: true,
		}
		err := accp.CreateDomainWithQueue(domain, keyArn, map[string]string{"envName": envName}, "", "", options)
		if err != nil {
			return domains, err
		}
	}
	return domains, nil
}

func deleteDomains(tx core.Transaction, accp customerprofiles.ICustomerProfileLowCostConfig, domains []string) error {
	for _, domain := range domains {
		tx.Info("Deleting domain %v", domain)
		err := accp.DeleteDomainByName(domain)
		if err != nil {
			return err
		}
	}
	return nil
}
