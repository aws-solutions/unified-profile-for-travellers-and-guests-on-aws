// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

// ********************************
// Please read:
// skipping this test since we do not currently use Entity Resolution
// if added back, the pipeline will likely fail due to a permissions issue until
// the CodeBuild role includes the principal "entityresolution.amazonaws.com"
// ********************************

// func TestToggleAiEntityResolution(t *testing.T) {
// 	t.Parallel()
// 	envCfg, err := config.LoadEnvConfig()
// 	if err != nil {
// 		t.Fatalf("error loading env config: %v", err)
// 	}
// 	domainName := randomDomain("ai_id_res")
// 	auroraCfg, err := SetupAurora(t)
// 	if err != nil {
// 		t.Errorf("[%s] Error setting up aurora: %+v ", t.Name(), err)
// 		return
// 	}
// 	// DynamoDB
// 	dynamoTableName := domainName
// 	dynamoCfg, err := db.InitWithNewTable(dynamoTableName, configTablePk, configTableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
// 	if err != nil {
// 		t.Errorf("[%s] error init with new table: %v", t.Name(), err)
// 	}
// 	err = dynamoCfg.WaitForTableCreation()
// 	if err != nil {
// 		t.Errorf("[%s] error waiting for table creation: %v", t.Name(), err)
// 	}

// 	//The database is maintained separately. The tables are managed by customerprofilelowcost
// 	dbName := domainName
// 	glueConfig := glue.Init(envCfg.Region, dbName, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
// 	err = glueConfig.CreateDatabase(dbName)
// 	if err != nil {
// 		t.Errorf("[%s] Error creating database: %v", t.Name(), err)
// 	}

// 	//output s3 bucket
// 	outputBucket := "ai-id-res-output"
// 	s3Config, err := s3.InitWithRandBucket(outputBucket, "", envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
// 	if err != nil {
// 		t.Errorf("[%s] Error creating s3 bucket: %v", t.Name(), err)
// 	}
// 	outputBucketName := s3Config.Bucket

// 	//input s3 bucket
// 	inputBucket := "ai-id-res-input"
// 	inputBucketName, err := s3Config.CreateRandomBucket(inputBucket)
// 	if err != nil {
// 		t.Errorf("[%s] Error creating s3 bucket: %v", t.Name(), err)
// 	}

// 	//role
// 	roleName := domainName + "_test_role"
// 	roleArn, err := createRoleTest(t, roleName)
// 	if err != nil {
// 		t.Errorf("[%s] Error creating role: %v", t.Name(), err)
// 	}
// 	iamConfig := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)

// 	kinesisMockCfg := kinesis.InitMock(nil, nil, nil, nil)

// 	mergeQueue := buildTestMergeQueue(t, envCfg.Region)

// 	initVals := CustomerProfileInitOptions{
// 		RoleArn:          roleArn,
// 		S3Bucket:         inputBucketName,
// 		GlueDb:           dbName,
// 		MergeQueueClient: mergeQueue,
// 	}

// 	lc := InitLowCost(envCfg.Region, auroraCfg, &dynamoCfg, kinesisMockCfg, "tah/upt/source/tah-core/customerprofile", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION, &initVals)

// 	err = lc.CreateDomainWithQueue(domainName, "", map[string]string{DOMAIN_CONFIG_ENV_KEY: "testenv"}, "", outputBucketName, DomainOptions{AiIdResolutionOn: true})
// 	if err != nil {
// 		t.Errorf("[%s] Error creating domain: %v", t.Name(), err)
// 	}

// 	objectTypeName := "basic_object"
// 	mapping := []FieldMapping{
// 		// Profile Object Unique Key
// 		{
// 			Type:   "STRING",
// 			Source: "_source.last_booking_id",
// 			Target: "_profile.Attributes.last_booking_id",
// 		},
// 		{
// 			Type:   "STRING",
// 			Source: "_source.address_mailing_state",
// 			Target: "_profile.MailingAddress.State",
// 		},
// 		{
// 			Type:    "STRING",
// 			Source:  "_source.first_name",
// 			Target:  "_profile.FirstName",
// 			KeyOnly: true,
// 		},
// 		{
// 			Type:    "STRING",
// 			Source:  "_source.middle_name",
// 			Target:  "_profile.MiddleName",
// 			KeyOnly: true,
// 		},
// 		{
// 			Type:        "STRING",
// 			Source:      "_source.last_name",
// 			Target:      "_profile.LastName",
// 			Searcheable: true,
// 		},
// 		// Profile Data
// 		{
// 			Type:        "STRING",
// 			Source:      "_source.traveller_id",
// 			Target:      "_profile.Attributes.profile_id",
// 			Searcheable: true,
// 			Indexes:     []string{"PROFILE"},
// 		},
// 		{
// 			Type:    "STRING",
// 			Source:  "_source.accp_object_id",
// 			Target:  "accp_object_id",
// 			Indexes: []string{"UNIQUE"},
// 			KeyOnly: true,
// 		},
// 	}

