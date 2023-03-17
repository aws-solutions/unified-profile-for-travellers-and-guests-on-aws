package maintainGluePartitions

import (
	"fmt"
	"sync"
	"tah/core/core"
	"tah/core/db"
	"tah/core/glue"
	"tah/ucp-sync/src/business-logic/model"
	"time"
)

var ORIGIN_DATE = "2010/01/01"
var CONFIG_PK_LAST_PARTITION_UPDATE = "ucp_sync_partition_status"
var CONFIG_SK_LAST_PARTITION_UPDATE_PREFIX = "table_"
var STATUS_SUCCESS = "success"
var STATUS_FAILURE = "failure"

func Run(tx core.Transaction, glueCfg glue.Config, configDb db.DBConfig, athenaTables []string, bucketNames []string) error {
	now := time.Now()
	var lastErr error
	wg := sync.WaitGroup{}
	wg.Add(len(athenaTables))
	for i, _ := range athenaTables {
		go func(i int) {
			lastUpdatedStatus, _ := getLastUpdateStatus(tx, configDb, athenaTables[i])
			t := lastUpdatedStatus.LastUpdated
			partitions := []glue.Partition{}
			for t.Before(now) {
				year := t.Format("2006")
				month := t.Format("01")
				day := t.Format("02")
				partitions = append(partitions, glue.Partition{
					Values:   []string{year, month, day},
					Location: "s3://" + bucketNames[i] + "/" + year + "/" + month + "/" + day,
				})
				t = t.AddDate(0, 0, 1)
			}
			errs := glueCfg.AddPartitionsToTable(athenaTables[i], partitions)
			if len(errs) > 0 {
				tx.Log("[GLUE_ERRORS] %v", errs)
				lastErr = errs[len(errs)-1]
			}
			updateStatus(tx, configDb, athenaTables[i], t, lastErr)
			wg.Done()
		}(i)
	}
	wg.Wait()
	return lastErr
}

func buildAddPartitionQuery(tx core.Transaction, t time.Time, tName string, bucketName string, dbName string) string {
	year := t.Format("2006")
	month := t.Format("01")
	day := t.Format("02")
	query := fmt.Sprintf("ALTER TABLE %s.%s ADD PARTITION (year='%s', month='%s', day='%s') location 's3://%s/%s/%s/%s'", dbName, tName, year, month, day, bucketName, year, month, day)
	tx.Log("Athena Query created: %v", query)
	return query
}

func getLastUpdateStatus(tx core.Transaction, configDb db.DBConfig, tableName string) (model.PartitionUpdateStatus, error) {
	origin, _ := time.Parse("2006/01/02", ORIGIN_DATE)
	status := model.PartitionUpdateStatus{}
	err := configDb.Get(CONFIG_PK_LAST_PARTITION_UPDATE, buildSk(tableName), &status)
	if err != nil {
		tx.Log("Could not fetch last updated status form dynamoDB: %v", err)
	}
	if status.LastUpdated.Before(origin) {
		tx.Log("No last updated date in config DB. Running update since origin %v", origin)
		status.LastUpdated = origin
	}
	return status, err
}

func updateStatus(tx core.Transaction, configDb db.DBConfig, tableName string, t time.Time, err error) error {
	status, _ := getLastUpdateStatus(tx, configDb, tableName)
	status.LastUpdated = t
	status.Pk = CONFIG_PK_LAST_PARTITION_UPDATE
	status.Sk = buildSk(tableName)
	if err != nil {
		status.Status = STATUS_FAILURE
		status.Error = err.Error()
	} else {
		status.Status = STATUS_SUCCESS
		status.Error = ""
	}
	tx.Log("Saving status: %v", status)
	_, err2 := configDb.Save(status)
	if err2 != nil {
		tx.Log("Error saving status: %v", err2)
	}
	return err2
}

func buildSk(tableName string) string {
	return CONFIG_SK_LAST_PARTITION_UPDATE_PREFIX + tableName
}
