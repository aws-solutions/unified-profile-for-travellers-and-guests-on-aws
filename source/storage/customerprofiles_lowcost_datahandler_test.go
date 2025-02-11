// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	"testing"
	"time"

	"github.com/google/uuid"
)

// this test validate the profile construction from interactions with respect to object type prioritiesvar MERGE_SERCH_RECORD_TESTS = []PutProfileObjectTests{
var now = time.Now()
var utcNow = now.UTC()
var TIME_LAYOUT = "Mon Jan 2 15:04:05 MST 2006"
var BUILD_PROFILE_FROM_INTERACTIONS_TEST = []PutProfileObjectTests{
	{
		ConnectID: "abcd",
		Name:      "single object",
		ProfileObjects: []map[string]interface{}{
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"timestamp":         now,
				"first_name":        "joe",
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
		},
		ExpectedMaster: map[string]string{
			"FirstName":          "joe",
			"FirstName_ts":       now.Format(TIME_LAYOUT),
			"FirstName_obj_type": OBJECT_TYPE_1,
		},
	},
	{
		ConnectID: "efgf",
		Name:      "more recent object",
		ProfileObjects: []map[string]interface{}{
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "joe",
				"timestamp":         now,
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "john",
				"timestamp":         now.AddDate(0, 0, 1),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jules",
				"timestamp":         now.AddDate(0, 0, 2),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jeff",
				"timestamp":         now.AddDate(0, 0, 3),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
		},
		ExpectedMaster: map[string]string{
			"FirstName":          "jeff",
			"FirstName_ts":       now.AddDate(0, 0, 3).Format(TIME_LAYOUT),
			"FirstName_obj_type": OBJECT_TYPE_1,
		},
	},
	{
		ConnectID: "ijkl",
		Name:      "object with higher priority",
		ProfileObjects: []map[string]interface{}{
			{
				"_object_type_name": OBJECT_TYPE_2,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "johnny",
				"timestamp":         now,
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jack",
				"timestamp":         now.AddDate(0, 0, 1),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jenny",
				"timestamp":         now.AddDate(0, 0, 2),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jessy",
				"timestamp":         now.AddDate(0, 0, 3),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
		},
		ExpectedMaster: map[string]string{
			"FirstName":          "johnny",
			"FirstName_ts":       now.Format(TIME_LAYOUT),
			"FirstName_obj_type": OBJECT_TYPE_2,
		},
	},
	{
		ConnectID: "mnop",
		Name:      "object with higher priority, object with lower prio, reordering",
		ProfileObjects: []map[string]interface{}{
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jack",
				"timestamp":         now.AddDate(0, 0, 3),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jenny",
				"timestamp":         now.AddDate(0, 0, 2),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_1,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "jessy",
				"timestamp":         now.AddDate(0, 0, 1),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_2,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "johnny",
				"timestamp":         now,
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
			{
				"_object_type_name": OBJECT_TYPE_3,
				"profile_id":        "1",
				"traveler_id":       "1",
				"accp_object_id":    "XX",
				"segment_id":        "XXXXXXXXXX",
				"from":              "pid1_from_1",
				"first_name":        "james",
				"timestamp":         now.AddDate(0, 0, 4),
				"last_updated":      utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
		},
		ExpectedMaster: map[string]string{
			"FirstName":          "johnny",
			"FirstName_ts":       now.Format(TIME_LAYOUT),
			"FirstName_obj_type": OBJECT_TYPE_2,
		},
	},
}

func TestDataHandlerSplitProfileMap(t *testing.T) {
	t.Parallel()
	toSplit := map[string]string{
		"FirstName": "joe",
		"LastName":  "doe",
		"Dob":       "2028-09-20",
		"Phone":     "1234567890",
		"Email":     "joe@example.com",
		"Address":   "123 Main St",
		"City":      "Anytown",
		"State":     "CA",
		"Zip":       "12345",
		"Country":   "USA",
	}
	for i := 0; i < len(toSplit); i++ {
		log.Printf("split map by %d", i)
		map1, map2 := splitProfileMap(toSplit, i)
		if len(map1) != i {
			t.Fatalf("Expected map1 to have %d elements, but got %d", i, len(map1))
		}
		if len(map2) != len(toSplit)-i {
			t.Fatalf("Expected map2 to have %d elements, but got %d", len(toSplit)-i, len(map2))
		}
		for k, v := range toSplit {
			if v != map1[k] && v != map2[k] {
				t.Fatalf("Expected key/Val %s=>%s to be in either map1 (%+v) or map2 (%+v)", k, v, map1, map2)
			}
		}
	}

}

func TestDataHandlerBuildProfileFromInteractions(t *testing.T) {
	t.Parallel()
	domain := randomDomain("merge_test")
	priority := []string{OBJECT_TYPE_2, OBJECT_TYPE_1}
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{ObjectTypePriority: priority})
	if err != nil {
		t.Fatalf("[%s][%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v ", t.Name(), err)
		}
	})
	mappings := []ObjectMapping{}
	for _, m := range OBJECT_MAPPINGS {
		mappings = append(mappings, m)
	}
	for _, test := range BUILD_PROFILE_FROM_INTERACTIONS_TEST {
		log.Printf("Testing objects: %+v", test.ProfileObjects)
		profile, lastUpdatedObjectTypes, lastUpdateTimestamps := dh.buildProfileFromInteractions(
			test.ConnectID,
			test.ProfileObjects,
			mappings,
			priority,
		)
		if profile.FirstName != test.ExpectedMaster["FirstName"] {
			t.Fatalf("[%s] Expected %s but got %s", t.Name(), test.ExpectedMaster["FirstName"], profile.FirstName)
		}
		if lastUpdatedObjectTypes["FirstName"] != test.ExpectedMaster["FirstName_obj_type"] {
			t.Fatalf(
				"[TestBuildProfileFromInteractions] Expected %s but got %s",
				test.ExpectedMaster["FirstName_obj_type"],
				lastUpdatedObjectTypes["FirstName"],
			)
		}
		if test.ExpectedMaster["FirstName_ts"] != "" {
			expectedTime, err := time.Parse(TIME_LAYOUT, test.ExpectedMaster["FirstName_ts"])
			if err != nil {
				t.Fatalf(
					"[TestBuildProfileFromInteractions] invalid expected update timestamp format in test case: %v (%v)",
					test.ExpectedMaster["FirstName_ts"],
					err,
				)

			}
			if lastUpdateTimestamps["FirstName"].Equal(expectedTime) {
				t.Fatalf(
					"[TestBuildProfileFromInteractions] Expected %s but got %s",
					test.ExpectedMaster["FirstName_ts"],
					lastUpdateTimestamps["FirstName"],
				)
			}
		}

	}

}

var MERGE_SERCH_RECORD_TESTS = []PutProfileObjectTests{
	{
		ID:             "1",
		Name:           "merging 2 profiles same type",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "1",
			"traveler_id":    "1",
			"accp_object_id": "11",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ObjectTypeName2: OBJECT_TYPE_1,
		ObjectData2: map[string]string{
			"profile_id":     "2",
			"traveler_id":    "2",
			"accp_object_id": "21",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "jack",
			"last_updated":   utcNow.UTC().Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "jack",
		},
	},
	{
		ID:             "2",
		Name:           "merging 2 profiles ",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "3",
			"traveler_id":    "3",
			"accp_object_id": "31",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "jeff",
			"last_updated":   utcNow.UTC().Format(INGEST_TIMESTAMP_FORMAT),
		},
		ObjectTypeName2: OBJECT_TYPE_2,
		ObjectData2: map[string]string{
			"profile_id":     "4",
			"traveler_id":    "4",
			"accp_object_id": "41",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "john",
			"last_updated":   utcNow.UTC().Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "jeff",
		},
	},
	{
		ID:             "3",
		Name:           "merging 2 profiles ",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "5",
			"traveler_id":    "5",
			"accp_object_id": "51",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "marc",
			"last_updated":   utcNow.UTC().Format(INGEST_TIMESTAMP_FORMAT),
		},
		ObjectTypeName2: OBJECT_TYPE_1,
		ObjectData2: map[string]string{
			"profile_id":     "6",
			"traveler_id":    "6",
			"accp_object_id": "61",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"last_updated":   time.Now().AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
			"first_name":     "henry",
		},
		ExpectedMaster: map[string]string{
			"FirstName": "marc",
		},
	},
}

// this test validates the logic of merging profile search records by preserving object type priority
func TestDataHandlerMergeSearchRecords(t *testing.T) {
	t.Parallel()
	domain := randomDomain("merge_test")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{ObjectTypePriority: []string{OBJECT_TYPE_2, OBJECT_TYPE_1}})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})
	for i, test := range MERGE_SERCH_RECORD_TESTS {
		profMap1, err := objectMapToProfileMap(test.ObjectData, OBJECT_MAPPINGS[test.ObjectTypeName])
		if err != nil {
			t.Fatalf("[%s] [test-%d] error creating profile level attribute map: %v", t.Name(), i, err)
		}
		profMap2, err := objectMapToProfileMap(test.ObjectData2, OBJECT_MAPPINGS[test.ObjectTypeName2])
		if err != nil {
			t.Fatalf("[%s] [test-%d] error creating profile level attribute map: %v", t.Name(), i, err)
		}
		objStr1, _ := json.Marshal(test.ObjectData)
		objStr2, _ := json.Marshal(test.ObjectData2)

		cid1, _, err := dh.InsertProfileObject(domain, test.ObjectTypeName, "accp_object_id", test.ObjectData, profMap1, string(objStr1))
		if err != nil {
			t.Fatalf("[%s] [test-%d] error inserting object 1: %v", t.Name(), i, err)
		}
		cid2, _, err := dh.InsertProfileObject(domain, test.ObjectTypeName2, "accp_object_id", test.ObjectData2, profMap2, string(objStr2))
		if err != nil {
			t.Fatalf("[%s] [test-%d] error inserting object 2: %v", t.Name(), i, err)
		}
		searchTableRes, err := dh.ShowProfileSearchTable(domain, 50)
		if err != nil {
			t.Fatalf("[%s] [test-%d] error showing search table: %v", t.Name(), i, err)
		}
		t.Logf("[%s][test-%d] search table:", t.Name(), i)
		for _, row := range searchTableRes {
			for key, val := range row {
				if val != nil {
					t.Logf("[%s] %s => %v", t.Name(), key, val)
				}
			}
		}

		mappings := []ObjectMapping{}
		for _, m := range OBJECT_MAPPINGS {
			mappings = append(mappings, m)
		}
		err = dh.MergeProfiles(domain, cid1, cid2, OBJECT_TYPE_NAMES, ProfileMergeContext{}, mappings, "test_id")
		if err != nil {
			t.Fatalf("[%s] [test-%d] error merging profiles: %v", t.Name(), i, err)
		}
		err = checkProfileLevelData(t, dh, domain, test)
		if err != nil {
			t.Fatalf("[%s] [test-%d] error checking profile level data: %v", t.Name(), i, err)
		}
	}

}

