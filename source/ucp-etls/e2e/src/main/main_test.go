// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	core "tah/upt/source/tah-core/core"
	db "tah/upt/source/tah-core/db"
	glue "tah/upt/source/tah-core/glue"
	iam "tah/upt/source/tah-core/iam"
	s3 "tah/upt/source/tah-core/s3"
	sqs "tah/upt/source/tah-core/sqs"

	awsglue "github.com/aws/aws-sdk-go/service/glue"

	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

var BOOKMARK_PK = "item_id"
var BOOKMARK_SK = "item_type"
var METRICS_UUID = "1111-2222-3333-4444"

type BusinessObjectTestConfig struct {
	ObjectName    string
	GlueJob       string
	TestFilePath  string
	TestFiles     []string
	GlueTableName string
	SourceBucket  string
	CrawlerName   string
	Expected      []Expected
}

type Expected struct {
	Prefix   string
	NCsv     int
	NRecords int
}

type Bookmark struct {
	Pk              string `json:"item_id"`
	Sk              string `json:"item_type"`
	Bookmark        string `json:"bookmark"`
	RecordProcessed int    `json:"records_processed"`
}

func TestMain(t *testing.T) {
	// Set up
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("error loading initial config %v", err)
	}
	domain := "etl_e2e_test_" + strings.ToLower(core.GenerateUniqueId())
	sqsClient := sqs.Init(envCfg.Region, "", "")
	queueUrl, err := sqsClient.Create("test-Queue-" + time.Now().Format("15-04-05"))
	if err != nil {
		t.Errorf("[TestMain]error creating queue: %v", err)
	}
	glueClient := glue.Init(envCfg.Region, infraCfg.GlueDbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	// we should not use the real bucket here not to generate thousand of errors since we don't create a domain
	targetBucketHandler, err := s3.InitWithRandBucket("etl-e2e-test", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[TestMain]error creating target bucket: %v", err)
	}
	iamHandler := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	policyArn, err := iamHandler.CreatePolicy(domain+"_s3_access", iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"s3:*"},
				Resource: []string{"arn:aws:s3:::" + targetBucketHandler.Bucket + "/*", "arn:aws:s3:::" + targetBucketHandler.Bucket},
			},
			{
				Effect:   "Allow",
				Action:   []string{"sqs:*"},
				Resource: []string{"*"},
			},
		},
	})
	if err != nil {
		t.Fatalf("[TestMain]error creating target bucket policy: %v", err)
	}
	err = iamHandler.AttachRolePolicy(infraCfg.GlueRoleName, policyArn)
	if err != nil {
		t.Fatalf("[TestMain]error attaching target bucket policy: %v", err)
	}
	t.Cleanup(func() {
		err := iamHandler.DetachRolePolicy(infraCfg.GlueRoleName, policyArn)
		if err != nil {
			t.Errorf("[TestMain]error detaching target bucket policy: %v", err)
		}
		err = iamHandler.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[TestMain]error deleting target bucket policy: %v", err)
		}
		err = targetBucketHandler.EmptyBucket()
		if err != nil {
			t.Errorf("[%s] error emptying bucket: %v", t.Name(), err)
		}
		err = targetBucketHandler.DeleteBucket(targetBucketHandler.Bucket)
		if err != nil {
			t.Errorf("[TestMain]error deleting target bucket: %v", err)
		}
	})

	dynamodbClient := db.Init(infraCfg.ConfigTableName, BOOKMARK_PK, BOOKMARK_SK, "", "")
	pk := "glue_job_bookmark"
	sk := "clickstream_" + domain
	sk2 := "hotel_booking_" + domain
	data := Bookmark{}

	year := "2022"
	month := "12"
	day := "01"
	hour := "09"

	bizObjectConfigs := []BusinessObjectTestConfig{
		{
			ObjectName:   "air_booking",
			GlueJob:      infraCfg.GlueJobNameAirBooking,
			TestFilePath: "../../../../test_data/air_booking/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketAirBooking,
			CrawlerName:   "glue_e2e_tests_air_booking",
			Expected: []Expected{{
				Prefix:   "air_booking",
				NCsv:     2,
				NRecords: 43,
			},
				{
					Prefix:   "ancillary_service",
					NCsv:     2,
					NRecords: 30,
				},
			},
		},
		{
			ObjectName:   "clickstream",
			GlueJob:      infraCfg.GlueJobNameClickstream,
			TestFilePath: "../../../../test_data/clickstream/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
				"customer_1.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketClickstream,
			CrawlerName:   "glue_e2e_tests_clickstream",
			Expected: []Expected{
				{
					Prefix:   "clickstream",
					NCsv:     1,
					NRecords: 13,
				},
			},
		},
		{
			ObjectName:   "guest_profile",
			GlueJob:      infraCfg.GlueJobNameGuestProfile,
			TestFilePath: "../../../../test_data/guest_profile/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
				"customer_1.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketGuestProfile,
			CrawlerName:   "glue_e2e_tests_guest_profile",
			Expected: []Expected{{
				Prefix:   "guest_profile",
				NCsv:     3,
				NRecords: 263,
			},
				{
					Prefix:   "loyalty_transaction",
					NCsv:     4,
					NRecords: 1945,
				},
			},
		},
		{
			ObjectName:   "hotel_booking",
			GlueJob:      infraCfg.GlueJobNameHotelBooking,
			TestFilePath: "../../../../test_data/hotel_booking/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
				"customer_1.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketHotelBooking,
			CrawlerName:   "glue_e2e_tests_hotel_booking",
			Expected: []Expected{{
				Prefix:   "hotel_booking",
				NCsv:     3,
				NRecords: 27,
			},
			},
		},
		{
			ObjectName:   "hotel_stay",
			GlueJob:      infraCfg.GlueJobNameHotelStay,
			TestFilePath: "../../../../test_data/hotel_stay/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketHotelStay,
			CrawlerName:   "glue_e2e_tests_hotel_stay",
			Expected: []Expected{{
				Prefix:   "hotel_stay_revenue_items",
				NCsv:     2,
				NRecords: 22,
			},
			},
		},
		{
			ObjectName:   "customer_service_interaction",
			GlueJob:      infraCfg.GlueJobNameCsi,
			TestFilePath: "../../../../test_data/customer_service_interaction/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketCsi,
			Expected: []Expected{{
				Prefix:   "customer_service_interaction",
				NCsv:     2,
				NRecords: 6,
			},
			},
		},
		{
			ObjectName:   "pax_profile",
			GlueJob:      infraCfg.GlueJobNamePaxProfile,
			TestFilePath: "../../../../test_data/pax_profile/",
			TestFiles: []string{
				"data1.jsonl",
				"data2.jsonl",
			},
			GlueTableName: "",
			SourceBucket:  infraCfg.BucketPaxProfile,
			CrawlerName:   "glue_e2e_tests_pax_profile",
			Expected: []Expected{{
				Prefix:   "pax_profile",
				NCsv:     2,
				NRecords: 206,
			},
				{
					Prefix:   "loyalty_transaction",
					NCsv:     4,
					NRecords: 1945,
				},
			},
		},
	}

	// Run ETL jobs
	var wg sync.WaitGroup
	testErrs := []string{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors
	wg.Add(len(bizObjectConfigs))
	tx := core.NewTransaction("etl_e2e", "", core.LogLevelDebug)
	for i := range bizObjectConfigs {
		log.Printf("[%v] 0-Create Glue table for object", bizObjectConfigs[i].ObjectName)
		glueTableName := "ucp_" + bizObjectConfigs[i].ObjectName + "_" + "etl_e2e_test_domain_" + strings.ToLower(core.GenerateUniqueId())
		bizObjectConfigs[i].GlueTableName = glueTableName
		log.Printf("[%v] 0.1-Loading schema file", bizObjectConfigs[i].ObjectName)
		schema, err0 := LoadSchema(&tx, bizObjectConfigs[i].ObjectName)
		if err0 != nil {
			log.Printf("Table Creation failed: %v", err0)
			t.Errorf("[MainTest] Could not fetch glue schemas: %v", err0)
			return
		}
		log.Printf("[%v] 0.1-Creating Table", bizObjectConfigs[i].ObjectName)
		err0 = glueClient.CreateTable(glueTableName, bizObjectConfigs[i].SourceBucket, []glue.PartitionKey{{Name: "year", Type: "int"}, {Name: "month", Type: "int"}, {Name: "day", Type: "int"}}, schema)
		if err0 != nil {
			log.Printf("Table Creation failed: %v", err0)
			t.Errorf("[MainTest] Could not create glue table for testing: %v", err0)
			return
		}
		go func(index int, c BusinessObjectTestConfig) {
			defer wg.Done()
			log.Printf("[%v] 1-Starting e2e tests", c.ObjectName)
			sourceBucketHandler := s3.Init(c.SourceBucket, "", envCfg.Region, "", "")

			log.Printf("[%v] 2-Uploading data to s3://%s/%s", c.ObjectName, c.SourceBucket, strings.Join([]string{year, month, day, hour}, "/"))
			for _, file := range c.TestFiles {
				err := sourceBucketHandler.UploadFile(year+"/"+month+"/"+day+"/"+hour+"/"+file, c.TestFilePath+file)
				if err != nil {
					testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] Could not upload files: %v", c.ObjectName, err))
					cancel()
					return
				}
			}
			select {
			case <-ctx.Done():
				log.Printf("[%v] Error in another test. stopping", c.ObjectName)
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}

			log.Printf("[%v] 3-Add Partitions to table", c.ObjectName)
			errs := glueClient.AddParquetPartitionsToTable(c.GlueTableName, []glue.Partition{{
				Values:   []string{year, month, day},
				Location: "s3://" + c.SourceBucket + "/" + strings.Join([]string{year, month, day}, "/"),
			}})
			if len(errs) > 0 {
				for _, e := range errs {
					testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error adding partition: %v", c.ObjectName, e))
				}
				cancel()
				return
			}
			select {
			case <-ctx.Done():
				log.Printf("[%v] Error in another test. stopping", c.ObjectName)
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}
			log.Printf("[%v] 6-Run job with modified source and targets", c.ObjectName)
			err := glueClient.RunJob(c.GlueJob, map[string]string{
				"--DEST_BUCKET":     targetBucketHandler.Bucket,
				"--SOURCE_TABLE":    c.GlueTableName,
				"--ERROR_QUEUE_URL": queueUrl,
				"--ACCP_DOMAIN":     domain,
				"--BIZ_OBJECT_NAME": c.ObjectName,
				"--DYNAMO_TABLE":    infraCfg.ConfigTableName,
				"--METRICS_UUID":    METRICS_UUID,
			})
			if err != nil {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error running job: %v", c.ObjectName, err))
				cancel()
				return
			}
			select {
			case <-ctx.Done():
				log.Printf("[%v] Error in another test. stopping", c.ObjectName)
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}
			log.Printf("[%v] 7-Wait for job run", c.ObjectName)

			status, err2 := WaitForJobRun(glueClient, domain, c.GlueJob, 300)
			if err2 != nil {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error waiting for job completion: %v", c.ObjectName, err2))
				cancel()
				return
			}
			if status != "SUCCEEDED" {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] Job completed with non success status %v", c.ObjectName, status))
				cancel()
				return
			}
			select {
			case <-ctx.Done():
				log.Printf("[%v] Error in another test. stopping", c.ObjectName)
				return // Error somewhere, terminate
			default: // Default is must to avoid blocking
			}
		}(i, bizObjectConfigs[i])
	}
	wg.Wait()
	log.Printf("Checking CSV files for all objects")
	for _, c := range bizObjectConfigs {
		log.Printf("[%v] Check csv File", c.ObjectName)
		for _, expected := range c.Expected {
			nRecs := 0
			log.Printf("[%v] Check expected csv in folder: %s", c.ObjectName, expected.Prefix)
			csvs, err1 := targetBucketHandler.Search(domain+"/"+expected.Prefix, 500)
			if err1 != nil {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error listing s3 buckets after ETLs: %v", c.ObjectName, err1))
			}
			csvs = filterCsvFiles(csvs)
			log.Printf("[TestGlue][%v] CSVs: %v", c.ObjectName, csvs)
			if len(csvs) != expected.NCsv {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] Expected %v csv files in %s, got %v", c.ObjectName, expected.NCsv, expected.Prefix, len(csvs)))
			}
			if len(csvs) > 0 {
				for _, csv := range csvs {
					data, err2 := targetBucketHandler.ParseCsvFromS3(csv)
					if err2 != nil {
						testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error listing s3 buckets after ETLs: %v", err2, c.ObjectName))
					}

					log.Printf("CSV has %v rows", len(data))
					nRecs += len(data)
					if len(data) == 0 {
						testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] invalid ETL output: should have at least one row", c.ObjectName))
					}

					columns := map[string]bool{}
					for _, col := range data[0] {
						columns[col] = true
					}
					//Looking for mandatory columns
					mandatoryColumns := []string{"traveller_id", "last_updated", "model_version", "accp_object_id"}
					log.Printf("[TestGlue][%v] Checking for mandatory columns %v in CSV file", c.ObjectName, mandatoryColumns)
					for _, colName := range mandatoryColumns {
						if !columns[colName] {
							testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] CSV is missing mandatory column '%v'", c.ObjectName, colName))
						}
					}
				}
			} else {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] No data created in CSV file", c.ObjectName))
			}
			if nRecs != expected.NRecords {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] Expected %v records in CSV file for %s, got %v", c.ObjectName, expected.NRecords, expected.Prefix, nRecs))
			}
		}
	}

	content, err := sqsClient.Get(sqs.GetMessageOptions{})
	if err != nil {
		t.Errorf("[TestMain] Error getting error queue content: %v", err)
	}
	if content.NMessages > 0 {
		log.Printf("SQS queue is not empty. content: %+v", content.Peek)
		t.Errorf("[TestMain] Error during ingestion: SQS has %v messages", content.NMessages)
	}

	if len(testErrs) > 0 {
		for _, testErr := range testErrs {
			t.Errorf(testErr)
		}
		t.Fatalf("[%s] see above errors", t.Name())
	} else {
		log.Printf("ETL succeeded for all business objects")
	}

	// Validate bookmark has been set
	err = dynamodbClient.Get(pk, sk, &data)
	if err != nil || data.Bookmark == "" {
		t.Fatalf("[TestMain] Error: unable to get initial bookmark data")
	} else {
		log.Printf("Initial bookmark data: %+v", data)
	}

	// Store initial count to compare against second run
	initialCount := data.RecordProcessed

	// Run one job again, expect only data from partitions since last run to be included
	log.Printf("Testing incremental job run with a bookmark")
	ts := time.Now().UTC().Format("2006-01-02-15")
	split := strings.Split(ts, "-")
	year2 := split[0]
	month2 := split[1]
	day2 := split[2]
	hour2 := split[3]
	c2 := BusinessObjectTestConfig{
		ObjectName:   "clickstream",
		GlueJob:      infraCfg.GlueJobNameClickstream,
		TestFilePath: "../../../../test_data/clickstream/",
		TestFiles: []string{
			"data1.jsonl",
		},
		GlueTableName: bizObjectConfigs[1].GlueTableName,
		SourceBucket:  infraCfg.BucketClickstream,
		CrawlerName:   "glue_e2e_tests_clickstream",
	}
	// Upload data to S3
	sourceBucketHandler := s3.Init(c2.SourceBucket, "", envCfg.Region, "", "")
	log.Printf("[MainTest] Uploading test file to S3: %v", year2+"/"+month2+"/"+day2+"/"+hour2+"/"+c2.TestFiles[0])
	err = sourceBucketHandler.UploadFile(year2+"/"+month2+"/"+day2+"/"+hour2+"/"+c2.TestFiles[0], c2.TestFilePath+c2.TestFiles[0])
	if err != nil {
		t.Errorf("[MainTest] Error uploading test file: %v", err)
	}
	// Create Glue table partition
	errs := glueClient.AddParquetPartitionsToTable(c2.GlueTableName, []glue.Partition{{
		Values:   []string{year2, month2, day2},
		Location: "s3://" + c2.SourceBucket + "/" + strings.Join([]string{year2, month2, day2}, "/"),
	}})
	if len(errs) > 0 {
		t.Errorf("[MainTest] Error(s) adding glue partitions: %v", errs)
	} else {
		log.Printf("Partition %v/%v/%v added to table %v", year2, month2, day2, c2.GlueTableName)
		parts, err2 := glueClient.GetPartitionsFromTable(c2.GlueTableName)
		if err2 != nil {
			t.Errorf("[MainTest] could not get partitions: %v", err2)
		}
		log.Printf("Partitions for table %v: %+v", c2.GlueTableName, parts)
	}
	log.Printf("Waiting 30 seconds for propagation")
	time.Sleep(30 * time.Second)
	// Run job
	err = glueClient.RunJob(c2.GlueJob, map[string]string{
		"--DEST_BUCKET":     targetBucketHandler.Bucket,
		"--SOURCE_TABLE":    c2.GlueTableName,
		"--ERROR_QUEUE_URL": queueUrl,
		"--ACCP_DOMAIN":     domain,
		"--BIZ_OBJECT_NAME": c2.ObjectName,
		"--DYNAMO_TABLE":    infraCfg.ConfigTableName,
		"--METRICS_UUID":    METRICS_UUID,
	})
	if err != nil {
		t.Errorf("[MainTest] Error running Glue job: %v", err)
	}
	// Wait for job to complete
	status, err := WaitForJobRun(glueClient, domain, c2.GlueJob, 300)

	if err != nil {
		t.Errorf("[MainTest] Error with Glue job: %v", err)
	}
	if status != "SUCCEEDED" {
		t.Errorf("[MainTest] Second Glue job run did not succeed")
	}
	// Verify expected records processed
	err = dynamodbClient.Get(pk, sk, &data)
	if err != nil || data.Bookmark == "" {
		t.Errorf("[TestMain] Error: unable to get bookmark data")
	} else {
		log.Printf("Second run bookmark data: %+v", data)
	}
	// Validate count of records processed is same as expected
	count := data.RecordProcessed
	if count == 0 || count >= initialCount {
		t.Errorf("[TestMain] The count of processed records on second run  (%v)  should be greater than the initial count (%v)", count, initialCount)
	}

	//Graceful Exit
	c3 := BusinessObjectTestConfig{
		ObjectName:    "hotel_booking",
		GlueJob:       infraCfg.GlueJobNameHotelBooking,
		GlueTableName: bizObjectConfigs[3].GlueTableName,
		SourceBucket:  infraCfg.BucketHotelBooking,
		CrawlerName:   "glue_e2e_tests_hotel_booking",
	}
	// Create Glue table partition
	errs = glueClient.AddParquetPartitionsToTable(c3.GlueTableName, []glue.Partition{{
		Values:   []string{year2, month2, day2},
		Location: "s3://" + c3.SourceBucket + "/" + strings.Join([]string{year2, month2, day2}, "/"),
	}})
	if len(errs) > 0 {
		t.Errorf("[MainTest] Error(s) adding glue partitions: %v", errs)
	} else {
		log.Printf("Partition %v/%v/%v added to table %v", year2, month2, day2, c3.GlueTableName)
		parts, err2 := glueClient.GetPartitionsFromTable(c3.GlueTableName)
		if err2 != nil {
			t.Errorf("[MainTest] could not get partitions: %v", err2)
		}
		log.Printf("Partitions for table %v: %+v", c3.GlueTableName, parts)
	}
	// Run job
	attr := map[string]string{
		"--DEST_BUCKET":     targetBucketHandler.Bucket,
		"--SOURCE_TABLE":    c3.GlueTableName,
		"--ERROR_QUEUE_URL": queueUrl,
		"--ACCP_DOMAIN":     domain,
		"--BIZ_OBJECT_NAME": c3.ObjectName,
		"--DYNAMO_TABLE":    infraCfg.ConfigTableName,
		"--METRICS_UUID":    METRICS_UUID,
	}
	log.Printf("Running job %v with attributes %+v", c3.GlueJob, attr)
	err = glueClient.RunJob(c3.GlueJob, attr)
	if err != nil {
		t.Errorf("[MainTest] Error running Glue job: %v", err)
	}
	// Wait for job to complete
	status, err = WaitForJobRun(glueClient, domain, c3.GlueJob, 300)
	if err != nil {
		t.Errorf("[MainTest] Error with Glue job: %v", err)
	}
	if status != "SUCCEEDED" {
		t.Errorf("[MainTest] Glue job run did not succeed")
	}
	// Verify expected records processed
	err = dynamodbClient.Get(pk, sk2, &data)
	if err != nil || data.Bookmark == "" {
		t.Errorf("[TestMain] Error: unable to get bookmark data")
	} else {
		log.Printf("Bookmark data: %+v", data)
	}
	// Validate count of records processed is same as expected
	count = data.RecordProcessed
	if count != 0 {
		t.Errorf("[TestMain] The count of processed records (%v) should be 0", count)
	}

	// Clean up test data
	log.Printf("Cleaning up")
	for _, c := range bizObjectConfigs {
		sourceBucketHandler := s3.Init(c.SourceBucket, "", envCfg.Region, "", "")
		log.Printf("[%v] Cleanup", c.ObjectName)
		log.Printf("[%v] Delete table %s", c.ObjectName, c.GlueTableName)
		err := glueClient.DeleteTable(c.GlueTableName)
		if err != nil {
			t.Errorf("[TestGlue][%v] Error deleting table %v", c.GlueTableName, errs)
		}
		err = sourceBucketHandler.EmptyBucket()
		if err != nil {
			t.Errorf("[TestGlue][%v] Error emptying bucket %v", c.ObjectName, err)
		}
	}
	err = targetBucketHandler.EmptyBucket()
	if err != nil {
		t.Errorf("[TestGlue]Error emptying target bucket %v", err)
	}

	err = sqsClient.Delete()
	if err != nil {
		t.Errorf("[TestGlue]error deleting queue: %v", err)
	}
	for _, cfg := range bizObjectConfigs {
		err := dynamodbClient.DeleteByKey(pk, cfg.ObjectName+"_"+domain)
		if err != nil {
			t.Errorf("[TestMain] Error deleting job bookmark for %v: %v", cfg.ObjectName, err)
		}
	}
}

