package validator

import (
	"log"
	"os"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/s3"
	common "tah/ucp/src/business-logic/common"
	model "tah/ucp/src/business-logic/model"
	"testing"
)

var UCP_REGION = getRegion()

func TestValidator(t *testing.T) {
	tx := core.NewTransaction("test_validator", "")
	uc := Usecase{Cx: common.Init(&tx, UCP_REGION)}
	bucketName := "test_bucket_name"
	object := "test/object.json"
	mappings := []customerprofiles.FieldMapping{
		customerprofiles.FieldMapping{
			Type:   "STRING",
			Source: "_source.creationChannelId",
			Target: "_order.Name",
		},
		customerprofiles.FieldMapping{
			Type:    "STRING",
			Source:  "_source.id",
			Target:  "_order.Attributes.confirmationNumber",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		customerprofiles.FieldMapping{
			Type:        "STRING",
			Source:      "_source.loyaltyId",
			Target:      "_profile.AccountNumber",
			Searcheable: true,
			//TODO: this index should go on a dedicated customer ID field
			Indexes: []string{"PROFILE"},
		},
	}
	data := [][]string{
		[]string{
			"model_version",
			"object_type",
			"creationChannelId",
			"id",
			"loyaltyId",
		},
		[]string{
			"1",
			"booking",
			"pms",
			"BD876SDH",
			"123456789",
		},
		[]string{
			"1",
			"",
			"",
			"",
			"",
		},
	}
	valErrs, err := uc.ValidateHeaderRow(object, bucketName, data[0], mappings)
	if err != nil {
		t.Errorf("Error during header validation: %+v", err)
	}
	if len(valErrs) > 0 {
		t.Errorf("Validator should not return errors for Valid CSV header %v but returns: %+v", data[0], valErrs)

	}
	valErrs, err = uc.ValidateDataRow(object, bucketName, 1, data[1], data[0], mappings)
	if err != nil {
		t.Errorf("Error during data row validation: %+v", err)
	}
	if len(valErrs) > 0 {
		t.Errorf("Validator should not return errors for Valid CSV row %v but returns: %+v", data[1], valErrs)

	}
	valErrs, err = uc.ValidateDataRow(object, bucketName, 1, data[2], data[0], mappings)
	if err != nil {
		t.Errorf("Error during data row validation: %+v", err)
	}
	if len(valErrs) != 4 {
		t.Errorf("Validator should not return 4 errors for CSV row %v but returns %+v erros: %+v", data[1], len(valErrs), valErrs)
	}
}

func TestEndToEndValidation(t *testing.T) {
	tx := core.NewTransaction("test_validator", "")
	uc := Usecase{Cx: common.Init(&tx, UCP_REGION)}
	mappings := []customerprofiles.FieldMapping{
		customerprofiles.FieldMapping{
			Type:   "STRING",
			Source: "_source.creationChannelId",
			Target: "_order.Name",
		},
		customerprofiles.FieldMapping{
			Type:    "STRING",
			Source:  "_source.id",
			Target:  "_order.Attributes.confirmationNumber",
			Indexes: []string{"UNIQUE", "ORDER"},
		},
		customerprofiles.FieldMapping{
			Type:        "STRING",
			Source:      "_source.loyaltyId",
			Target:      "_profile.AccountNumber",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
	}
	log.Printf("Testing CSV Validation")
	s3c, err := s3.InitWithRandBucket("ucp-test-accp-validation", "", UCP_REGION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}
	log.Printf("Uploading CSV ../../../test_assets/tes_validator.csv")
	err = s3c.UploadFile("bookings/tes_validator.csv", "../../../test_assets/tes_validator.csv")
	if err != nil {
		t.Errorf("[TestEndToEndValidation] Failed to upload CSV file: %+v", err)
	}
	valErrs, err2 := uc.ValidateAccpRecords(model.PaginationOptions{Page: 0, PageSize: 100}, s3c.Bucket, "bookings", mappings)
	if err2 != nil {
		t.Errorf("Error during CSV file validation: %+v", err2)
	}
	if len(valErrs) != 4 {
		t.Errorf("Validator should return 4 errors for CSV but returns %+v errors: %+v", len(valErrs), valErrs)
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestEndToEndValidation] Failed to empty and delete bucket %+v", err)
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
