package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	db "tah/core/db"
	glue "tah/core/glue"
	s3 "tah/core/s3"
	sqs "tah/core/sqs"
	"testing"
	"time"
)

var UCP_REGION = getRegion()
var DYNAMODB_CONFIG_TABLE_NAME = os.Getenv("DYNAMODB_CONFIG_TABLE_NAME")
var GLUE_ROLE_NAME = os.Getenv("GLUE_ROLE_NAME")
var GLUE_DB_NAME = os.Getenv("GLUE_DB_NAME")
var TEST_BUCKET_AIR_BOOKING = os.Getenv("TEST_BUCKET_AIR_BOOKING")
var TEST_BUCKET_HOTEL_BOOKINGS = os.Getenv("TEST_BUCKET_HOTEL_BOOKINGS")
var TEST_BUCKET_PAX_PROFILES = os.Getenv("TEST_BUCKET_PAX_PROFILES")
var TEST_BUCKET_GUEST_PROFILES = os.Getenv("TEST_BUCKET_GUEST_PROFILES")
var TEST_BUCKET_STAY_REVENUE = os.Getenv("TEST_BUCKET_STAY_REVENUE")
var TEST_BUCKET_CLICKSTREAM = os.Getenv("TEST_BUCKET_CLICKSTREAM")
var TEST_BUCKET_ACCP_IMPORT = os.Getenv("TEST_BUCKET_ACCP_IMPORT")
var TEST_TABLE_AIR_BOOKING = os.Getenv("TEST_TABLE_AIR_BOOKING")
var TEST_TABLE_HOTEL_BOOKINGS = os.Getenv("TEST_TABLE_HOTEL_BOOKINGS")
var TEST_TABLE_PAX_PROFILES = os.Getenv("TEST_TABLE_PAX_PROFILES")
var TEST_TABLE_GUEST_PROFILES = os.Getenv("TEST_TABLE_GUEST_PROFILES")
var TEST_TABLE_STAY_REVENUE = os.Getenv("TEST_TABLE_STAY_REVENUE")
var TEST_TABLE_CLICKSTREAM = os.Getenv("TEST_TABLE_CLICKSTREAM")
var GLUE_JOB_NAME_AIR_BOOKING = os.Getenv("GLUE_JOB_NAME_AIR_BOOKING")
var GLUE_JOB_NAME_HOTEL_BOOKINGS = os.Getenv("GLUE_JOB_NAME_HOTEL_BOOKINGS")
var GLUE_JOB_NAME_PAX_PROFILES = os.Getenv("GLUE_JOB_NAME_PAX_PROFILES")
var GLUE_JOB_NAME_GUEST_PROFILES = os.Getenv("GLUE_JOB_NAME_GUEST_PROFILES")
var GLUE_JOB_NAME_STAY_REVENUE = os.Getenv("GLUE_JOB_NAME_STAY_REVENUE")
var GLUE_JOB_NAME_CLICKSTREAM = os.Getenv("GLUE_JOB_NAME_CLICKSTREAM")

var BOOKMARK_PK = "item_id"
var BOOKMARK_SK = "item_type"

type BusinessObjectTestConfig struct {
	ObjectName    string
	GlueJob       string
	TestFilePath  string
	TestFiles     []string
	GlueTableName string
	SourceBucket  string
	CrawlerName   string
	TargetPrefix  string
	Multiline     bool
}