var SchemaPathMap = map[string]string{
	"air_booking":                  "tah-common-glue-schemas/air_booking.glue.json",
	"clickstream":                  "tah-common-glue-schemas/clickevent.glue.json",
	"guest_profile":                "tah-common-glue-schemas/guest_profile.glue.json",
	"hotel_booking":                "tah-common-glue-schemas/hotel_booking.glue.json",
	"hotel_stay":                   "tah-common-glue-schemas/hotel_stay_revenue.glue.json",
	"pax_profile":                  "tah-common-glue-schemas/pax_profile.glue.json",
	"customer_service_interaction": "tah-common-glue-schemas/customer_service_interaction.glue.json",
}

func LoadSchema(tx *core.Transaction, bizObjectName string) (glue.Schema, error) {
	tx.Info("Loading schema for: %s", bizObjectName)
	projectRootDir, err := config.AbsolutePathToProjectRootDir()
	if err != nil {
		return glue.Schema{}, err
	}
	relativeTahCommonPath := "source/tah-common/"
	filePath := projectRootDir + relativeTahCommonPath + SchemaPathMap[bizObjectName]
	tx.Info("Reading file %s", filePath)
	schemaBytes, err := os.ReadFile(filePath)
	if err != nil {
		return glue.Schema{}, err
	}
	schema, err := glue.ParseSchema(string(schemaBytes))
	if err != nil {
		return glue.Schema{}, err
	}
	return schema, nil
}

