// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aurora

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	core "tah/upt/source/tah-core/core"
	"time"

	pgxv5 "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DEFAULT_POSTGRESQL_PORT = "5432"

// Init function for mysql driver
func InitPostgresql(host string, dbname string, user string, pwd string, poolMaxCon int, logLevel core.LogLevel) *PostgresDBConfig {
	tx := core.NewTransaction("AURORA-POSTGRESQL", "", logLevel)
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require pool_max_conns=%d statement_cache_capacity=0 default_query_exec_mode=exec",
		host,
		DEFAULT_POSTGRESQL_PORT,
		user,
		pwd,
		dbname,
		poolMaxCon,
	)

	// Connect to database
	//tx.Log("Connecting to Aurora PostgreSQL cluster")
	//db, err := sql.Open("postgres", psqlInfo)
	//if err != nil {
	//	log.Fatal(err)
	//}
	dbpool, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		tx.Error("ERROR: Cannot connect to PostgreSQL cluster on Aurora: %+v \n", err)
		return &PostgresDBConfig{conErr: err}
	}
	dbcfg := PostgresDBConfig{
		ConnPool: dbpool,
	}

	// Note: sql.Open() validates the connection string, but does NOT immediately connect to the DB or indicate
	// a connection has been made. If we want to explicitly test a connection, we can optionally call db.Ping().
	// https://pkg.go.dev/database/sql#Open
	tx.Debug("Valid PostgreSQL connection pool provided. DB config: %+v\n", dbcfg)
	return &dbcfg
}

func BuildStartTransactionPostgreSql() string {
	return "START TRANSACTION;"
}
func BuildCommitPostgreSql() string {
	return "COMMIT;"
}
func BuildRollbackPostgreSql() string {
	return "ROLLBACK;"
}

func BuildGetPIDSql() string {
	return "SELECT pg_backend_pid();"
}

func (dbcfg *PostgresDBConfig) Engine() string {
	return dbcfg.engine
}
func (dbcfg *PostgresDBConfig) ConErr() error {
	return dbcfg.conErr
}

func (dbcfg *PostgresDBConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "aurora"
	dbcfg.Tx = tx
}

func (dbcfg *PostgresDBConfig) StartTransaction() error {
	_, err := dbcfg.Query(BuildStartTransactionPostgreSql())
	return err
}
func (dbcfg *PostgresDBConfig) Commit() error {
	_, err := dbcfg.Query(BuildCommitPostgreSql())
	return err
}
func (dbcfg *PostgresDBConfig) Rollback() error {
	_, err := dbcfg.Query(BuildRollbackPostgreSql())
	return err
}

func (dbcfg *PostgresDBConfig) HasTable(name string) bool {
	panic("Not implemented")
}

func (dbcfg *PostgresDBConfig) GetColumns(tableName string) ([]Column, error) {
	res, err := dbcfg.Query(fmt.Sprintf("SHOW COLUMNS FROM %s;", tableName))
	//dbcfg.Tx.Log("[GetColumns] Get Columns result %v\n", res)
	columns := []Column{}
	for _, rec := range res {
		for _, val := range rec {
			columns = append(columns, Column{Field: val.(string)})
		}
	}
	return columns, err
}

