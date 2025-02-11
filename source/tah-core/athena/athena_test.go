// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package athena

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/s3"
	"testing"
)

func TestCreateDeleteWorkgroup(t *testing.T) {
	testName := "TestCreateDeleteWorkgroup"
	testPostFix := core.GenerateUniqueId()

	//	Create Base Resources
	var outputBucket s3.S3Config
	var athenaCfg Config
	createBaseResources(t, testName, testPostFix, &outputBucket, &athenaCfg)

	nonDefaultWorkgroup := "non-default-workgroup-" + testPostFix
	err := athenaCfg.CreateWorkGroup(nonDefaultWorkgroup)
	if err != nil {
		t.Fatalf("[%s] error creating workgroup: %v", testName, err)
	}
	t.Cleanup(func() {
		err = athenaCfg.DeleteWorkGroup(nonDefaultWorkgroup)
		if err != nil {
			t.Errorf("[%s] error deleting workgroup: %v", testName, err)
		}
	})

	err = athenaCfg.CreateWorkGroup("")
	if err != nil {
		t.Fatalf("[%s] error creating workgroup: %v", testName, err)
	}
	t.Cleanup(func() {
		err = athenaCfg.DeleteWorkGroup("")
		if err != nil {
			t.Errorf("[%s] error deleting workgroup: %v", testName, err)
		}
	})
}

func TestCreateDeletePreparedStatement(t *testing.T) {
	testName := "TestCreateDeletePreparedStatement"
	testPostFix := core.GenerateUniqueId()

	//	Create Base Resources
	var outputBucket s3.S3Config
	var athenaCfg Config
	createBaseResources(t, testName, testPostFix, &outputBucket, &athenaCfg)

	err := athenaCfg.CreateWorkGroup(athenaCfg.Workgroup)
	if err != nil {
		t.Fatalf("[%s] error creating workgroup: %v", testName, err)
	}
	t.Cleanup(func() { athenaCfg.DeleteWorkGroup(athenaCfg.Workgroup) })

	testStatementName := "test_statement_" + testPostFix
	err = athenaCfg.CreatePreparedStatement("select 1", testStatementName)
	if err != nil {
		t.Fatalf("[%s] error creating prepared statement: %v", testName, err)
	}
	t.Cleanup(func() {
		err = athenaCfg.DeletePreparedStatement(testStatementName)
		if err != nil {
			t.Errorf("[%s] error deleting prepared statement: %v", testName, err)
		}
	})
}

func TestExecutePreparedStatement(t *testing.T) {

	testName := "TestCreateDeletePreparedStatement"
	testPostFix := core.GenerateUniqueId()

	//	Create Base Resources
	var outputBucket s3.S3Config
	var athenaCfg Config
	createBaseResources(t, testName, testPostFix, &outputBucket, &athenaCfg)

	err := athenaCfg.CreateWorkGroup(athenaCfg.Workgroup)
	if err != nil {
		t.Fatalf("[%s] error creating workgroup: %v", testName, err)
	}
	t.Cleanup(func() { athenaCfg.DeleteWorkGroup(athenaCfg.Workgroup) })

	err = athenaCfg.CreatePreparedStatement("select 1", "test_statement_1")
	if err != nil {
		t.Fatalf("[%s] error creating prepared statement: %v", testName, err)
	}
	t.Cleanup(func() { athenaCfg.DeletePreparedStatement("test_statement_1") })

	_, err = athenaCfg.ExecutePreparedStatement("test_statement_1", []*string{}, Option{})
	if err != nil {
		t.Fatalf("[%s] error executing prepared statement: %v", testName, err)
	}
}

func createBaseResources(t *testing.T, testName string, testPostFix string, outputBucket *s3.S3Config, athenaCfg *Config) {
	//	Create Athena Output Bucket
	bucket, err := s3.InitWithRandBucket("test-athena-output-bucket", "", core.GetTestRegion(), "", "")
	if err != nil {
		t.Fatalf("[%s] error creating athena output bucket: %v", testName, err)
	}
	t.Cleanup(func() { outputBucket.EmptyAndDelete() })

	cfg := Init("test-db-"+testPostFix, "test-table-"+testPostFix, "test-workgroup-"+testPostFix, core.GetTestRegion(), "s3://"+bucket.Bucket, "", "", core.LogLevelDebug)

	*outputBucket = bucket
	*athenaCfg = cfg
}