// test concurrent insert of 100 record in master table and ensure only one connect ID is generated
func TestConsistentConnectIdGeneration(t *testing.T) {
	t.Parallel()
	domain := randomDomain("consistency_test")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})
	var wg sync.WaitGroup
	var cidMutex sync.Mutex
	cids := []string{}
	errors := []error{}
	nInserts := 50
	wg.Add(nInserts)
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	for i := 0; i < nInserts; i++ {
		go func(index int) {
			defer wg.Done()
			connectID, _, err := dh.InsertProfileInMasterTable(tx, domain, "profileID")
			if err != nil {
				log.Printf("Error inserting profile id in master table: %v", err)
				errors = append(errors, err)
				t.Errorf("[%s] Error inserting profileID in master table: %+v", t.Name(), err)
			} else {
				cidMutex.Lock()
				defer cidMutex.Unlock()
				cids = append(cids, connectID)
			}
		}(i)
	}
	wg.Wait()
	res, err := dh.ShowMasterTable(domain, 100)
	if err != nil {
		t.Fatalf("[%s] Error showing master table: %+v", t.Name(), err)
	}
	for _, row := range res {
		cid := row["connect_id"].(string)
		originalCid := row["original_connect_id"].(string)
		if cid != originalCid {
			t.Fatalf("[%s] Inconsistency between originalCid: %s, cid: %s", t.Name(), originalCid, cid)
		}
	}
	log.Printf("cids: %v", cids)
	if len(cids) == 0 {
		t.Fatalf("[%s] Error inserting object no cid created", t.Name())
	}
	if len(errors) > 0 {
		t.Fatalf("[%s] error inserting some records: %+v", t.Name(), errors)
	}
	if len(cids) != nInserts {
		t.Fatalf("[%s] inconsistent number of cid: should have %d and not %d", t.Name(), nInserts, len(cids))
	}

	for _, cid := range cids {
		if cid != cids[0] {
			t.Errorf("[%s] Inconsistency detected: cids: %+v", t.Name(), cids)
		}
	}
	dh.AurSvc.PrintPoolStats()

}

func TestValidatePutProfileObjectInputs(t *testing.T) {
	h := DataHandler{}

	err := h.validatePutProfileObjectInputs("", "", "", map[string]string{}, map[string]string{}, "")
	if err == nil || err.Error() != "no profile ID provided in the request" {
		t.Fatalf("Expected error 'no profile ID provided in the request' but got %v", err)
	}
	err = h.validatePutProfileObjectInputs("profile_id", "", "", map[string]string{}, map[string]string{}, "")
	if err == nil || err.Error() != "no object type name provided in the request" {
		t.Fatalf("Expected error 'no object type name provided in the request' but got %v", err)
	}
	err = h.validatePutProfileObjectInputs("profile_id", "objectTypeName", "", map[string]string{}, map[string]string{}, "")
	if err == nil || err.Error() != "no unique object key provided in the request" {
		t.Fatalf("Expected error 'no unique object key provided in the request' but got %v", err)
	}
	// no object data
	err = h.validatePutProfileObjectInputs("profile_id", "objectTypeName", "uniqueObjectKey", map[string]string{}, map[string]string{}, "")
	if err == nil || err.Error() != "no object data provided in the request" {
		t.Fatalf("Expected error 'no object data provided in the request' but got %v", err)
	}
	// no profile data
	err = h.validatePutProfileObjectInputs("profile_id", "objectTypeName", "uniqueObjectKey", map[string]string{"key": "val"}, map[string]string{}, "")
	if err == nil || err.Error() != "no profile data provided in the request" {
		t.Fatalf("Expected error 'no profile data provided in the request' but got %v", err)
	}
	// too many fields
	objMap := map[string]string{}
	for i := 0; i < MAX_FIELDS_IN_OBJECT_TYPE+1; i++ {
		objMap[strconv.Itoa(i)] = "val"
	}
	// no object data
	err = h.validatePutProfileObjectInputs("profile_id", "objectTypeName", "uniqueObjectKey", objMap, map[string]string{"key": "val"}, "")
	// no profile data
	if err == nil || err.Error() != fmt.Sprintf("too many fields in object type. Maximum is %d", MAX_FIELDS_IN_OBJECT_TYPE) {
		t.Fatalf("Expected error 'too many fields in object type. Maximum is %d' but got %v", MAX_FIELDS_IN_OBJECT_TYPE, err)
	}
}

// this test ingest N objects concurrently for the same traveler and validate that we only create 1 upt_id
func TestDataHandlerConsistentInsertObject(t *testing.T) {
	t.Parallel()
	domain := randomDomain("consistency_test")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	nObject := 30
	objs := []map[string]string{}
	for i := 0; i < nObject; i++ {
		objs = append(
			objs,
			map[string]string{
				"profile_id":     "pid1",
				"traveler_id":    "pid1",
				"accp_object_id": "obj_id_" + strconv.Itoa(i),
				"first_name":     "john",
			},
		)
	}
	objStrs := []string{}
	for _, o := range objs {
		objStr, _ := json.Marshal(o)
		objStrs = append(objStrs, string(objStr))
	}
	profMaps := []map[string]string{}
	for _, o := range objs {
		profMap, err := objectMapToProfileMap(o, OBJECT_MAPPINGS[OBJECT_TYPE_1])
		if err != nil {
			t.Fatalf("[%s] error creating profile level attribute map: %v", t.Name(), err)
		}
		profMaps = append(profMaps, profMap)
	}
	cids := []string{}
	errors := []error{}
	res, _ := dh.AurSvc.Query("SHOW TRANSACTION ISOLATION LEVEL;")
	log.Printf("isolation level: %v", res)
	wg.Add(len(objs))
	for i, o := range objs {
		go func(index int, object map[string]string) {
			defer wg.Done()
			profMap := profMaps[index]
			objStr := objStrs[index]
			log.Printf("Inserting object %+v with profile data %v and string: %s", object, profMap, string(objStr))
			cid, _, err := dh.InsertProfileObject(domain, OBJECT_TYPE_1, "accp_object_id", object, profMap, string(objStr))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
				t.Errorf("[%s] Error inserting object: %+v", t.Name(), err)
			} else {
				cids = append(cids, cid)
			}
		}(i, o)
	}
	wg.Wait()
	log.Printf("cids: %v", cids)
	if len(cids) == 0 {
		t.Fatalf("[%s] Error inserting object no cid created", t.Name())
	}
	if len(errors) > 0 {
		t.Fatalf("[%s] error inserting some records: %+v", t.Name(), errors)
	}
	if len(cids) != len(objs) {
		t.Fatalf("[%s] inconsistent number of cid: should have %d and not %d", t.Name(), len(objs), len(cids))
	}

	for _, cid := range cids {
		if cid != cids[0] {
			t.Errorf("[%s] Inconsistency detected: cids: %+v", t.Name(), cids)
		}
	}
	res, err = dh.ShowInteractionTable(domain, OBJECT_TYPE_1, 50)
	if len(res) != len(objs) {
		t.Fatalf("[%s] Same profile id, different objects, number of records as objects: %+v", t.Name(), res)
	}
}

