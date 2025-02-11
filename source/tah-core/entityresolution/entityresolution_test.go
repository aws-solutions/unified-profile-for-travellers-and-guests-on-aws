// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package entityresolution

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

type TableSchema struct {
	IpAddress           string `json:"ip_address"`
	FirstName           string `json:"first_name"`
	MiddleName          string `json:"middle_name"`
	LastName            string `json:"last_name"`
	GuestProfileAddress string `json:"guest_profile_address"`
	ConnectId           string `json:"connect_id"`
}

func TestCreateSchemaMapping(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	glueDb := "test_db"
	entityResolutionClient := Init(envCfg.Region, glueDb, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	entityResolutionClient.SetTx(tx)
	schemaName := "test_schema_name"
	fieldMappings := []EntityResolutionFieldMapping{
		{
			Source: "clickstream_ip_address",
			Target: "ip_address",
		},
		{
			Source: "clickstream_first_name",
			Target: "NAME_FIRST",
		},
		{
			Source: "clickstream_middle_name",
			Target: "NAME_MIDDLE",
		},
		{
			Source: "guest_profile_address",
			Target: "ADDRESS_COUNTRY",
		},
		{
			Source: "connect_id",
			Target: "UNIQUE_ID",
		},
	}
	schemaMappingOutput, err := entityResolutionClient.CreateSchema(fieldMappings, schemaName)
	if err != nil {
		t.Fatalf("[%s] error creating schema: %v", t.Name(), err)
	}
	t.Logf("[%s] schema mapping name: %v", t.Name(), schemaMappingOutput)

	t.Cleanup(func() {
		message, err := entityResolutionClient.DeleteSchema(schemaName)
		if err != nil {
			t.Errorf("[%s] error deleting schema: %v", t.Name(), err)
		}
		t.Logf("[%s] schema mapping output: %v", t.Name(), message)
	})

	if schemaMappingOutput != schemaName {
		t.Errorf("[%s] schema mapping name does not match: %v", t.Name(), schemaMappingOutput)
	}
}

func TestCreateSchemasInParallel(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	// Create a wait group
	var wg sync.WaitGroup

	// Number of schemas to create
	const numSchemas = 10

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	fieldMappings := []EntityResolutionFieldMapping{
		{
			Source: "clickstream_ip_address",
			Target: "ip_address",
		},
		{
			Source: "clickstream_first_name",
			Target: "NAME_FIRST",
		},
		{
			Source: "clickstream_middle_name",
			Target: "NAME_MIDDLE",
		},
		{
			Source: "guest_profile_address",
			Target: "ADDRESS_COUNTRY",
		},
		{
			Source: "connect_id",
			Target: "UNIQUE_ID",
		},
	}

	glueDb := "test_db"
	entityResolutionClient := Init(envCfg.Region, glueDb, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	entityResolutionClient.SetTx(tx)

	// Create schemas in goroutines
	for i := 0; i < numSchemas; i++ {
		// Add a wait group counter for each schema
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			schemaName := fmt.Sprintf("test_schema_%d", i)
			_, err := entityResolutionClient.CreateSchema(fieldMappings, schemaName)
			if err != nil {
				t.Errorf("error creating schema: %v", err)
			}
		}(i)
	}

	t.Cleanup(func() {
		var deleteWg sync.WaitGroup

		for i := 0; i < numSchemas; i++ {
			deleteWg.Add(1)
			go func(i int) {
				defer deleteWg.Done()

				schemaName := fmt.Sprintf("test_schema_%d", i)
				_, err := entityResolutionClient.DeleteSchema(schemaName)
				if err != nil {
					t.Errorf("error deleting schema: %v", err)
				}
			}(i)
		}
		deleteWg.Wait()
	})

	wg.Wait()
}

func TestAiMapping(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	dbName := "tah_core_test_db_entity"
	tableName := "tah_core_test_table_entity_resolution"
	schemaName := "test_schema_name_ai"
	workflowName := "entity-resolution-test-workflow"
	s3c, err := s3.InitWithRandBucket(
		"tah-core-test-entity-resolution",
		"",
		envCfg.Region,
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
	)
	if err != nil {
		t.Fatalf("Could not initialize random bucket %+v", err)
	}
	t.Cleanup(func() {
		err = s3c.EmptyAndDelete()
		if err != nil {
			t.Errorf("[%s] error deleting s3 bucket: %v", t.Name(), err)
		}
	})
	bucketName := s3c.Bucket
	outputBucket, err := s3c.CreateRandomBucket("tah-core-entity-resolution-output")
	if err != nil {
		t.Fatalf("Could not create random bucket %+v", err)
	}
	t.Cleanup(func() {
		err = s3c.EmptySpecificBucket(outputBucket)
		if err != nil {
			t.Errorf("[%s] error emptying s3 bucket: %v", t.Name(), err)
		}
		err = s3c.DeleteBucket(outputBucket)
		if err != nil {
			t.Errorf("[%s] error deleting s3 bucket: %v", t.Name(), err)
		}
	})
	tableSchemaList := GenerateTableSchemaList()
	file, err := os.Create("data.jsonl")
	if err != nil {
		t.Fatalf("Could not create file %+v", err)
	}
	defer file.Close()
	t.Cleanup(func() {
		err = os.Remove("data.jsonl")
		if err != nil {
			t.Errorf("[%s] error deleting file: %v", t.Name(), err)
		}
	})
	encoder := json.NewEncoder(file)

	//iterate through schema list, add lines to table
	for _, schema := range tableSchemaList {
		err = encoder.Encode(schema)
		if err != nil {
			t.Fatalf("Could not encode %+v", err)
		}
	}

	err = s3c.UploadFile("data.jsonl", "data.jsonl")
	if err != nil {
		t.Fatalf("Could not upload file %+v", err)
	}

	newTestRoleName := "er-test-role-" + strings.ToLower(core.GenerateUniqueId())
	newTestRoleArn, err := createRoleTest(t, newTestRoleName)
	if err != nil {
		t.Fatalf("Could not create role %+v", err)
	}
	//Need to wait for role to propagate
	time.Sleep(10 * time.Second)

	//Granting Entity Resolution Control of Buckets
	resource := []string{"arn:aws:s3:::" + outputBucket, "arn:aws:s3:::" + outputBucket + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"entityresolution.amazonaws.com"}}
	err = s3c.AddPolicy(outputBucket, resource, actions, principal)
	if err != nil {
		t.Fatalf("Could not add policy %+v", err)
	}

	resource = []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	err = s3c.AddPolicy(bucketName, resource, actions, principal)
	if err != nil {
		t.Fatalf("Could not add policy %+v", err)
	}
	//setting up glue

	schemaRawData := `{
		"columns": [
			{
				"name": "ip_address",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "first_name",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "middle_name",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "last_name",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "guest_profile_address",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "connect_id",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			}
		]
	}`
	glueClient := glue.Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Fatalf("[%s] error creating database: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteDatabase(dbName)
		if err != nil {
			t.Errorf("[%s] error deleting database: %v", t.Name(), err)
		}
	})
	schemaStruct, err := glue.ParseSchema(schemaRawData)
	if err != nil {
		t.Fatalf("[%s] error parsing schema string: %v", t.Name(), err)
	}
	err = glueClient.CreateTable(tableName, bucketName+"/", []glue.PartitionKey{}, schemaStruct)
	if err != nil {
		t.Fatalf("[%s] error creating table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueClient.DeleteTable(tableName)
		if err != nil {
			t.Errorf("[%s] Error deleting table: %v", t.Name(), err)
		}
	})

	//creating schema for entity resolution
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	fieldMappings := []EntityResolutionFieldMapping{
		{
			Source: "ip_address",
			Target: "ip_address",
		},
		{
			Source: "first_name",
			Target: "NAME_FIRST",
		},
		{
			Source: "middle_name",
			Target: "NAME_MIDDLE",
		},
		{
			Source: "last_name",
			Target: "NAME_LAST",
		},
		{
			Source: "guest_profile_address",
			Target: "ADDRESS_COUNTRY",
		},
		{
			Source: "connect_id",
			Target: "UNIQUE_ID",
		},
	}
	entityResolutionClient := Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	entityResolutionClient.SetTx(tx)
	_, err = entityResolutionClient.CreateSchema(fieldMappings, schemaName)
	if err != nil {
		t.Fatalf("[%s] error creating schema: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		_, err = entityResolutionClient.DeleteSchema(schemaName)
		if err != nil {
			t.Errorf("[%s] error deleting schema: %v", t.Name(), err)
		}
	})
	//creating workflow for entity resolution
	workflowOutputName, err := entityResolutionClient.CreateMatchingWorkflow(
		workflowName,
		schemaName,
		[]string{tableName},
		newTestRoleArn,
		outputBucket,
		"connect_id",
	)
	if err != nil {
		t.Fatalf("[%s] error creating workflow: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		_, err = entityResolutionClient.DeleteWorkflow(workflowName)
		if err != nil {
			t.Errorf("[%s] error deleting workflow: %v", t.Name(), err)
		}
	})
	if workflowOutputName != workflowName {
		t.Errorf("[%s] wrong output name: %s", t.Name(), workflowOutputName)
	}
}

