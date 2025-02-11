// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package db

import (
	"context"
	"fmt"
	"log"
	"tah/upt/source/tah-core/core"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoLockable struct {
	ID       string `json:"id"`
	IsLocked bool   `json:"isLocked"`
}

type TestObject struct {
	PK       string  `json:"pk"`
	SK       string  `json:"sk"`
	Count    float64 `json:"object_count"`
	BoolKey  bool    `json:"boolKey"`
	Score    string  `json:"score"`
	Constant string  `json:"constant"`
}

// Note: we are not testing the actual consistency here (this would be hard to do). we are just testing that the
// ConsistentGet function correctly retrieved an object just inserted
func TestConsistentRead(t *testing.T) {
	cfg, err := InitWithNewTable("consistent-read", "id", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] error init with new table: %v", t.Name(), err)
	}
	t.Cleanup(func() {
		log.Printf("Cleaning up")
		err = cfg.DeleteTable(cfg.TableName)
		if err != nil {
			t.Errorf("[DeleteTable] Error deleting test table %v", err)
		}
	})
	cfg.WaitForTableCreation()
	val := fmt.Sprintf("item-%v", time.Now())
	item := map[string]string{
		"id": val,
		"sk": "sk",
	}
	_, err = cfg.Save(item)
	if err != nil {
		t.Fatalf("[%s] error saving item %s into DynamoDB Table: %v", t.Name(), val, err)
	}
	readObj := map[string]string{}
	err = cfg.ConsistentGet(val, "sk", &readObj)
	if err != nil {
		t.Fatalf("[%s] error fetching item %s from DynamoDB Table: %v", t.Name(), val, err)
	}
	if readObj["id"] != val || readObj["sk"] != "sk" {
		t.Fatalf("[%s] invalid object %+v retrieved from Dynamo", t.Name(), readObj)
	}
}

/*
	NOTE: we should always prefer using a defined structure so we know how to
	interact with the data.

	However, we don't always know how the data will be structured.
	In that case we can instead use an generic stucture.

	Using defined struct:
	testObject := []TestObject{}
	cfg.Get(testPk, testSk, &testObject)
	to access count => testObject.Count

	Using generic type:
	testObject := []map[string]interface{}{}
	cfg.Get(testPk, testSk, &testObject)
	to access count => testObject["object_count"]

	The data can still be accessed with obj["object_count"], but we
	don't know what type of data to expect and can get a nil pointer
	exception if we try to use an attribute that does not exist (i.e. obj["foo"])
*/

func TestGsi(t *testing.T) {
	t.Parallel()

	tableName := "tah-core-unit-tests-gsi"
	tablePk := "pk"
	tableSk := "sk"
	indexName := "reverse_index"
	indexPk := "constant"
	indexSk := "score"
	log.Printf("TestGsi]  Create test table")
	cfg := Init(tableName, tablePk, tableSk, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)

	err := cfg.CreateTableWithOptions(tableName, tablePk, tableSk, TableOptions{
		GSIs: []GSI{
			{
				IndexName: indexName,
				PkName:    indexPk,
				PkType:    "S",
				SkName:    indexSk,
				SkType:    "S",
			},
		},
	})
	if err != nil {
		t.Errorf("[TestGsi] error creating table with GSI: %v", err)
	}

	cfg.WaitForTableCreation()
	items := []TestObject{
		{
			PK:       "p1",
			SK:       "s1",
			Score:    "score_3",
			BoolKey:  true,
			Constant: "constant",
		},
		{
			PK:       "p2",
			SK:       "s2",
			Score:    "score_2",
			BoolKey:  true,
			Constant: "constant",
		},
		{
			PK:       "p3",
			SK:       "s3",
			Score:    "score_1",
			BoolKey:  false,
			Constant: "constant",
		},
	}

	err = cfg.SaveMany(items)
	if err != nil {
		t.Errorf("[TestGsi] error saving items to table %v", err)
	}

	// Test using an object's structure (when we know what the data should look like)
	testObject := TestObject{}
	err = cfg.Get("p1", "s1", &testObject)
	if err != nil {
		t.Errorf("[TestGsi] error getting item with specific struct in table %v", err)
	}
	log.Printf("Test Object returned with specified struct: %v", testObject)

	log.Printf("Waiting 10 seconds for index to populate")
	time.Sleep(10 * time.Second)
	found := []map[string]interface{}{}
	err = cfg.FindStartingWithAndFilterWithIndex(
		"constant",
		"score_",
		&found,
		QueryOptions{
			Index: DynamoIndex{
				Name: indexName,
				Pk:   indexPk,
				Sk:   indexSk,
			},
			ReverseOrder: true,
			ProjectionExpression: DynamoProjectionExpression{
				Projection: []string{"pk"},
			},
		},
	)
	if err != nil {
		t.Errorf("[TestGsi] error finding items in index %v", err)
	}
	if len(found) != len(items) {
		t.Errorf("[TestGsi] Index should have %v items but found %v", len(items), len(found))
	} else {
		if found[0]["pk"] != "p1" {
			t.Errorf("[TestGsi] p3 should be returned in first position in index %v", found)
		}
		if _, ok := found[0]["sk"]; ok {
			t.Errorf("[TestGsi] p3 should not have a sk in index %v", found)
		}
		if _, ok := found[0]["score"]; ok {
			t.Errorf("[TestGsi] p3 should not have a score in index %v", found)
		}
	}
	log.Printf("Test Objects returned: %+v", found)

	found = []map[string]interface{}{}
	err = cfg.FindStartingWithAndFilterWithIndex("constant", "score_", &found, QueryOptions{Index: DynamoIndex{
		Name: indexName,
		Pk:   indexPk,
		Sk:   indexSk,
	}, ReverseOrder: true,
		Filter: DynamoFilterExpression{
			EqualBool: []DynamoBoolFilterCondition{{Key: "falsePositive", Value: false}},
		}})
	if err != nil {
		t.Errorf("Search with filter error: %v", err)
	}
	if len(found) > 1 {
		t.Errorf("Search with filter should return 1 item and has %v: %+v", len(found), found)
	}

	log.Printf("5. Deleteing table")
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("[TestGsi] Error deleting test table %v", err)
	}
}

