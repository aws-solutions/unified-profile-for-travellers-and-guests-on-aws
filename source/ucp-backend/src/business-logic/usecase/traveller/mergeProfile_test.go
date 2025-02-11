// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"encoding/json"
	"log"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"

	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/lambda"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	asyncModel "tah/upt/source/ucp-common/src/model/async/usecase"

	commonModel "tah/upt/source/ucp-common/src/model/admin"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

func TestFilterOutOverlapingRq(t *testing.T) {
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{Tx: &tx}
	uc := NewMergeProfile()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	mergeRqs := []commonModel.MergeRq{
		{SourceProfileID: "a1", TargetProfileID: "t1"},
		{SourceProfileID: "a2", TargetProfileID: "t2"},
		{SourceProfileID: "t1", TargetProfileID: "a5"},
		{SourceProfileID: "a3", TargetProfileID: "t3"},
		{SourceProfileID: "b6", TargetProfileID: "t1"},
		{SourceProfileID: "t1", TargetProfileID: "a1"},
	}
	res := uc.filterOutOverlapingRq(mergeRqs)
	if len(res) != 3 {
		t.Errorf("Expected 3, got %d", len(res))
	}
	if res[0].SourceProfileID != "a1" || res[0].TargetProfileID != "t1" {
		t.Errorf("Expected a1, t1, got %s, %s", res[0].SourceProfileID, res[0].TargetProfileID)
	}
	if res[1].SourceProfileID != "a2" || res[1].TargetProfileID != "t2" {
		t.Errorf("Expected a1, t1, got %s, %s", res[1].SourceProfileID, res[1].TargetProfileID)
	}
	if res[2].SourceProfileID != "a3" || res[2].TargetProfileID != "t3" {
		t.Errorf("Expected a1, t1, got %s, %s", res[2].SourceProfileID, res[2].TargetProfileID)
	}
}

func TestMergeProfile(t *testing.T) {
	log.Printf("TestMergeProfile: initialize test resources")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	domain := customerprofiles.Domain{Name: "my_test_domain"}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}

	pk := "domain_sourceProfileId"
	sk := "match_targetProfileId"
	indexName := "matchesByConfidenceScore"
	indePkName := "runId"
	indexSkName := "scoreTargetId"
	tableName := "merge_profile_test_table_" + time.Now().Format("20060102150405")
	matchDB := db.Init(tableName, pk, sk, "", "")
	err := matchDB.CreateTableWithOptions(tableName, pk, sk, db.TableOptions{
		GSIs: []db.GSI{
			{
				IndexName: indexName,
				PkName:    indePkName,
				PkType:    "S",
				SkName:    indexSkName,
				SkType:    "S",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error initializing match db: %s", err)
	}
	t.Cleanup(func() { matchDB.DeleteTable(tableName) })
	matchDB.WaitForTableCreation()
	configTableName := "create-delete-test-" + core.GenerateUniqueId()
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Error initializing config db: %s", err)
	}
	t.Cleanup(func() { configDbClient.DeleteTable(configTableName) })
	configDbClient.WaitForTableCreation()

	var accp = customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings)
	asyncLambdaConfig := lambda.InitMock("ucpAsync")

	reg := registry.Registry{Accp: accp, Tx: &tx, MatchDB: &matchDB, ConfigDB: &configDbClient, AsyncLambda: asyncLambdaConfig}
	reg.AddEnv("ACCP_DOMAIN_NAME", domain.Name)
	uc := NewMergeProfile()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{
		Body: `{
			"mergeRq": [{
				"source": "abcde",
				"target": "fghig"
			}]
		}`,
	}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "MergeProfile", err0)
	}
	uc.reg.SetAppAccessPermission(constant.MergeProfilePermission)

	err0 = uc.ValidateRequest(wrapper)
	if err0 == nil {
		t.Errorf("[%s] request should be denied if user does not have the right permissin group", "MergeProfile")
	}
	uc.reg.DataAccessPermission = "*/*"
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] request should be authorized fopr admin user but got error %v", "MergeProfile", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "MergeProfile", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "MergeProfile", err2)
	}
	tx.Debug("Api Gateway response", apiRes)
}

