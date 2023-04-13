package maintainGluePartitions

import (
	"errors"
	"fmt"
	"sync"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/db"
	"tah/core/glue"
	common "tah/ucp-common/src/constant/admin"
	commonModel "tah/ucp-common/src/model/admin"
	adminSvc "tah/ucp-common/src/services/admin"
	"tah/ucp-sync/src/business-logic/model"
	"time"
)

var CONFIG_PK_LAST_PARTITION_UPDATE = "ucp_sync_partition_status"
var CONFIG_SK_LAST_PARTITION_UPDATE_PREFIX = "table_"
var STATUS_SUCCESS = "success"
var STATUS_FAILURE = "failure"

type TableConfig struct {
	TableName  string
	DomainName string
	BucketName string
	JobName    string
}

func Run(tx core.Transaction, glueCfg glue.Config, configDb db.DBConfig, bucketNames map[string]string, accpCfg customerprofiles.CustomerProfileConfig, env string, originDate string, jobs map[string]string) ([]error, []error) {
	now := time.Now()
	partitionErrors := []error{}
	glueJobsErrors := []error{}
	wg := sync.WaitGroup{}

	tx.Log("1- searching domains for env %v", env)
	domains, err := adminSvc.SearchDomain(tx, accpCfg, env)
	if err != nil {
		tx.Log("Error searching domains: %v", err)
		return partitionErrors, glueJobsErrors
	}
	tx.Log("2-Creating athena Table for %v domains and %v business objects", len(domains), len(common.BUSINESS_OBJECTS))
	configs := buildTableConfig(domains, env, bucketNames, jobs)
	tx.Log("3-Updating partitions for %v athena table", len(configs))
	wg.Add(len(configs))
	for i, config := range configs {
		tableName := config.TableName
		domainName := config.DomainName
		bucketName := config.BucketName
		jobName := config.JobName
		go func(i int, tName string, domain string, bucket string) {
			tx.Log("[%v] 1-Getting last update status", tName)
			lastUpdatedStatus, err := getLastUpdateStatus(tx, configDb, tName, originDate)
			if err != nil {
				tx.Log("Coudl not get last status: %v", err)
				partitionErrors = append(partitionErrors, err)
				wg.Done()
				return
			}
			tx.Log("[%v]Last updated time is %v", tName, lastUpdatedStatus.LastUpdated)

			tx.Log("[%v] 2-Adding Partitions", tName)
			partitions, updatedTimeStamp := buildPartitions(lastUpdatedStatus.LastUpdated, now, bucket, domain)
			errs := glueCfg.AddPartitionsToTable(tName, partitions)
			var statusErr error
			if len(errs) > 0 {
				tx.Log("[%v] errors adding partions to table: %v", tName, errs)
				statusErr = errors.New(fmt.Sprintf("%v errors while adding patitions to glue table %v: %v", len(errs), tName, errs))
				partitionErrors = append(partitionErrors, statusErr)
			}
			tx.Log("[%v] 3-Updating status", tName)
			updateStatus(tx, configDb, tName, updatedTimeStamp, statusErr, originDate)
			tx.Log("[%v] 4-Start Glue Job %v", tName, jobName)
			err = glueCfg.RunJob(jobName, map[string]string{
				"--DEST_BUCKET":  bucket,
				"--SOURCE_TABLE": tName,
				"--ACCP_DOMAIN":  domain,
			})
			if err != nil {
				tx.Log("Coudl not start Glue job %v: %v", jobName, err)
				glueJobsErrors = append(glueJobsErrors, err)
			}
			wg.Done()
		}(i, tableName, domainName, bucketName)
	}
	wg.Wait()
	return partitionErrors, glueJobsErrors
}

func buildPartitions(lastupdated time.Time, now time.Time, bucketName string, domain string) ([]glue.Partition, time.Time) {
	partitions := []glue.Partition{}
	t := lastupdated
	for t.Before(now.AddDate(0, 0, -1)) {
		year := t.Format("2006")
		month := t.Format("01")
		day := t.Format("02")
		partitions = append(partitions, glue.Partition{
			Values:   []string{year, month, day},
			Location: "s3://" + bucketName + "/" + domain + "/" + year + "/" + month + "/" + day,
		})
		t = t.AddDate(0, 0, 1)
	}
	return partitions, t
}

func buildTableConfig(domains []commonModel.Domain, env string, bucketNames map[string]string, jobs map[string]string) []TableConfig {
	configs := []TableConfig{}
	for _, domain := range domains {
		for _, bo := range common.BUSINESS_OBJECTS {
			configs = append(configs, TableConfig{
				TableName:  buildAthenaTableName(env, bo, domain),
				DomainName: domain.Name,
				BucketName: bucketNames[bo.Name],
				JobName:    jobs[bo.Name],
			})
		}
	}
	return configs
}

func getLastUpdateStatus(tx core.Transaction, configDb db.DBConfig, tableName string, originDate string) (model.PartitionUpdateStatus, error) {
	status := model.PartitionUpdateStatus{}
	origin, err := time.Parse("2006/01/02", originDate)
	if err != nil {
		tx.Log("Invalid origin date: %v", err)
		return status, err
	}
	err = configDb.Get(CONFIG_PK_LAST_PARTITION_UPDATE, buildSk(tableName), &status)
	if err != nil {
		tx.Log("No updated status in DynamoDB. returning origin date %v: %v", origin, err)
		return model.PartitionUpdateStatus{LastUpdated: origin}, nil
	}
	if status.LastUpdated.Before(origin) {
		tx.Log("No last updated date in config DB. Running update since origin %v", origin)
		status.LastUpdated = origin
	}
	return status, err
}

func updateStatus(tx core.Transaction, configDb db.DBConfig, tableName string, t time.Time, err error, originDate string) error {
	status, _ := getLastUpdateStatus(tx, configDb, tableName, originDate)
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

func buildAthenaTableName(env string, bo commonModel.BusinessObject, domain commonModel.Domain) string {
	return "ucp_" + env + "_" + bo.Name + "_" + domain.Name
}