func TestCrud(t *testing.T) {
	t.Parallel()

	log.Printf("TestCrud]  Create test table")
	cfg, err := InitWithNewTable("tah-core-unit-tests-crud", "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestCrud] error init with new table: %v", err)
	}
	testPk := "1"
	testSk := "1"
	cfg.WaitForTableCreation()
	items := []TestObject{
		{
			PK:    testPk,
			SK:    testSk,
			Count: 1,
		},
		{
			PK:    "1",
			SK:    "2",
			Count: 2,
		},
		{
			PK:    "1",
			SK:    "3",
			Count: 3,
		},
	}

	err = cfg.SaveMany(items)
	if err != nil {
		t.Errorf("[TestCrud] error saving items to table %v", err)
	}

	// Test using an object's structure (when we know what the data should look like)
	testObject := TestObject{}
	err = cfg.Get(testPk, testSk, &testObject)
	if err != nil {
		t.Errorf("[TestCrud] error getting item with specific struct in table %v", err)
	}
	log.Printf("Test Object returned with specified struct: %v", testObject)

	found := []map[string]interface{}{}
	_, err = cfg.FindAll(&found, nil)
	if err != nil {
		t.Errorf("[TestCrud] error finding items in table %v", err)
	}
	if len(found) != len(items) {
		t.Errorf("[TestCrud] table should have %v items but found %v", len(items), len(found))
	}
	err = cfg.DeleteByKey(testPk, testSk)
	if err != nil {
		t.Errorf("[TestCrud] error deleting item by key from table %v", err)
	}
	_, err = cfg.FindAll(&found, nil)
	if err != nil {
		t.Errorf("[TestCrud] error finding items in table %v", err)
	}
	if len(found) != len(items)-1 {
		t.Errorf("[TestCrud] table should have %v items after deletion but found %v", len(items)-1, len(found))
	}
	err = cfg.DeleteAll()
	if err != nil {
		t.Errorf("[TestCrud] error deleting items from table %v", err)
	}
	found = []map[string]interface{}{}
	_, err = cfg.FindAll(&found, nil)
	if err != nil {
		t.Errorf("[TestCrud] error finding items in table %v", err)
	}
	if len(found) != 0 {
		t.Errorf("[TestCrud] table should have %v items but found %v", 0, len(found))
	}
	log.Printf("5. Deleteing table")
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("[DeleteTable] Error deleting test table %v", err)
	}
}

