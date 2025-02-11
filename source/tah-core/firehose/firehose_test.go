// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package firehose

import (
	"strings"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/glue"
	"tah/upt/source/tah-core/iam"
	"tah/upt/source/tah-core/s3"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
)

func TestDynamicPartitioning(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	streamName := "tah_core_test_stream_dynamic_part_" + core.GenerateUniqueId()
	testRec1 := "{\"key\":\"test record 1\", \"folder\":\"f1\", \"subfolder\":\"sf1\"}"
	testRec2 := "{\"key\":\"test record 2\", \"folder\":\"f1\", \"subfolder\":\"sf2\"}"
	testRec3 := "{\"key\":\"test record 3\", \"folder\":\"f2\", \"subfolder\":\"sf3\"}"
	testRec4 := "{\"key\":\"test record 4\", \"folder\":\"f2\", \"subfolder\":\"sf4\"}"
	testRec5 := "{\"key\":\"test record 5\", \"folder\":\"f2\", \"subfolder\":\"sf5\"}"

	s3c, err := s3.InitWithRandBucket("tah-core-test-bucket-firehose", "", envCfg.Region, "", "")
	if err != nil {
		t.Fatalf("[%s] Failed create %+v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = s3c.EmptyAndDelete()
		if err != nil {
			t.Errorf("[%s] Failed delete %+v", t.Name(), err)
		}
	})

	name := s3c.Bucket
	bucketArn := "arn:aws:s3:::" + name
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:PutObject"}
	principal := map[string][]string{"Service": {"firehose.amazonaws.com"}}
	err = s3c.AddPolicy(name, resources, actions, principal)
	if err != nil {
		t.Fatalf("[%s] error adding bucket policy %+v", t.Name(), err)
	}

	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleName := "firehose-unit-test-role-dyn-partitioning-" + strings.ToLower(core.GenerateUniqueId())
	roleArn, policyArn, err := iamCfg.CreateRoleWithActionsResource(roleName, "firehose.amazonaws.com", []string{"s3:PutObject",
		"s3:PutObjectAcl",
		"s3:ListBucket"}, resources)
	if err != nil {
		t.Fatalf("[%s] error %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = iamCfg.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[%s] error deleting policy %v", t.Name(), err)
		}
	})
	t.Cleanup(func() {
		err = iamCfg.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[%s] error deleting role whith attached policies policies %v", t.Name(), err)
		}
	})

	t.Logf("Waiting 10 seconds role to propagate")
	time.Sleep(10 * time.Second)

	t.Logf("0-Creating stream %v", streamName)
	cfg, err := InitAndCreateWithPartitioning(streamName, bucketArn, roleArn, []PartitionKeyVal{{Key: "folder", Val: ".folder"}, {
		Key: "subfolder", Val: ".subfolder"}}, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error creating stream: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		t.Logf("4-Deleting stream %v", streamName)
		err = cfg.Delete(streamName)
		if err != nil {
			t.Errorf("[%s] error deleting stream: %v", t.Name(), err)
		}
	})

	_, err = cfg.WaitForStreamCreation(streamName, 120)
	if err != nil {
		t.Fatalf("[%s] timeout waiting for stream creation: %v", t.Name(), err)
	}
	t.Logf("1-Describing Stream %v", streamName)
	stream, err := cfg.Describe(streamName)
	if err != nil {
		t.Fatalf("[%s] error describing stream: %v", t.Name(), err)
	}
	t.Logf("Stream: %+v created", stream)

	t.Logf("2-Sending 2 records to stream %v", streamName)
	records := []Record{
		{
			Data: testRec1,
		},
		{
			Data: testRec2,
		},
		{
			Data: testRec3,
		},
		{
			Data: testRec4,
		},
		{
			Data: testRec5,
		}}
	errs, err := cfg.PutRecords(streamName, records)
	if err != nil {
		t.Fatalf("[%s] error sending data to stream: %v. %+v", t.Name(), err, errs)
	}
	err = WaitForData(s3c, 600)
	if err != nil {
		t.Fatalf("[%s] error waiting for data: %v", t.Name(), err)
	}
	res, err := s3c.Search("", 100)
	t.Logf("S3 bucket constains %v items: %+v", len(res), res)
	if err != nil {
		t.Fatalf("Could not list results in S3")
	}
	if len(res) != 5 {
		t.Errorf("Bucket Should have 5 item not %v", len(res))
	}
	found := map[string]bool{}
	for _, r := range res {
		found[r[:28]] = true
		data, err := s3c.GetTextObj(r)
		if err != nil {
			t.Errorf("Could not get object from S3")
		}
		t.Logf("found: %+v", data)
	}

	if len(found) != 5 {
		t.Errorf("Should have found 5 records in S3")
	}
	for _, f := range []string{"data/folder=f1/subfolder=sf1", "data/folder=f1/subfolder=sf2", "data/folder=f2/subfolder=sf3", "data/folder=f2/subfolder=sf4", "data/folder=f2/subfolder=sf5"} {
		if !found[f] {
			t.Errorf("Record with prefix %v not found in S3", f)
		}
	}
}