/*
Stat is a snapshot of Pool statistics.
see https://pkg.go.dev/github.com/jackc/pgx/v4/pgxpool#Pool.Config
AcquireCount returns the cumulative count of successful acquires from the pool.
AcquireDuration returns the total duration of all successful acquires from the pool.
AcquiredConns returns the number of currently acquired connections in the pool.
CanceledAcquireCount returns the cumulative count of acquires from the pool that were canceled by a context.
ConstructingConns returns the number of conns with construction in progress in the pool.
EmptyAcquireCount returns the cumulative count of successful acquires from the pool that waited for a resource to be released or constructed because the pool was empty.
IdleConns returns the number of currently idle conns in the pool.
MaxConns returns the maximum size of the pool.
MaxIdleDestroyCount returns the cumulative count of connections destroyed because they exceeded MaxConnIdleTime.
MaxLifetimeDestroyCount returns the cumulative count of connections destroyed because they exceeded MaxConnLifetime.
NewConnsCount returns the cumulative count of new connections opened.
*/
func (dbcfg *PostgresDBConfig) GetPoolStats() map[string]interface{} {
	statsMap := map[string]interface{}{}
	stats := dbcfg.ConnPool.Stat()
	statsMap["AcquireCount"] = stats.AcquireCount()
	statsMap["AcquireDuration"] = stats.AcquireDuration()
	statsMap["AcquiredConns"] = stats.AcquiredConns()
	statsMap["CanceledAcquireCount"] = stats.CanceledAcquireCount()
	statsMap["ConstructingConns"] = stats.ConstructingConns()
	statsMap["EmptyAcquireCount"] = stats.EmptyAcquireCount()
	statsMap["IdleConns"] = stats.IdleConns()
	statsMap["MaxConns"] = stats.MaxConns()
	statsMap["MaxIdleDestroyCount"] = stats.MaxIdleDestroyCount()
	statsMap["MaxLifetimeDestroyCount"] = stats.MaxLifetimeDestroyCount()
	statsMap["NewConnsCount"] = stats.NewConnsCount()
	statsMap["TotalConns"] = stats.TotalConns()
	return statsMap
}

func (dbcfg *PostgresDBConfig) PrintPoolStats() {
	stats := dbcfg.GetPoolStats()
	dbcfg.Tx.Debug("Connection Pool stats: ")
	for name, val := range stats {
		dbcfg.Tx.Debug("%s : %+v", name, val)
	}
}

func (dbcfg *PostgresDBConfig) AcquireConnection(tx core.Transaction) (DBConnection, error) {
	dbcfg.Tx.Info("[postgresql] Acquiring DB connection (trace_id: %s)", tx.TransactionID)
	conn, err := dbcfg.ConnPool.Acquire(context.Background())
	if err != nil {
		dbcfg.Tx.Error("[postgresql] Error while acquiring DB connection: %+v", err)
		return DBConnection{}, err
	}
	return DBConnection{Conn: conn, Tx: tx}, nil
}

func (conn *DBConnection) Release() {
	conn.Tx.Info("[postgresql] Releasing DB connection")
	conn.Conn.Release()
}

func (conn *DBConnection) StartTransaction() error {
	conn.Tx.Info("[postgresql] Starting transaction")
	_, err := conn.Query(BuildStartTransactionPostgreSql())
	if err != nil {
		conn.Tx.Error("[postgresql] Error while starting transaction: %+v", err)
	}
	return err
}

func (conn *DBConnection) CommitTransaction() error {
	conn.Tx.Info("[postgresql] Commiting transaction")
	_, err := conn.Query(BuildCommitPostgreSql())
	if err != nil {
		conn.Tx.Error("[postgresql] Error while committing transaction: %+v", err)
	}
	return err
}
func (conn *DBConnection) RollbackTransaction() error {
	conn.Tx.Info("[postgresql] Rolling back transaction")
	_, err := conn.Query(BuildRollbackPostgreSql())
	if err != nil {
		conn.Tx.Error("[postgresql] Error while rolling back transaction: %+v", err)
	}
	return err
}

func (dbcfg *DBConnection) GetPID() (int, error) {
	res, err := dbcfg.Query(BuildGetPIDSql())
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, fmt.Errorf("no pid returned")
	}
	pid, ok := res[0]["pg_backend_pid"].(int32)
	if !ok {
		return 0, fmt.Errorf("cannot convert %v to int32", reflect.TypeOf(res[0]["pg_backend_pid"]))
	}
	return int(pid), err
}