func TestDynamoDB(t *testing.T) {
	t.Parallel()

	log.Printf("INIT FUNCTION TEST")
	log.Printf("*******************")

	cfg := Init("table", "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	log.Printf("1.Testing Init Function")
	if cfg.TableName != "table" {
		t.Errorf("[Init] Error while initializing the dynamoDB table")
	}

	log.Printf("2.Testing Validate Function")
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("[validate] Validate DynamoDB config error: %v", err)
	}
	log.Printf("3. Create test table")
	cfg, err = InitWithNewTable("tah-core-unit-tests", "id", "", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[InitWithNewTable] error init with new table: %v", err)
	}
	cfg.WaitForTableCreation()
	item := DynamoLockable{
		ID:       "id-" + time.Now().Format("2006-01-02-15:04:05"),
		IsLocked: false,
	}
	log.Printf("4. Create test item")
	_, err = cfg.Save(item)
	if err != nil {
		t.Errorf("[InitWithNewTable] error saving item to table %v", err)
	}
	log.Printf("4. Locking item")
	err = cfg.Lock(item.ID, "", "isLocked")
	if err != nil {
		t.Errorf("[InitWithNewTable] error locking item %v", err)
	}
	log.Printf("4. Trying to lock again")
	err = cfg.Lock(item.ID, "", "isLocked")
	if err == nil {
		t.Errorf("[InitWithNewTable] locking and already locked item should not succeed and return an error")
	}
	log.Printf("5. Unlocking item")
	err = cfg.Unlock(item.ID, "", "isLocked")
	if err != nil {
		t.Errorf("[InitWithNewTable] Unlock returns and error: %v", err)
	}
	log.Printf("5. Trying to unlocking item again")
	err = cfg.Unlock(item.ID, "", "isLocked")
	if err == nil {
		t.Errorf("[InitWithNewTable] Unlock shouold return and erroor after the second call")
	}

	err = cfg.LockWithTTL(item.ID, "", "isLocked", 10)
	if err == nil {
		t.Errorf("[InitWithNewTable] locking an iten with TTL on a non TTL enabled table  shoudl fail")
	}

	log.Printf("5. Deleteing table")
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("[DeleteTable] Error deleting test table %v", err)
	}
}

type DynamoLockableWithTTL struct {
	ID       string `json:"id"`
	IsLocked bool   `json:"isLocked"`
	TTL      bool   `json:"myTestTTL"`
}

func TestDynamoDBLockWithTTL(t *testing.T) {
	t.Parallel()

	log.Printf("3. Create test table")
	cfg, err := InitWithNewTableAndTTL("tah-core-unit-tests-ttl", "id", "", "myTestTTL", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[InitWithNewTable] error init with new table: %v", err)
	}
	cfg.WaitForTableCreation()
	item := DynamoLockableWithTTL{
		ID:       "id-" + time.Now().Format("2006-01-02-15:04:05"),
		IsLocked: false,
	}
	log.Printf("4. Locking item")
	err = cfg.LockWithTTL(item.ID, "", "isLocked", 5)
	if err != nil {
		t.Errorf("[InitWithNewTable] error locking item %v", err)
	}
	log.Printf("4. Trying to lock again")
	err = cfg.LockWithTTL(item.ID, "", "isLocked", 5)
	if err == nil {
		t.Errorf("[InitWithNewTable] locking an already locked item should not succeed and return an error")
	}
	log.Printf("5. Waiting 6 second")
	time.Sleep(6 * time.Second)
	log.Printf("5.Locking item with expired ttls")
	err = cfg.LockWithTTL(item.ID, "", "isLocked", 5)
	if err != nil {
		t.Errorf("[InitWithNewTable] Item TTL shoudl have expired and alloow lock")
	}
	log.Printf("5. Unlocking item")
	err = cfg.Unlock(item.ID, "", "isLocked")
	if err != nil {
		t.Errorf("[InitWithNewTable] Unlock returns and error: %v", err)
	}
	log.Printf("5. Trying to unlocking item again")
	err = cfg.Unlock(item.ID, "", "isLocked")
	if err == nil {
		t.Errorf("[InitWithNewTable] Unlock shouold return and erroor after the second call")
	}
	log.Printf("5. Deleteing table")
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("[DeleteTable] Error deleting test table %v", err)
	}
}

