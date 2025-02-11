// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	testutils "tah/upt/source/e2e"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	"tah/upt/source/tah-core/s3"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	commonadmin "tah/upt/source/ucp-common/src/model/admin"
	commontrav "tah/upt/source/ucp-common/src/model/traveler"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	"tah/upt/source/ucp-common/src/utils/config"
	"tah/upt/source/ucp-common/src/utils/utils"
	uptSdk "tah/upt/source/ucp-sdk/src"

	"testing"
	"time"
)

type Services struct {
	matchBucket    s3.S3Config
	matchDb        db.DBConfig
	portalDb       db.DBConfig
	realTimeStream *kinesis.Config
}

func randomDomain(prefix string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(prefix), strings.ToLower(core.GenerateUniqueId()))
}

func TestNonExistentProfile(t *testing.T) {
	var fakeDomain = randomDomain("match_e2e")
	var PathToTestFile = "../../test_data/test_matching/CustomerProfilesPendingMatches-fakedomain.csv"
	var s3Prefix = "Matches/" + fakeDomain + "/fakeId"
	var FileId = "test-file.csv"
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %s", err)
	}
	s3Config := s3.Init(infraCfg.BucketMatch, "", envCfg.Region, "", "")
	dbConfig := db.Init(infraCfg.MatchTableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	portalDb := db.Init(infraCfg.PortalConfigTableName, "config_item", "config_item_category", "", "")

	dynamoRecord := model.DynamoDomainConfig{
		Pk:       "match_threshold",
		Sk:       fakeDomain,
		Value:    "0",
		IsActive: false,
	}

	_, err = portalDb.Save(dynamoRecord)
	if err != nil {
		t.Fatalf("error saving config: %s", err)
	}

	data, err := os.ReadFile(PathToTestFile)
	if err != nil {
		t.Errorf("Error reading file: %s", err)
	}
	err = s3Config.Save(s3Prefix, FileId, data, "text/csv")
	if err != nil {
		t.Errorf("Error saving file: %s", err)
	}

	csvData, err := s3Config.ParseCsvFromS3(s3Prefix + "/" + FileId)
	if err != nil {
		t.Errorf("Error parsing csv: %s", err)
	}
	//Providing 5 seconds for lambda to process the file to dynamo
	time.Sleep(time.Second * 5)
	//The convention I am using is to test on the first profile in the csv
	exampleSourceProfileId := csvData[0][0]
	matchPairs, err := travSvc.SearchMatches(dbConfig, fakeDomain, exampleSourceProfileId)
	if err != nil {
		t.Errorf("Error searching matches: %s", err)
	}
	t.Logf("Number of matches: %d", len(matchPairs))
	if len(matchPairs) != 0 {
		t.Errorf("There should be no matches since the profiles do not exists in LCS or ACCP, unexpected number of matches: %d", len(matchPairs))
	}

	//Cleanup
	for _, line := range csvData {
		sourceId := line[0]
		t.Logf("Source Id is %s", sourceId)
		err = travSvc.DeleteMatches(dbConfig, fakeDomain, sourceId)
		if err != nil {
			t.Errorf("Error deleting matches: %s", err)
		}
	}

	err = s3Config.EmptyBucket()
	if err != nil {
		t.Errorf("Error emptying bucket: %s", err)
	}
}