func TestMain(t *testing.T) {
	// Set up
	glueClient := glue.Init(UCP_REGION, GLUE_DB_NAME)
	targetBucketHandler := s3.Init(TEST_BUCKET_ACCP_IMPORT, "", UCP_REGION)
	sqsClient := sqs.Init(UCP_REGION)
	domain := "etl_e2e_test_domain"
	queueUrl, err := sqsClient.Create("test-Queue-" + time.Now().Format("15-04-05"))
	if err != nil {
		t.Errorf("[TestMain]error creating queue: %v", err)
	}
	dynamodbClient := db.Init(DYNAMODB_CONFIG_TABLE_NAME, BOOKMARK_PK, BOOKMARK_SK)
	pk := "glue_job_bookmark"
	sk := "clickstream_" + domain
	data := make(map[string]interface{})

	year := "2022"
	month := "12"
	day := "01"
	hour := "09"
	bizObjectConfigs := []BusinessObjectTestConfig{
		BusinessObjectTestConfig{
			ObjectName:   "air_booking",
			GlueJob:      GLUE_JOB_NAME_AIR_BOOKING,
			TestFilePath: "../../../../test_data/air_booking/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_AIR_BOOKING,
			SourceBucket:  TEST_BUCKET_AIR_BOOKING,
			CrawlerName:   "glue_e2e_tests_air_booking",
			TargetPrefix:  "air_booking",
		},
		BusinessObjectTestConfig{
			ObjectName:   "clickstream",
			GlueJob:      GLUE_JOB_NAME_CLICKSTREAM,
			TestFilePath: "../../../../test_data/clickstream/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_CLICKSTREAM,
			SourceBucket:  TEST_BUCKET_CLICKSTREAM,
			CrawlerName:   "glue_e2e_tests_clickstream",
			TargetPrefix:  "clickstream",
			Multiline:     true,
		},
		BusinessObjectTestConfig{
			ObjectName:   "guest_profile",
			GlueJob:      GLUE_JOB_NAME_GUEST_PROFILES,
			TestFilePath: "../../../../test_data/guest_profile/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_GUEST_PROFILES,
			SourceBucket:  TEST_BUCKET_GUEST_PROFILES,
			CrawlerName:   "glue_e2e_tests_guest_profile",
			TargetPrefix:  "guest_profile",
		},
		BusinessObjectTestConfig{
			ObjectName:   "hotel_booking",
			GlueJob:      GLUE_JOB_NAME_HOTEL_BOOKINGS,
			TestFilePath: "../../../../test_data/hotel_booking/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_HOTEL_BOOKINGS,
			SourceBucket:  TEST_BUCKET_HOTEL_BOOKINGS,
			CrawlerName:   "glue_e2e_tests_hotel_booking",
			TargetPrefix:  "hotel_booking",
		},
		BusinessObjectTestConfig{
			ObjectName:   "hotel_stay",
			GlueJob:      GLUE_JOB_NAME_STAY_REVENUE,
			TestFilePath: "../../../../test_data/hotel_stay/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_STAY_REVENUE,
			SourceBucket:  TEST_BUCKET_STAY_REVENUE,
			CrawlerName:   "glue_e2e_tests_hotel_stay",
			TargetPrefix:  "hotel_stay_revenue_items",
		},
		BusinessObjectTestConfig{
			ObjectName:   "pax_profile",
			GlueJob:      GLUE_JOB_NAME_PAX_PROFILES,
			TestFilePath: "../../../../test_data/pax_profile/",
			TestFiles: []string{
				"data1.json",
				"data2.json",
			},
			GlueTableName: TEST_TABLE_PAX_PROFILES,
			SourceBucket:  TEST_BUCKET_PAX_PROFILES,
			CrawlerName:   "glue_e2e_tests_pax_profile",
			TargetPrefix:  "pax_profile",
		},
	}

	// Run ETL jobs
	var wg sync.WaitGroup
	testErrs := []string{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure it's called to release resources even if no errors
	wg.Add(len(bizObjectConfigs))
	for i, config := range bizObjectConfigs {
		go func(index int, c BusinessObjectTestConfig) {
			defer wg.Done()
			log.Printf("[%v] 1-Starting e2e tests", c.ObjectName)
			sourceBucketHandler := s3.Init(c.SourceBucket, "", UCP_REGION)

			log.Printf("[%v] 2-Uploading data to s3://%s/%s", c.ObjectName, c.SourceBucket, strings.Join([]string{year, month, day, hour}, "/"))
			for _, file := range c.TestFiles {
				//Fo single linek json object we unprerttyfy them since gluue cann't process pererttified Json.
				//For multiline jsonl files we keep them as is. For noo only c
				if !c.Multiline {
					unprettyfy(c.TestFilePath + file)
				}
				err := sourceBucketHandler.UploadFile(year+"/"+month+"/"+day+"/"+hour+"/"+file, c.TestFilePath+file)
				if err != nil {
					testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] Cound not upload files: %v", c.ObjectName, err))
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
			errs := glueClient.AddPartitionsToTable(c.GlueTableName, []glue.Partition{glue.Partition{
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
				"--DEST_BUCKET":     TEST_BUCKET_ACCP_IMPORT,
				"--SOURCE_TABLE":    c.GlueTableName,
				"--ERROR_QUEUE_URL": queueUrl,
				"--ACCP_DOMAIN":     domain,
				"--BIZ_OBJECT_NAME": c.ObjectName,
				"--DYNAMO_TABLE":    DYNAMODB_CONFIG_TABLE_NAME,
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
			status, err2 := glueClient.WaitForJobRun(c.GlueJob, 600)
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
			log.Printf("[%v] 8-Check csv File", c.ObjectName)
			csvs, err1 := targetBucketHandler.Search(domain+"/"+c.TargetPrefix, 500)
			if err1 != nil {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error listing s3 buckets after ETLs: %v", c.ObjectName, err1))
				cancel()
				return
			}
			log.Printf("[TestGlue][%v] CSVs: %v", c.ObjectName, csvs)
			if len(csvs) > 0 {
				for _, csv := range csvs {
					if strings.HasSuffix(csv, ".csv") {
						data, err2 := targetBucketHandler.ParseCsvFromS3(csv)
						if err2 != nil {
							testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] error listing s3 buckets after ETLs: %v", err2, c.ObjectName))
							cancel()
							return
						}
						select {
						case <-ctx.Done():
							log.Printf("[%v] Error in another test. stopping", c.ObjectName)
							return // Error somewhere, terminate
						default: // Default is must to avoid blocking
						}
						log.Printf("CSV has %v rows", len(data))
						log.Printf("CSV data: %v", data)
						if len(data) == 0 {
							testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] invalid ETL output: should have at least one row", c.ObjectName))
							cancel()
							return
						}
						//checking the error queue
						//TODO

						columns := map[string]bool{}
						for _, col := range data[0] {
							columns[col] = true
						}
						//Looking for mandatory columns
						mandatoryColumns := []string{"traveller_id", "last_updated", "model_version"}
						log.Printf("[TestGlue][%v] Checking for mandatory columns %v in CSV file", c.ObjectName, mandatoryColumns)
						for _, colName := range mandatoryColumns {
							if !columns[colName] {
								testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] CSV is missisng mandatory column '%v'", c.ObjectName, colName))
							}
						}

						select {
						case <-ctx.Done():
							log.Printf("[%v] Error in another test. stopping", c.ObjectName)
							return // Error somewhere, terminate
						default: // Default is must to avoid blocking
						}
					}
				}
			} else {
				testErrs = append(testErrs, fmt.Sprintf("[TestGlue][%v] No data created in CSV file", c.ObjectName))
				cancel()
				return
			}
		}(i, config)
	}
	wg.Wait()
	if len(testErrs) > 0 {
		for _, testErr := range testErrs {
			t.Errorf(testErr)
		}
	} else {
		log.Printf("All Tests successfully completed")
	}

	// Validate bookmark has been set
	err = dynamodbClient.Get(pk, sk, &data)
	if err != nil || data["bookmark"] == "" {
		t.Errorf("[TestMain] Error: unable to get initial bookmark data")
	} else {
		log.Printf("Initial bookmark data: %v", data)
	}
	// Validate count of records processed is same as expected
	count := int(data["records_processed"].(float64))
	expectedCount := 23 // TODO: improve
	if count != expectedCount {
		t.Errorf("[TestMain] The count of processed records on initial run does not match expected.")
	}
	// Run one job again, expect only data from partitions since last run to be included
	log.Printf("Testing incremental job run with a bookmark")
	ts := time.Now().Format("2006-01-02-15")
	split := strings.Split(ts, "-")
	year2 := split[0]
	month2 := split[1]
	day2 := split[2]
	hour2 := split[3]
	c2 := BusinessObjectTestConfig{
		ObjectName:   "clickstream",
		GlueJob:      GLUE_JOB_NAME_CLICKSTREAM,
		TestFilePath: "../../../../test_data/clickstream/",
		TestFiles: []string{
			"data1.json",
		},
		GlueTableName: TEST_TABLE_CLICKSTREAM,
		SourceBucket:  TEST_BUCKET_CLICKSTREAM,
		CrawlerName:   "glue_e2e_tests_clickstream",
		TargetPrefix:  "clickstream",
		Multiline:     true,
	}
	// Upload data to S3
	sourceBucketHandler := s3.Init(c2.SourceBucket, "", UCP_REGION)
	err = sourceBucketHandler.UploadFile(year2+"/"+month2+"/"+day2+"/"+hour2+"/"+c2.TestFiles[0], c2.TestFilePath+c2.TestFiles[0])
	if err != nil {
		t.Errorf("[MainTest] Error uploading test file: %v", err)
	}
	// Create Glue table partition
	errs := glueClient.AddPartitionsToTable(c2.GlueTableName, []glue.Partition{glue.Partition{
		Values:   []string{year2, month2, day2},
		Location: "s3://" + c2.SourceBucket + "/" + strings.Join([]string{year2, month2, day2}, "/"),
	}})
	if len(errs) > 0 {
		t.Errorf("[MainTest] Error(s) adding glue partitions: %v", errs)
	}
	// Run job
	err = glueClient.RunJob(c2.GlueJob, map[string]string{
		"--DEST_BUCKET":     TEST_BUCKET_ACCP_IMPORT,
		"--SOURCE_TABLE":    c2.GlueTableName,
		"--ERROR_QUEUE_URL": queueUrl,
		"--ACCP_DOMAIN":     domain,
		"--BIZ_OBJECT_NAME": c2.ObjectName,
		"--DYNAMO_TABLE":    DYNAMODB_CONFIG_TABLE_NAME,
	})
	if err != nil {
		t.Errorf("[MainTest] Error running Glue job: %v", err)
	}
	// Wait for job to complete
	status, err := glueClient.WaitForJobRun(c2.GlueJob, 600)
	if err != nil {
		t.Errorf("[MainTest] Error with Glue job: %v", err)
	}
	if status != "SUCCEEDED" {
		t.Errorf("[MainTest] Second Glue job run did not succeed")
	}
	// Verify expected records processed
	err = dynamodbClient.Get(pk, sk, &data)
	if err != nil || data["bookmark"] == "" {
		t.Errorf("[TestMain] Error: unable to get bookmark data")
	} else {
		log.Printf("Second run bookmark data: %v", data)
	}
	// Validate count of records processed is same as expected
	count = int(data["records_processed"].(float64))
	expectedCount = 9 // TODO: improve
	if count != expectedCount {
		t.Errorf("[TestMain] The count of processed records on second run is not correct.")
	}

	log.Printf("Cleaning up")
	for _, c := range bizObjectConfigs {
		sourceBucketHandler := s3.Init(c.SourceBucket, "", UCP_REGION)
		log.Printf("[%v] Cleanup", c.ObjectName)
		errs := glueClient.RemovePartitionsFromTable(c.GlueTableName, []glue.Partition{glue.Partition{
			Values: []string{year, month, day},
		}})
		if len(errs) > 0 {
			t.Errorf("[TestGlue][%v] Error deleting partitions %v", c.ObjectName, errs)
		}

		err := sourceBucketHandler.EmptyBucket()
		if err != nil {
			t.Errorf("[TestGlue][%v] Error emptying bucket %v", c.ObjectName, err)
		}
	}
	errs = glueClient.RemovePartitionsFromTable(c2.GlueTableName, []glue.Partition{
		glue.Partition{Values: []string{year2, month2, day2}},
	})
	if len(errs) > 0 {
		t.Errorf("[TestMain] Error deleting partition for second run: %v", errs)
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

// unprettyfy json file before upload
func unprettyfy(path string) error {
	log.Printf("Unprettyfying: %v", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	log.Printf("before unprettyfy: %v", string(data))
	unprettyfied := strings.Replace(string(data), "\n", "", -1)
	unprettyfied = strings.Replace(unprettyfied, "\t", "", -1)
	log.Printf("after unprettyfy: %v", unprettyfied)
	err = ioutil.WriteFile(path, []byte(unprettyfied), 0777)
	return err
}

// TODO: move this somewhere centralized
func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
