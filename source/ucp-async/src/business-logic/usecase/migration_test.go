// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"context"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/ucp-common/src/feature"
	"tah/upt/source/ucp-common/src/migrate"
	model "tah/upt/source/ucp-common/src/model/async"
	"tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

func TestCpTravelerIdMappingMigration(t *testing.T) {
	t.Parallel()

	targetVersion := feature.AccpTravelerIdMappingIntroducedVersion
	originalVersion := targetVersion - 1
	if originalVersion != 1 {
		t.Fatalf("unexpected migration version: %d", feature.AccpTravelerIdMappingIntroducedVersion)
	}

	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	resources := testSetup(t, t.Name())
	domainName := resources.testDomain
	mockCognitoFunctions(resources, domainName)
	env := "test-env"
	createDomainUsecase := InitCreateDomain()

	streamName := "stream_" + core.GenerateUniqueId()
	kcfg, err := kinesis.InitAndCreate(streamName, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("error initializing kinesis client: %v", err)
	}
	t.Cleanup(func() {
		err = kcfg.Delete(kcfg.Stream)
		if err != nil {
			t.Errorf("error deleting stream: %v", err)
		}
	})
	description, err := kcfg.Describe()
	if err != nil {
		t.Fatalf("error describing stream: %v", err)
	}
	streamArn := description.Arn

	createDomainBody := usecase.CreateDomainBody{
		KmsArn:                resources.keyArn,
		AccpDlq:               resources.qUrl,
		Env:                   env,
		MatchBucketName:       resources.bucketNameMatch,
		AccpSourceBucketName:  resources.bucketName,
		AccpDestinationStream: resources.destinationStream.Arn,
		DomainName:            resources.testDomain,
		HotelBookingBucket:    "hotel-booking-bucket",
		AirBookingBucket:      "air-booking-bucket",
		GuestProfileBucket:    "guest-profile-bucket",
		PaxProfileBucket:      "pax-profile-bucket",
		StayRevenueBucket:     "stay-revenue-bucket",
		ClickstreamBucket:     "clickstream-bucket",
		CSIBucket:             "customer-service-interaction-bucket",
		GlueSchemaPath:        resources.glueSchemaPath,
		RequestedVersion:      originalVersion,
	}
	payload := model.AsyncInvokePayload{
		EventID:       "test-event-id",
		Usecase:       model.USECASE_CREATE_DOMAIN,
		TransactionID: "test-tx-id",
		Body:          createDomainBody,
	}
	services := model.Services{
		AccpConfig:        resources.profiles,
		ConfigDB:          resources.configDbClient,
		ErrorDB:           resources.errorDbClient,
		SolutionsConfig:   resources.solutionsConfig,
		CognitoConfig:     resources.cognitoClient,
		GlueConfig:        resources.glueClient,
		PrivacyResultsDB:  resources.privacyDbClient,
		RetryLambdaConfig: resources.lambdaConfig,
		PortalConfigDB:    resources.portalConfigDbClient,
		AsyncLambdaConfig: resources.asyncLambdaConfig,
		Env: map[string]string{
			"S3_EXPORT_BUCKET":              resources.s3ExportBucketName,
			"GLUE_EXPORT_TABLE_NAME":        resources.travelerTableName,
			"TRAVELER_S3_ROOT_PATH":         "profiles",
			"CUSTOMER_PROFILE_STORAGE_MODE": "true",
			"DYNAMO_STORAGE_MODE":           "true",
			"CP_EXPORT_STREAM":              streamArn,
		},
	}
	err = createDomainUsecase.Execute(payload, services, resources.tx)
	if err != nil {
		t.Fatalf("[%s] Error executing create domain usecase %v", t.Name(), err)
	}

	t.Cleanup(func() {
		deleteDomainUsecase := InitDeleteDomain()
		deleteDomainBody := usecase.DeleteDomainBody{
			Env:        env,
			DomainName: domainName,
		}
		payload = model.AsyncInvokePayload{
			EventID:       "test-event-delete",
			Usecase:       model.USECASE_DELETE_DOMAIN,
			TransactionID: "test-tx-id-create",
			Body:          deleteDomainBody,
		}
		services = model.Services{
			AccpConfig:       resources.profiles,
			ConfigDB:         resources.configDbClient,
			SolutionsConfig:  resources.solutionsConfig,
			CognitoConfig:    resources.cognitoClient,
			GlueConfig:       resources.glueClient,
			PrivacyResultsDB: resources.privacyDbClient,
			PortalConfigDB:   resources.portalConfigDbClient,
			Env: map[string]string{
				"GLUE_EXPORT_TABLE_NAME": resources.travelerTableName,
				"TRAVELER_S3_ROOT_PATH":  "profiles",
			},
		}
		err = deleteDomainUsecase.Execute(payload, services, resources.tx)
		if err != nil {
			t.Errorf("[%s] Error executing delete domain usecase %v", t.Name(), err)
		}
	})

	lcs := resources.profiles
	err = lcs.SetDomain(domainName)
	if err != nil {
		t.Fatalf("error setting domain: %v", err)
	}
	version, err := lcs.GetVersion()
	if err != nil {
		t.Fatalf("error getting version: %v", err)
	}
	expectedOriginalVersion := feature.FeatureSetVersion{Version: originalVersion, CompatibleVersion: originalVersion}
	if version != expectedOriginalVersion {
		t.Fatalf("expected version to be %v, got %v", expectedOriginalVersion, version)
	}

	accp := customerprofiles.InitWithDomain(domainName, envCfg.Region, "", "")
	mappings, err := accp.GetMappings()
	if err != nil {
		t.Fatalf("error getting mappings: %v", err)
	}
	for _, mapping := range mappings {
		for _, field := range mapping.Fields {
			if field.Source == "_source.traveller_id" {
				t.Fatalf("found mapping with unexpected field: %v", mapping.Name)
			}
		}
	}

	migrations, err := migrate.InitMigrations(context.Background(), resources.tx, migrate.Services{Accp: accp}, true)
	if err != nil {
		t.Fatalf("error initializing migrations: %v", err)
	}
	migrationFramework, err := migrate.InitMigrationFramework(lcs, migrations, resources.tx, migrate.AbortOnFailure)
	if err != nil {
		t.Fatalf("error initializing migration framework: %v", err)
	}
	ctx := context.Background()
	err = migrationFramework.Migrate(ctx, targetVersion)
	if err != nil {
		t.Fatalf("error migrating to target version: %v", err)
	}

	version, err = lcs.GetVersion()
	if err != nil {
		t.Fatalf("error getting version: %v", err)
	}
	expectedMigratedVersion := feature.FeatureSetVersion{Version: targetVersion, CompatibleVersion: targetVersion}
	if version != expectedMigratedVersion {
		t.Fatalf("expected version to be %v, got %v", expectedMigratedVersion, version)
	}

	mappings, err = accp.GetMappings()
	if err != nil {
		t.Fatalf("error getting mappings: %v", err)
	}
	for _, mapping := range mappings {
		if mapping.Name == "profileFieldOjectType" {
			// no traveler id mapping on profile level
			continue
		}
		foundMappingForTravelerId := false
		for _, field := range mapping.Fields {
			if field.Source == "_source.traveller_id" {
				foundMappingForTravelerId = true
				break
			}
		}
		if !foundMappingForTravelerId {
			t.Errorf("missing mapping for traveler ID: %v", mapping.Name)
		}
	}

	err = migrationFramework.Migrate(ctx, originalVersion)
	if err != nil {
		t.Fatalf("error migrating to original version: %v", err)
	}

	version, err = lcs.GetVersion()
	if err != nil {
		t.Fatalf("error getting version: %v", err)
	}
	if version != expectedOriginalVersion {
		t.Fatalf("expected version to be %v, got %v", expectedOriginalVersion, version)
	}

	mappings, err = accp.GetMappings()
	if err != nil {
		t.Fatalf("error getting mappings: %v", err)
	}
	// todo: examine more robustly (possibly delegate to a more specific migration test)
	for _, mapping := range mappings {
		for _, field := range mapping.Fields {
			if field.Source == "_source.traveller_id" {
				t.Errorf("found mapping with unexpected field: %v", mapping.Name)
			}
		}
	}

	// todo: check domain version
}

// todo: test no-op migration
