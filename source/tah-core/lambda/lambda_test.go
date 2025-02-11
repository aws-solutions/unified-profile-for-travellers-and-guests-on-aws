// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lambda

import (
	"encoding/json"
	"log"
	core "tah/upt/source/tah-core/core"
	db "tah/upt/source/tah-core/db"
	iam "tah/upt/source/tah-core/iam"
	s3 "tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

/*******
* LAMBDA
*****/

type EnablementSession struct {
	ContactID string `json:"contactID"`
}

func (p EnablementSession) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}
func TestLambda(t *testing.T) {
	var lambdaSvc IConfig = Init("testFunction", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	res := EnablementSession{}
	_, err := lambdaSvc.Invoke(map[string]string{"key": "value"}, res, false)
	if err == nil {
		t.Errorf("[core][Lambda] Should return Error while invoking invalid lambda")
	}
	//t.Errorf("[core][Lambda] Error while invoking lambda %v", res2)
}

func TestEventSourceMapping(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	tableName := "dynamoTestStream" + time.Now().Format("20060102150405")
	lambdaName := "testLambdaFunction" + time.Now().Format("20060102150405")
	roleName := "testRoleName" + time.Now().Format("20060102150405")
	s3Key := "code.zip"

	lambdaSvc := Init("testFunction", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	dbCfg := db.Init(tableName, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	s3c, err := s3.InitWithRandBucket("lambda-test-code-bucket", "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}

	roleArn, policyArn, err := iamCfg.CreateRoleWithActionsResource(roleName, "lambda.amazonaws.com", []string{
		"dynamodb:DescribeStream",
		"dynamodb:GetRecords",
		"dynamodb:GetShardIterator",
		"dynamodb:ListStreams",
	}, []string{"arn:aws:dynamodb:" + envCfg.Region + ":*:table/" + tableName + "/stream/*"})
	if err != nil {
		t.Errorf("[TestIam] error creating lambda execution role %v", err)
	}
	log.Printf("Waiting 10s for IAM role to propagate")
	time.Sleep(10 * time.Second)

	err = dbCfg.CreateTableWithOptions(tableName, "pk", "sk", db.TableOptions{
		StreamEnabled: true,
	})
	if err != nil {
		t.Errorf("error creating table with stream enabled: %v", err)
	}
	dbCfg.WaitForTableCreation()
	streamArn, err := dbCfg.GetStreamArn()
	if err != nil || streamArn == "" {
		t.Errorf("error getting stream arn: %v", err)
	}
	log.Printf("Uploading Lambda code ../../test_data/ssi_core/code.zip")
	err = s3c.UploadFile(s3Key, "../../test_data/ssi_core/code.zip")
	if err != nil {
		t.Errorf("Failed to upload code file: %+v", err)
	}

	err = lambdaSvc.CreateFunction(lambdaName, roleArn, s3c.Bucket, s3Key, LambdaOptions{})
	if err != nil {
		t.Errorf("error creating function: %v", err)
	}

	err = lambdaSvc.CreateDynamoEventSourceMapping(streamArn, lambdaName)
	if err != nil {
		t.Errorf("Could not create event stream mapping: %v", err)
	}

	mappings, err := lambdaSvc.SearchDynamoEventSourceMappings(streamArn, lambdaName)
	if err != nil {
		t.Errorf("Could not search event stream mappings: %v", err)
	}
	uuid := ""
	if len(mappings) != 1 {
		t.Errorf("Expected 1 mapping, got %v", len(mappings))
	} else {
		uuid = mappings[0].UUID
	}

	err = lambdaSvc.DeleteDynamoEventSourceMapping(uuid)
	if err != nil {
		t.Errorf("Could not delete event stream mapping: %v", err)
	}

	log.Printf("5. Deleting iam role")
	err = iamCfg.DetachRolePolicy(roleName, policyArn)
	if err != nil {
		t.Errorf("[TestIam] error detaching policy from role %v", err)
	}
	err = iamCfg.DeletePolicy(policyArn)
	if err != nil {
		t.Errorf("[TestIam] error deleting policy %v", err)
	}
	err = iamCfg.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestIam] error deleting role %v", err)
	}

	log.Printf("5. Deleting Dynamo table")
	err = dbCfg.DeleteTable(dbCfg.TableName)
	if err != nil {
		t.Errorf("[TestGsi] Error deleting test table %v", err)
	}
	log.Printf("5. Deleting Lambda function")
	err = lambdaSvc.DeleteFunction(lambdaName)
	if err != nil {
		t.Errorf("[TestGsi] Error deleting test table %v", err)
	}
	log.Printf("5. Deleting S3 bucket")
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}

}

func TestMock(t *testing.T) {
	var lambdaSvc IConfig = InitMock("testFunction")
	lambdaSvc.CreateDynamoEventSourceMapping("testArn", lambdaSvc.Get("FunctionName"))
	mappings, _ := lambdaSvc.SearchDynamoEventSourceMappings("testArn", "testFunction")
	if len(mappings) != 1 {
		t.Errorf("Expected 1 mapping in mock after creation, got %v", len(mappings))
	}
	err := lambdaSvc.DeleteDynamoEventSourceMapping(mappings[0].UUID)
	if err != nil {
		t.Errorf("Could not delete event stream mapping: %v", err)
	}
	err = lambdaSvc.DeleteDynamoEventSourceMapping("invalid uuid")
	if err == nil {
		t.Errorf("Expected error while deleting mapping with invalid uuid")
	}

}