// 	err = lc.CreateMapping(objectTypeName, domainName, mapping)
// 	if err != nil {
// 		t.Errorf("[%s] CreateMapping failed: %s", t.Name(), err)
// 	}

// 	_, err = lc.EntityConfig.GetMatchingWorkflow(domainName)
// 	if err != nil {
// 		t.Errorf("[%s] Error getting matching workflow, should exist: %v", t.Name(), err)
// 	}

// 	err = lc.ToggleAiIdentityResolution(domainName, inputBucketName, roleArn)
// 	if err != nil {
// 		t.Errorf("[%s] ToggleAiIdentityResolution failed: %s", t.Name(), err)
// 	}

// 	_, err = lc.EntityConfig.GetMatchingWorkflow(domainName)
// 	if err == nil {
// 		t.Errorf("[%s] Matching workflow should not exist: %v", t.Name(), err)
// 	}

// 	//creating and deleting again
// 	err = lc.ToggleAiIdentityResolution(domainName, inputBucketName, roleArn)
// 	if err != nil {
// 		t.Errorf("[%s] ToggleAiIdentityResolution failed: %s", t.Name(), err)
// 	}

// 	_, err = lc.EntityConfig.GetMatchingWorkflow(domainName)
// 	if err != nil {
// 		t.Errorf("[%s] Matching workflow should not exist: %v", t.Name(), err)
// 	}

// 	err = lc.ToggleAiIdentityResolution(domainName, inputBucketName, roleArn)
// 	if err != nil {
// 		t.Errorf("[%s] ToggleAiIdentityResolution failed: %s", t.Name(), err)
// 	}

// 	_, err = lc.EntityConfig.GetMatchingWorkflow(domainName)
// 	if err == nil {
// 		t.Errorf("[%s] Matching workflow should not exist: %v", t.Name(), err)
// 	}

// 	// Cleanup
// 	err = lc.DeleteDomain()
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteDomain failed: %s", err)
// 	}
// 	err = dynamoCfg.DeleteTable(dynamoTableName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteTable failed: %s", err)
// 	}
// 	err = s3Config.EmptySpecificBucket(outputBucketName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteBucket failed: %s", err)
// 	}
// 	err = s3Config.DeleteBucket(outputBucketName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteBucket failed: %s", err)
// 	}
// 	err = s3Config.EmptySpecificBucket(inputBucketName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteBucket failed: %s", err)
// 	}
// 	err = s3Config.DeleteBucket(inputBucketName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteBucket failed: %s", err)
// 	}
// 	err = iamConfig.DeleteRole(roleName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteRole failed: %s", err)
// 	}
// 	err = glueConfig.DeleteDatabase(dbName)
// 	if err != nil {
// 		t.Errorf("[TestFlippingAIEntityResolution] DeleteDatabase failed: %s", err)
// 	}
// }

// func createRoleTest(t *testing.T, roleName string) (string, error) {
// 	iamClient := iam.Init(core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
// 	//link to document where permissions came from
// 	//https://docs.aws.amazon.com/entityresolution/latest/userguide/security-iam-awsmanpol.html
// 	policyDoc := iam.PolicyDocument{
// 		Version: "2012-10-17",
// 		Statement: []iam.StatementEntry{
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"entityresolution:*"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"glue:GetSchema", "glue:SearchTables", "glue:GetSchemaByDefinition", "glue:GetSchemaVersion", "glue:GetSchemaVersionsDiff", "glue:GetDatabase", "glue:GetDatabases", "glue:GetTable", "glue:GetTables", "glue:GetTableVersion", "glue:GetTableVersions"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"s3:ListAllMyBuckets"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"s3:ListBucket", "s3:GetBucketLocation", "s3:ListBucketVersions", "s3:GetBucketVersioning"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"tag:GetTagKeys", "tag:GetTagValues"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"kms:DescribeKey", "kms:ListAliases"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"iam:ListRoles"},
// 				Resource: []string{"*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"iam:PassRole"},
// 				Resource: []string{"arn:aws:iam::*:role/*entityresolution*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"events:PutRule", "events:DeleteRule", "events:PutTargets"},
// 				Resource: []string{"arn:aws:events:*:*:rule/entity-resolution-automatic*"},
// 			},
// 			{
// 				Effect:   "Allow",
// 				Action:   []string{"dataexchange:GetDataSet"},
// 				Resource: []string{"*"},
// 			},
// 		},
// 	}

// 	roleArn, policyArn, err := iamClient.CreateRoleWithPolicy(roleName, "entityresolution.amazonaws.com", policyDoc)
// 	if err != nil {
// 		return "", err
// 	}
// 	t.Cleanup(func() {
// 		err = iamClient.DeleteRole(roleName)
// 		if err != nil {
// 			t.Errorf("Error deleting role: %v", err)
// 		}
// 		err = iamClient.DeletePolicy(policyArn)
// 		if err != nil {
// 			t.Errorf("Error deleting policy: %v", err)
// 		}
// 	})
// 	return roleArn, err
// }