func TestAutoMergeSimple(t *testing.T) {
	var domainName = randomDomain("match_simple")
	uptSdkHandler, services := CreateTestResources(t, domainName, "0.8", true)

	// 1. Create profiles
	_, profileId1 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "clickstream", "session", "data1")
	_, profileId2 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data1")
	_, profileId3 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "guest_profile", "", "data1")

	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{{ids: []string{profileId1, profileId2, profileId3}, score: "0.90", group: 0}}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	saveCSV(t, services.matchBucket, domainName, fileId)

	// WaitForExpectedCondition waits first; will allow time for lambda to process CSV
	err := testutils.WaitForExpectedCondition(func() bool {
		// merged profiles should have the same connectId
		connectId1, err := uptSdkHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		connectId2, err := uptSdkHandler.GetProfileId(domainName, profileId2)
		if err != nil {
			t.Logf("[%s] error searching profileId2: %s %v", t.Name(), profileId2, err)
			return false
		}

		connectId3, err := uptSdkHandler.GetProfileId(domainName, profileId3)
		if err != nil {
			t.Logf("[%s] error searching profileId3: %s %v", t.Name(), profileId3, err)
			return false
		}
		// Match pair should not exist for auto merged match
		if doesMatchPairExists(t, services.matchDb, domainName, connectId1) {
			t.Fatalf("Match pair found for %s, profile should have been auto merged", connectId1)
		}
		if doesMatchPairExists(t, services.matchDb, domainName, connectId2) {
			t.Fatalf("Match pair found for %s, profile should have been auto merged", connectId2)
		}
		if doesMatchPairExists(t, services.matchDb, domainName, connectId3) {
			t.Fatalf("Match pair found for %s, profile should have been auto merged", connectId3)
		}
		if connectId1 != connectId2 || connectId2 != connectId3 {
			t.Logf("[%s] profiles do not have the same parent\n\tprofileId1: %s -> connectId1: %s\n\tprofileId2: %s -> connectId2: %s\n\tprofileId3: %s -> connectId3: %s",
				t.Name(),
				profileId1, connectId1,
				profileId2, connectId2,
				profileId3, connectId3)
			return false
		}
		return true
	}, 5, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}
}

func TestAutoMergeMixed(t *testing.T) {
	// Setup
	var domainName = randomDomain("match_mixed")
	uptSdkHandler, services := CreateTestResources(t, domainName, "0.8", true)

	// 1. Create profiles
	_, profileId1 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "clickstream", "session", "data1")
	_, profileId2 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data1")
	_, profileId3 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "clickstream", "session", "data2")
	_, profileId4 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data2")

	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{
		{ids: []string{profileId1, profileId2}, score: "0.90", group: 0},
		{ids: []string{profileId3, profileId4}, score: "0.65", group: 1},
	}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	saveCSV(t, services.matchBucket, domainName, fileId)

	// 4. Validate; WaitForExpectedCondition waits first
	err := testutils.WaitForExpectedCondition(func() bool {
		// merged profiles should have the same connectId
		connectId1, err := uptSdkHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		connectId2, err := uptSdkHandler.GetProfileId(domainName, profileId2)
		if err != nil {
			t.Logf("[%s] error searching profileId2: %s %v", t.Name(), profileId2, err)
			return false
		}
		// Match pair should not exist for auto merged match
		if doesMatchPairExists(t, services.matchDb, domainName, connectId1) {
			t.Fatalf("Match pair found for %s, profile should have been auto merged", connectId1)
			return false
		}
		if doesMatchPairExists(t, services.matchDb, domainName, connectId2) {
			t.Fatalf("Match pair found for %s, profile should have been auto merged", connectId2)
			return false
		}
		if connectId1 != connectId2 {
			t.Logf("[%s] profiles do not have the same parent\n\tprofileId1: %s -> connectId1: %s\n\tprofileId2: %s -> connectId2: %s",
				t.Name(),
				profileId1, connectId1,
				profileId2, connectId2)
			return false
		}
		return true
	}, 7, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}

	// WaitForExpectedCondition waits first; will allow time for lambda to process CSV
	err = testutils.WaitForExpectedCondition(func() bool {
		connectId3, err := uptSdkHandler.GetProfileId(domainName, profileId3)
		if err != nil {
			t.Logf("[%s] error searching profileId3: %s %v", t.Name(), profileId3, err)
			return false
		}

		connectId4, err := uptSdkHandler.GetProfileId(domainName, profileId4)
		if err != nil {
			t.Logf("[%s] error searching profileId4: %s %v", t.Name(), profileId4, err)
			return false
		}
		// Match pair should exist for match pair below threshold
		if !doesMatchPairExists(t, services.matchDb, domainName, connectId3) {
			t.Logf("Match pair not found for %s, profile should not have been auto merged", connectId3)
			return false
		}

		// since pair was not auto merged, they should not have the same connectId
		if connectId3 == connectId4 {
			t.Logf("profiles should not be auto merged")
			return false
		}
		return true
	}, 5, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}
}

