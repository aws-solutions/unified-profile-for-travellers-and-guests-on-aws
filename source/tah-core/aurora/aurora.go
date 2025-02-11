// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aurora

import (
	sql "database/sql"
	"errors"
	"fmt"
	core "tah/upt/source/tah-core/core"
	secret "tah/upt/source/tah-core/secret"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
	"github.com/jackc/pgx/v5/pgxpool"
)

//////////////////////
// MODEL
////////////////////////

type DBConfig interface {
	Query(query string) ([]map[string]interface{}, error)
	StartTransaction() error
	Commit() error
	Rollback() error
	ConErr() error
	HasTable(name string) bool
	GetColumns(tableName string) ([]Column, error)
	Engine() string
	SetTx(tx core.Transaction)
}

type DBConnection struct {
	Conn *pgxpool.Conn
	Tx   core.Transaction
}

type Cluster struct {
	ID             string
	Status         string
	Arn            string
	ReaderEndpoint string
	WriterEndpoint string
}

type Instance struct {
	ID     string
	Status string
}

type PostgresDBConfig struct {
	//not thread-safe
	Db *sql.DB
	//thread save
	ConnPool *pgxpool.Pool
	engine   string
	conErr   error //connection error
	Aurora   *AuroraConfig
	Tx       core.Transaction //service transaction ID
}

// Config for data services
type DataServiceDBConfig struct {
	Db               *rdsdataservice.RDSDataService
	Aurora           *AuroraConfig
	CurTransactionId string //for now we keep only one transaction at the teim for compatibility purpose with GORM implementation

}

type Column struct {
	Field string `gorm:"column:Field"`
	Type  string `gorm:"column:Type"`
}

type AuroraConfig struct {
	RdsSvc         *rds.RDS
	SecretStoreArn string
	ClusterArn     string
	DbName         string
	Tx             core.Transaction //service transaction ID
	Region         string
	ReaderEndpoint string
	WriterEndpoint string
}

func InitAurora(region, solutionId, solutionVersion string, logLevel core.LogLevel) *AuroraConfig {
	client := core.CreateClient(solutionId, solutionVersion)
	mySession := session.Must(session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: client}))
	dbcfg := &AuroraConfig{
		RdsSvc: rds.New(mySession),
		Tx:     core.NewTransaction("aurora", "", logLevel),
		Region: region,
	}
	return dbcfg
}

func (dbcfg *AuroraConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "AURORA"
	dbcfg.Tx = tx
}

// mostly used for test purpose so we don't need to support all the options
func (dbcfg *AuroraConfig) CreateTestCluster(id string, uname string, pwd string, dbName string) (Cluster, error) {
	dbcfg.Tx.Info("Creating tests cluster: %v", id)

	dbcfg.Tx.Debug("0. Creating secret: %v", id)
	sm := secret.InitWithRegion("", dbcfg.Region, "", "")
	arn, err := sm.CreateAuroraSecret("aurora-"+id, uname, pwd)
	if err != nil {
		dbcfg.Tx.Error("Error creating secret: %v", err)
		return Cluster{}, err
	}
	dbcfg.SecretStoreArn = arn

	dbcfg.Tx.Debug("1. Creating cluster: %v", id)
	input := &rds.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(id),
		Engine:              aws.String("aurora-postgresql"),
		EngineVersion:       aws.String("15.3"),
		MasterUserPassword:  aws.String(pwd),
		MasterUsername:      aws.String(uname),
		EnableHttpEndpoint:  aws.Bool(true),
		DatabaseName:        aws.String(dbName),
		// PubliclyAccessible:  aws.Bool(true), // Configuration option is not valid for PostgreSQL, must set manually
		ServerlessV2ScalingConfiguration: &rds.ServerlessV2ScalingConfiguration{
			MaxCapacity: aws.Float64(2),
			MinCapacity: aws.Float64(1),
		},
	}

	_, err = dbcfg.RdsSvc.CreateDBCluster(input)
	if err != nil {
		dbcfg.Tx.Error("Error creating DB cluster: %v", err)
		return Cluster{}, err
	}

	dbcfg.Tx.Debug("2. Creating intstance: %v", id)
	instanceInput := &rds.CreateDBInstanceInput{
		DBClusterIdentifier:  aws.String(id),
		DBInstanceIdentifier: aws.String(id + "-instance-1"),
		Engine:               aws.String("aurora-postgresql"),
		DBInstanceClass:      aws.String("db.serverless"),
	}

	instanceOut, err := dbcfg.RdsSvc.CreateDBInstance(instanceInput)
	if err != nil {
		dbcfg.Tx.Error("Error creating DB cluster: %v", err)
		return Cluster{}, err
	}
	return Cluster{ID: id, Status: *instanceOut.DBInstance.DBInstanceStatus}, nil
}