func (conn *DBConnection) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	query, args, err := prepQuery(conn.Tx, query, args...)
	if err != nil {
		conn.Tx.Error("[postgresql] Error while preparing query: %+v", err)
		return []map[string]interface{}{}, err
	}
	rows, err := conn.Conn.Query(context.Background(), query, args...) // Note: Ignoring errors for brevity
	if err != nil {
		conn.Tx.Error("Error while running query %s: %v", query, err)
		return []map[string]interface{}{}, err
	}
	return parseQueryResults(conn.Tx, rows)
}

func prepQuery(tx core.Transaction, query string, args ...interface{}) (string, []interface{}, error) {
	// Validate the number query parameters matches provided number of args
	// e.g. "SELECT * FROM table WHERE id = ?" => 1 query parameter is expected
	expectedArgs := strings.Count(query, "?")
	if len(args) != expectedArgs {
		tx.Error("[postgresql] Query parameter mismatch: expected %d, got %d", expectedArgs, len(args))
		return query, args, fmt.Errorf(
			"number of provided arguments (%d) does not match number of query parameters (%d)",
			len(args),
			expectedArgs,
		)
	}

	// Replace ? with $n; n being the index of the argument, starting at 1
	// https://www.postgresql.org/docs/current/xfunc-sql.html#XFUNC-SQL-FUNCTION-ARGUMENTS
	for i := range args {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i+1), 1)
	}

	tx.Debug("[postgresql] Executing Query %s with args %d", query, len(args))
	return query, args, nil
}

func parseQueryResults(tx core.Transaction, rows pgxv5.Rows) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)
	columns := rows.FieldDescriptions()
	length := len(columns)
	for rows.Next() {
		current := makeResultReceiverPosgresql(length)
		if err := rows.Scan(current...); err != nil {
			tx.Error("[postgresql] error scanning result: %s", err)
			return nil, err
		}
		value := make(map[string]interface{})
		for i := 0; i < length; i++ {
			key := columns[i].Name
			val := *(current[i]).(*interface{})
			if val == nil {
				value[key] = nil
				continue
			}
			switch v := val.(type) {
			case int64:
				value[key] = v
			case int32:
				value[key] = v
			case float64:
				value[key] = v
			case bool:
				value[key] = v
			case string:
				value[key] = v
			case time.Time:
				value[key] = v
			case []uint8:
				value[key] = string(v)
			default:
				tx.Warn("[postgresql] unsupported data type %v for value '%+v' now", reflect.TypeOf(val), val)
			}
		}
		results = append(results, value)
	}
	rows.Close()
	return results, rows.Err()
}

// Execute a query and return the results
//
// Parameterized queries are supported. A ? is used as a placeholder for a query parameter.
// The provided args are used to replace the placeholders in the query, allowing us to safely
// handle 1/ data provided by users, protecting from sql injection 2/ escaping quotation marks, etc.
//
// https://go.dev/doc/database/sql-injection
//
//	e.g. "SELECT * FROM table WHERE id = ?" => 1 query parameter included, 1 arg for id expected
func (dbcfg *PostgresDBConfig) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	query, args, err := prepQuery(dbcfg.Tx, query, args...)
	if err != nil {
		dbcfg.Tx.Error("[postgresql] Error while preparing query: %+v", err)
		return []map[string]interface{}{}, err
	}
	rows, err := dbcfg.ConnPool.Query(context.Background(), query, args...) // Note: Ignoring errors for brevity
	if err != nil {
		dbcfg.Tx.Error("Error while running query %s: %v", query, err)
		return []map[string]interface{}{}, err
	}
	return parseQueryResults(dbcfg.Tx, rows)

}

func (dbcfg *PostgresDBConfig) Ping() error {
	return dbcfg.ConnPool.Ping(context.Background())
}

func makeResultReceiverPosgresql(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{} = struct{}{}
		result = append(result, &current)
	}
	return result
}

func (dbcfg *PostgresDBConfig) Close() {
	dbcfg.ConnPool.Close()
}