func TestProfilePagination(t *testing.T) {
	t.Parallel()
	domain := randomDomain("pagination")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	nObject := 31
	objs := []map[string]string{}
	now := time.Now()
	for i := 0; i < nObject; i++ {
		objs = append(
			objs,
			map[string]string{
				"profile_id":     "pagination_test_profile",
				"traveler_id":    "pagination_test_profile",
				"accp_object_id": "obj_id_" + strconv.Itoa(i),
				"first_name":     "john",
				"timestamp":      now.AddDate(0, 0, i).Format(TIMESTAMP_FORMAT),
				"last_updated":   now.AddDate(0, 0, i).Format(INGEST_TIMESTAMP_FORMAT),
			},
		)
	}
	objStrs := []string{}
	for _, o := range objs {
		objStr, _ := json.Marshal(o)
		objStrs = append(objStrs, string(objStr))
	}
	profMaps := []map[string]string{}
	for _, o := range objs {
		profMap, err := objectMapToProfileMap(o, OBJECT_MAPPINGS[OBJECT_TYPE_1])
		if err != nil {
			t.Fatalf("[%s] error creating profile level attribute map: %v", t.Name(), err)
		}
		profMaps = append(profMaps, profMap)
	}
	cids := []string{}
	errors := []error{}
	wg.Add(len(objs))
	for i, o := range objs {
		go func(index int, object map[string]string) {
			defer wg.Done()
			profMap := profMaps[index]
			objStr := objStrs[index]
			log.Printf("Inserting object %+v with profile data %v and string: %s", object, profMap, string(objStr))
			cid, _, err := dh.InsertProfileObject(domain, OBJECT_TYPE_1, "accp_object_id", object, profMap, string(objStr))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
				t.Errorf("[%s] Error inserting object: %+v", t.Name(), err)
			} else {
				cids = append(cids, cid)
			}
		}(i, o)
	}
	wg.Wait()
	log.Printf("cids: %v", cids)
	if len(cids) == 0 {
		t.Fatalf("no cid returned after ingesting %d profile objects", nObject)
	}
	conn, err := dh.AurSvc.AcquireConnection(dh.Tx)
	if err != nil {
		t.Fatalf("[%s] Error acquiring connection: %+v", t.Name(), err)
	}
	defer conn.Release()
	mappings := []ObjectMapping{}
	for _, objMapping := range OBJECT_MAPPINGS {
		mappings = append(mappings, objMapping)
	}

	//No pagination
	prof, _, _, err := dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{})
	if len(prof.ProfileObjects) != nObject {
		t.Fatalf("profile retrieved without pagination should have %d objects and not %d", nObject, len(prof.ProfileObjects))
	}
	for _, obj := range prof.ProfileObjects {
		if obj.TotalCount != int64(nObject) {
			t.Fatalf("Object should have total count set to %d and not %d", nObject, obj.TotalCount)
		}
	}

	//first mage (most recents)
	pageSize := 10
	prof, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: pageSize, Page: 0},
	})
	if len(prof.ProfileObjects) != 10 {
		t.Fatalf("profile retrieved wit pagination: 10:0 should have %d objects and not %d", 10, len(prof.ProfileObjects))
	}
	for i, obj := range prof.ProfileObjects {
		lastUpdated, ok := obj.AttributesInterface["timestamp"].(time.Time)
		if !ok {
			t.Fatalf("profile object timestamp attribute should be of type time.Time. have %+v", obj.AttributesInterface["timestamp"])
		}
		expectedDate := now.AddDate(0, 0, nObject-pageSize+i)
		if lastUpdated.Format("2006-01-02T15:04:05") != expectedDate.Format("2006-01-02T15:04:05") {
			t.Fatalf("Paginated object (page 0) should be returned ordered by recency (timestamp: %v should be equal to %v). Full list %+v", lastUpdated, expectedDate, prof.ProfileObjects)
		}
		expectedID := "obj_id_" + strconv.Itoa(nObject-pageSize+i)
		if obj.ID != expectedID {
			t.Fatalf("Paginated object  (page 0) should be returned ordered by recency (id: %v should be equal to %v). Full list %+v", obj.ID, expectedID, prof.ProfileObjects)
		}
	}

	//next Page
	prof, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: pageSize, Page: 1},
	})
	if len(prof.ProfileObjects) != 10 {
		t.Fatalf("profile retrieved wit pagination: 10:1 should have %d objects and not %d", 10, len(prof.ProfileObjects))
	}
	for i, obj := range prof.ProfileObjects {
		lastUpdated, ok := obj.AttributesInterface["timestamp"].(time.Time)
		if !ok {
			t.Fatalf("profile object timestamp attribute should be of type time.Time. have %+v", obj.AttributesInterface["timestamp"])
		}
		expectedDate := now.AddDate(0, 0, nObject-pageSize*2+i)
		if lastUpdated.Format("2006-01-02T15:04:05") != expectedDate.Format("2006-01-02T15:04:05") {
			t.Fatalf("Paginated object (page 1) should be returned ordered by recency (timestamp: %v should be equal to %v). Full list %+v", lastUpdated, expectedDate, prof.ProfileObjects)
		}
		expectedID := "obj_id_" + strconv.Itoa(nObject-pageSize*2+i)
		if obj.ID != expectedID {
			t.Fatalf("Paginated object (page 1) should be returned ordered by recency (id: %v should be equal to %v). Full list %+v", obj.ID, expectedID, prof.ProfileObjects)
		}
	}

	//last Page
	prof, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: pageSize, Page: 3},
	})
	if len(prof.ProfileObjects) != 1 {
		t.Fatalf("profile retrieved wit pagination: 10:1 should have %d objects and not %d", 10, len(prof.ProfileObjects))
	}
	obj := prof.ProfileObjects[0]
	lastUpdated, ok := obj.AttributesInterface["timestamp"].(time.Time)
	if !ok {
		t.Fatalf("profile object timestamp attribute should be of type time.Time. have %+v", obj.AttributesInterface["timestamp"])
	}
	if lastUpdated.Format("2006-01-02T15:04:05") != now.Format("2006-01-02T15:04:05") {
		t.Fatalf("Paginated object (last page) should be return one object only with time stamp %v and not %v. Object: %v", lastUpdated, now.Format("2006-01-02T15:04:05"), prof.ProfileObjects)
	}
	expectedID := "obj_id_0"
	if obj.ID != expectedID {
		t.Fatalf("Paginated object (last page) should be returned ordered by recency (id: %v should be equal to %v). Full list %+v", obj.ID, expectedID, prof.ProfileObjects)
	}

	//out of bound page
	prof, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: pageSize, Page: 100},
	})
	if len(prof.ProfileObjects) != 0 {
		t.Fatalf("out of bound pagination should return empty objects and not %+v", prof.ProfileObjects)
	}

	//negative page
	_, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: 1, Page: -1},
	})
	if err == nil {
		t.Fatalf("negative page should return an error")
	}

	//too large page size
	_, _, _, err = dh.buildProfile(conn, domain, cids[0], map[string]string{
		OBJECT_TYPE_1: "",
	}, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{
		{ObjectType: OBJECT_TYPE_1, PageSize: MAX_INTERACTIONS_PER_PROFILE + 1, Page: -1},
	})
	if err == nil {
		t.Fatalf("page size greater than MAX_INTERACTIONS_PER_PROFILE should return an error")
	}
}

func TestSameObjectIdDifferentProfiles(t *testing.T) {
	t.Parallel()
	domain := randomDomain("same_obj_dif_pro")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	nObject := 30
	objs := []map[string]string{}
	for i := 0; i < nObject-1; i++ {
		objs = append(
			objs,
			map[string]string{
				"profile_id":     "pid1" + strconv.Itoa(i),
				"traveler_id":    "pid1" + strconv.Itoa(i),
				"accp_object_id": "obj_id_standard",
				"first_name":     "john",
				"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
			},
		)
	}
	objs = append(
		objs,
		map[string]string{
			"profile_id":     "pid1" + strconv.Itoa(nObject),
			"traveler_id":    "pid1" + strconv.Itoa(nObject),
			"accp_object_id": "obj_id_standard",
			"first_name":     "Michael",
			"timestamp":      time.Now().AddDate(5, 0, 0).Format(TIMESTAMP_FORMAT),
			"last_updated":   utcNow.AddDate(0, 0, 1).Format(INGEST_TIMESTAMP_FORMAT),
		},
	)
	objStrs := []string{}
	for _, o := range objs {
		objStr, _ := json.Marshal(o)
		objStrs = append(objStrs, string(objStr))
	}
	profMaps := []map[string]string{}
	for _, o := range objs {
		profMap, err := objectMapToProfileMap(o, OBJECT_MAPPINGS[OBJECT_TYPE_1])
		if err != nil {
			t.Fatalf("[%s] error creating profile level attribute map: %v", t.Name(), err)
		}
		profMaps = append(profMaps, profMap)
	}
	cids := []string{}
	errors := []error{}
	res, _ := dh.AurSvc.Query("SHOW TRANSACTION ISOLATION LEVEL;")
	log.Printf("isolation level: %v", res)
	wg.Add(len(objs))
	for i, o := range objs {
		go func(index int, object map[string]string) {
			defer wg.Done()
			profMap := profMaps[index]
			objStr := objStrs[index]
			log.Printf("Inserting object %+v with profile data %v and string: %s", object, profMap, string(objStr))
			cid, _, err := dh.InsertProfileObject(domain, OBJECT_TYPE_1, "accp_object_id", object, profMap, string(objStr))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
				t.Errorf("[%s] Error inserting object: %+v", t.Name(), err)
			} else {
				cids = append(cids, cid)
			}
		}(i, o)
	}
	wg.Wait()
	log.Printf("cids: %v", cids)
	if len(cids) == 0 {
		t.Fatalf("[%s] Error inserting object no cid created", t.Name())
	}
	if len(errors) > 0 {
		t.Fatalf("[%s] error inserting some records: %+v", t.Name(), errors)
	}
	if len(cids) != len(objs) {
		t.Fatalf("[%s] inconsistent number of cid: should have %d and not %d", t.Name(), len(objs), len(cids))
	}

	mappings := []ObjectMapping{}
	for _, m := range OBJECT_MAPPINGS {
		mappings = append(mappings, m)
	}
	for i, cid := range cids {
		if i != 0 {
			if cid == cids[0] {
				t.Fatalf("[%s] Inconsistency detected: cids: %+v", t.Name(), cids)
			}
			err = dh.MergeProfiles(domain, cids[0], cid, OBJECT_TYPE_NAMES, ProfileMergeContext{}, mappings, "test_id")
			if err != nil {
				t.Fatalf("[%s] [test-%d] error merging profiles: %v", t.Name(), i, err)
			}
		}
	}
	res, err = dh.ShowInteractionTable(domain, OBJECT_TYPE_1, 50)
	//do not reconcile on put, reconcile on retrieval
	if len(res) != len(objs) {
		t.Fatalf("[%s] Inconsistency detected: interaction table: %+v", t.Name(), res)
	}

	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	conn, err := dh.AurSvc.AcquireConnection(tx)
	if err != nil {
		t.Fatalf("[%s] Error acquiring connection: %+v", t.Name(), err)
	}
	defer conn.Release()
	objectTypeNameToId := make(map[string]string, len(OBJECT_TYPE_NAMES))
	for _, objectType := range OBJECT_TYPE_NAMES {
		objectTypeNameToId[objectType] = ""
	}
	prof, _, _, err := dh.buildProfile(conn, domain, cids[0], objectTypeNameToId, mappings, map[string]interface{}{}, []string{}, []PaginationOptions{})
	//do not reconcile on put, reconcile on retrieval
	if err != nil {
		t.Fatalf("[%s] Error building profile: %+v", t.Name(), err)
	}
	if prof.FirstName != "Michael" {
		t.Fatalf("[%s] Inconsistency detected: profile first name: %+v, should be Michael", t.Name(), prof.FirstName)
	}
	//only 1 when we retrieve, retrieve reconciliation
	if len(prof.ProfileObjects) != 1 {
		t.Fatalf("[%s] Inconsistency detected: profile objects: %+v, should have only 1", t.Name(), prof.ProfileObjects)
	}
	objectFirstName, ok := prof.ProfileObjects[0].Attributes["first_name"]
	if !ok || objectFirstName != "Michael" {
		t.Fatalf("[%s] Inconsistency detected: profile object first name: %+v, should be Michael", t.Name(), objectFirstName)
	}

}