func TestMarkFalsePositiveProfile(t *testing.T) {
	log.Printf("TestMergeProfile: initialize test resources")
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	domain := customerprofiles.Domain{Name: "my_test_domain"}
	domains := []customerprofiles.Domain{domain}
	profile := profilemodel.Profile{}
	profiles := []profilemodel.Profile{}
	mappings := []customerprofiles.ObjectMapping{}
	pk := "domain_sourceProfileId"
	sk := "match_targetProfileId"
	indexName := "matchesByConfidenceScore"
	indePkName := "runId"
	indexSkName := "scoreTargetId"
	tableName := "mark_false_positive_test_table_" + time.Now().Format("20060102150405")
	matchDB := db.Init(tableName, pk, sk, "", "")
	err := matchDB.CreateTableWithOptions(tableName, pk, sk, db.TableOptions{
		GSIs: []db.GSI{
			{
				IndexName: indexName,
				PkName:    indePkName,
				PkType:    "S",
				SkName:    indexSkName,
				SkType:    "S",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error initializing db: %s", err)
	}
	t.Cleanup(func() { matchDB.DeleteTable(tableName) })
	matchDB.WaitForTableCreation()
	configTableName := "create-delete-test-" + core.GenerateUniqueId()
	configDbClient, err := db.InitWithNewTable(configTableName, "item_id", "item_type", "", "")
	if err != nil {
		t.Fatalf("Error initializing config db: %s", err)
	}
	t.Cleanup(func() { configDbClient.DeleteTable(configTableName) })
	configDbClient.WaitForTableCreation()

	var accp = customerprofiles.InitMock(&domain, &domains, &profile, &profiles, &mappings)
	asyncLambdaConfig := lambda.InitMock("ucpAsync")
	reg := registry.Registry{Accp: accp, Tx: &tx, MatchDB: &matchDB, ConfigDB: &configDbClient, AsyncLambda: asyncLambdaConfig}
	reg.AddEnv("ACCP_DOMAIN_NAME", domain.Name)
	uc := NewMergeProfile()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)

	sourceID := "abcdesdlfheowinfqsdf3"
	targetID := "083h9298dgwiegfowieh"

	mergeRq := []commonModel.MergeRq{{SourceProfileID: sourceID, TargetProfileID: targetID}}
	eventId := uuid.NewString()

	payload := uc.buildAsyncPayload(domain.Name, eventId, mergeRq)
	if payload.EventID == "" {
		t.Errorf("EventID should not be empty")
	}
	if payload.TransactionID == "" {
		t.Errorf("TransactionID should not be empty")
	}
	if payload.Usecase != "mergeProfiles" {
		t.Errorf("Usecase should be mergeProfiles")
	}
	if payload.Body.(asyncModel.MergeProfilesBody).Domain != domain.Name {
		t.Errorf("Domain should be %s but got %s", domain.Name, payload.Body.(asyncModel.MergeProfilesBody).Domain)
	}
	if len(payload.Body.(asyncModel.MergeProfilesBody).Rq) != 1 {
		t.Errorf("Rq should be 1 but got %d", len(payload.Body.(asyncModel.MergeProfilesBody).Rq))
	}

	jsonRq, _ := json.Marshal(mergeRq)
	rq := events.APIGatewayProxyRequest{
		Body: `{
			"mergeRq": ` + string(jsonRq) + `
		}`,
	}

	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "MergeProfile", err0)
	}
	log.Printf("Running use case without any match in teh DB")
	_, err = uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] No error should be returned if profile does not exists to allow merging ad-hoc profiles. Got error %v", "MergeProfile", err)
	}

	log.Printf("Inserting profile pair for testing")
	mp0 := travSvc.BuildMatchPairRecord(sourceID, targetID, "0999999", domain.Name, false)
	mp2 := travSvc.BuildMatchPairRecord(targetID, sourceID, "0999999", domain.Name, false)
	mp3 := travSvc.BuildMatchPairRecord(targetID, "sources_2", "0999999", domain.Name, false)
	mp4 := travSvc.BuildMatchPairRecord(targetID, "sources_3", "0999999", domain.Name, false)
	err = matchDB.SaveMany([]commonModel.MatchPair{mp0, mp2, mp3, mp4})
	if err != nil {
		t.Errorf("Error inserting profile pair for testing: %s", err)
	}
	log.Printf("re-Running use case with a match in the DB")
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] errror merging profiles: %v", "MergeProfile", err)
	}

	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "MergeProfile", err2)
	}
	tx.Debug("Api Gateway response", apiRes)
	for _, mp := range []commonModel.MatchPair{mp0, mp2, mp3, mp4} {
		mpAfterMerge := commonModel.MatchPair{}
		matchDB.Get(mp.Pk, mp.Sk, &mpAfterMerge)
		if !mpAfterMerge.MergeInProgress {
			t.Errorf("MergeInProgress should be true after the sync part of the mage request. Got %+v", mpAfterMerge)
		}
	}
}
