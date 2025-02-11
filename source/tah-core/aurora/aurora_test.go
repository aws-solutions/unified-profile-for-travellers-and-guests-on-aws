// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package aurora

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"tah/upt/source/tah-core/core"
	secret "tah/upt/source/tah-core/secret"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*******
* Aurora
*****/

// This test validate the behavior of the connection pool
// we create multiple go routine (more than max conn to ensure connection reuse)
// for each goroutine we start a transaction, do 2 inserts and commit.
// we expect to see a dedicated PID for each go routing except in the case of connection reuse
// if the connection is reused we expect each transaction to be handled sequentially (and we validate it with timestamps)
// (in other words the same DB process can handle 2 goroutine but never concurrently, always sequentially )
func TestAuroraPosgresConcurency(t *testing.T) {
	postgres, err := SetupAuroraPostgres(t)
	if err != nil {
		log.Fatalf("Error setting up aurora postgres: %+v", err)
	}
	tableName := "test_transactions" + core.GenerateUniqueId()
	//choose a go routing number greater than connection pool max_conn to test connection reuse
	numGoroutines := 60
	_, err = postgres.Query(
		`CREATE TABLE IF NOT EXISTS ` + tableName + ` ( id SERIAL PRIMARY KEY, goroutine_id INT, tx_pid INT, current_pid INT, operation VARCHAR(255) )`,
	)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	t.Cleanup(func() {
		log.Printf("cleaning up")
		_, err := postgres.Query("DROP TABLE " + tableName) // Clean up table after test
		if err != nil {
			log.Fatalf("Error dropping table: %v", err)
		}
	})

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	var errs []error
	mu := sync.Mutex{}
	pidStartEnd := map[string][]time.Time{}
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int, tst *testing.T) {
			defer wg.Done()
			conn, err := postgres.AcquireConnection(core.NewTransaction("con_"+strconv.Itoa(goroutineID), "", core.LogLevelDebug))
			if err != nil {
				conn.Tx.Error("error acquiring connection")
				errs = append(errs, err)
			}
			err = conn.StartTransaction()
			if err != nil {
				errs = append(errs, err)
			}
			pgBackendPID, err := conn.GetPID()
			pid := strconv.Itoa(int(pgBackendPID))
			mu.Lock()
			pidStartEnd[pid] = append(pidStartEnd[pid], time.Now())
			mu.Unlock()
			if err != nil {
				postgres.Rollback()
				errs = append(errs, err)
			}
			_, err = conn.Query(
				"INSERT INTO "+tableName+" (goroutine_id, tx_pid, current_pid, operation) VALUES (?, ?, pg_backend_pid(), ?)",
				goroutineID,
				pgBackendPID,
				"insert1_"+strconv.Itoa(goroutineID),
			)
			if err != nil {
				postgres.Rollback()
				errs = append(errs, err)
			}
			_, err = conn.Query(
				"INSERT INTO "+tableName+" (goroutine_id, tx_pid, current_pid, operation) VALUES (?, ?, pg_backend_pid(), ?)",
				goroutineID,
				pgBackendPID,
				"insert2_"+strconv.Itoa(goroutineID),
			)
			if err != nil {
				postgres.Rollback()
				errs = append(errs, err)
			}
			err = conn.CommitTransaction()
			mu.Lock()
			pidStartEnd[pid] = append(pidStartEnd[pid], time.Now())
			mu.Unlock()
			if err != nil {
				errs = append(errs, err)
			}
			conn.Release()
			postgres.PrintPoolStats()
		}(i, t)
	}
	wg.Wait() // Validate results rows,
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("Error: %v", err)
		}
		t.Fatalf("Errors occurred during test")
	}
	for pg, times := range pidStartEnd {
		log.Printf("pg: %s, times: %+v", pg, times)
	}
	res, err := postgres.Query("SELECT goroutine_id, tx_pid, current_pid, operation FROM " + tableName)
	log.Printf("test_transactions dump:")
	for _, row := range res {
		log.Printf(
			"goroutine_id: %d, tx_pid: %d, current_pid: %d, operation: %s",
			row["goroutine_id"],
			row["tx_pid"],
			row["current_pid"],
			row["operation"],
		)
	}
	if err != nil {
		t.Fatalf("Error querying test_transactions: %v", err)
	}
	operationByRoutinePID := make(map[string][]string)

	for _, row := range res {
		goroutineID, ok := row["goroutine_id"].(int32)
		if !ok {
			t.Fatalf("Error scanning goroutine_id: not an int")
		}
		txPid, ok := row["tx_pid"].(int32)
		if !ok {
			t.Fatalf("Error scanning tx_pid: not an int")
		}
		pgBackendPID, ok := row["current_pid"].(int32)
		if !ok {
			t.Fatalf("Error scanning current_pid: not an int")
		}
		operation, ok := row["operation"].(string)
		if !ok {
			t.Fatalf("Error scanning operation: not a string")
		}
		goroutineIDStr := strconv.Itoa(int(goroutineID))
		pgBackendPIDStr := strconv.Itoa(int(pgBackendPID))
		operationByRoutinePID[goroutineIDStr+"-"+pgBackendPIDStr] = append(
			operationByRoutinePID[goroutineIDStr+"-"+pgBackendPIDStr],
			operation,
		)
		if txPid != pgBackendPID {
			t.Fatalf("Expected tx_pid and current_pid to be equal, got %d and %d", txPid, pgBackendPID)
		}
	}
	log.Printf("operationByRoutinePID: %v", operationByRoutinePID)
	if len(operationByRoutinePID) != numGoroutines {
		t.Fatalf("Expected %d goroutines, got %d", numGoroutines, len(operationByRoutinePID))
	}
	pids := map[string]bool{}
	for grpid, ops := range operationByRoutinePID {
		routinePid := strings.Split(grpid, "-")
		if pids[routinePid[1]] {
			log.Printf("Found duplicate PID %v. checking start and end", routinePid[1])
			previousEventTime := time.Time{}
			for i, ts := range pidStartEnd[routinePid[1]] {
				log.Printf("[pid-%s] event %d at %v", routinePid[1], i, ts)
				if previousEventTime.After(ts) {
					t.Fatalf(
						"Timestamp not in order for pid %v (%+v). this means the same DB process handled 2 go routines concurrently",
						routinePid[1],
						pidStartEnd[routinePid[1]],
					)
				}
			}
		}
		pids[routinePid[1]] = true
		if len(ops) != 2 {
			t.Fatalf("Expected 2 operations for goroutine %s, got %d (%+v)", grpid, len(ops), ops)
		}
		if ops[0] != "insert1_"+routinePid[0] {
			t.Fatalf("Expected operation 1 to be 'insert1_%s', got '%s'", routinePid[0], ops[0])
		}
		if ops[1] != "insert2_"+routinePid[0] {
			t.Fatalf("Expected operation 2 to be 'insert2_%s', got '%s'", routinePid[0], ops[1])
		}
	}
	log.Printf("Test passed")
}