func TestSameObjectIdSameProfile(t *testing.T) {
	t.Parallel()
	domain := randomDomain("same_obj_dif_pro")
	dh, err := SetupDataPlaneTest(t, domain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%s] [%s] Error setting up data place tests: %+v ", t.Name(), domain, err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(domain, dh)
		if err != nil {
			t.Errorf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
		}
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	nObject := 30
	objs := []map[string]string{}
	for i := 0; i < nObject; i++ {
		objs = append(
			objs,
			map[string]string{
				"profile_id":     "pid1",
				"traveler_id":    "pid1",
				"accp_object_id": "obj_id_standard",
				"first_name":     "john",
			},
		)
	}
	objStrs := []string{}
	for _, o := range objs {
		objStr, _ := json.Marshal(o)
		objStrs = append(objStrs, string(objStr))
	}
	profMaps := []map[string]string{}
	for _, o := range objs {
		profMap, err := objectMapToProfileMap(o, OBJECT_MAPPINGS[OBJECT_TYPE_1])
		if err != nil {
			t.Fatalf("[%s] error creating profile level attribute map: %v", t.Name(), err)
		}
		profMaps = append(profMaps, profMap)
	}
	cids := []string{}
	errors := []error{}
	res, _ := dh.AurSvc.Query("SHOW TRANSACTION ISOLATION LEVEL;")
	log.Printf("isolation level: %v", res)
	wg.Add(len(objs))
	for i, o := range objs {
		go func(index int, object map[string]string) {
			profMap := profMaps[index]
			objStr := objStrs[index]
			log.Printf("Inserting object %+v with profile data %v and string: %s", object, profMap, string(objStr))
			cid, _, err := dh.InsertProfileObject(domain, OBJECT_TYPE_1, "accp_object_id", object, profMap, string(objStr))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, err)
				t.Errorf("[%s] Error inserting object: %+v", t.Name(), err)
			} else {
				cids = append(cids, cid)
			}
			wg.Done()
		}(i, o)

	}
	wg.Wait()
	log.Printf("cids: %v", cids)
	if len(cids) == 0 {
		t.Fatalf("[%s] Error inserting object no cid created", t.Name())
	}
	if len(errors) > 0 {
		t.Fatalf("[%s] error inserting some records: %+v", t.Name(), errors)
	}
	if len(cids) != len(objs) {
		t.Fatalf("[%s] inconsistent number of cid: should have %d and not %d", t.Name(), len(objs), len(cids))
	}

	for _, cid := range cids {
		if cid != cids[0] {
			t.Errorf("[%s] Inconsistency detected: cids: %+v", t.Name(), cids)
		}
	}
	res, err = dh.ShowInteractionTable(domain, OBJECT_TYPE_1, 50)
	//do not reconcile on put, reconcile on retrieval
	if len(res) != 1 {
		t.Fatalf("[%s] same object and id, should have all collapsed: %+v", t.Name(), res)
	}
}

/**
 Test definition struct to test object priority function. use as as follow:
	{
		Name: "Higher priority should updates",
		Ts: "2020-01-01T00:00:00Z",
		LastObj: "object_2",
		IncomingObj: "object_3",
		IncomingTs: "2020-01-01T00:00:00Z",
		Expected: true
	},

*/

type ShouldUpdateFnTest struct {
	Name        string // give a name to teh test that makes it easy to see in teh log what you are testing
	Ts          string // timestamp of the existing record
	LastObj     string // Name of the Last Object type that updated the record
	IncomingObj string // Name of the incoming object type
	IncomingTs  string // timestamp of incoming record
	Expected    bool   // whether the we expect tha the update would go though or not (true mean, incoming object should overwrite)
}

func TestDataHandlerObjectPrioFn(t *testing.T) {
	t.Parallel()
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		t.Fatalf("[%s] Error setting up aurora: %+v", t.Name(), err)
		return
	}
	dh := DataHandler{
		AurSvc:       auroraCfg,
		MetricLogger: cloudwatch.NewMetricLogger("test/namespace"),
	}
	testDomain := randomDomain("prio_fn_test")
	objectTypePriorities := []string{"object_1", "object_2", "object_3"}
	err = dh.CreateObjectTypeMappingFunction(testDomain, objectTypePriorities)
	if err != nil {
		t.Fatalf("[%s] error creating object type mapping function %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dh.DeleteObjectTypeMappingFunction(testDomain)
		if err != nil {
			t.Errorf("[%s] error deleting object type mapping function %v", t.Name(), err)
		}
	})

	expected := map[string]int32{
		"object_1": 1,
		"object_2": 2,
		"object_3": 3,
	}
	for _, objectType := range objectTypePriorities {
		data, err := dh.AurSvc.Query("SELECT " + PRIO_FUNCTION_PREFIX + "_" + testDomain + "('" + objectType + "')")
		if err != nil {
			t.Fatalf("[%s] error running query with function %s: %v", t.Name(), PRIO_FUNCTION_PREFIX, err)
		}
		if len(data) > 0 {
			val, ok := data[0][PRIO_FUNCTION_PREFIX+"_"+testDomain].(int32)
			if !ok {
				t.Fatalf(
					"[TestObjectPrioFn] wrong object type priority returned for object type %s. could not cast %+v into float64 ",
					objectType,
					data[0][PRIO_FUNCTION_PREFIX+"_"+testDomain],
				)
			}
			if val != expected[objectType] {
				t.Fatalf(
					"[%s] wrong object type priority returned for object type %s. Expected %v got %+v",
					t.Name(),
					objectType,
					expected[objectType],
					data[0][PRIO_FUNCTION_PREFIX+"_"+testDomain],
				)
			}
		} else {
			t.Fatalf("[%s] prio function test did not return any result", t.Name())
		}
	}

	err = dh.CreateShouldUpdateFunction(testDomain)
	if err != nil {
		t.Fatalf("[%s] error creating should update function %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = dh.DeleteShouldUpdateFunction(testDomain)
		if err != nil {
			t.Errorf("[%s] error deleting should update function %v", t.Name(), err)
		}
	})

	fnName := shouldUpdateFnName(testDomain)

	var shouldUpdateFnTests = []ShouldUpdateFnTest{
		{
			Name:        "Higher priority should updates",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_2",
			IncomingObj: "object_3",
			IncomingTs:  "2020-01-01T00:00:00Z",
			Expected:    true,
		},
		{
			Name:        "Lower priority should not updates",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_3",
			IncomingObj: "object_2",
			IncomingTs:  "2020-01-01T00:00:00Z",
			Expected:    false,
		},
		{
			Name:        "Equal priority more recent should updates",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_2",
			IncomingObj: "object_2",
			IncomingTs:  "2021-01-01T00:00:00Z",
			Expected:    true,
		},
		{
			Name:        "Equal priority older should not updates",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_2",
			IncomingObj: "object_2",
			IncomingTs:  "2019-01-01T00:00:00Z",
			Expected:    false,
		},
		{
			Name:        "Equal priority current timestamp",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_3",
			IncomingObj: "object_3",
			IncomingTs:  "LOCALTIMESTAMP",
			Expected:    true,
		},
		{
			Name:        "unprioritized incoming object should not update",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "object_1",
			IncomingObj: "unprio_obj",
			IncomingTs:  "LOCALTIMESTAMP",
			Expected:    false,
		},
		{
			Name:        "unprioritized last updated object should update",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "unprio_obj",
			IncomingObj: "object_1",
			IncomingTs:  "LOCALTIMESTAMP",
			Expected:    true,
		},
		{
			Name:        "unprioritized both object more recent should update",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "unprio_obj",
			IncomingObj: "unprio_obj",
			IncomingTs:  "2021-01-01T00:00:00Z",
			Expected:    true,
		},
		{
			Name:        "unprioritized both object older should not update",
			Ts:          "2020-01-01T00:00:00Z",
			LastObj:     "unprio_obj",
			IncomingObj: "unprio_obj",
			IncomingTs:  "2019-01-01T00:00:00Z",
			Expected:    false,
		},
		{
			Name:        "field never updated before",
			Ts:          "NULL",
			LastObj:     "NULL",
			IncomingObj: "unprio_obj",
			IncomingTs:  "2019-01-01T00:00:00Z",
			Expected:    true,
		},
	}

	for _, test := range shouldUpdateFnTests {
		if test.Ts != "LOCALTIMESTAMP" && test.Ts != "NULL" {
			test.Ts = fmt.Sprintf(`TIMESTAMP '%s'`, test.Ts)
		}
		if test.IncomingTs != "LOCALTIMESTAMP" {
			test.IncomingTs = fmt.Sprintf(`TIMESTAMP '%s'`, test.IncomingTs)
		}
		if test.LastObj != "NULL" {
			test.LastObj = fmt.Sprintf(`'%s'`, test.LastObj)
		}
		sql := fmt.Sprintf(`SELECT %s(%s,%s,%s,'%s')`, fnName, test.Ts, test.LastObj, test.IncomingTs, test.IncomingObj)
		data, err := dh.AurSvc.Query(sql)
		if err != nil {
			t.Fatalf("[%s] error running query with function %s: %v", t.Name(), PRIO_FUNCTION_PREFIX, err)
		}
		if len(data) > 0 {
			val, ok := data[0][fnName].(bool)
			if !ok {
				log.Printf("Test '%s' FAILED", test.Name)
				t.Fatalf(
					"[%s] wrong shouldUpdate value returned for test '%s'. could not cast %+v into bool ",
					t.Name(),
					test.Name,
					data[0][fnName],
				)
			}
			if val != test.Expected {
				log.Printf("Test '%s' FAILED", test.Name)
				t.Fatalf(
					"[%s] wrong shouldUpdate value returned for test '%s'. Expected %v got %+v",
					t.Name(),
					test.Name,
					test.Expected,
					data[0][fnName],
				)
			}
		}
		log.Printf("Test '%s' PASSED", test.Name)
	}

}