func TestAutoMergeTurnedOff(t *testing.T) {
	var domainName = randomDomain("match_off")
	uptSdkHandler, services := CreateTestResources(t, domainName, "0", false)

	// 1. Create profiles
	_, profileId1 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "clickstream", "session", "data1")
	_, profileId2 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data1")
	time.Sleep(time.Second * 10)

	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{{ids: []string{profileId1, profileId2}, score: "0.90", group: 0}}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	saveCSV(t, services.matchBucket, domainName, fileId)

	// WaitForExpectedCondition waits first; will allow time for lambda to process CSV
	err := testutils.WaitForExpectedCondition(func() bool {
		// merged profiles should have the different connectId
		connectId1, err := uptSdkHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		connectId2, err := uptSdkHandler.GetProfileId(domainName, profileId2)
		if err != nil {
			t.Logf("[%s] error searching profileId2: %s %v", t.Name(), profileId2, err)
			return false
		}
		// Match pair should exist
		if !doesMatchPairExists(t, services.matchDb, domainName, connectId1) {
			t.Logf("Match pair found for %s, profile should have been auto merged", connectId1)
			return false
		}

		if connectId1 == connectId2 {
			t.Log("profiles have the same parent and shouldn't")
			return false
		}
		return true
	}, 5, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}
}

func TestProfileDataDiff(t *testing.T) {
	var domainName = randomDomain("match_diff")
	uptSdkHandler, services := CreateTestResources(t, domainName, "0", false)

	// 1. Create profiles
	_, profileId1 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data1")
	_, profileId2 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "guest_profile", "", "data1")
	time.Sleep(time.Second * 20)

	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{{ids: []string{profileId1, profileId2}, score: "0.90", group: 0}}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	saveCSV(t, services.matchBucket, domainName, fileId)

	// 4. validate; WaitForExpectedCondition waits first
	var matchPairs []commonadmin.MatchPair
	err := testutils.WaitForExpectedCondition(func() bool {
		connectId1, err := uptSdkHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		matchPairs, err = travSvc.SearchMatches(services.matchDb, domainName, connectId1)
		if err != nil || len(matchPairs) == 0 {
			t.Logf("Match pair not found for %s", connectId1)
			return false
		}
		return true
	}, 5, 15*time.Second)
	if err != nil || len(matchPairs) < 1 {
		t.Fatalf("[%s] error waiting for condition: %v\n\tlen(matchPairs): %d", t.Name(), err, len(matchPairs))
	}
	compareDiffWithExpectedDiff(t, domainName, uptSdkHandler, profileId1, profileId2, matchPairs[0].ProfileDataDiff)
}

func TestMergeUsingProfileId(t *testing.T) {
	// Setup
	domainName := randomDomain("match_profileid")
	uptSdkHandler, services := CreateTestResources(t, domainName, "0.8", true)

	// 1. Create profiles
	_, profileId1 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "hotel_booking", "holder", "data1")
	_, profileId2 := createProfile(t, domainName, services.realTimeStream, uptSdkHandler, "guest_profile", "", "data1")

	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{
		{ids: []string{profileId1, profileId2}, score: "0.90", group: 0},
	}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	saveCSV(t, services.matchBucket, domainName, fileId)

	// Validate
	// WaitForExpectedCondition waits first; will allow time for
	// Matchlambda to process CSV and MergeLambda to perform merge
	err := testutils.WaitForExpectedCondition(func() bool {
		// merged profiles should have the same connectId
		connectId1, err := uptSdkHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		connectId2, err := uptSdkHandler.GetProfileId(domainName, profileId2)
		if err != nil {
			t.Logf("[%s] error searching profileId2: %s %v", t.Name(), profileId2, err)
			return false
		}

		// Match pair should not exist for auto merged match
		if doesMatchPairExists(t, services.matchDb, domainName, connectId1) {
			t.Logf("[%s] Match pair found for %s, profile should have been auto merged", t.Name(), connectId1)
			return false
		}
		if doesMatchPairExists(t, services.matchDb, domainName, connectId2) {
			t.Logf("[%s] Match pair found for %s, profile should have been auto merged", t.Name(), connectId2)
			return false
		}

		if connectId1 != connectId2 {
			t.Logf("[%s] profiles do not have the same parent\n\tconnectId1: %s\n\tconnectId2: %s", t.Name(), connectId1, connectId2)
			return false
		}
		return true
	}, 5, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}
}