func filterCsvFiles(folders []string) []string {
	filtered := make([]string, 0)
	for _, f := range folders {
		if strings.HasSuffix(f, ".csv") {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// we can't use the upt SDK here because to override the domain parameters. we should move this code to full e2e on a domain
func WaitForJobRun(glueHandler glue.Config, domainName, jobName string, timeoutSeconds int) (string, error) {
	var waitDelay = 5
	log.Printf("Waiting for job %s run to complete", jobName)
	_, status, err := glueHandler.GetJobRunStatusWithFilter(jobName, map[string]string{"--ACCP_DOMAIN": domainName})
	if err != nil {
		return glue.JOB_RUN_STATUS_UNKNOWN, err
	}
	it := 0
	for status != awsglue.JobRunStateSucceeded {
		log.Printf("Job %s run Status: %v Waiting 5 seconds before checking again", jobName, status)
		time.Sleep(time.Duration(waitDelay) * time.Second)
		_, status, err = glueHandler.GetJobRunStatusWithFilter(jobName, map[string]string{"--ACCP_DOMAIN": domainName})
		if err != nil {
			return glue.JOB_RUN_STATUS_UNKNOWN, err
		}
		if status == "FAILED" {
			log.Printf("Job %s run Status: %v. Completed", jobName, status)
			return status, nil
		}
		it += 1
		if it*waitDelay >= timeoutSeconds {
			return glue.JOB_RUN_STATUS_UNKNOWN, fmt.Errorf("Job %s wait timed out after %v seconds", jobName, it*5)
		}
	}
	log.Printf("Job %s run Status: %v. Completed", jobName, status)
	return status, err
}