type DomainTests struct {
	DomainName    string
	ObjectypePrio []string
	Tests         []PutProfileObjectTests
}

type PutProfileObjectTests struct {
	ID             string
	Name           string
	ObjectTypeName string
	ObjectData     map[string]string
	//for builfprofilefrominteraction test
	ConnectID      string
	ProfileObjects []map[string]interface{}
	//used for merge tests
	ObjectTypeName2 string
	ObjectData2     map[string]string
	ExpectedMaster  map[string]string
	ExpectedObject  map[string]string
	ExpectedError   string
	ExpectedCounts  map[string]int64
}

var DOMAIN_NAME_1 = "dh_test_with_prio_" + strings.ToLower(core.GenerateUniqueId())
var DOMAIN_NAME_2 = "dh_test_no_prio_" + strings.ToLower(core.GenerateUniqueId())
var OBJECT_TYPE_1 = "object_type_1"
var OBJECT_TYPE_2 = "object_type_2"
var OBJECT_TYPE_3 = "object_type_3"
var LARGE_OBJECT_TYPE = "large_object_type"

const advancedSearchObjectTypeName = "adv_search_obj_type"

var OBJECT_TYPE_NAMES = []string{OBJECT_TYPE_1, OBJECT_TYPE_2, OBJECT_TYPE_3}
var OBJECT_MAPPINGS = map[string]ObjectMapping{
	OBJECT_TYPE_1: {
		Name: OBJECT_TYPE_1, Fields: []FieldMapping{
			{
				Type:    "STRING",
				Source:  "_source.traveler_id",
				Target:  "traveler_id",
				Indexes: []string{"PROFILE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{"UNIQUE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.segment_id",
				Target:  "segment_id",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.from",
				Target: "_profile.Attributes.from",
			},
			{
				Type:   "STRING",
				Source: "_source.first_name",
				Target: "_profile.FirstName",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.last_updated",
				Target: LAST_UPDATED_FIELD,
			},
		},
	},
	OBJECT_TYPE_2: {
		Name: OBJECT_TYPE_2, Fields: []FieldMapping{
			{
				Type:    "STRING",
				Source:  "_source.traveler_id",
				Target:  "traveler_id",
				Indexes: []string{"PROFILE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{"UNIQUE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.segment_id",
				Target:  "segment_id",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.from",
				Target: "_profile.Attributes.from",
			},
			{
				Type:   "STRING",
				Source: "_source.first_name",
				Target: "_profile.FirstName",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.last_updated",
				Target: LAST_UPDATED_FIELD,
			},
		},
	},
	OBJECT_TYPE_3: {
		Name: OBJECT_TYPE_3, Fields: []FieldMapping{
			{
				Type:    "STRING",
				Source:  "_source.traveler_id",
				Target:  "traveler_id",
				Indexes: []string{"PROFILE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{"UNIQUE"},
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.segment_id",
				Target:  "Attributes.segment_id",
				KeyOnly: true,
			},
			{
				Type:    "STRING",
				Source:  "_source.to",
				Target:  "_profile.Attributes.to",
				KeyOnly: true,
			},
			{
				Type:   "STRING",
				Source: "_source.first_name",
				Target: "_profile.FirstName",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.last_updated",
				Target: LAST_UPDATED_FIELD,
			},
		},
	},
	advancedSearchObjectTypeName: {
		Name: advancedSearchObjectTypeName, Fields: []FieldMapping{
			{
				Type:    MappingTypeString,
				Source:  "_source.traveler_id",
				Target:  "traveler_id",
				Indexes: []string{INDEX_PROFILE},
				KeyOnly: true,
			},
			{
				Type:    MappingTypeString,
				Source:  "_source.accp_object_id",
				Target:  "accp_object_id",
				Indexes: []string{INDEX_UNIQUE},
				KeyOnly: true,
			},
			{
				Type:   MappingTypeString,
				Source: "_source.first_name",
				Target: "_profile.FirstName",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.last_name",
				Target: "_profile.LastName",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.personal_email",
				Target: "_profile.PersonalEmailAddress",
			},
			{
				Type:   MappingTypeString,
				Source: "_source.business_email",
				Target: "_profile.BusinessEmailAddress",
			},
		},
	},
	LARGE_OBJECT_TYPE: createLargeObjectMapping(),
}

var DATA_PLANE_TESTS_WITHOUT_OBJECT_PRIO = []PutProfileObjectTests{
	{
		ID:             "1",
		Name:           "Insert object 1: should be found",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "joe",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           1,
		},
	},
	{
		ID:             "2",
		Name:           "Insert same object with other name (first name update)",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_2",
			"segment_id":     "pid1_seg_2",
			"from":           "pid1_from_2",
			"first_name":     "john",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "john",
		},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           2,
		},
	},
	{
		ID:             "3",
		Name:           "Insert other object with other name (no first name update)",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_3",
			"segment_id":     "pid1_seg_3",
			"from":           "pid1_from_3",
			"first_name":     "jeff",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "jeff",
		},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
		},
	},
	{
		ID:             "4",
		Name:           "Insert object with newer timestamp",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			//we add this to account for the time difference since existing timestamp are aurora-generated possibly in another region
			"last_updated": time.Now().AddDate(0, 0, 1).Format(INGEST_TIMESTAMP_FORMAT),
			"first_name":   "Joe",
		},
		ExpectedMaster: map[string]string{
			"FirstName":       "Joe",
			"Attributes.from": "pid1_from_1",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
		},
	},
	{
		ID:             "5",
		Name:           "Insert object with older timestamp",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"last_updated":   time.Now().AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_10",
			"first_name":     "Georges",
		},
		ExpectedMaster: map[string]string{
			"FirstName":       "Joe",
			"Attributes.from": "pid1_from_1",
		},
		ExpectedObject: map[string]string{},
	},
	{
		ID:             "6",
		Name:           "Insert object with empty name",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "Joe",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
		},
	},
	{
		ID:             "7",
		Name:           "Insert object with to value, which has not been added previously",
		ObjectTypeName: OBJECT_TYPE_3,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_7",
			"segment_id":     "pid1_seg_1",
			"to":             "TO",
			"first_name":     "",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName":     "Joe",
			"Attributes.to": "TO",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
			OBJECT_TYPE_3:           1,
		},
	},
	{
		ID:             "8",
		Name:           "Large Object Insert",
		ObjectTypeName: LARGE_OBJECT_TYPE,
		ObjectData:     createLargeObject("jerry", "louis", "pid9999", "oid1"),
		ExpectedMaster: map[string]string{
			"FirstName": "jerry",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			LARGE_OBJECT_TYPE: 1,
		},
	},
}