// mostly used for test purpose so we don't need to support all the options
func (dbcfg *AuroraConfig) WaitForClusterCreation(id string, timeout int) error {
	dbcfg.Tx.Info("[aurora] Waiting for aurora cluster %s to be created", id)
	max := timeout / 5
	for i := 0; i < max; i++ {
		cluster, err := dbcfg.GetCluster(id)
		instance, err2 := dbcfg.GetClusterInstance(id + "-instance-1")
		if err != nil {
			dbcfg.Tx.Error("[aurora] Error retrieving aurora cluster %s", err)
			return err
		}
		if err2 != nil {
			dbcfg.Tx.Error("[aurora] Error retrieving aurora instance %s", err)
			return err
		}
		if cluster.Status == "available" && instance.Status == "available" {
			dbcfg.Tx.Info("[aurora] cluster %s and instance successfully created.", id)
			return nil
		}
		dbcfg.Tx.Debug("[aurora] cluster status is : %s and instance status is %s, waiting 5 seconds", cluster.Status, instance.Status)
		time.Sleep(5 * time.Second)
	}
	return errors.New("[aurora] timeout waiting for aurora cluster creation")
}

func (dbcfg *AuroraConfig) GetCluster(id string) (Cluster, error) {
	if id == "" {
		return Cluster{}, fmt.Errorf("please provide a non empty cluster ID")
	}
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}

	result, err := dbcfg.RdsSvc.DescribeDBClusters(input)
	if err != nil {
		dbcfg.Tx.Error("Error describing DB cluster: %v", err)
		return Cluster{}, err
	}

	input2 := &rds.DescribeDBClusterEndpointsInput{
		DBClusterIdentifier: aws.String(id),
	}
	endpointRes, err := dbcfg.RdsSvc.DescribeDBClusterEndpoints(input2)
	if err != nil {
		dbcfg.Tx.Error("Error describing DB cluster encpoints: %v", err)
		return Cluster{}, err
	}
	if len(endpointRes.DBClusterEndpoints) == 0 {
		dbcfg.Tx.Error("No cluster encpoint found with ID %v", id)
		return Cluster{}, fmt.Errorf("no cluster found for cluster id: %v", id)
	}
	for _, endpoint := range endpointRes.DBClusterEndpoints {
		if *endpoint.EndpointType == "READER" {
			dbcfg.ReaderEndpoint = *endpoint.Endpoint
		}
		if *endpoint.EndpointType == "WRITER" {
			dbcfg.WriterEndpoint = *endpoint.Endpoint

		}
	}
	return Cluster{ID: id, Status: *result.DBClusters[0].Status, Arn: *result.DBClusters[0].DBClusterArn, ReaderEndpoint: dbcfg.ReaderEndpoint, WriterEndpoint: dbcfg.WriterEndpoint}, nil
}

func (dbcfg *AuroraConfig) GetClusterInstance(instanceID string) (Instance, error) {
	if instanceID == "" {
		return Instance{}, fmt.Errorf("please provide a non empty cluster ID")
	}
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceID),
	}

	result, err := dbcfg.RdsSvc.DescribeDBInstances(input)
	if err != nil {
		dbcfg.Tx.Error("Error describing DB cluster: %v", err)
		return Instance{}, err
	}
	if len(result.DBInstances) == 0 {
		dbcfg.Tx.Error("No instance found with ID %v", instanceID)
		return Instance{}, fmt.Errorf("no instance found with id: %v", instanceID)
	}
	return Instance{ID: instanceID, Status: *result.DBInstances[0].DBInstanceStatus}, nil
}

func (dbcfg *AuroraConfig) Exist(id string) bool {
	_, err := dbcfg.GetCluster(id)
	if err != nil {
		dbcfg.Tx.Error("[exist] Error retrieving aurora cluster %s", err)
		return false
	}
	return true
}

func (dbcfg *AuroraConfig) DeleteTestCluster(id string) error {
	input := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(id),
		SkipFinalSnapshot:   aws.Bool(true),
	}

	_, err := dbcfg.RdsSvc.DeleteDBCluster(input)
	if err != nil {
		dbcfg.Tx.Error("Error deleting DB cluster: %v", err)
		return err
	}
	return nil
}

//////////////////////
// UTILITY FUNCTIONS
////////////////////////

func WrapVal(val string) string {
	return "\"" + val + "\""
}

func WrapField(field string) string {
	return "`" + field + "`"
}
