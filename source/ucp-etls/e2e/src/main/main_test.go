package main

import (
	"log"
	"os"
	glue "tah/core/glue"
	s3 "tah/core/s3"
	"testing"
	"time"
)

var UCP_REGION = getRegion()
var GLUE_ROLE_NAME = os.Getenv("GLUE_ROLE_NAME")
var GLUE_DB_NAME = os.Getenv("GLUE_DB_NAME")
var TEST_BUCKET_AIR_BOOKING = os.Getenv("TEST_BUCKET_AIR_BOOKING")
var TEST_BUCKET_ACCP_IMPORT = os.Getenv("TEST_BUCKET_ACCP_IMPORT")
var GLUE_JOB_NAME_AIR_BOOKING = os.Getenv("GLUE_JOB_NAME_AIR_BOOKING")

type BusinessObjectTestConfig struct {
	ObjectName    string
	GlueJob       string
	TestFilePath  string
	TestFiles     []string
	GlueTableName string
	SourceBucket  string
	CrawlerName   string
}

func TestMain(t *testing.T) {
	schedule := "cron(15 12 * * ? *)"
	glueClient := glue.Init(UCP_REGION, GLUE_DB_NAME)

	bizObjectConfigs := []BusinessObjectTestConfig{
		BusinessObjectTestConfig{
			ObjectName:   "air_booking",
			GlueJob:      GLUE_JOB_NAME_AIR_BOOKING,
			TestFilePath: "../../../air_booking/testData/",
			TestFiles: []string{
				"5M71Q4.json",
				"PT4UBE.json",
			},
			GlueTableName: "air_booking_" + time.Now().Format("15_04_05"),
			SourceBucket:  TEST_BUCKET_AIR_BOOKING,
			CrawlerName:   "glue_e2e_tests_air_booking",
		},
	}

	for _, c := range bizObjectConfigs {
		log.Printf("1-Running e2e tests for business object %v", c.ObjectName)
		sourceBucketHandler := s3.Init(c.SourceBucket, "", UCP_REGION)
		targetBucketHandler := s3.Init(TEST_BUCKET_ACCP_IMPORT, "", UCP_REGION)
		log.Printf("2-Uploading data to S3")
		for _, file := range c.TestFiles {
			err := sourceBucketHandler.UploadFile(c.GlueTableName+"/2023/12/01/09/"+file, c.TestFilePath+file)
			if err != nil {
				t.Errorf("Cound not crerate bucket to unit test UCP %v", err)
			}
		}
		log.Printf("3-Create cralwer")
		err := glueClient.CreateSimpleS3Crawler(c.CrawlerName, GLUE_ROLE_NAME, schedule, "s3://"+c.SourceBucket+"/"+c.GlueTableName+"/")
		if err != nil {
			t.Errorf("[TestGlue] error creating trigger: %v", err)
		}
		log.Printf("4-Run Crawler")
		err = glueClient.RunCrawler(c.CrawlerName)
		if err != nil {
			t.Errorf("[TestGlue] error running crawler: %v", err)
		}
		log.Printf("5-Wait for Crawler")
		status, err2 := glueClient.WaitForCrawlerRun(c.CrawlerName, 300)
		if err2 != nil {
			t.Errorf("[TestGlue] error running crawler: %v", err2)
		}
		if status != "SUCCEEDED" {
			t.Errorf("[TestGlue] crawler completed with non success status %v", status)
		}
		log.Printf("6-Run job with modified source and targets")
		err = glueClient.RunJob(c.GlueJob, map[string]string{
			"--DEST_BUCKET":  TEST_BUCKET_ACCP_IMPORT,
			"--SOURCE_TABLE": c.GlueTableName,
		})
		if err != nil {
			t.Errorf("[TestGlue] error running job: %v", err)
		}
		log.Printf("7-Wait for job run")
		status, err = glueClient.WaitForJobRun(c.GlueJob, 600)
		if err != nil {
			t.Errorf("[TestGlue] error waiting for job completion: %v", err)
		}
		if status != "SUCCEEDED" {
			t.Errorf("[TestGlue] Job completed with non success status %v", status)
		}
		log.Printf("8-Check csv File")
		csvs, err1 := targetBucketHandler.Search("")
		if err1 != nil {
			t.Errorf("[TestGlue] error wlisting s3 buckets after ETLs: %v", err1)
		}
		log.Printf("CSVs: %v", csvs)
		if len(csvs) > 0 {
			for _, csv := range csvs {
				data, err2 := targetBucketHandler.ParseCsvFromS3(csv)
				if err2 != nil {
					t.Errorf("[TestGlue] error wlisting s3 buckets after ETLs: %v", err2)
				}
				log.Printf("CSV has %v rows", len(data))
				log.Printf("CSV data: %v", data)
				if len(data) == 0 {
					t.Errorf("[TestGlue] invalid ETL output: should have at least one row %v", err2)
				}
				//looking for an error header in the CSV that woudl indicate that an exception has occured in the trasformer
				for i, col := range data[0] {
					if col == "error" {
						t.Errorf("[TestGlue] invalid ETL output: data has an error column: %v", err2)
						for j, row := range data {
							if row[i] != "" {
								t.Errorf("[TestGlue] Error at row %v and col %v : %v", j, i, row[i])
							}
						}
					}
				}
			}
		} else {
			t.Errorf("[TestGlue] No data created in CSV file")
		}
		log.Printf("9-Cleanup")
		err = glueClient.DeleteCrawlerIfExists(c.CrawlerName)
		if err != nil {
			t.Errorf("[TestGlue] error deleting crawler: %v", err)
		}
		err = sourceBucketHandler.EmptyBucket()
		if err != nil {
			t.Errorf("Error deleting bucket %v", err)
		}
		err = targetBucketHandler.EmptyBucket()
		if err != nil {
			t.Errorf("Error deleting bucket %v", err)
		}
	}
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
