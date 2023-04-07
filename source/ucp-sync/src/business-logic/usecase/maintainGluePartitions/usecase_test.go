package maintainGluePartitions

import (
	"os"
	"strings"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/db"
	"tah/core/glue"
	"tah/core/kms"
	common "tah/ucp-common/src/constant/admin"
	commonModel "tah/ucp-common/src/model/admin"
	testutils "tah/ucp-common/src/utils/test"
	"testing"
	"time"
	//model "cloudrack-lambda-core/config/model"
)

var UCP_REGION = testutils.GetTestRegion()

var buckets = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: "s3_hotel_booking",
	common.BIZ_OBJECT_AIR_BOOKING:   "s3_air_booking",
	common.BIZ_OBJECT_GUEST_PROFILE: "s3_guest_profile",
	common.BIZ_OBJECT_PAX_PROFILE:   "s3_pax_profile",
	common.BIZ_OBJECT_CLICKSTREAM:   "s3_clickstream",
	common.BIZ_OBJECT_STAY_REVENUE:  "s3_stay_revenue",
}
var jobs = map[string]string{
	common.BIZ_OBJECT_HOTEL_BOOKING: os.Getenv("GLUE_JOB_NAME_HOTEL_BOOKINGS"),
	common.BIZ_OBJECT_AIR_BOOKING:   os.Getenv("GLUE_JOB_NAME_AIR_BOOKING"),
	common.BIZ_OBJECT_GUEST_PROFILE: os.Getenv("GLUE_JOB_NAME_GUEST_PROFILES"),
	common.BIZ_OBJECT_PAX_PROFILE:   os.Getenv("GLUE_JOB_NAME_PAX_PROFILES"),
	common.BIZ_OBJECT_CLICKSTREAM:   os.Getenv("GLUE_JOB_NAME_CLICKSTREAM"),
	common.BIZ_OBJECT_STAY_REVENUE:  os.Getenv("GLUE_JOB_NAME_STAY_REVENUE"),
}

func TestBuildTableConfig(t *testing.T) {
	domains := []commonModel.Domain{commonModel.Domain{Name: "test_dom_1"}, commonModel.Domain{Name: "test_dom_2"}}
	env := "test_env"
	configs := buildTableConfig(domains, env, buckets, jobs)
	nTableExpected := len(domains) * len(common.BUSINESS_OBJECTS)
	if len(configs) != nTableExpected {
		t.Errorf("[TestBuildTableConfig] %v tables should be created", nTableExpected)
	}
	for i, domain := range domains {
		for j, bo := range common.BUSINESS_OBJECTS {
			expected := "ucp_" + env + "_" + bo.Name + "_" + domain.Name
			if configs[i*len(common.BUSINESS_OBJECTS)+j].TableName != expected {
				t.Errorf("[TestBuildTableConfig] table config at pos (%v,%v) should be %v", i, j, expected)
			}
			if configs[i*len(common.BUSINESS_OBJECTS)+j].JobName != jobs[bo.Name] {
				t.Errorf("[TestBuildTableConfig] job name in config at pos (%v,%v) should be %v", i, j, jobs[bo.Name])
			}
		}
	}
}

func TestBuildPartitions(t *testing.T) {
	nDaysSinceLastRun := 30
	origin := time.Now().AddDate(0, 0, -nDaysSinceLastRun)
	partitions, updatedTs := buildPartitions(origin, time.Now(), "ucp_syn_test_bucket", "ucp_syn_test_domain")
	if len(partitions) != nDaysSinceLastRun {
		t.Errorf("[TestbuildPartitions] Shoudl create %v partitions", partitions)
	}
	if updatedTs.Format("2006-01-02") != time.Now().Format("2006-01-02") {
		t.Errorf("[TestbuildPartitions] Updatde time stamp should be %v and not %v", time.Now().Format("2006-01-02"), updatedTs.Format("2006-01-02"))
	}
	if len(partitions) != nDaysSinceLastRun {
		orYear := origin.Format("2006")
		orMonth := origin.Format("01")
		orDay := origin.Format("02")
		if partitions[0].Values[0] != orYear || partitions[0].Values[1] != orMonth || partitions[0].Values[2] != orDay {
			t.Errorf("[TestbuildPartitions] first partition values Shoudl be %v", []string{orYear, orMonth, orDay})
		}
		expectedLocation := strings.Join([]string{"ucp_syn_test_bucket/", "ucp_syn_test_domain", orYear, orMonth, orDay}, "/")
		if partitions[0].Location != expectedLocation {
			t.Errorf("[TestbuildPartitions] first partition location Shoudl be %v", expectedLocation)
		}
	}
}

func TestInvalidDateFormat(t *testing.T) {
	tx := core.NewTransaction("ucp_sync_test", "")
	_, err := getLastUpdateStatus(tx, db.DBConfig{}, "table_name", "2004-03-27")
	if err == nil {
		t.Errorf("[TestInvalidDateFormat] getLastUpdateStatus shoudl return an error with invalid date format %v", "2004-03-27")
	}
}

