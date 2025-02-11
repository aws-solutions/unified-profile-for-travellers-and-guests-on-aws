// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"encoding/json"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/sqs"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	accpmappings "tah/upt/source/ucp-common/src/model/accp-mappings"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/test"
	"tah/upt/source/ucp-common/src/utils/utils"
	"testing"
)

func TestTraveler(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("[%s] failed to load env config", t.Name())
	}

	// Create ACCP event stream destination
	testDomain := "traveler_" + strings.ToLower(core.GenerateUniqueId())
	destinationStreamName := "accp-event-stream-traveler-test-" + core.GenerateUniqueId()
	kinesisConfig, err := kinesis.InitAndCreate(destinationStreamName, envCfg.Region, "", "", core.LogLevelDebug)
	if err != nil {
		t.Fatalf("[%s] Error creating kinesis config %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = kinesisConfig.Delete(destinationStreamName)
		if err != nil {
			t.Errorf("[%s] Error deleting stream %v", t.Name(), err)
		}
	})
	_, err = kinesisConfig.WaitForStreamCreation(300)
	if err != nil {
		t.Fatalf("[%s] Error waiting for kinesis config to be created %v", t.Name(), err)
	}
	lcsConfigTable, err := db.InitWithNewTable("lcs-test-"+testDomain, "pk", "sk", "", "")
	if err != nil {
		t.Fatalf("error creating lcs config table: %v", err)
	}
	t.Cleanup(func() {
		err = lcsConfigTable.DeleteTable(lcsConfigTable.TableName)
		if err != nil {
			t.Errorf("error deleting lcs config table: %v", err)
		}
	})
	err = lcsConfigTable.WaitForTableCreation()
	if err != nil {
		t.Fatalf("error waiting for lcs config table: %v", err)
	}
	mergeQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = mergeQueueClient.CreateRandom("mergerTravelerTest")
	if err != nil {
		t.Errorf("error creating merge queue: %v", err)
	}
	t.Cleanup(func() {
		err = mergeQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting merge queue: %v", err)
		}
	})
	cpWriterQueueClient := sqs.Init(envCfg.Region, "", "")
	_, err = cpWriterQueueClient.CreateRandom("cpWriterTravelerTest")
	if err != nil {
		t.Errorf("error creating cp writer queue: %v", err)
	}
	t.Cleanup(func() {
		err = cpWriterQueueClient.Delete()
		if err != nil {
			t.Errorf("error deleting cp writer queue: %v", err)
		}
	})
	initParams := test.ProfileStorageParams{
		LcsConfigTable:      &lcsConfigTable,
		LcsKinesisStream:    kinesisConfig,
		MergeQueueClient:    &mergeQueueClient,
		CPWriterQueueClient: &cpWriterQueueClient,
	}
	profileStorage, err := test.InitProfileStorage(initParams)
	if err != nil {
		t.Fatalf("[%s] failed to init profile storage: %v", t.Name(), err)
	}

	// Create domain
	options := customerprofiles.DomainOptions{
		AiIdResolutionOn:       false,
		RuleBaseIdResolutionOn: true,
	}
	err = profileStorage.CreateDomainWithQueue(testDomain, "", map[string]string{"envName": "travelerTest"}, "", "", options)
	if err != nil {
		t.Fatalf("[%s] Error creating domain %v", t.Name(), err)
	}

	// Defer deleting domain
	t.Cleanup(func() {
		err = profileStorage.DeleteDomain()
		if err != nil {
			t.Errorf("[%s] Error deleting domain", t.Name())
		}
	})

	// Create mappings
	for _, rec := range admin.ACCP_RECORDS {
		objectMapping := accpmappings.ACCP_OBJECT_MAPPINGS[rec.Name]()
		err = profileStorage.CreateMapping(rec.Name, "test mapping for "+rec.Name, objectMapping.Fields)
		if err != nil {
			t.Errorf("[%s] Error creating mapping %v", t.Name(), err)
		}
	}

	// Ingest profiles
	obj1, err := json.Marshal(
		map[string]interface{}{
			"accp_object_id":               "ROM05|MEA0GZX5SC|G0WBL6G1JL|2023-12-15|QAHSGEEHQD",
			"add_on_codes":                 "",
			"add_on_descriptions":          "",
			"add_on_names":                 "",
			"address_billing_city":         "",
			"address_billing_country":      "",
			"address_billing_is_primary":   false,
			"address_billing_line1":        "",
			"address_billing_line2":        "",
			"address_billing_line3":        "",
			"address_billing_line4":        "",
			"address_billing_postal_code":  "",
			"address_billing_province":     "",
			"address_billing_state":        "",
			"address_business_city":        "",
			"address_business_country":     "",
			"address_business_is_primary":  false,
			"address_business_line1":       "",
			"address_business_line2":       "",
			"address_business_line3":       "",
			"address_business_line4":       "",
			"address_business_postal_code": "",
			"address_business_province":    "",
			"address_business_state":       "",
			"address_city":                 "Miami",
			"address_country":              "",
			"address_is_primary":           false,
			"address_line1":                "100 Wallaby Lane",
			"address_line2":                "",
			"address_line3":                "",
			"address_line4":                "",
			"address_mailing_city":         "Indianapolis",
			"address_mailing_country":      "IQ",
			"address_mailing_is_primary":   false,
			"address_mailing_line1":        "8874 Lightsstad",
			"address_mailing_line2":        "more content line 2",
			"address_mailing_line3":        "more content line 3",
			"address_mailing_line4":        "more content line 4",
			"address_mailing_postal_code":  "17231",
			"address_mailing_province":     "AB",
			"address_mailing_state":        "",
			"address_postal_code":          "",
			"address_province":             "",
			"address_state":                "",
			"address_type":                 "mailing",
			"attribute_codes":              "MINI_BAR|BALCONY",
			"attribute_descriptions":       "Mini bar with local snacks and beverages|Large balcony with chairs and a table",
			"attribute_names":              "Mini bar|Balcony",
			"booker_id":                    "1354804801",
			"booking_id":                   "MEA0GZX5SC",
			"cc_cvv":                       "",
			"cc_exp":                       "",
			"cc_name":                      "",
			"cc_token":                     "",
			"cc_type":                      "",
			"check_in_date":                "2023-12-15",
			"company":                      "Panjiva",
			"creation_channel_id":          "gds",
			"crs_id":                       "",
			"date_of_birth":                "1985-05-31",
			"email":                        "",
			"email_business":               "",
			"email_primary":                "",
			"first_name":                   "Flossie",
			"gds_id":                       "",
			"gender":                       "other",
			"honorific":                    "",
			"hotel_code":                   "ROM05",
			"job_title":                    "Executive",
			"language_code":                "",
			"language_name":                "",
			"last_booking_id":              "MEA0GZX5SC",
			"last_name":                    "Wolf",
			"last_update_channel_id":       "",
			"last_updated":                 "2024-04-09T09:39:49.341150Z",
			"last_updated_by":              "Lucinda Lindgren",
			"last_updated_partition":       "2024-04-09-09",
			"middle_name":                  "Noe",
			"model_version":                "1.1.0",
			"n_guests":                     1,
			"n_nights":                     1,
			"nationality_code":             "",
			"nationality_name":             "",
			"object_type":                  "hotel_booking",
			"payment_type":                 "cash",
			"phone":                        "",
			"phone_business":               "",
			"phone_home":                   "+1234567890",
			"phone_mobile":                 "",
			"phone_primary":                "",
			"pms_id":                       "",
			"product_id":                   "QAHSGEEHQD",
			"pronoun":                      "he",
			"rate_plan_code":               "XMAS_WEEK",
			"rate_plan_description":        "Special rate for the chrismas season",
			"rate_plan_name":               "Xmas special rate",
			"room_type_code":               "DBL",
			"room_type_description":        "Room with Double bed",
			"room_type_name":               "Double room",
			"segment_id":                   "G0WBL6G1JL",
			"status":                       "confirmed",
			"total_segment_after_tax":      289.18982,
			"total_segment_before_tax":     251.4694,
			"traveller_id":                 "1354804801",
			"tx_id":                        "34cc8864-5484-4e6f-bca7-711b8856ee5f",
		},
	)
	obj2, err := json.Marshal(
		map[string]interface{}{
			"accp_object_id":               "ROM05|MEA0GZX5SC|G0WBL6G1JL|2023-12-15|QAHSGEEHQD",
			"add_on_codes":                 "",
			"add_on_descriptions":          "",
			"add_on_names":                 "",
			"address_billing_city":         "",
			"address_billing_country":      "",
			"address_billing_is_primary":   false,
			"address_billing_line1":        "",
			"address_billing_line2":        "",
			"address_billing_line3":        "",
			"address_billing_line4":        "",
			"address_billing_postal_code":  "",
			"address_billing_province":     "",
			"address_billing_state":        "",
			"address_business_city":        "",
			"address_business_country":     "",
			"address_business_is_primary":  false,
			"address_business_line1":       "",
			"address_business_line2":       "",
			"address_business_line3":       "",
			"address_business_line4":       "",
			"address_business_postal_code": "",
			"address_business_province":    "",
			"address_business_state":       "",
			"address_city":                 "Miami",
			"address_country":              "",
			"address_is_primary":           false,
			"address_line1":                "100 Wallaby Lane",
			"address_line2":                "",
			"address_line3":                "",
			"address_line4":                "",
			"address_mailing_city":         "Indianapolis",
			"address_mailing_country":      "IQ",
			"address_mailing_is_primary":   false,
			"address_mailing_line1":        "",
			"address_mailing_line2":        "more content line 2",
			"address_mailing_line3":        "more content line 3",
			"address_mailing_line4":        "more content line 4",
			"address_mailing_postal_code":  "17231",
			"address_mailing_province":     "AB",
			"address_mailing_state":        "",
			"address_postal_code":          "",
			"address_province":             "",
			"address_state":                "",
			"address_type":                 "mailing",
			"attribute_codes":              "MINI_BAR|BALCONY",
			"attribute_descriptions":       "Mini bar with local snacks and beverages|Large balcony with chairs and a table",
			"attribute_names":              "Mini bar|Balcony",
			"booker_id":                    "1354804801",
			"booking_id":                   "MEA0GZX5SC",
			"cc_cvv":                       "",
			"cc_exp":                       "",
			"cc_name":                      "",
			"cc_token":                     "",
			"cc_type":                      "",
			"check_in_date":                "2023-12-15",
			"company":                      "Panjiva",
			"creation_channel_id":          "gds",
			"crs_id":                       "",
			"date_of_birth":                "1985-05-31",
			"email":                        "",
			"email_business":               "",
			"email_primary":                "",
			"first_name":                   "Flossie",
			"gds_id":                       "",
			"gender":                       "other",
			"honorific":                    "",
			"hotel_code":                   "ROM05",
			"job_title":                    "Executive",
			"language_code":                "",
			"language_name":                "",
			"last_booking_id":              "MEA0GZX5SC",
			"last_name":                    "Wolf",
			"last_update_channel_id":       "",
			"last_updated":                 "2024-04-09T09:39:49.341150Z",
			"last_updated_by":              "Lucinda Lindgren",
			"last_updated_partition":       "2024-04-09-09",
			"middle_name":                  "Noe",
			"model_version":                "1.1.0",
			"n_guests":                     1,
			"n_nights":                     1,
			"nationality_code":             "",
			"nationality_name":             "",
			"object_type":                  "hotel_booking",
			"payment_type":                 "cash",
			"phone":                        "",
			"phone_business":               "",
			"phone_home":                   "+1234567890",
			"phone_mobile":                 "",
			"phone_primary":                "",
			"pms_id":                       "",
			"product_id":                   "QAHSGEEHQD",
			"pronoun":                      "he",
			"rate_plan_code":               "XMAS_WEEK",
			"rate_plan_description":        "Special rate for the chrismas season",
			"rate_plan_name":               "Xmas special rate",
			"room_type_code":               "DBL",
			"room_type_description":        "Room with Double bed",
			"room_type_name":               "Double room",
			"segment_id":                   "G0WBL6G1JL",
			"status":                       "confirmed",
			"total_segment_after_tax":      289.18982,
			"total_segment_before_tax":     251.4694,
			"traveller_id":                 "1354804801",
			"tx_id":                        "34cc8864-5484-4e6f-bca7-711b8856ee5f",
		},
	)
	if err != nil {
		t.Errorf("[%s] Error marshaling profile %v", t.Name(), err)
	}
	err = profileStorage.PutProfileObject(string(obj1), "hotel_booking")
	if err != nil {
		t.Errorf("[%s] Error ingesting profile %v", t.Name(), err)
	}
	err = profileStorage.PutProfileObject(string(obj2), "hotel_booking")
	if err != nil {
		t.Errorf("[%s] Error ingesting profile %v", t.Name(), err)
	}

	// Search profiles
	profiles, err := profileStorage.SearchProfiles("Address.Address1", []string{"100 Wallaby Lane"})
	if err != nil {
		t.Errorf("[%s] Error searching profiles %v", t.Name(), err)
	}
	if len(profiles) != 1 {
		t.Errorf("[%s] Expected 1 profile from address, got %d", t.Name(), len(profiles))
	}
	profiles, err = profileStorage.SearchProfiles("MailingAddress.Address1", []string{"8874 Lightsstad"})
	if err != nil {
		t.Errorf("[%s] Error searching profiles %v", t.Name(), err)
	}
	if len(profiles) != 1 {
		t.Errorf("[%s] Expected 1 profile from mailing address, got %d", t.Name(), len(profiles))
	}

	uc := NewSearchProfile()
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	uc.SetTx(&tx)
	handlers := registry.ServiceHandlers{
		Accp: profileStorage,
	}
	reg := registry.NewRegistry(envCfg.Region, core.LogLevelDebug, handlers)
	uc.SetRegistry(&reg)
	reqWrapper := model.RequestWrapper{
		SearchRq: model.SearchRq{
			AddressType: "MailingAddress",
			Address1:    "8874 Lightsstad",
		},
	}
	resWrapper, err := uc.Run(reqWrapper)
	if err != nil {
		t.Errorf("[%s] Error running usecase %v", t.Name(), err)
	}
	profilesRes := utils.SafeDereference(resWrapper.Profiles)
	if len(profilesRes) != 1 {
		t.Errorf("[%s] Expected 1 profile from mailing address, got %d", t.Name(), len(profilesRes))
	}
}
