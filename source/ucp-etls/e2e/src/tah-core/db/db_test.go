package db

import (
	"log"
	"os"
	"testing"
	"time"
)

var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

type DynamoLockable struct {
	ID       string `json:"id"`
	IsLocked bool   `json:"isLocked"`
}

func TestDynamoDB(t *testing.T) {
	log.Printf("INIT FUNCTION TEST")
	log.Printf("*******************")

	cfg := Init("table", "pk", "sk")
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
	cfg, err = InitWithNewTable("tah-core-unit-tests", "id", "")
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
	log.Printf("3. Create test table")
	cfg, err := InitWithNewTableAndTTL("tah-core-unit-tests-ttl", "id", "", "myTestTTL")
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
