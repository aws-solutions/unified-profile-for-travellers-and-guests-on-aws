package e2epipeline

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	util "tah/upt/source/e2e"
	"tah/upt/source/e2e/generators"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/s3"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

type MatchData struct {
	ids   []string
	score string
	group int
}

func TestMergeUsingAccpId(t *testing.T) {
	// Setup
	kinesisRate := 1000

	envConfig, infraConfig, err := config.LoadConfigs()
	if err != nil {
		t.Fatalf("Error loading configs: %v", err)
	}

	domainName, uptHandler, _, cpCache, realTimeStream, err := util.InitializeEndToEndTestEnvironment(t, infraConfig, envConfig)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}

	_, err = uptHandler.UpdatePromptConfig(domainName, model.DomainSetting{
		MatchConfig: model.DomainConfig{
			Value:    "0.8",
			IsActive: true,
		},
		PromptConfig: model.DomainConfig{
			Value:    "",
			IsActive: false,
		},
	})
	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}

	matchBucket := s3.Init(infraConfig.BucketMatch, "", envConfig.Region, "", "")
	matchDb := db.Init(infraConfig.MatchTableName, "domain_sourceProfileId", "match_targetProfileId", "", "")

	// 1. Create profiles
	profileId1 := "travellerId1"
	profileId2 := "travellerId2"
	hotelRecords, err := generators.GenerateSimpleHotelStayRecords(profileId1, domainName, 1)
	if err != nil {
		t.Fatalf("Error generating hotel record: %v", err)
	}
	guestRecords, err := generators.GenerateSimpleGuestProfileRecords(profileId2, domainName, 1)
	if err != nil {
		t.Fatalf("Error generating guest profile record: %v", err)
	}
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, hotelRecords)
	util.SendKinesisBatch(t, realTimeStream, kinesisRate, guestRecords)

	//	Add buffer time to allow AltProfId merge rule to settle the ConnectIDs so that the match Lambda grabs the latest one
	time.Sleep(time.Second * 20)

	accpId1s, err := cpCache.SearchProfiles("profile_id", []string{profileId1})
	if err != nil || len(accpId1s) < 1 {
		t.Fatalf("[%s] error finding accp profile: %s", t.Name(), err)
	}
	accpId1 := accpId1s[0].ProfileId

	accpId2s, err := cpCache.SearchProfiles("profile_id", []string{profileId2})
	if err != nil || len(accpId2s) < 1 {
		t.Fatalf("error finding accp profile: %s", err)
	}
	accpId2 := accpId2s[0].ProfileId
	t.Logf("[%s] accpId1: %s\taccpId2: %s", t.Name(), accpId1, accpId2)
	// If CP profile is not indexed, we do not retry here (unlike the cp-writer)
	// this is because the profile might be in another database
	// it is not an issue because the matches happen a few times a week and does not
	// happen with real time ingestion, but the tests do
	// thus, we need to add delay here to allow the profile to be indexed
	time.Sleep(time.Second * 10)
	// 2. Create match csv
	fileId := fmt.Sprintf("custom-match-%s.csv", domainName)
	writeCSV([]MatchData{
		{ids: []string{accpId1, accpId2}, score: "0.90", group: 0},
	}, fileId)
	t.Cleanup(func() {
		os.Remove(fileId)
	})
	// 3. save csv
	csvData := saveCSV(t, matchBucket, domainName, fileId)

	//Validate
	// WaitForExpectedCondition waits first; will allow time for lambda to process CSV
	err = util.WaitForExpectedCondition(func() bool {
		// Match pair should not exist for auto merged match
		connectId1, err := uptHandler.GetProfileId(domainName, profileId1)
		if err != nil {
			t.Logf("[%s] error searching profileId1: %s %v", t.Name(), profileId1, err)
			return false
		}

		connectId2, err := uptHandler.GetProfileId(domainName, profileId2)
		if err != nil {
			t.Logf("[%s] error searching profileId2: %s %v", t.Name(), profileId2, err)
			return false
		}

		if doesMatchPairExists(t, matchDb, domainName, csvData[0][0]) {
			t.Logf("Match pair found for %s, profile should have been auto merged", accpId1)
			return false
		}
		if doesMatchPairExists(t, matchDb, domainName, csvData[1][0]) {
			t.Logf("Match pair found for %s, profile should have been auto merged", accpId2)
			return false
		}

		// merged profiles should have the same connectId
		if connectId1 != connectId2 {
			t.Logf("[%s] profiles do not have the same parent\n\tconnectId1: %s\n\tconnectId2: %s", t.Name(), connectId1, connectId2)
			return false
		}
		return true
	}, 1, 15*time.Second)
	if err != nil {
		t.Fatalf("[%s] error waiting for expected condition: %v", t.Name(), err)
	}
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

func doesMatchPairExists(t *testing.T, dbConfig db.DBConfig, domainName string, id string) bool {
	matchPairs, err := travSvc.SearchMatches(dbConfig, domainName, id)
	if err != nil {
		t.Errorf("Error searching matches: %s", err)
	}
	return len(matchPairs) != 0
}
