// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maintainGluePartitions

import (
	"fmt"
	"log"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/glue"
	common "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
	adminSvc "tah/upt/source/ucp-common/src/services/admin"
	"tah/upt/source/ucp-sync/src/business-logic/model"
	"time"
)

var CONFIG_PK_LAST_PARTITION_UPDATE = "ucp_sync_partition_status"
var CONFIG_SK_LAST_PARTITION_UPDATE_PREFIX = "table_"
var STATUS_SUCCESS = "success"
var STATUS_FAILURE = "failure"

type TableConfig struct {
	TableName          string
	DomainName         string
	BucketName         string
	JobName            string
	BusinessObjectName string
}

func Run(cfg model.AWSConfig) ([]error, []error) {
	tx := cfg.Tx
	bucketNames := cfg.BizObjectBucketNames
	jobs := cfg.JobNames
	accpCfg := cfg.Accp
	env := cfg.Env["LAMBDA_ENV"]
	skipJobRun := cfg.Env["SKIP_JOB_RUN"]
	methodOfInvocation := cfg.Env["METHOD_OF_INVOCATION"]
	requestedDomains := cfg.Domains

	now := time.Now().UTC()
	partitionErrors := []error{}
	glueJobsErrors := []error{}

	tx.Debug("1- searching domains for env %v", env)
	domains, err := adminSvc.SearchDomain(tx, accpCfg, env)
	if err != nil {
		tx.Error("Error searching domains: %v", err)
		return partitionErrors, glueJobsErrors
	}
	if len(requestedDomains) > 0 {
		tx.Debug("Overriding domain list with manual request: %+v", requestedDomains)
		domains = requestedDomains
	}
	if cfg.RequestedJob != "" {
		tx.Debug("Requested job: %+v. only this job will be run", cfg.RequestedJob)
		tx.Debug("2-Creating athena Table for %v domains and 1 business object", len(domains))
	} else {
		tx.Debug("2-Creating athena Table for %v domains and %v business objects", len(domains), len(common.BUSINESS_OBJECTS))
	}
	configs := buildTableConfig(domains, env, bucketNames, jobs, cfg.RequestedJob)
	tx.Debug("3-Updating partitions for %v athena table", len(configs))

	for _, config := range configs {
		tx.Debug("Creation partition and running job for Table: %v", config.TableName)
		//Check if either running in manual mode or scheduled runs disabled
		if skipJobRun != "true" || methodOfInvocation == common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
			partitionErrors, glueJobsErrors = addAndUpdate(cfg, partitionErrors, glueJobsErrors, config, now, methodOfInvocation)
		} else {
			tx.Debug("Scheduled Runs disabled. Run jobs on demand only")
		}
	}
	return partitionErrors, glueJobsErrors
}

func addAndUpdate(cfg model.AWSConfig, partitionErrors []error, glueJobsErrors []error, config TableConfig, now time.Time, methodOfInvocation string) ([]error, []error) {
	tx := cfg.Tx
	glueCfg := cfg.Glue
	configDb := cfg.Dynamo
	tName := config.TableName
	domain := config.DomainName
	bucket := config.BucketName
	jobName := config.JobName
	boName := config.BusinessObjectName
	originDate := cfg.Env["ORIGIN_DATE"]
	accpBucket := cfg.Env["CONNECT_PROFILE_SOURCE_BUCKET"]
	metricsUuid := cfg.Env["METRRIC_UUID"]

	tx.Debug("[%v] 1-Getting last update status", tName)
	//we always use the latest update status even in manual override mode to avoid recreating partitions
	lastUpdatedStatus, err := getLastUpdateStatus(tx, configDb, tName, originDate, "")
	if err != nil {
		tx.Error("Could not get last status: %v", err)
		partitionErrors = append(partitionErrors, err)
		return partitionErrors, glueJobsErrors
	}
	tx.Debug("[%v]Last updated time is %v", tName, lastUpdatedStatus.LastUpdated)

	tx.Debug("[%v] 2-Adding Partitions", tName)
	partitions, updatedTimeStamp := buildPartitions(lastUpdatedStatus.LastUpdated, now, bucket, domain)
	errs := glueCfg.AddParquetPartitionsToTable(tName, partitions)
	var statusErr error
	if len(errs) > 0 {
		tx.Error("[%v] errors adding partions to table: %v", tName, errs)
		statusErr = fmt.Errorf("%v errors while adding patitions to glue table %v: %v", len(errs), tName, errs)
		partitionErrors = append(partitionErrors, statusErr)
	}

	tx.Debug("[%v] 3-Updating status", tName)
	updateStatus(tx, configDb, tName, updatedTimeStamp, statusErr, originDate, methodOfInvocation)

	tx.Debug("[%v] 4-Start Glue Job %v", tName, jobName)
	if methodOfInvocation == common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
		tx.Debug("Manual run. reseting job bookmark before running")
		err := resetJobBookmark(tx, configDb, boName, domain)
		if err != nil {
			tx.Error("Could not reset bookmark: %v", err)
		}
	}
	err = glueCfg.RunJob(jobName, map[string]string{
		"--DEST_BUCKET":  accpBucket,
		"--SOURCE_TABLE": tName,
		"--ACCP_DOMAIN":  domain,
		"--METRICS_UUID": metricsUuid,
	})
	if err != nil {
		tx.Error("Could not start Glue job %v: %v", jobName, err)
		glueJobsErrors = append(glueJobsErrors, err)
		return partitionErrors, glueJobsErrors
	}
	return partitionErrors, glueJobsErrors
}