func TestReplaceTable(t *testing.T) {
	t.Parallel()

	tn := "replace-table-test"
	cfg, err := InitWithNewTable(tn, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Errorf("[TestReplaceTable] error init with new table: %v", err)
	}
	err = cfg.WaitForTableCreation()
	if err != nil {
		t.Errorf("[TestReplaceTable] error waiting for table creation: %v", err)
	}
	getTableInput := &dynamodb.DescribeTableInput{
		TableName: &tn,
	}
	oldTable, err := cfg.DbService.DescribeTable(context.TODO(), getTableInput)
	if err != nil {
		t.Errorf("[TestReplaceTable] error describing old table: %v", err)
	}
	err = cfg.ReplaceTable()
	if err != nil {
		t.Errorf("[TestReplaceTable] error replacing table: %v", err)
	}
	newTable, err := cfg.DbService.DescribeTable(context.TODO(), getTableInput)
	if err != nil {
		t.Errorf("[TestReplaceTable] error describing new table: %v", err)
	}
	if *oldTable.Table.TableName != *newTable.Table.TableName ||
		*oldTable.Table.TableArn != *newTable.Table.TableArn ||
		len(oldTable.Table.KeySchema) != len(newTable.Table.KeySchema) ||
		len(oldTable.Table.AttributeDefinitions) != len(newTable.Table.AttributeDefinitions) ||
		oldTable.Table.BillingModeSummary.BillingMode != newTable.Table.BillingModeSummary.BillingMode {
		t.Errorf("[TestReplaceTable] unexpected table definition after replacing table")
	}
	err = cfg.DeleteTable(tn)
	if err != nil {
		t.Errorf("[TestReplaceTable] error deleting table: %v", err)
	}
}

func TestFindByPk(t *testing.T) {
	t.Parallel()

	testName := "TestFindByPk"
	testPostFix := core.GenerateUniqueId()

	//	Create Dynamo Table
	cfg, err := InitWithNewTable(testName+testPostFix, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error init with new table: %v", testName, err)
	}
	cfg.WaitForTableCreation()
	t.Cleanup(func() { cfg.DeleteTable(cfg.TableName) })

	err = cfg.SaveMany([]map[string]string{
		{
			"pk": "english",
			"sk": "hello",
		},
		{
			"pk": "english",
			"sk": "goodbye",
		},
		{
			"pk": "french",
			"sk": "bonjour",
		},
		{
			"pk": "spanish",
			"sk": "hola",
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error saving items: %v", testName, err)
	}

	var results []TestObject
	err = cfg.FindByPk("english", &results)
	if err != nil {
		t.Fatalf("[%s] Error finding items: %v", testName, err)
	}
	if len(results) != 2 {
		t.Fatalf("[%s] Expected 2 results, got %d", testName, len(results))
	}
	for _, item := range results {
		if item.SK != "hello" && item.SK != "goodbye" {
			t.Fatalf("[%s] Unexpected item: %v", testName, item)
		}
	}
}

func TestUpdateTableElements(t *testing.T) {
	t.Parallel()

	testName := "TestFindByPk"
	testPostFix := core.GenerateUniqueId()

	//	Create Dynamo Table
	cfg, err := InitWithNewTable(testName+testPostFix, "pk", "sk", core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	if err != nil {
		t.Fatalf("[%s] Error init with new table: %v", testName, err)
	}
	cfg.WaitForTableCreation()
	t.Cleanup(func() { cfg.DeleteTable(cfg.TableName) })

	err = cfg.SaveMany([]map[string]string{
		{
			"pk":       "english",
			"sk":       "hello",
			"Score":    "American",
			"Constant": "hi",
		},
		{
			"pk": "english",
			"sk": "goodbye",
		},
		{
			"pk": "french",
			"sk": "bonjour",
		},
		{
			"pk": "spanish",
			"sk": "hola",
		},
	})
	if err != nil {
		t.Fatalf("[%s] Error saving items: %v", testName, err)
	}

	var results []TestObject
	err = cfg.FindByPk("english", &results)
	if err != nil {
		t.Fatalf("[%s] Error finding items: %v", testName, err)
	}
	if len(results) != 2 {
		t.Fatalf("[%s] Expected 2 results, got %d", testName, len(results))
	}
	for _, item := range results {
		if item.SK != "hello" && item.SK != "goodbye" {
			t.Fatalf("[%s] Unexpected item: %v", testName, item)
		}
	}

	updateFields := map[string]interface{}{
		"Score":        "Australian",
		"Constant":     "G'day mate",
		"status":       "active",
		"exists":       "yes",
		"describe":     "this is a test",
		"data":         "reserved_words",
		"object_count": 1000000,
	}

	err = cfg.UpdateItems("english", "hello", updateFields)
	if err != nil {
		t.Fatalf("[%s] Error updating items: %v", testName, err)
	}

	var newResults []TestObject
	err = cfg.FindStartingWith("english", "hello", &newResults)
	if err != nil {
		t.Fatalf("[%s] Error finding items: %v", testName, err)
	}

	if (newResults[0].Score != "Australian") || (newResults[0].Constant != "G'day mate") || (newResults[0].Count != 1000000) {
		t.Fatalf("[%s] Unexpected item, update failed: %v", testName, newResults[0])
	}

}
