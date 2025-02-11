// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package datasync

import (
	"log"
	"testing"
	"time"

	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
)

func TestDataSync(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	// Create required resources
	sourceBucketPrefix := "bookings"
	destinationBucketPrefix := "air_booking"
	fileName := "csv1.csv"
	// Create buckets
	sourceBucketCfg, err := s3.InitWithRandBucket("datasync-source-bucket", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("Error creating source test bucket: %v", err)
	}
	err = sourceBucketCfg.UploadFile(sourceBucketPrefix+"/2023/05/26/"+fileName, "../../test_data/ssi_core/csv1.csv")
	if err != nil {
		t.Errorf("Error uploading test file: %v", err)
	}
	destinationBucketCfg, err := s3.InitWithRandBucket("datasync-destination-bucket", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("Error creating destination test bucket: %v", err)
	}
	sourceBucketName := sourceBucketCfg.Bucket
	destinationBucketName := destinationBucketCfg.Bucket
	// Create IAM roles
	iamClient := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	sourceBucketPolicyDoc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"s3:GetBucketLocation", "s3:ListBucket", "s3:ListBucketMultipartUploads"},
				Resource: []string{core.S3BucketNameToARN(sourceBucketName)},
			},
			{
				Effect:   "Allow",
				Action:   []string{"s3:AbortMultipartUpload", "s3:DeleteObject", "s3:GetObject", "s3:ListMultipartUploadParts", "s3:PutObjectTagging", "s3:GetObjectTagging", "s3:PutObject"},
				Resource: []string{core.S3BucketNameToARN(sourceBucketName) + "/*"},
			},
		},
	}
	sourceRoleName := "test-source-datasync-role"
	sourceRoleArn, sourcePolicyArn, err := iamClient.CreateRoleWithPolicy(sourceRoleName, "datasync.amazonaws.com", sourceBucketPolicyDoc)
	if err != nil {
		t.Errorf("Error creating role: %v", err)
	}
	destinationBucketPolicyDoc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"s3:GetBucketLocation", "s3:ListBucket", "s3:ListBucketMultipartUploads"},
				Resource: []string{core.S3BucketNameToARN(destinationBucketName)},
			},
			{
				Effect:   "Allow",
				Action:   []string{"s3:AbortMultipartUpload", "s3:DeleteObject", "s3:GetObject", "s3:ListMultipartUploadParts", "s3:PutObjectTagging", "s3:GetObjectTagging", "s3:PutObject"},
				Resource: []string{core.S3BucketNameToARN(destinationBucketName) + "/*"},
			},
		},
	}
	destinationRoleName := "test-destination-datasync-role"
	destinationRoleArn, destinationPolicyArn, err := iamClient.CreateRoleWithPolicy(destinationRoleName, "datasync.amazonaws.com", destinationBucketPolicyDoc)
	if err != nil {
		t.Errorf("Error creating role: %v", err)
	}
	log.Printf("Waiting for new roles and policies to propagate")
	time.Sleep(15 * time.Second)

	// Test Datasync functions
	datasyncClient := InitRegion(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, core.LogLevelDebug)
	// Create source and target S3 locations
	sourceLocationArn, err := datasyncClient.CreateS3Location(sourceBucketName, sourceBucketPrefix, sourceRoleArn)
	if err != nil {
		t.Errorf("Error creating location: %v", err)
	}
	destinationLocationArn, err := datasyncClient.CreateS3Location(destinationBucketName, destinationBucketPrefix, destinationRoleArn)
	if err != nil {
		t.Errorf("Error creating location: %v", err)
	}
	// List DataSync locations
	locations, err := datasyncClient.ListLocations()
	if err != nil {
		t.Errorf("Error listing locations: %v", err)
	}
	if !core.ContainsString(locations, sourceLocationArn) || !core.ContainsString(locations, destinationLocationArn) {
		t.Errorf("Unexpected location list")
	}
	// Create DataSync task
	taskArn, err := datasyncClient.CreateTask("test-datasync-task", sourceLocationArn, destinationLocationArn, nil)
	if err != nil {
		t.Errorf("Error creating task: %v", err)
	}
	// List tasks and check for newly created task
	tasks, err := datasyncClient.ListTasks()
	if err != nil {
		t.Errorf("Error listing tasks: %v", err)
	}
	taskFound := false
	for _, task := range tasks {
		if task.ARN == taskArn {
			taskFound = true
		}
	}
	if !taskFound {
		t.Errorf("Task not found in list")
	}
	// Start task execution
	_, err = datasyncClient.StartTaskExecution(taskArn)
	// executionArn, err := datasyncClient.StartTaskExecution(taskArn)
	if err != nil {
		t.Errorf("Error starting task execution: %v", err)
	}
	// ***** Testing the full run can take up to 30 minutes and not practical to test here *****
	// // Wait for task to finish, max of 30 minutes
	// i, max, duration, done := 0, 360, 5*time.Second, false
	// for ; i < max; i++ {
	// 	executions, err := datasyncClient.ListTaskExecutions()
	// 	if done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		t.Errorf("Error listing task executions: %v", err)
	// 		break
	// 	}
	// 	for _, execution := range executions {
	// 		if execution.ExecutionArn == executionArn && execution.ExecutionStatus == TaskExecutionStatusSuccess {
	// 			done = true
	// 			break
	// 		}
	// 	}
	// 	log.Printf("Waiting for task to finish...")
	// 	time.Sleep(duration)
	// }
	// if i == max {
	// 	t.Errorf("Task did not complete in time")
	// }
	// Delete task and locations
	err = datasyncClient.DeleteTask(taskArn)
	if err != nil {
		t.Errorf("Error deleting task: %v", err)
	}
	err = datasyncClient.DeleteLocation(sourceLocationArn)
	if err != nil {
		t.Errorf("Error deleting location: %v", err)
	}
	err = datasyncClient.DeleteLocation(destinationLocationArn)
	if err != nil {
		t.Errorf("Error deleting location: %v", err)
	}

	// Clean up resources
	err = sourceBucketCfg.EmptyAndDelete()
	if err != nil {
		t.Errorf("Error cleaning up source test bucket: %v", err)
	}
	err = destinationBucketCfg.EmptyAndDelete()
	if err != nil {
		t.Errorf("Error cleaning up destination test bucket: %v", err)
	}
	err = iamClient.DetachRolePolicy(sourceRoleName, sourcePolicyArn)
	if err != nil {
		t.Errorf("Error detaching source policy: %v", err)
	}
	err = iamClient.DetachRolePolicy(destinationRoleName, destinationPolicyArn)
	if err != nil {
		t.Errorf("Error detaching destination policy: %v", err)
	}
	err = iamClient.DeletePolicy(sourcePolicyArn)
	if err != nil {
		t.Errorf("Error deleting source policy: %v", err)
	}
	err = iamClient.DeletePolicy(destinationPolicyArn)
	if err != nil {
		t.Errorf("Error deleting destination policy: %v", err)
	}
	err = iamClient.DeleteRole(sourceRoleName)
	if err != nil {
		t.Errorf("Error deleting source role: %v", err)
	}
	err = iamClient.DeleteRole(destinationRoleName)
	if err != nil {
		t.Errorf("Error deleting destination role: %v", err)
	}
}