// modify at will to design test
func GenerateTableSchemaList() []TableSchema {
	initialSchema := TableSchema{
		IpAddress:           "1000001",
		FirstName:           "Michael",
		MiddleName:          "J",
		LastName:            "Fox",
		GuestProfileAddress: "Canada",
		ConnectId:           "XXX",
	}
	secondSchema := TableSchema{
		IpAddress:           "1000000",
		FirstName:           "Michael",
		MiddleName:          "J",
		LastName:            "Fox",
		GuestProfileAddress: "Canada",
		ConnectId:           "XXY",
	}
	thirdSchema := TableSchema{
		IpAddress:           "1000000",
		FirstName:           "Michael",
		MiddleName:          "R",
		LastName:            "Fox",
		GuestProfileAddress: "Canada",
		ConnectId:           "XXZ",
	}
	return []TableSchema{initialSchema, secondSchema, thirdSchema}
}

func createRoleTest(t *testing.T, roleName string) (string, error) {
	t.Helper()

	iamClient := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	//link to document where permissions came from
	//https://docs.aws.amazon.com/entityresolution/latest/userguide/security-iam-awsmanpol.html
	policyDoc := iam.PolicyDocument{
		Version: "2012-10-17",
		Statement: []iam.StatementEntry{
			{
				Effect:   "Allow",
				Action:   []string{"entityresolution:*"},
				Resource: []string{"*"},
			},
			{
				Effect: "Allow",
				Action: []string{
					"glue:GetSchema",
					"glue:SearchTables",
					"glue:GetSchemaByDefinition",
					"glue:GetSchemaVersion",
					"glue:GetSchemaVersionsDiff",
					"glue:GetDatabase",
					"glue:GetDatabases",
					"glue:GetTable",
					"glue:GetTables",
					"glue:GetTableVersion",
					"glue:GetTableVersions",
				},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"s3:ListAllMyBuckets"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"s3:ListBucket", "s3:GetBucketLocation", "s3:ListBucketVersions", "s3:GetBucketVersioning"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"tag:GetTagKeys", "tag:GetTagValues"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"kms:DescribeKey", "kms:ListAliases"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"iam:ListRoles"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"iam:PassRole"},
				Resource: []string{"arn:aws:iam::*:role/*entityresolution*"},
				// TODO: add condition
			},
			{
				Effect:   "Allow",
				Action:   []string{"events:PutRule", "events:DeleteRule", "events:PutTargets"},
				Resource: []string{"arn:aws:events:*:*:rule/entity-resolution-automatic*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"dataexchange:GetDataSet"},
				Resource: []string{"*"},
			},
			{
				Effect:   "Allow",
				Action:   []string{"*"},
				Resource: []string{"*"},
			},
		},
	}

	roleArn, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "entityresolution.amazonaws.com", policyDoc)
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		err = iamClient.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[%s] Error deleting policy: %v", t.Name(), err)
		}
	})
	t.Cleanup(func() {
		err = iamClient.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[%s] Error deleting role: %v", t.Name(), err)
		}
	})
	return roleArn, err
	// TODO: attach role
}