func createLargeObject(firstName, lastName, tid, objId string) map[string]string {
	data := map[string]string{}
	data["first_name"] = firstName
	data["last_name"] = lastName
	data["traveler_id"] = tid
	data["profile_id"] = tid
	data["accp_object_id"] = objId
	largeObjectSize := MAX_FIELDS_IN_OBJECT_TYPE - 100
	if largeObjectSize < 0 {
		largeObjectSize = 100
	}
	for i := 0; i < largeObjectSize; i++ {
		data[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	return data
}

func createLargeObjectMapping() ObjectMapping {
	fieldMappings := []FieldMapping{}
	fieldMappings = append(fieldMappings, FieldMapping{
		Type:   "STRING",
		Source: "_source.first_name",
		Target: "_profile.FirstName",
	})
	fieldMappings = append(fieldMappings, FieldMapping{
		Type:   "STRING",
		Source: "_source.last_name",
		Target: "_profile.LastName",
	})
	fieldMappings = append(fieldMappings, FieldMapping{
		Type:    "STRING",
		Source:  "_source.traveler_id",
		Target:  "traveler_id",
		Indexes: []string{"PROFILE"},
		KeyOnly: true,
	})
	fieldMappings = append(fieldMappings, FieldMapping{
		Type:    "STRING",
		Source:  "_source.accp_object_id",
		Target:  "accp_object_id",
		Indexes: []string{"UNIQUE"},
		KeyOnly: true,
	})

	largeObjectSize := MAX_FIELDS_IN_OBJECT_TYPE - 100
	if largeObjectSize < 0 {
		largeObjectSize = 100
	}
	for i := 0; i < largeObjectSize; i++ {
		fieldMappings = append(fieldMappings, FieldMapping{
			Type:   "STRING",
			Source: fmt.Sprintf("_source.field_%d", i),
			Target: fmt.Sprintf("_profile.Attributes.field_%d", i),
		})
	}
	return ObjectMapping{
		Name:   LARGE_OBJECT_TYPE,
		Fields: fieldMappings,
	}
}

var DATA_PLANE_TESTS_WITH_OBJECT_PRIO = []PutProfileObjectTests{
	{
		ID:             "1",
		Name:           "Insert object 1: should be found",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "CDG",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName":       "joe",
			"Attributes.from": "CDG",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 1,
			OBJECT_TYPE_1:           1,
		},
	},
	{
		ID:             "2",
		Name:           "Insert object 2: should be found",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid2",
			"traveler_id":    "pid2",
			"accp_object_id": "obj_id_2",
			"from":           "BOS",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName":       "joe",
			"Attributes.from": "BOS",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
		},
	},
	{
		ID:             "3",
		Name:           "Update object 2: from should be updated",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid2",
			"traveler_id":    "pid2",
			"accp_object_id": "obj_id_2",
			"from":           "ATL",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName":       "joe",
			"Attributes.from": "ATL",
		},
		ExpectedObject: map[string]string{},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
		},
	},
	{
		ID:             "4",
		Name:           "Insert invalid object: no pid",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "",
			"traveler_id":    "pid2",
			"accp_object_id": "obj_id_2",
			"from":           "NYC",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "joe",
		},
		ExpectedObject: map[string]string{},
		ExpectedError:  "no profile ID provided in the request",
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
		},
	},
	{
		ID:             "5",
		Name:           "Insert invalid object: invalid field",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "pid2",
			"traveler_id":    "pid2",
			"accp_object_id": "obj_id_2",
			"invalid_from":   "LAX",
			"first_name":     "joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "joe",
		},
		ExpectedObject: map[string]string{},
		ExpectedError:  "a database-level error occurred while inserting object-level data",
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
		},
	},
	{
		ID:             "6",
		Name:           "Insert object with higher priority (first name update)",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_2",
			"segment_id":     "pid1_seg_2",
			"from":           "pid1_from_2",
			"first_name":     "john",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "john",
		},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
		},
	},
	{
		ID:             "7",
		Name:           "Insert object with lower priority (no first name update)",
		ObjectTypeName: OBJECT_TYPE_3,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_3",
			"segment_id":     "pid1_seg_3",
			"to":             "CDG",
			"first_name":     "jeff",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "john",
		},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           1,
			OBJECT_TYPE_3:           1,
		},
	},
	{
		ID:             "8",
		Name:           "Insert object with no first name (test first name not overwritten)",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_3",
			"segment_id":     "pid1_seg_3",
			"from":           "pid1_from_3",
			"first_name":     "",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "john",
		},
		ExpectedCounts: map[string]int64{
			countTableProfileObject: 2,
			OBJECT_TYPE_1:           2,
			OBJECT_TYPE_2:           2,
			OBJECT_TYPE_3:           1,
		},
	},
	{
		ID:             "9",
		Name:           "Insert object with newer timestamp",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "Joe",
			"last_updated":   utcNow.Format(INGEST_TIMESTAMP_FORMAT),
		},
		ExpectedMaster: map[string]string{
			"FirstName": "Joe",
		},
		ExpectedObject: map[string]string{},
	},
	{
		ID:             "10",
		Name:           "Insert object with older timestamp (first name should not be updated)",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid1",
			"traveler_id":    "pid1",
			"last_updated":   time.Now().AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "Georges",
		},
		ExpectedMaster: map[string]string{
			"FirstName": "Joe",
		},
		ExpectedObject: map[string]string{},
	},
	{
		ID:             "11",
		Name:           "Insert object with older timestamp and higher priority (pid2 first name should be updated)",
		ObjectTypeName: OBJECT_TYPE_2,
		ObjectData: map[string]string{
			"profile_id":     "pid2",
			"traveler_id":    "pid2",
			"last_updated":   time.Now().AddDate(-1, 0, 0).Format(INGEST_TIMESTAMP_FORMAT),
			"accp_object_id": "obj_id_1",
			"segment_id":     "pid2_seg_1",
			"from":           "pid2_from_1",
			"first_name":     "Georges",
		},
		ExpectedMaster: map[string]string{
			"FirstName": "Georges",
		},
		ExpectedObject: map[string]string{},
	},
}

var ALL_DH_TESTS = []DomainTests{
	{
		DomainName:    DOMAIN_NAME_1,
		ObjectypePrio: []string{OBJECT_TYPE_1, OBJECT_TYPE_2},
		Tests:         DATA_PLANE_TESTS_WITH_OBJECT_PRIO,
	},
	{
		DomainName:    DOMAIN_NAME_2,
		ObjectypePrio: []string{},
		Tests:         DATA_PLANE_TESTS_WITHOUT_OBJECT_PRIO,
	},
}

func SetupDataPlaneTest(t *testing.T, domainName string, domainOptions DomainOptions) (DataHandler, error) {
	auroraCfg, err := SetupAurora(t)
	if err != nil {
		return DataHandler{}, err
	}
	dh := DataHandler{
		AurSvc:       auroraCfg,
		MetricLogger: cloudwatch.NewMetricLogger("test/namespace"),
	}
	mappings := []ObjectMapping{}
	for _, objMapping := range OBJECT_MAPPINGS {
		mappings = append(mappings, objMapping)
	}
	err = dh.CreateProfileMasterTable(domainName)
	if err != nil {
		dh.Tx.Error("[TestInsertQuery] error creating profile master table %v", err)
		return DataHandler{}, err
	}
	err = dh.CreateProfileSearchIndexTable(domainName, mappings, domainOptions)
	if err != nil {
		dh.Tx.Error("[TestInsertQuery] error creating profile search table %v", err)
		return DataHandler{}, err
	}
	err = dh.CreateRecordCountTable(domainName)
	if err != nil {
		dh.Tx.Error("[TestInsertQuery] error creating profile count table %v", err)
		return DataHandler{}, err
	}
	err = dh.InitializeCount(domainName, []string{countTableProfileObject})
	if err != nil {
		dh.Tx.Error("[TestInsertQuery] error initializing profile count %v", err)
		return DataHandler{}, err
	}
	err = dh.CreateProfileHistoryTable(domainName)
	if err != nil {
		dh.Tx.Error("[TestInsertQuery] error creating profile history table %v", err)
		return DataHandler{}, err
	}

	for _, objMapping := range OBJECT_MAPPINGS {
		err = dh.CreateInteractionTable(domainName, objMapping)
		if err != nil {
			dh.Tx.Error("[TestInsertQuery] error executing query: %v", err)
			return DataHandler{}, err
		}
		err = dh.InitializeCount(domainName, []string{objMapping.Name})
		if err != nil {
			dh.Tx.Error("[TestInsertQuery] error initializing object type count for object %s: %v", objMapping.Name, err)
			return DataHandler{}, err
		}
		err = dh.CreateObjectHistoryTable(domainName, objMapping.Name)
		if err != nil {
			dh.Tx.Error("[TestInsertQuery] error executing query: %v", err)
			return DataHandler{}, err
		}
	}
	return dh, err
}

func CleanupDataPlaneTest(domainName string, dh DataHandler) error {
	for _, objMapping := range OBJECT_MAPPINGS {
		err := dh.DeleteInteractionTable(domainName, objMapping.Name)
		if err != nil {
			dh.Tx.Error("[CleanupDataPlaneTest] error interaction interaction table: %v", err)
			return err
		}
		err = dh.DeleteObjectHistoryTable(domainName, objMapping.Name)
		if err != nil {
			dh.Tx.Error("[CleanupDataPlaneTest] error deleting interaction history table: %v", err)
			return err
		}
	}
	err := dh.DeleteRecordCountTable(domainName)
	if err != nil {
		dh.Tx.Error("[CleanupDataPlaneTest] error deleting profile count table %v", err)
		return err
	}
	err = dh.DeleteProfileMasterTable(domainName)
	if err != nil {
		dh.Tx.Error("[CleanupDataPlaneTest] error deleting profile master table %v", err)
		return err
	}
	err = dh.DeleteProfileSearchIndexTable(domainName)
	if err != nil {
		dh.Tx.Error("[CleanupDataPlaneTest] error deleting profile search table %v", err)
		return err
	}
	err = dh.DeleteProfileHistoryTable(domainName)
	if err != nil {
		dh.Tx.Error("[CleanupDataPlaneTest] error deleting profile history table %v", err)
		return err
	}
	//closing connection to avoid "could not receive data from client: Connection reset by peer" in aurora log
	return nil
}

