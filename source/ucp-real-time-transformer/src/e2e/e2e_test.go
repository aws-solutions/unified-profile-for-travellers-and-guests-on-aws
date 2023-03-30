package kinesis

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"tah/core/core"
	customerprofiles "tah/core/customerprofiles"
	kinesis "tah/core/kinesis"
	kms "tah/core/kms"
	"testing"
	"time"
)

var KINESIS_NAME_REAL_TIME = os.Getenv("KINESIS_NAME_REAL_TIME")
var LAMBDA_NAME_REAL_TIME = os.Getenv("LAMBDA_NAME_REAL_TIME")
var KINESIS_NAME_OUTPUT_REAL_TIME = os.Getenv("KINESIS_NAME_OUTPUT_REAL_TIME")
var UCP_REGION = getRegion()

var fieldMappings = customerprofiles.FieldMappings{
	{
		Type:   "STRING",
		Source: "_source.model_version",
		Target: "_order.Attributes.model_version",
	},
	// Profile Data
	{
		Type:        "STRING",
		Source:      "_source.traveller_id",
		Target:      "_profile.Attributes.profile_id",
		Searcheable: true,
		Indexes:     []string{"PROFILE"},
	},
	//Order Data
	{
		Type:    "STRING",
		Source:  "_source.event_timestamp",
		Target:  "_order.Attributes.timestamp",
		Indexes: []string{"UNIQUE", "ORDER"},
	},
}

type BusinessObjectRecord struct {
	Domain       string                 `json:"domain"`
	ObjectType   string                 `json:"objectType"`
	ModelVersion string                 `json:"modelVersion"`
	Data         map[string]interface{} `json:"data"`
}

func TestKinesis(t *testing.T) {
	log.Printf("E2EDataStream] Running E2E test for data stream")
	incfg := kinesis.Init(KINESIS_NAME_REAL_TIME, UCP_REGION)
	outcfg := kinesis.Init(KINESIS_NAME_OUTPUT_REAL_TIME, UCP_REGION)
	outcfg.InitConsumer(kinesis.ITERATOR_TYPE_LATEST)
	kmsc := kms.Init(UCP_REGION)
	customCfg := customerprofiles.Init(UCP_REGION)
	domainName := "test-domain-water-" + core.GeneratUniqueId()

	keyArn, err1 := kmsc.CreateKey("kinesis-unit-test-key")
	if err1 != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err1)
	}
	err := customCfg.CreateDomain(domainName, true, keyArn, map[string]string{"key": "value"})
	if err != nil {
		t.Errorf("Error creating domain")
	}
	err8 := customCfg.CreateMapping("clickstream", "description of the test mapping", fieldMappings)
	if err8 != nil {
		t.Errorf("Error creating mapping %v", err8)
	}

	log.Printf("E2EDataStream] 1- sending records to kinesis stream %v", KINESIS_NAME_REAL_TIME)
	clickstreamRecord := buildRecord("clickstream", "clickstream", domainName)
	clickstreamSerialized, _ := json.Marshal(clickstreamRecord)
	err, errs := incfg.PutRecords([]kinesis.Record{
		kinesis.Record{
			Pk:   clickstreamRecord.ObjectType + "-" + clickstreamRecord.ModelVersion + "-1",
			Data: string(clickstreamSerialized),
		},
	})
	if err != nil {
		t.Errorf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}

	log.Printf("E2EDataStream] Waiting for profile creation")
	//TODO: extends this to other objects and and fix this ugly code below
	profile_id := ((clickstreamRecord.Data["session"].(map[string]interface{}))["id"]).(string)
	err3 := waitForProfile(customCfg, profile_id, 60)
	if err3 != nil {
		t.Errorf("Could not find profile with ID %s. Error: %v", profile_id, err3)
	}

	log.Printf("E2EDataStream] Cleaning up")
	err4 := customCfg.DeleteDomain()
	if err4 != nil {
		t.Errorf("Error deleting domain %v", err4)
	}
}

func waitForProfile(customCfg customerprofiles.CustomerProfileConfig, profileID string, timeout int) error {
	log.Printf("Waiting %v seconds for profile ID %v creation", timeout, profileID)
	it := 0
	for it*5 < timeout {
		accpProfileID, err := customCfg.GetProfileId(profileID)
		if err != nil {
			log.Printf("Error getting profile, exception thrown %v", err)
			return err
		}
		if accpProfileID != "" {
			log.Printf("Found profile ID: %v", accpProfileID)
			return nil
		}
		log.Printf("Not found. Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return errors.New("Could not find profile ID: timout expired")
}

func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}

func buildRecord(objectType string, folder string, domain string) BusinessObjectRecord {
	content, err := os.ReadFile("../../../test_data/" + folder + "/data1.json")
	log.Printf("%s", string(content))
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return BusinessObjectRecord{}
	}

	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         v,
	}

}
