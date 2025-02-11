// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type InfraConfig struct {
	// DynamoDB
	LcsConfigTableName    string `json:"lcsConfigTableName"`
	LcsConfigTablePk      string `json:"lcsConfigTablePk"`
	LcsConfigTableSk      string `json:"lcsConfigTableSk"`
	ConfigTableName       string `json:"dynamodbConfigTableName"`
	PortalConfigTableName string `json:"dynamodbPortalConfigTableName"`
	MatchTableName        string `json:"dynamoTable"`
	// Glue
	GlueDbName              string `json:"glueDBname"`
	GlueRoleName            string `json:"ucpDataAdminRoleName"`
	GlueJobNameAirBooking   string `json:"customerJobNameairbooking"`
	GlueJobNameHotelBooking string `json:"customerJobNamehotelbooking"`
	GlueJobNamePaxProfile   string `json:"customerJobNamepaxprofile"`
	GlueJobNameGuestProfile string `json:"customerJobNameguestprofile"`
	GlueJobNameHotelStay    string `json:"customerJobNamehotelstay"`
	GlueJobNameClickstream  string `json:"customerJobNameclickstream"`
	GlueJobNameCsi          string `json:"customerJobNamecustomerserviceinteraction"`
	// S3
	BucketAirBooking     string `json:"customerBucketairbooking"`
	BucketHotelBooking   string `json:"customerBuckethotelbooking"`
	BucketPaxProfile     string `json:"customerBucketpaxprofile"`
	BucketGuestProfile   string `json:"customerBucketguestprofile"`
	BucketHotelStay      string `json:"customerBuckethotelstay"`
	BucketClickstream    string `json:"customerBucketclickstream"`
	BucketCsi            string `json:"customerBucketcustomerserviceinteraction"`
	BucketAccpImport     string `json:"connectProfileImportBucketOut"`
	BucketIngestorBackup string `json:"realTimeFeedBucket"`
	BucketProfileOutput  string `json:"outputBucket"`
	BucketMatch          string `json:"matchBucket"`
	// Kinesis
	ChangeProcessorStreamName  string `json:"changeProcessorStreamName"`
	RealTimeIngestorStreamName string `json:"kinesisStreamNameRealTime"`
	RealTimeOutputStreamName   string `json:"kinesisStreamOutputNameRealTime"`
	// SQS
	MergeQueueUrl    string `json:"mergeQueueUrl"`
	CPWriterQueueUrl string `json:"lcsCpWriterQueueUrl"`
	// Lambda
	RealTimePythonLambdaName string `json:"lambdaFunctionNameRealTime"`
	MergerLambdaName         string `json:"mergerLambdaName"`
	// Cognito
	CognitoUserPoolId    string `json:"cognitoUserPoolId"`
	CognitoClientId      string `json:"cognitoClientId"`
	CognitoTokenEndpoint string `json:"tokenEndpoint"`
	// API Gateway
	ApiBaseUrl string `json:"ucpApiUrl"`
	// EventBridge
	EventBridgeBusName string `json:"travellerEventBus"`
	// CloudFormation parameters
	RealTimeBackupEnabled string `json:"realTimeBackupEnabled"`
	// Ssm Namespace
	SsmParamNamespace string `json:"ssmParamNamespace"`
	//cache modes
	CustomerProfileStorageMode string `json:"customerProfileStorageMode"`
	DynamoStorageMode          string `json:"dynamoStorageMode"`
	// CP Export
	CpExportStream string `json:"cpExportStream"`
}

type EnvConfig struct {
	AuroraClusterName string `json:"auroraClusterName"`
	AuroraDbName      string `json:"auroraDbName"`
	AuroraEndpoint    string `json:"auroraEndpoint"`
	AuroraSecretName  string `json:"auroraSecretName"`
	ArtifactBucket    string `json:"artifactBucket"`
	ConnectorBucket   string `json:"connectorBucket"`
	EnvName           string `json:"localEnvName"`
	Region            string `json:"region"`
	AccountID         string `json:"accountId"`
}

func LoadConfigs() (EnvConfig, InfraConfig, error) {
	ec, err := LoadEnvConfig()
	if err != nil {
		return EnvConfig{}, InfraConfig{}, err
	}
	if ec.ArtifactBucket == "" || ec.Region == "" {
		return EnvConfig{}, InfraConfig{}, fmt.Errorf("artifact bucket or region not provided by local config")
	}
	ic, err := LoadInfraConfig(ec.ArtifactBucket, ec.EnvName, ec.Region)
	if err != nil {
		return EnvConfig{}, InfraConfig{}, err
	}

	return ec, ic, nil
}

func LoadEnvConfig() (EnvConfig, error) {
	// Open env.json
	rootDir, err := AbsolutePathToProjectRootDir()
	if err != nil {
		return EnvConfig{}, err
	}
	envFile, err := os.Open(rootDir + "source/env.json")
	if err != nil {
		return EnvConfig{}, err
	}
	defer envFile.Close()

	// Parse into struct
	var env EnvConfig
	decoder := json.NewDecoder(envFile)
	err = decoder.Decode(&env)

	return env, err
}

func LoadInfraConfig(bucketName string, envName string, region string) (InfraConfig, error) {
	infraConfigKey := fmt.Sprintf("config/ucp-config-%s.json", envName)

	s3Client := InitS3Client(bucketName, "", region, "", "")

	var infraConfig InfraConfig
	err := s3Client.Get(infraConfigKey, &infraConfig)

	return infraConfig, err
}

// Get the absolute path to the project's root directory with ending '/' included
//
//	"/path/to/Unified-profile-for-travellers-and-guests-on-aws/"
func AbsolutePathToProjectRootDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	var index int
	if strings.Contains(cwd, "deployment") {
		index = strings.LastIndex(cwd, "deployment")
	} else {
		//using 'strings.Index' fails if the solution in contained in a folder called source we use Last index
		index = strings.LastIndex(cwd, "source")
	}
	return cwd[:index], nil
}

// We use config-utils to test the core S3 package. We also need S3 functionality to load the stack infra config.
// In order to avoid circular dependency issues, we duplicate some of the S3 package here.
// We only use basic functionality for testing purposes, so a bit of duplicate code here is fine.

type S3Config struct {
	Svc       *s3.S3
	Region    string
	Bucket    string
	AccessKey string
	Secret    string
	Token     string
	Path      string
	Tx        core.Transaction
}

func InitS3Client(bucket, path, region, solutionId, solutionVersion string) S3Config {
	cfg := aws.NewConfig().WithRegion(region)
	session, _ := session.NewSession()
	svc := s3.New(session, cfg)
	return S3Config{
		Svc:    svc,
		Bucket: bucket,
		Region: region,
		Path:   path}
}

func (s3c S3Config) Get(key string, object interface{}) error {
	session, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	downloader := s3manager.NewDownloader(session)
	buf := aws.NewWriteAtBuffer([]byte{})
	s3c.Tx.Info("DOWNLOADING FROM S3 bucket %s at key %s", s3c.Bucket, s3c.Path+"/"+key)
	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s3c.Bucket),
			Key:    aws.String(s3c.Path + "/" + key),
		})
	s3c.Tx.Debug("Downloaded %v Bytes", numBytes)
	s3c.Tx.Debug("[S3] Object Downloaded %s", buf.Bytes())
	json.Unmarshal(buf.Bytes(), object)
	if err != nil {
		s3c.Tx.Error("[S3] ERROR while downloading object %+v", err)
	}
	return err
}