func TestDataHandlerInsertObject(t *testing.T) {
	t.Parallel()
	var dh DataHandler
	var err error
	t.Cleanup(func() {
		for _, domTest := range ALL_DH_TESTS {
			log.Printf("Cleanup data for domain %v", domTest.DomainName)
			err = CleanupDataPlaneTest(domTest.DomainName, dh)
			if err != nil {
				t.Fatalf("[%s] Error cleaning up data place tests: %+v", t.Name(), err)
			}
		}
		dh.AurSvc.Close()
	})
	for _, domTest := range ALL_DH_TESTS {
		dh, err = SetupDataPlaneTest(t, domTest.DomainName, DomainOptions{ObjectTypePriority: domTest.ObjectypePrio})
		if err != nil {
			t.Fatalf("[TestInsertObject][%s] Error setting up data place tests: %+v ", domTest.DomainName, err)
		}

		for _, test := range domTest.Tests {
			log.Printf("********************************************")
			log.Printf("*Test %s: '%s'", test.ID, test.Name)
			log.Printf("********************************************")
			obj, _ := json.Marshal(test.ObjectData)
			profMap, err := objectMapToProfileMap(test.ObjectData, OBJECT_MAPPINGS[test.ObjectTypeName])
			if err != nil {
				t.Fatalf(
					"[TestInsertObject][%s][test-%s]  error creating profile level attribute map: %v",
					domTest.DomainName,
					test.ID,
					err,
				)
				break
			}
			connectID, _, err := dh.InsertProfileObject(
				domTest.DomainName,
				test.ObjectTypeName,
				"accp_object_id",
				test.ObjectData,
				profMap,
				string(obj),
			)
			//in case we expect an error from this test case
			if test.ExpectedError != "" {
				if err == nil || err.Error() != test.ExpectedError {
					t.Fatalf(
						"[TestInsertObject][%s][test-%s] invalid error while inserting object. should have '%v' and not '%v'",
						domTest.DomainName,
						test.ID,
						test.ExpectedError,
						err,
					)
				}
			} else {
				//no error is expected for this test case
				if err != nil {
					t.Fatalf("[%s] [%s][test-%s] unexpected error executing query: %v", t.Name(), domTest.DomainName, test.ID, err)
				}
				if connectID == "" {
					t.Fatalf("[%s] [%s][test-%s]  connect id returned by InsertProfileObject should not be empty", t.Name(), domTest.DomainName, test.ID)
				}

				err = checkProfileLevelData(t, dh, domTest.DomainName, test)
				if err != nil {
					t.Fatalf("%v", err)
				}
				err = checkObjectLevelData(t, dh, domTest.DomainName, test)
				if err != nil {
					t.Fatalf("%v", err)
				}
				for k, v := range test.ExpectedCounts {
					counts, err := dh.GetAllObjectCounts(domTest.DomainName)
					if err != nil {
						t.Fatalf("[%s] [%s][test-%s]  error getting profile count: %v", t.Name(), domTest.DomainName, test.ID, err)
					}
					if counts[k] != v {
						t.Fatalf("[%s] [%s][test-%s]  expected count for %s is %d and not %d", t.Name(), domTest.DomainName, test.ID, k, v, counts[k])
					}
				}
				//testing search

				cids, err := dh.FindConnectIDByObjectField(domTest.DomainName, test.ObjectTypeName, "accp_object_id", []string{test.ObjectData["accp_object_id"]})
				if err != nil {
					t.Fatalf("[%s] [%s][test-%s]  error searching for connect id: %v", t.Name(), domTest.DomainName, test.ID, err)
				}
				if len(cids) == 0 {
					t.Fatalf("[%s] [%s][test-%s]  no connect id found for object %s", t.Name(), domTest.DomainName, test.ID, test.ObjectData["accp_object_id"])
				}
			}
		}
		//Testing profile partition:
		partition, err := dh.GetProfilePartition(domTest.DomainName, "00000000-0000-0000-0000-000000000000", "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")
		log.Printf("partition: %+v", partition)
		if err != nil {
			t.Fatalf("[%s] Error getting profile partition: %v", domTest.DomainName, err)
		}
		if len(partition) == 0 {
			t.Fatalf("[%s] Error getting profile partition: no partition returned", domTest.DomainName)
		}
		//getting the full search table to evaluate the number of profile and compare against partition
		searchTableRes, err := dh.ShowProfileSearchTable(domTest.DomainName, 50)
		if err != nil {
			t.Fatalf("[%s] Error getting profile search table: %v", domTest.DomainName, err)
		}
		if len(searchTableRes) != len(partition) {
			t.Fatalf("[%s] partition should have %d records and not %d", domTest.DomainName, len(partition), len(searchTableRes))
		}
		log.Printf("Successfully completed test for %v", domTest.DomainName)
	}
}

func checkProfileLevelData(t *testing.T, dh DataHandler, domain string, test PutProfileObjectTests) error {
	masterTableRes, err := dh.ShowMasterTable(domain, 50)
	if err != nil {
		return fmt.Errorf("[%s] [%s][test-%s] Error selecting master table: %v", t.Name(), domain, test.ID, err)
	}
	searchTableRes, err := dh.ShowProfileSearchTable(domain, 50)
	if err != nil {
		return fmt.Errorf("[%s] [%s][test-%s] Error selecting search table: %v", t.Name(), domain, test.ID, err)
	}
	log.Printf("Search table export:")
	for _, row := range searchTableRes {
		log.Printf("------------------------------------------------------------------------")
		for field, val := range row {
			if val != nil {
				log.Printf("'%v':'%v'", field, val)
			}
		}
	}
	masterFound := false
	searchRecFound := false
	for _, row := range masterTableRes {
		if row["profile_id"] == test.ObjectData["profile_id"] {
			masterFound = true
			connectID := row["connect_id"]
			for _, searchRow := range searchTableRes {
				if searchRow["connect_id"] == connectID {
					searchRecFound = true
					for k, v := range test.ExpectedMaster {
						if searchRow[k] != v {
							return fmt.Errorf(
								"[TestInsertObject][%s][test-%s] Expected value for %s is %s and not %s for profile %s (connect ID %s)",
								domain,
								test.ID,
								k,
								v,
								searchRow[k],
								test.ObjectData["profile_id"],
								connectID,
							)
						}
					}
				}
			}
			if !searchRecFound {
				return fmt.Errorf(
					"[TestInsertObject][%s][test-%s] Connect id %s not found in profile search table",
					domain,
					test.ID,
					connectID,
				)
			}
		}
	}
	if !masterFound {
		return fmt.Errorf(
			"[TestInsertObject][%s][test-%s] Profile %s not found in master table",
			domain,
			test.ObjectData["profile_id"],
			test.ID,
		)
	}
	dh.Tx.Info("[TestInsertObject][%s][test-%s]  Successfully validated profile-level data", domain, test.ID)
	return nil
}

func checkObjectLevelData(t *testing.T, dh DataHandler, domain string, test PutProfileObjectTests) error {
	res, err := dh.ShowInteractionTable(domain, test.ObjectTypeName, 50)
	if err != nil {
		return fmt.Errorf("[%s] [%s][test-%s] Error selecting interaction table: %v", t.Name(), domain, test.ID, err)
	}
	found := false
	for _, row := range res {
		if row["profile_id"] == test.ObjectData["profile_id"] {
			found = true
			for k, v := range test.ExpectedObject {
				if row[k] != v {
					return fmt.Errorf(
						"[TestInsertObject][%s][test-%s] Expected value for %s is %s and not %s",
						domain,
						test.ID,
						k,
						v,
						row[k],
					)
				}
			}
		}
	}
	if !found {
		return fmt.Errorf(
			"[TestInsertObject][%s][test-%s] Object %s not found in interaction table",
			domain,
			test.ID,
			test.ObjectData["accp_object_id"],
		)
	}
	dh.Tx.Info("[TestInsertObject][%s][test-%s]  Successfully validated object-level data", domain, test.ID)
	return nil
}

func TestInteractionTablePartitionPage(t *testing.T) {
	t.Parallel()

	testDomain := randomDomain("int_page_test")
	dh, err := SetupDataPlaneTest(t, testDomain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%v] error setting up data plane test: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(testDomain, dh)
		if err != nil {
			t.Errorf("[%v] error cleaning up data plane test: %v", t.Name(), err)
		}
	})

	connectIds := []uuid.UUID{}
	connectIdsToObjIds := map[uuid.UUID]string{}

	for i := 0; i < 4; i++ {
		obj := map[string]string{
			"profile_id":     uuid.New().String(),
			"traveler_id":    "pid1",
			"accp_object_id": fmt.Sprintf("obj_%d", i),
			"segment_id":     "pid1_seg_1",
			"from":           "CDG",
			"first_name":     "joe",
		}
		objJson, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("[%v] error marshalling object: %v", t.Name(), err)
		}
		profMap, err := objectMapToProfileMap(obj, OBJECT_MAPPINGS[OBJECT_TYPE_1])
		if err != nil {
			t.Fatalf("[%v] error creating profile map: %v", t.Name(), err)
		}

		connectId, _, err := dh.InsertProfileObject(testDomain, OBJECT_TYPE_1, "accp_object_id", obj, profMap, string(objJson))
		if err != nil {
			t.Fatalf("[%v] error inserting object: %v", t.Name(), err)
		}

		connectUuid, err := uuid.Parse(connectId)
		if err != nil {
			t.Fatalf("[%v] error parsing connect id: %v", t.Name(), err)
		}
		connectIds = append(connectIds, connectUuid)

		connectIdsToObjIds[connectUuid] = obj["accp_object_id"]
	}

	slices.SortFunc(connectIds, func(a, b uuid.UUID) int {
		aStr := a.String()
		bStr := b.String()
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		} else {
			return 0
		}
	})

	firstPartition := UuidPartition{LowerBound: connectIds[0], UpperBound: connectIds[1]}
	secondPartition := UuidPartition{LowerBound: connectIds[2], UpperBound: connectIds[3]}

	objs, err := dh.ShowInteractionTablePartitionPage(testDomain, OBJECT_TYPE_1, firstPartition, uuid.Nil, 1)
	if err != nil {
		t.Fatalf("[%v] error showing interaction table partition page: %v", t.Name(), err)
	}
	if len(objs) != 1 {
		t.Fatalf("[%v] expected 1 object, got %v", t.Name(), len(objs))
	}
	if objs[0]["accp_object_id"] != connectIdsToObjIds[connectIds[0]] {
		t.Fatalf("[%v] expected accp_object_id to be %v, got %v", t.Name(), connectIdsToObjIds[connectIds[0]], objs[0]["accp_object_id"])
	}

	objs, err = dh.ShowInteractionTablePartitionPage(testDomain, OBJECT_TYPE_1, firstPartition, connectIds[0], 1)
	if err != nil {
		t.Fatalf("[%v] error showing interaction table partition page: %v", t.Name(), err)
	}
	if len(objs) != 1 {
		t.Fatalf("[%v] expected 1 object, got %v", t.Name(), len(objs))
	}
	if objs[0]["accp_object_id"] != connectIdsToObjIds[connectIds[1]] {
		t.Fatalf("[%v] expected accp_object_id to be %v, got %v", t.Name(), connectIdsToObjIds[connectIds[0]], objs[0]["accp_object_id"])
	}

	objs, err = dh.ShowInteractionTablePartitionPage(testDomain, OBJECT_TYPE_1, firstPartition, connectIds[1], 1)
	if err != nil {
		t.Fatalf("[%v] error showing interaction table partition page: %v", t.Name(), err)
	}
	if len(objs) != 0 {
		t.Fatalf("[%v] expected 0 objects, got %v", t.Name(), len(objs))
	}

	objs, err = dh.ShowInteractionTablePartitionPage(testDomain, OBJECT_TYPE_1, secondPartition, uuid.Nil, 1)
	if err != nil {
		t.Fatalf("[%v] error showing interaction table partition page: %v", t.Name(), err)
	}
	if len(objs) != 1 {
		t.Fatalf("[%v] expected 1 object, got %v", t.Name(), len(objs))
	}
	if objs[0]["accp_object_id"] != connectIdsToObjIds[connectIds[2]] {
		t.Fatalf("[%v] expected accp_object_id to be %v, got %v", t.Name(), connectIdsToObjIds[connectIds[2]], objs[0]["accp_object_id"])
	}

	objs, err = dh.ShowInteractionTablePartitionPage(testDomain, OBJECT_TYPE_1, secondPartition, connectIds[2], 1)
	if err != nil {
		t.Fatalf("[%v] error showing interaction table partition page: %v", t.Name(), err)
	}
	if len(objs) != 1 {
		t.Fatalf("[%v] expected 1 object, got %v", t.Name(), len(objs))
	}
	if objs[0]["accp_object_id"] != connectIdsToObjIds[connectIds[3]] {
		t.Fatalf("[%v] expected accp_object_id to be %v, got %v", t.Name(), connectIdsToObjIds[connectIds[3]], objs[0]["accp_object_id"])
	}
}