// TODO: update test to dynamically create resources
func TestAuroraPosgres(t *testing.T) {
	postgres, err := SetupAuroraPostgres(t)
	if err != nil {
		log.Fatalf("Error setting up aurora postgres: %+v", err)
	}
	testDataTypeParsing(t, postgres)
	testVulnerableQuery(t, postgres)
	testParameterizedQuery(t, postgres)

}

func SetupAuroraPostgres(t *testing.T) (*PostgresDBConfig, error) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}

	if envCfg.AuroraClusterName == "" {
		t.Fatalf("error AuroraClusterName is not set in env.json")
	}
	clusterID := envCfg.AuroraClusterName

	if envCfg.AuroraDbName == "" {
		t.Fatalf("error AuroraDbName is not set in env.json")
	}
	dbName := envCfg.AuroraDbName

	uname := "tah_core_user"
	pwd := core.RandomPassword(40)
	auroraCfg := InitAurora(envCfg.Region, "", "", core.LogLevelDebug)
	auroraCfg.SetTx(core.NewTransaction(t.Name(), "", core.LogLevelDebug))
	secretCfg := secret.InitWithRegion("", envCfg.Region, "", "")
	var postgres *PostgresDBConfig

	cluster, err := auroraCfg.GetCluster(clusterID)
	if err != nil {
		log.Printf("Test cluster does not exist. Creating it now.")
		cluster, err = auroraCfg.CreateTestCluster(clusterID, uname, pwd, dbName)
		if err != nil {
			log.Printf("[Aurora] Error creating cluster: %+v ", err)
			return postgres, err
		}
		err = auroraCfg.WaitForClusterCreation(cluster.ID, 600)
		if err != nil {
			log.Printf("[Aurora] Error waiting for cluster: %+v ", err)
			return postgres, err
		}
	}

	log.Printf("Test cluster exists. Running tests.")
	secretArn, err := secretCfg.GetSecretArn("aurora-" + clusterID)
	if err != nil {
		log.Printf("[Aurora] Error getting secret arn: %+v ", err)
		return postgres, err
	}
	pwd = secret.InitWithRegion(secretArn, envCfg.Region, "", "").Get("password")
	log.Printf("Cluster arn for accessing aurora: %s", cluster.Arn)
	postgres = InitPostgresql(cluster.WriterEndpoint, dbName, uname, pwd, 50, core.LogLevelDebug)
	if postgres == nil {
		log.Printf("[Aurora] Error initializing postgresql")
		return postgres, err
	}
	if postgres.conErr != nil {
		log.Printf("[Aurora] Error connecting to postgresql: %+v ", postgres.conErr)
		return postgres, err
	}
	err = postgres.Ping()
	if err != nil {
		log.Printf("[Aurora] Error pinging postgresql: %+v ", err)
		return postgres, err
	}
	res, err := postgres.Query("SELECT * FROM pg_catalog.pg_tables;")
	if err != nil {
		log.Printf("[Aurora] Query Error: %+v ", err)
		return postgres, err
	}
	log.Printf("Query result: %+v", res)
	return postgres, err

}