func TestDataFormatConversion(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	streamName := "tah_core_test_stream_data_format_conversion_" + strings.ToLower(core.GenerateUniqueId())
	testRec1 := "{\"key\":\"test record 1\", \"folder\":\"f1\", \"subfolder\":\"sf1\"}"
	testRec2 := "{\"key\":\"test record 2\", \"folder\":\"f1\", \"subfolder\":\"sf2\"}"
	testRec3 := "{\"key\":\"test record 3\", \"folder\":\"f2\", \"subfolder\":\"sf3\"}"
	testRec4 := "{\"key\":\"test record 4\", \"folder\":\"f2\", \"subfolder\":\"sf4\"}"
	testRec5 := "{\"key\":\"test record 5\", \"folder\":\"f2\", \"subfolder\":\"sf5\"}"

	s3c, err := s3.InitWithRandBucket("tah-core-test-bucket-firehose", "", envCfg.Region, "", "")
	if err != nil {
		t.Fatalf("[%s] Failed create %+v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = s3c.EmptyAndDelete()
		if err != nil {
			t.Errorf("[%s] Failed delete %+v", t.Name(), err)
		}
	})

	name := s3c.Bucket
	bucketArn := "arn:aws:s3:::" + name
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:PutObject"}
	principal := map[string][]string{"Service": {"firehose.amazonaws.com"}}
	if err = s3c.AddPolicy(name, resources, actions, principal); err != nil {
		t.Fatalf("[%s] error adding bucket policy %+v", t.Name(), err)
	}

	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	glueDbName := "firehose_test_format_conversion_db_" + strings.ToLower(core.GenerateUniqueId())
	glueCfg := glue.Init(envCfg.Region, glueDbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	err = glueCfg.CreateDatabase(glueDbName)
	if err != nil {
		t.Fatalf("[%s] Error creating Glue DB: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueCfg.DeleteDatabase(glueDbName)
		if err != nil {
			t.Errorf("[%s] error deleting database %v", t.Name(), err)
		}
	})
	schema, err := glue.ParseSchema(`{
		"columns": [
			{
				"name": "key",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "folder",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			},
			{
				"name": "subfolder",
				"type": {
					"isPrimitive": true,
					"inputString": "string"
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("[%s] error creating schema: %v", t.Name(), err)
	}
	glueTableName := "firehose_test_format_conversion_table_" + strings.ToLower(core.GenerateUniqueId())
	err = glueCfg.CreateParquetTable(glueTableName, name, []glue.PartitionKey{}, schema)
	if err != nil {
		t.Fatalf("[%s] Error creating Glue table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = glueCfg.DeleteTable(glueTableName)
		if err != nil {
			t.Errorf("[%s] error deleting table %v", t.Name(), err)
		}
	})
	glueRoleName := "test_glue_role-" + strings.ToLower(core.GenerateUniqueId())
	glueRoleArn, gluePolicyArn, err := iamCfg.CreateRoleWithActionsResource(glueRoleName, "firehose.amazonaws.com", []string{
		"glue:GetTable",
		"glue:GetTableVersion",
		"glue:GetTableVersions",
	}, []string{
		"arn:aws:glue:" + envCfg.Region + ":*:catalog",
		"arn:aws:glue:" + envCfg.Region + ":*:database/" + glueDbName,
		"arn:aws:glue:" + envCfg.Region + ":*:table/" + glueDbName + "/" + glueTableName,
	})
	if err != nil {
		t.Fatalf("[%s] Error creating Glue role: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = iamCfg.DeletePolicy(gluePolicyArn)
		if err != nil {
			t.Errorf("[%s] error deleting policy %v", t.Name(), err)
		}
	})
	t.Cleanup(func() {
		err = iamCfg.DeleteRole(glueRoleName)
		if err != nil {
			t.Errorf("[%s] error deleting role with attached policies %v", t.Name(), err)
		}
	})

	roleName := "firehose-unit-test-role-data-format-conversion-" + strings.ToLower(core.GenerateUniqueId())
	roleArn, policyArn, err := iamCfg.CreateRoleWithActionsResource(
		roleName,
		"firehose.amazonaws.com",
		[]string{
			"s3:PutObject",
			"s3:PutObjectAcl",
			"s3:ListBucket",
		},
		resources)
	if err != nil {
		t.Fatalf("[%s] error %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = iamCfg.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[%s] error deleting policy %v", t.Name(), err)
		}
	})
	t.Cleanup(func() {
		err = iamCfg.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[%s] error deleting role with attached policies %v", t.Name(), err)
		}
	})

	t.Logf("Waiting 10 seconds role to propagate")
	time.Sleep(10 * time.Second)

	t.Logf("0-Creating stream %v", streamName)
	cfg, err := InitAndCreateWithPartitioningAndDataFormatConversion(
		streamName,
		bucketArn,
		roleArn,
		[]PartitionKeyVal{{Key: "folder", Val: ".folder"}, {Key: "subfolder", Val: ".subfolder"}},
		envCfg.Region,
		core.TEST_SOLUTION_ID,
		core.TEST_SOLUTION_VERSION,
		&firehose.DataFormatConversionConfiguration{
			SchemaConfiguration: &firehose.SchemaConfiguration{
				Region:       &envCfg.Region,
				DatabaseName: &glueCfg.DbName,
				TableName:    &glueTableName,
				RoleARN:      &glueRoleArn,
				VersionId:    aws.String("LATEST"),
			},
			Enabled: aws.Bool(true),
			InputFormatConfiguration: &firehose.InputFormatConfiguration{
				Deserializer: &firehose.Deserializer{
					HiveJsonSerDe: &firehose.HiveJsonSerDe{},
				},
			},
			OutputFormatConfiguration: &firehose.OutputFormatConfiguration{
				Serializer: &firehose.Serializer{
					ParquetSerDe: &firehose.ParquetSerDe{
						Compression: aws.String("GZIP"),
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("[%s] error creating stream: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		t.Logf("4-Deleting stream %v", streamName)
		err = cfg.Delete(streamName)
		if err != nil {
			t.Errorf("[%s] error deleting stream: %v", t.Name(), err)
		}
	})

	_, err = cfg.WaitForStreamCreation(streamName, 120)
	if err != nil {
		t.Fatalf("[%s] timeout waiting for stream creation: %v", t.Name(), err)
	}
	t.Logf("1-Describing Stream %v", streamName)
	stream, err := cfg.Describe(streamName)
	if err != nil {
		t.Fatalf("[%s] error describing stream: %v", t.Name(), err)
	}
	t.Logf("Stream: %+v created", stream)

	t.Logf("2-Sending 5 records to stream %v", streamName)
	records := []Record{
		{
			Data: testRec1,
		},
		{
			Data: testRec2,
		},
		{
			Data: testRec3,
		},
		{
			Data: testRec4,
		},
		{
			Data: testRec5,
		},
	}
	errs, err := cfg.PutRecords(streamName, records)
	if err != nil {
		t.Fatalf("[%s] error sending data to stream: %v. %+v", t.Name(), err, errs)
	}
	err = WaitForData(s3c, 600)
	if err != nil {
		t.Fatalf("[%s] error waiting for data: %v", t.Name(), err)
	}

	t.Log("Waiting 10 seconds to allow remaining records to arrive")
	time.Sleep(10 * time.Second)

	t.Logf("3-Listing S3 bucket %v", name)
	res, err := s3c.Search("", 100)
	t.Logf("S3 bucket constains %v items: %+v", len(res), res)
	if err != nil {
		t.Fatalf("Could not list results in S3")
	}

	if len(res) != 5 {
		t.Errorf("Bucket Should have 5 item not %v", len(res))
	}

	found := map[string]bool{}
	for _, r := range res {
		found[r[:28]] = true
		data, err := s3c.GetParquetObj(r)
		if err != nil {
			t.Errorf("Could not get object from S3")
		}
		t.Logf("found: %+v", data)
	}

	if len(found) != 5 {
		t.Errorf("Should have found 5 record in S3")
	}

	for _, f := range []string{"data/folder=f1/subfolder=sf1", "data/folder=f1/subfolder=sf2", "data/folder=f2/subfolder=sf3", "data/folder=f2/subfolder=sf4", "data/folder=f2/subfolder=sf5"} {
		if !found[f] {
			t.Errorf("Record with prefix %v not found in S3", f)
		}
	}
}

func TestStreamCrud(t *testing.T) {
	t.Parallel()

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	streamName := "tah_core_test_stream_firehose_" + strings.ToLower(core.GenerateUniqueId())
	testRec1 := "{\"key\":\"test record 1\"}"
	testRec2 := "{\"key\":\"test record 2\"}"

	s3c, err := s3.InitWithRandBucket("tah-core-test-bucket-firehose", "", envCfg.Region, "", "")
	if err != nil {
		t.Fatalf("[%s] Failed create %+v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = s3c.EmptyAndDelete()
		if err != nil {
			t.Errorf("[%s] Failed delete %+v", t.Name(), err)
		}
	})
	name := s3c.Bucket
	bucketArn := "arn:aws:s3:::" + name
	resources := []string{bucketArn, bucketArn + "/*"}
	actions := []string{"s3:PutObject"}
	principal := map[string][]string{"Service": {"firehose.amazonaws.com"}}
	err = s3c.AddPolicy(name, resources, actions, principal)
	if err != nil {
		t.Fatalf("[%s] error adding bucket policy %+v", t.Name(), err)
	}

	iamCfg := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	roleName := "firehose-unit-test-role-" + strings.ToLower(core.GenerateUniqueId())
	roleArn, policyArn, err := iamCfg.CreateRoleWithActionsResource(roleName, "firehose.amazonaws.com", []string{"s3:PutObject",
		"s3:PutObjectAcl",
		"s3:ListBucket"}, resources)
	if err != nil {
		t.Fatalf("[%s] error %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = iamCfg.DeletePolicy(policyArn)
		if err != nil {
			t.Errorf("[%s] error deleting policy %v", t.Name(), err)
		}
	})
	t.Cleanup(func() {
		err = iamCfg.DeleteRole(roleName)
		if err != nil {
			t.Errorf("[%s] error deleting role %v", t.Name(), err)
		}
	})

	t.Logf("Waiting 10 seconds role to propagate")
	time.Sleep(10 * time.Second)

	t.Logf("0-Creating stream %v", streamName)
	cfg, err := InitAndCreate(streamName, bucketArn, roleArn, envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error creating stream: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = cfg.Delete(streamName)
		if err != nil {
			t.Errorf("[%s] error deleting stream: %v", t.Name(), err)
		}
	})
	_, err = cfg.WaitForStreamCreation(streamName, 120)
	if err != nil {
		t.Fatalf("[%s] timeout waiting for stream creation: %v", t.Name(), err)
	}
	t.Logf("1-Describing Stream %v", streamName)
	stream, err := cfg.Describe(streamName)
	if err != nil {
		t.Fatalf("[%s] error describing stream: %v", t.Name(), err)
	}
	t.Logf("Stream: %+v created", stream)

	t.Logf("2-Sending 2 records to stream %v", streamName)
	records := []Record{
		{
			Data: testRec1,
		},
		{
			Data: testRec2,
		}}
	errs, err := cfg.PutRecords(streamName, records)
	if err != nil {
		t.Fatalf("[%s] error sending data to stream: %v. %+v", t.Name(), err, errs)
	}
	err = WaitForData(s3c, 600)
	if err != nil {
		t.Fatalf("[%s] error waiting for data: %v", t.Name(), err)
	}

	res, err := s3c.Search("", 100)
	t.Logf("S3 bucket contains %v items: %+v", len(res), res)
	if err != nil {
		t.Fatalf("Could not list results in S3")
	}
	if len(res) == 0 {
		t.Fatalf("Bucket Should have at least 1 item")
	}
	found := map[string]bool{}
	for _, r := range res {
		data, err := s3c.GetTextObj(r)
		if err != nil {
			t.Errorf("Could not get object from S3")
		}
		objs := strings.Split(data, "\n")
		for _, obj := range objs {
			if obj == testRec1 || obj == testRec2 {
				found[obj] = true
				t.Logf("Found record in S3: %v", obj)
			} else {
				t.Errorf("Record not found in S3: %v", obj)
			}
		}
	}

	if len(found) != 2 {
		t.Errorf("Should have found 2 records in S3")
	}
}