var advancedProfileSearchTestProfiles = []map[string]string{
	{
		"profile_id":     uuid.New().String(),
		"traveler_id":    "1",
		"accp_object_id": "obj_1",
		"first_name":     "Bob",
		"last_name":      "John",
		"personal_email": "bobjohn@example.com",
	},
	{
		"profile_id":     uuid.New().String(),
		"traveler_id":    "2",
		"accp_object_id": "obj_2",
		"first_name":     "Bob",
		"last_name":      "John",
		"personal_email": "bobbyjangles@example.com",
		"business_email": "SirBobJohnIII@example.com",
	},
	{
		"profile_id":     uuid.New().String(),
		"traveler_id":    "3",
		"accp_object_id": "obj_3",
		"first_name":     "Terry",
		"last_name":      "John",
		"business_email": "elephantsRcool@example.com",
	},
}

// Search on Columns maps to OBJECT_MAPPINGS[advancedSearchObjectTypeName] above
var advancedProfileSearchTestData = []struct {
	name                string
	searchCriteria      []BaseCriteria
	expectedError       bool
	expectedResultCount int
}{
	{
		name:           "No search criteria",
		searchCriteria: []BaseCriteria{},
		expectedError:  true,
	},
	{
		name: "Search by last name",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
		},
		expectedResultCount: 3,
	},
	{
		name: "Search by last name and first name",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "FirstName", Operator: EQ, Value: "Bob"},
		},
		expectedResultCount: 2,
	},
	{
		name: "Search by last, first, and personal email",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "FirstName", Operator: EQ, Value: "Bob"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "PersonalEmailAddress", Operator: EQ, Value: "bobjohn@example.com"},
		},
		expectedResultCount: 1,
	},
	{
		name: "Search by personal or business email address",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "PersonalEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
			SearchOperator{Operator: OR},
			SearchCriterion{Column: "BusinessEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
		},
		expectedResultCount: 1,
	},
	{
		name: "Search by last name and personal/business email addresses",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{Column: "PersonalEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
					SearchOperator{Operator: OR},
					SearchCriterion{Column: "BusinessEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
				},
			},
		},
		expectedResultCount: 1,
	},
	{
		name: "Search by last name, first name, and personal/business email addresses",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "FirstName", Operator: EQ, Value: "Bob"},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{Column: "PersonalEmailAddress", Operator: EQ, Value: "SirBobJohnIII@example.com"},
					SearchOperator{Operator: OR},
					SearchCriterion{Column: "BusinessEmailAddress", Operator: EQ, Value: "SirBobJohnIII@example.com"},
				},
			},
		},
		expectedResultCount: 1,
	},
	{
		name: "Search by last name, first name, and personal/business email addresses; no results",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "FirstName", Operator: EQ, Value: "Bob"},
			SearchOperator{Operator: AND},
			SearchGroup{
				Criteria: []BaseCriteria{
					SearchCriterion{Column: "PersonalEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
					SearchOperator{Operator: OR},
					SearchCriterion{Column: "BusinessEmailAddress", Operator: EQ, Value: "elephantsRcool@example.com"},
				},
			},
		},
		expectedResultCount: 0,
	},
	{
		name: "Search by last name and 'like' personal email address",
		searchCriteria: []BaseCriteria{
			SearchCriterion{Column: "LastName", Operator: EQ, Value: "John"},
			SearchOperator{Operator: AND},
			SearchCriterion{Column: "PersonalEmailAddress", Operator: LIKE, Value: "%bob%"},
		},
		expectedResultCount: 2,
	},
}

func TestAdvancedProfileSearch(t *testing.T) {
	t.Parallel()

	testDomain := randomDomain("adv_prof_search")
	dh, err := SetupDataPlaneTest(t, testDomain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%v] error setting up data plane test: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(testDomain, dh)
		if err != nil {
			t.Errorf("[%v] error cleaning up data plane test: %v", t.Name(), err)
		}
	})

	//	Ingest sample profiles
	for _, prof := range advancedProfileSearchTestProfiles {
		profMap, err := objectMapToProfileMap(prof, OBJECT_MAPPINGS[advancedSearchObjectTypeName])
		if err != nil {
			t.Fatalf("[%v] error creating profile map: %v", t.Name(), err)
		}

		objJson, err := json.Marshal(prof)
		if err != nil {
			t.Fatalf("[%s] error marshalling object: %v", t.Name(), err)
		}

		_, _, err = dh.InsertProfileObject(testDomain, advancedSearchObjectTypeName, "accp_object_id", prof, profMap, string(objJson))
		if err != nil {
			t.Fatalf("[%v] error inserting object: %v", t.Name(), err)
		}
	}

	for _, testData := range advancedProfileSearchTestData {
		t.Run(testData.name, func(t *testing.T) {
			rows, err := dh.AdvancedSearchProfilesByProfileLevelField(testDomain, testData.searchCriteria)
			if testData.expectedError {
				if err == nil {
					t.Fatalf("[%v] expected error searching profiles, got nil", t.Name())
				}
			} else {
				if err != nil {
					t.Fatalf("[%v] error searching profiles: %v", t.Name(), err)
				}
				if len(rows) != testData.expectedResultCount {
					t.Fatalf("[%v] expected %v results, got %v", t.Name(), testData.expectedResultCount, len(rows))
				}
			}
		})
	}
}

func TestSplitProfileMap(t *testing.T) {
	t.Parallel()

	testMap := make(map[string]string)
	maxMapSize := 5
	for i := 0; i < maxMapSize; i++ {
		testMap[strconv.Itoa(i)] = ""
	}
	testMap["new"] = "random"
	upsertMap, updateMap := splitProfileMap(testMap, maxMapSize)
	if len(upsertMap) != 1 || len(updateMap) != 0 {
		t.Fatalf("[%v] should filter out any empty values", t.Name())
	}
}

func TestInsertApostrophe(t *testing.T) {
	t.Parallel()
	testDomain := randomDomain("insert_apostrophe")
	dh, err := SetupDataPlaneTest(t, testDomain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%v] error setting up data plane test: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(testDomain, dh)
		if err != nil {
			t.Errorf("[%v] error cleaning up data plane test: %v", t.Name(), err)
		}
	})

	test := PutProfileObjectTests{
		ID:             "1",
		Name:           "ingesting apostrophe value",
		ObjectTypeName: OBJECT_TYPE_1,
		ObjectData: map[string]string{
			"profile_id":     "1",
			"traveler_id":    "1",
			"accp_object_id": "11",
			"segment_id":     "pid1_seg_1",
			"from":           "pid1_from_1",
			"first_name":     "O'Connor",
		},
		ExpectedMaster: map[string]string{
			"FirstName": "O'Connor",
		},
	}

	profMap, err := objectMapToProfileMap(test.ObjectData, OBJECT_MAPPINGS[test.ObjectTypeName])
	if err != nil {
		t.Fatalf("[%s] error creating profile level attribute map: %v", t.Name(), err)
	}
	objStr, _ := json.Marshal(test.ObjectData)
	_, _, err = dh.InsertProfileObject(testDomain, test.ObjectTypeName, "accp_object_id", test.ObjectData, profMap, string(objStr))
	if err != nil {
		t.Fatalf("[%s] error inserting object 1: %v", t.Name(), err)
	}

	searchTableRes, err := dh.ShowProfileSearchTable(testDomain, 50)
	if err != nil {
		t.Fatalf("[%s] error showing search table: %v", t.Name(), err)
	}
	t.Logf("[%s] search table:", t.Name())
	for _, row := range searchTableRes {
		for key, val := range row {
			if val != nil {
				t.Logf("[%s] %s => %v", t.Name(), key, val)
			}
		}
	}
	err = checkProfileLevelData(t, dh, testDomain, test)
	if err != nil {
		t.Fatalf("[%s] error checking profile level data: %v", t.Name(), err)
	}
}

func TestDeleteObjectHistoryNoProfile(t *testing.T) {
	t.Parallel()
	testDomain := randomDomain("del_hist_no_prof")
	dh, err := SetupDataPlaneTest(t, testDomain, DomainOptions{})
	if err != nil {
		t.Fatalf("[%v] error setting up data plane test: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		err = CleanupDataPlaneTest(testDomain, dh)
		if err != nil {
			t.Errorf("[%v] error cleaning up data plane test: %v", t.Name(), err)
		}
	})

	conn, err := dh.AurSvc.AcquireConnection(dh.Tx)
	if err != nil {
		t.Fatalf("[%s] Error acquiring connection: %+v", t.Name(), err)
	}
	defer conn.Release()
	err = dh.DeleteProfileObjectsFromObjectHistoryTable(conn, testDomain, "air_booking", "fake_conn_id")
	if err != nil {
		t.Fatalf("[%v] function should fail gracefully with no profiles found: %v", t.Name(), err)
	}
}