func testDataTypeParsing(t *testing.T, postgres *PostgresDBConfig) {
	// Create table
	tableName := "type_test_table"
	createVulnerableTableQuery := fmt.Sprintf(`CREATE TABLE %s (num FLOAT);`, tableName)
	_, err := postgres.Query(createVulnerableTableQuery)
	assert.NoError(t, err, "create table query should not return an error")
	t.Cleanup(func() { postgres.Query(fmt.Sprintf("DROP TABLE %s;", tableName)) })

	// Try inserting number
	testNum := 0.124567889
	invalidInsertQuery := fmt.Sprintf(`INSERT INTO %s (num) VALUES (?);`, tableName)
	_, err = postgres.Query(invalidInsertQuery, testNum)
	assert.NoError(t, err, "inserting number should work")

	// Select all via 1=1 attack
	selectQuery := fmt.Sprintf(`SELECT * FROM %s;`, tableName)

	res, err := postgres.Query(selectQuery)
	assert.NoError(t, err, "select query should not return an error")
	assert.Equal(t, 1, len(res), "expect 1 row, got %s", len(res))
	assert.Equal(t, testNum, res[0]["num"], "expect %s row, got %s", testNum, res[0]["num"])
}

// Test the usage of vulnerably queries
func testVulnerableQuery(t *testing.T, postgres *PostgresDBConfig) {
	// Create table
	createVulnerableTableQuery := `
		CREATE TABLE vulnerable_test_table (
    		name VARCHAR(255)
		);
	`
	_, err := postgres.Query(createVulnerableTableQuery)
	assert.NoError(t, err, "create table query should not return an error")
	t.Cleanup(func() { postgres.Query("DROP TABLE vulnerable_test_table;") })

	// Try inserting unescaped single quotes
	invalidInsertQuery := fmt.Sprintf(`
		INSERT INTO vulnerable_test_table (name)
		VALUES
			('%s'),
			('%s');
	`, "regular", "'single_quote'")
	res, err := postgres.Query(invalidInsertQuery)
	log.Printf("res: %+v, err: %+v", res, err)
	assert.Error(t, err, "not expected to handle single quotes")

	res, err = postgres.Query("SELECT * from vulnerable_test_table")
	log.Printf("res: %+v, err: %+v", res, err)
	assert.NoError(t, err, "select query should not return an error")
	assert.Len(t, res, 0, "expect 0 rows, got %s", len(res))
	_, err = postgres.Query("SELECT \"non-existant_field\" from vulnerable_test_table")
	assert.Error(t, err, "not expected to handle non existant field")

	// Insert valid data
	validInsertQuery := "INSERT INTO vulnerable_test_table (name) VALUES ('regular');"
	_, err = postgres.Query(validInsertQuery)
	assert.NoError(t, err, "insert query should not return an error")

	// Select all via 1=1 attack
	name := "'regular' OR 1=1"
	selectAllAttack := fmt.Sprintf(`
		SELECT *
		FROM vulnerable_test_table
		WHERE name = %s;
	`, name)

	res, err = postgres.Query(selectAllAttack)
	assert.NoError(t, err, "select query should not return an error")
	assert.Len(t, res, 1, "expect 1 row, got %s", len(res))
}

// Test the usage of parameterized query to protect against SQL injection.
func testParameterizedQuery(t *testing.T, postgres *PostgresDBConfig) {
	// Create table
	createSafeTableQuery := `
		CREATE TABLE safe_test_table (
    		name VARCHAR(255)
		);
	`
	_, err := postgres.Query(createSafeTableQuery)
	assert.NoError(t, err, "create table query should not return an error")
	t.Cleanup(func() { postgres.Query("DROP TABLE safe_test_table;") })

	// Insert data into the table
	insertQuery := `
		INSERT INTO safe_test_table (name)
		VALUES
			(?),
			(?);
	`
	_, err = postgres.Query(insertQuery, "regular", "'single_quote'")
	assert.NoError(t, err, "insert query should not return an error")

	// Valid select
	selectQuery := `
		SELECT *
		FROM safe_test_table
		WHERE name = ?;
	`
	res, err := postgres.Query(selectQuery, "regular")
	assert.NoError(t, err, "select query should not return an error")
	assert.Len(t, res, 1, "select query should return 1 row")

	// SQL injection select (1=1 always true attack)
	res, err = postgres.Query(selectQuery, "regular OR 1=1")
	assert.NoError(t, err, "select injection query should not return an error")
	assert.Len(t, res, 0, "select injection query should return 0 results")

	// Perform SQL injection to delete the table
	injectedSqlParameterized := "safe_test_table; DROP TABLE ; --"
	parameterizedQuery := `
		SELECT *
		FROM safe_test_table
		WHERE name = ?;
	`
	_, err = postgres.Query(parameterizedQuery, injectedSqlParameterized)
	assert.NoError(t, err, "parameterized query should not return an error")

	// Validate expected results (safe table should exist and unsafe table should not exist).
	res, err = postgres.Query("SELECT * FROM pg_catalog.pg_tables;")
	assert.NoError(t, err, "select all tables query should not return an error")
	var safeExists bool
	for _, row := range res {
		if row["tablename"] == "safe_test_table" {
			safeExists = true
			break
		}
	}
	assert.True(t, safeExists, "safe_test_table table should still exist")
}