func buildPartitions(lastupdated time.Time, now time.Time, bucketName string, domain string) ([]glue.Partition, time.Time) {
	partitions := []glue.Partition{}
	t := lastupdated
	for t.Before(now) {
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

func buildTableConfig(domains []commonModel.Domain, env string, bucketNames map[string]string, jobs map[string]string, requestedJob string) []TableConfig {
	configs := []TableConfig{}
	for _, domain := range domains {
		for _, bo := range common.BUSINESS_OBJECTS {
			if requestedJob == "" || requestedJob == jobs[bo.Name] {
				configs = append(configs, TableConfig{
					TableName:          buildAthenaTableName(env, bo, domain),
					DomainName:         domain.Name,
					BucketName:         bucketNames[bo.Name],
					JobName:            jobs[bo.Name],
					BusinessObjectName: bo.Name,
				})
			}
		}
	}
	log.Printf("Final table config list to eb run: %+v", configs)
	return configs
}

func getLastUpdateStatus(tx core.Transaction, configDb db.DBConfig, tableName string, originDate string, methodOfInvocation string) (model.PartitionUpdateStatus, error) {
	status := model.PartitionUpdateStatus{}
	origin, err := time.Parse("2006/01/02", originDate)
	if err != nil {
		tx.Error("Invalid origin date: %v", err)
		return status, err
	}
	if methodOfInvocation == common.CLOUDWATCH_EVENT_DETAIL_UCP_MANUAL_TRIGGER {
		log.Printf("Job run  manually requested: starting at origin date: %v", origin)
		status.LastUpdated = origin
		return status, nil
	}
	err = configDb.Get(CONFIG_PK_LAST_PARTITION_UPDATE, buildSk(tableName), &status)
	if err != nil {
		tx.Error("No updated status in DynamoDB. returning origin date %v: %v", origin, err)
		return model.PartitionUpdateStatus{LastUpdated: origin}, nil
	}
	if status.LastUpdated.Before(origin) {
		tx.Debug("No last updated date in config DB. Running update since origin %v", origin)
		status.LastUpdated = origin
	}
	return status, err
}

func updateStatus(tx core.Transaction, configDb db.DBConfig, tableName string, t time.Time, err error, originDate string, methodOfInvocation string) error {
	status, _ := getLastUpdateStatus(tx, configDb, tableName, originDate, methodOfInvocation)
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
	tx.Debug("Saving status: %v", status)
	_, err2 := configDb.Save(status)
	if err2 != nil {
		tx.Error("Error saving status: %v", err2)
	}
	return err2
}

func resetJobBookmark(tx core.Transaction, configDb db.DBConfig, bizObjectName string, domainName string) error {
	tx.Info("Resetting bookmark for glue job %v_%v", bizObjectName, domainName)
	pk := "glue_job_bookmark"
	sk := bizObjectName + "_" + domainName
	_, err := configDb.Delete(model.JobBookmark{Pk: pk, Sk: sk})
	if err != nil {
		tx.Error("Error deleting bookmark: %v", err)
	}
	return err
}

func buildSk(tableName string) string {
	return CONFIG_SK_LAST_PARTITION_UPDATE_PREFIX + tableName
}

func buildAthenaTableName(env string, bo commonModel.BusinessObject, domain commonModel.Domain) string {
	return "ucp_" + env + "_" + bo.Name + "_" + domain.Name
}