func getReflectedStrVal(field reflect.Value) string {
	if field.IsValid() && field.CanInterface() {
		reflectedVal := field.Interface()
		if strVal, ok := reflectedVal.(string); ok {
			return strVal
		}
	}
	return ""
}

func compareDiffWithExpectedDiff(t *testing.T, domainName string, uptSdkHandler uptSdk.UptHandler, profileId1 string, profileId2 string, profileDataDiff []commonadmin.DataDiff) {
	res, err := uptSdkHandler.SearchProfile(domainName, "travellerId", []string{profileId1})
	profiles := utils.SafeDereference(res.Profiles)
	if err != nil || len(profiles) != 1 {
		t.Fatalf("%s", err)
	}
	profile1Obj := reflect.ValueOf(profiles[0])

	res2, err := uptSdkHandler.SearchProfile(domainName, "travellerId", []string{profileId2})
	profiles2 := utils.SafeDereference(res2.Profiles)
	if err != nil || len(profiles2) != 1 {
		t.Fatalf("%s", err)
	}
	profile2Obj := reflect.ValueOf(profiles2[0])
	if len(profileDataDiff) == 0 {
		t.Fatalf("Profile data diff is missing")
	}

	for _, diff := range profileDataDiff {
		if profile1Obj.FieldByName(diff.FieldName).Kind() == reflect.String {
			sourceVal := getReflectedStrVal(profile1Obj.FieldByName(diff.FieldName))
			targetVal := getReflectedStrVal(profile2Obj.FieldByName(diff.FieldName))

			if sourceVal != diff.SourceValue || targetVal != diff.TargetValue {
				t.Fatalf("Profile diff is incorrect")
			}
		} else if strings.HasPrefix(diff.FieldName, "Address") {
			validateAddressDiff(t, diff, profiles[0].HomeAddress, profiles2[0].HomeAddress)
		}
	}
}

func validateAddressDiff(t *testing.T, diff commonadmin.DataDiff, sourceAddr commontrav.Address, targetAddr commontrav.Address) {
	addrType := strings.Split(diff.FieldName, ".")[1]
	if addrType == "Address1" {
		if sourceAddr.Address1 != diff.SourceValue || targetAddr.Address1 != diff.TargetValue {
			t.Fatalf("Profile diff is incorrect")
		}
	} else if addrType == "Address2" {
		if sourceAddr.Address2 != diff.SourceValue || targetAddr.Address2 != diff.TargetValue {
			t.Fatalf("Profile diff is incorrect")
		}
	} else if addrType == "Address3" {
		if sourceAddr.Address3 != diff.SourceValue || targetAddr.Address3 != diff.TargetValue {
			t.Fatalf("Profile diff is incorrect")
		}
	} else if addrType == "City" {
		if sourceAddr.City != diff.SourceValue || targetAddr.City != diff.TargetValue {
			t.Fatalf("Profile diff is incorrect")
		}
	}
}

func doesMatchPairExists(t *testing.T, dbConfig db.DBConfig, domainName string, id string) bool {
	matchPairs, err := travSvc.SearchMatches(dbConfig, domainName, id)
	if err != nil {
		t.Errorf("Error searching matches: %s", err)
	}
	return len(matchPairs) != 0
}