func TestRun(t *testing.T) {
	tx := core.NewTransaction("ucp_sync_test", "")
	envName := "ucp_syn_test_env_" + time.Now().Format("2006-01-02-15-04-05")
	dbName := "ucp_sync_test_db"
	glueClient := glue.Init(UCP_REGION, dbName)
	kmsc := kms.Init(UCP_REGION)
	accp := customerprofiles.Init(UCP_REGION)
	nDaysSinceLastRun := 30
	origin := time.Now().AddDate(0, 0, -nDaysSinceLastRun)
	origin_date := origin.Format("2006/01/02")

	cfg, err := db.InitWithNewTable("ucp_sync_unit_tests", "item_id", "item_type")
	if err != nil {
		t.Errorf("[TestRun] error init with new table: %v", err)
	}
	err = glueClient.CreateDatabase(dbName)
	if err != nil {
		t.Errorf("[TestRun] error creating database: %v", err)
	}
	keyArn, err0 := kmsc.CreateKey("tah-unit-test-key")
	if err0 != nil {
		t.Errorf("Could not create KMS key to unit test UCP %v", err0)
	}
	domains, err1 := createDomains(tx, accp, keyArn, envName)
	if err1 != nil {
		t.Errorf("Could not create domains to unit test UCP %v", err1)
	}
	tables, err2 := createTables(tx, glueClient, domains, envName, buckets)
	if err2 != nil {
		t.Errorf("[TestRun] error creating tables: %v", err2)
	}
	for _, tb := range tables {
		status, err3 := getLastUpdateStatus(tx, cfg, tb, origin_date)
		if err3 != nil {
			t.Errorf("[TestRun] error getLastUpdateStatus: %v", err3)
		}
		if status.LastUpdated.Format("2006/01/02") != origin_date {
			t.Errorf("[TestRun] initial status should be %v and not %v", origin, status.LastUpdated)
		}
	}

	partitionErrors, glueJobsErrors := Run(tx, glueClient, cfg, buckets, accp, envName, origin_date, jobs)
	if len(partitionErrors) > 0 {
		t.Errorf("[TestRun] Error during partitions updates: %v", partitionErrors)
	}
	if len(glueJobsErrors) > 0 {
		t.Errorf("[TestRun] Error during glue Job start: %v", glueJobsErrors)
	}
	for _, tb := range tables {
		status, err3 := getLastUpdateStatus(tx, cfg, tb, origin_date)
		if err3 != nil {
			t.Errorf("[TestRun] error getLastUpdateStatus: %v", err3)
		}
		now := time.Now()
		expected := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if status.LastUpdated.Format("2006/01/02") != expected.Format("2006/01/02") {
			t.Errorf("[TestRun] final status should be %v and not %v", expected.Format("2006/01/02"), status.LastUpdated.Format("2006/01/02"))
		}
	}

	tx.Log("Cleanup")
	err = deleteDomains(tx, accp, domains)
	if err != nil {
		t.Errorf("Error deleting domains %v", err)
	}
	err = deleteTables(tx, glueClient, tables)
	if err != nil {
		t.Errorf("Error deleting tables %v", err)
	}
	err = kmsc.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting key %v", err)
	}
	err = cfg.DeleteTable(cfg.TableName)
	if err != nil {
		t.Errorf("Error deleting dynamoDB table %v", err)
	}
	err = glueClient.DeleteDatabase(dbName)
	if err != nil {
		t.Errorf("[TestRun] error deleting database: %v", err)
	}
}

func createTables(tx core.Transaction, glueCfg glue.Config, domains []string, env string, buckets map[string]string) ([]string, error) {
	names := []string{}
	for _, domain := range domains {
		for _, bo := range common.BUSINESS_OBJECTS {
			tName := "ucp_" + env + "_" + bo.Name + "_" + domain
			tx.Log("Creating table %v", tName)
			err := glueCfg.CreateTable(tName, buckets[bo.Name], map[string]string{"year": "int", "month": "int", "day": "int"})
			if err != nil {
				return names, err
			}
			names = append(names, tName)
		}
	}
	return names, nil
}

func createDomains(tx core.Transaction, accp customerprofiles.CustomerProfileConfig, keyArn string, envName string) ([]string, error) {
	domains := []string{}
	domains = append(domains, "ucp_sync_test_"+time.Now().Format("2006-01-02-15-04-05"))
	domains = append(domains, "ucp_sync_test_2"+time.Now().Format("2006-01-02-15-04-05"))

	for _, domain := range domains {
		tx.Log("Creating domain %v", domain)
		err := accp.CreateDomainWithQueue(domain, true, keyArn, map[string]string{"envName": envName}, "")
		if err != nil {
			return domains, err
		}
	}
	return domains, nil
}

func deleteDomains(tx core.Transaction, accp customerprofiles.CustomerProfileConfig, domains []string) error {
	for _, domain := range domains {
		tx.Log("Deleting domain %v", domain)
		err := accp.DeleteDomainByName(domain)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteTables(tx core.Transaction, glueCfg glue.Config, tables []string) error {
	for _, table := range tables {
		tx.Log("Deleting table %v", table)
		err := glueCfg.DeleteTable(table)
		if err != nil {
			return err
		}
	}
	return nil
}