func CreateTestResources(t *testing.T, domainName string, matchThreshold string, isMatchActive bool) (uptSdk.UptHandler, Services) {
	envCfg, infraCfg, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %s", err)
	}

	// Upt Sdk
	uptHandler, err := uptSdk.Init(infraCfg)
	if err != nil {
		t.Fatalf("Error initializing SDK: %v", err)
	}
	err = uptHandler.SetupTestAuth(domainName)
	if err != nil {
		t.Fatalf("Error setting up auth: %v", err)
	}
	t.Cleanup(func() {
		err := uptHandler.CleanupAuth(domainName)
		if err != nil {
			t.Errorf("Error cleaning up auth: %v", err)
		}
	})

	// Create Domain
	err = uptHandler.CreateDomainAndWait(domainName, 300)
	if err != nil {
		t.Fatalf("Error creating domain: %v", err)
	}
	t.Cleanup(func() {
		_, err = uptHandler.DeleteDomain(domainName)
		if err != nil {
			t.Errorf("Error deleting domain %v", err)
		}
	})

	_, err = uptHandler.UpdatePromptConfig(domainName, model.DomainSetting{
		MatchConfig: model.DomainConfig{
			Value:    matchThreshold,
			IsActive: isMatchActive,
		},
		PromptConfig: model.DomainConfig{
			Value:    "",
			IsActive: false,
		},
	})
	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}

	matchBucket := s3.Init(infraCfg.BucketMatch, "", envCfg.Region, "", "")
	matchDb := db.Init(infraCfg.MatchTableName, "domain_sourceProfileId", "match_targetProfileId", "", "")
	portalDb := db.Init(infraCfg.PortalConfigTableName, "config_item", "config_item_category", "", "")
	realTimeStream := kinesis.Init(infraCfg.RealTimeIngestorStreamName, envCfg.Region, "", "", core.LogLevelDebug)

	return uptHandler, Services{
		matchBucket:    matchBucket,
		matchDb:        matchDb,
		portalDb:       portalDb,
		realTimeStream: realTimeStream,
	}
}

type PartialModeOptions struct {
	Fields []string `json:"fields"`
}

type BusinessObjectRecord struct {
	Domain             string                 `json:"domain"`
	ObjectType         string                 `json:"objectType"`
	ModelVersion       string                 `json:"modelVersion"`
	Data               map[string]interface{} `json:"data"`
	Mode               string                 `json:"mode"`
	MergeModeProfileID string                 `json:"mergeModeProfileId"`
	PartialModeOptions PartialModeOptions     `json:"partialModeOptions"`
}

func buildRecord(objectType string, folder string, file string, domain string, mode string, line int) BusinessObjectRecord {
	newMode := ""
	if mode != "" {
		newMode = "_" + mode
	}
	content, err := os.ReadFile("../../test_data/" + folder + "/" + file + newMode + ".jsonl")
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return BusinessObjectRecord{}
	}
	content = []byte(strings.Split(string(content), "\n")[line])

	var v map[string]interface{}
	json.Unmarshal(content, &v)

	return BusinessObjectRecord{
		Domain:       domain,
		ObjectType:   objectType,
		ModelVersion: "1",
		Data:         v,
		Mode:         mode,
	}
}

func createProfile(t *testing.T, domainName string, realTimeStream *kinesis.Config, uptSdkHandler uptSdk.UptHandler, recordType, attr, file string) (string, string) {
	record := buildRecord(recordType, recordType, file, domainName, "", 0)
	var profileId = ""
	if attr == "" {
		profileId = (record.Data["id"]).(string)
	} else {
		profileId = (record.Data[attr].(map[string]interface{}))["id"].(string)
	}
	record1Serialized, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("error building record, %v", err)
	}

	err, errs := realTimeStream.PutRecords([]kinesis.Record{{
		Pk:   record.ObjectType + "-" + record.ModelVersion + "-1",
		Data: string(record1Serialized),
	}})
	if err != nil {
		t.Fatalf("error sending data to stream: %v. %+v", err, errs)
	}

	traveler, err := uptSdkHandler.WaitForProfile(domainName, profileId, 90)
	if err != nil {
		t.Fatalf("Could not find profile with ID %s for clickstream rec. Error: %v", profileId, err)
	}

	return traveler.ConnectID, profileId
}

type MatchData struct {
	ids   []string
	score string
	group int
}

func writeCSV(matches []MatchData, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, match := range matches {
		for _, id := range match.ids {
			err := writer.Write([]string{id, strconv.Itoa(match.group), match.score})
			if err != nil {
				log.Fatalln("error writing record to file", err)
			}
		}
	}
}

func saveCSV(t *testing.T, s3Config s3.S3Config, domainName string, fileId string) [][]string {
	var s3Prefix = "Matches/" + domainName + "/fakeId"
	data, err := os.ReadFile(fileId)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	err = s3Config.Save(s3Prefix, fileId, data, "text/csv")
	if err != nil {
		t.Fatalf("Error saving file: %s", err)
	}
	csvData, err := s3Config.ParseCsvFromS3(s3Prefix + "/" + fileId)
	if err != nil {
		t.Fatalf("Error parsing csv: %s", err)
	}

	return csvData
}
