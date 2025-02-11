// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package athena

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
)

var ATHENA_MAX_RESULTS = 1000

type Config struct {
	Client         *athena.Athena
	DBName         string
	TableName      string
	Workgroup      string
	OutputLocation string
	Tx             core.Transaction
}

type Option struct {
	NAdditionalPages int //how many pages should be returned
}

func Init(dbName string, tableName string, workgroupName string, region string, outputLocation string, solutionId, solutionVersion string, logLevel core.LogLevel) Config {
	mySession := session.Must(session.NewSession())
	// Create a Athena client from just a session.
	client := core.CreateClient(solutionId, solutionVersion)
	cfg := aws.NewConfig().WithHTTPClient(client).WithRegion(region)
	svc := athena.New(mySession, cfg)
	tx := core.NewTransaction("ATHENA", "", logLevel)
	return Config{
		Client:         svc,
		DBName:         dbName,
		TableName:      tableName,
		Workgroup:      workgroupName,
		OutputLocation: outputLocation,
		Tx:             tx,
	}
}

// Create an Athena Workgroup. If no name is provided, the default workgroup name is used.
func (c Config) CreateWorkGroup(workgroupName string) error {
	if workgroupName == "" {
		workgroupName = c.Workgroup
	}
	if workgroupName == "" {
		c.Tx.Error("Workgroup name is required")
		return errors.New("workgroup name is required")
	}
	c.Tx.Info("Creating Athena Workgroup: %s", c.Workgroup)
	_, err := c.Client.CreateWorkGroup(&athena.CreateWorkGroupInput{
		Name: aws.String(workgroupName),
	})
	if err != nil {
		c.Tx.Error("Error creating athena workgroup: %s", err.Error())
		return err
	}

	return nil
}

// Deletes an Athena workgroup. If no name is provided, the default workgroup name is used.
func (c Config) DeleteWorkGroup(workgroupName string) error {
	if workgroupName == "" {
		workgroupName = c.Workgroup
	}
	if workgroupName == "" {
		c.Tx.Error("Workgroup name is required")
		return errors.New("workgroup name is required")
	}
	c.Tx.Info("Deleting Athena Workgroup: %s", c.Workgroup)
	_, err := c.Client.DeleteWorkGroup(&athena.DeleteWorkGroupInput{
		RecursiveDeleteOption: aws.Bool(true),
		WorkGroup:             aws.String(c.Workgroup),
	})
	if err != nil {
		c.Tx.Error("Error deleting athena workgroup: %s", err.Error())
		return err
	}
	return nil
}

// Create a Prepared Statement for use in Athena
func (c Config) CreatePreparedStatement(queryStatement string, statementName string) error {
	c.Tx.Debug("Validating statement name")
	match, err := regexp.MatchString("^[a-zA-Z_][a-zA-Z0-9_@:]{1,256}$", statementName)
	if err != nil {
		c.Tx.Error("Error validating statement name: %s", err.Error())
		return err
	}
	if !match {
		c.Tx.Error("Invalid statement name: %s\nStatement name must satisfy regular expression pattern: [a-zA-Z_][a-zA-Z0-9_@:]{1,256}", statementName)
		return err
	}
	//constraint:
	c.Tx.Info("Creating Prepared Statement: %s", statementName)
	_, err = c.Client.CreatePreparedStatement(&athena.CreatePreparedStatementInput{
		QueryStatement: aws.String(queryStatement),
		StatementName:  aws.String(statementName),
		WorkGroup:      aws.String(c.Workgroup),
	})
	if err != nil {
		c.Tx.Error("Error creating Prepared Statement: %s", err.Error())
		return err
	}
	return nil
}

// Deletes a specified Prepared Statement
func (c Config) DeletePreparedStatement(statementName string) error {
	c.Tx.Info("Deleting Prepared Statement: %s", statementName)
	_, err := c.Client.DeletePreparedStatement(&athena.DeletePreparedStatementInput{
		StatementName: aws.String(statementName),
		WorkGroup:     aws.String(c.Workgroup),
	})
	if err != nil {
		c.Tx.Error("Error deleting Prepared Statement: %s", err.Error())
		return err
	}

	return nil
}

// Execute a previously created prepared statement
func (c Config) ExecutePreparedStatement(statementName string, executionParameters []*string, option Option) ([]map[string]string, error) {
	//1-Ensure Prepared Statement Exists
	c.Tx.Debug("Checking if Prepared Statement Exists: %s", statementName)
	_, err := c.Client.GetPreparedStatement(&athena.GetPreparedStatementInput{
		StatementName: aws.String(statementName),
		WorkGroup:     aws.String(c.Workgroup),
	})
	if err != nil {
		return []map[string]string{}, err
	}

	//2-Start Athena Query Execution
	c.Tx.Info("Starting Prepared Statement Execution: %s", statementName)
	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(fmt.Sprintf("EXECUTE %s", statementName)),
		WorkGroup:   aws.String(c.Workgroup),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String(c.OutputLocation),
		},
	}
	if len(executionParameters) > 0 {
		input.ExecutionParameters = executionParameters
	}
	out, err := c.Client.StartQueryExecution(input)
	if err != nil {
		return []map[string]string{}, err
	}

	//3-Wait for query completion and return results
	c.Tx.Info("Waiting for query completion and tabulating results")
	return c.parseQueryExecutionOutput(out, option)
}

// Run a specified Query on Athena.
func (c Config) Query(sql string, option Option) ([]map[string]string, error) {
	//1-Start Athena Query Execution
	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(sql),
		WorkGroup:   aws.String(c.Workgroup),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String(c.OutputLocation),
		},
	}
	out, err := c.Client.StartQueryExecution(input)
	if err != nil {
		return []map[string]string{}, err
	}

	//2-Wait for query completion and return results
	return c.parseQueryExecutionOutput(out, option)
}

// Drops and Re-Adds all existing Partitions on the configured Athena database and table.
func (c Config) SyncPartitions() error {
	partitionsResult, err := c.ShowPartitions()
	if err != nil {
		return err
	}

	partitions := []string{}
	for _, row := range partitionsResult {
		partition := row["partition"]
		partitionParts := strings.Split(partition, "=")
		if len(partitionParts) == 2 {
			partition = fmt.Sprintf(`%s="%s"`, partitionParts[0], strings.Trim(strings.TrimSpace(partitionParts[1]), `"`))
			partitions = append(partitions, partition)
		}
	}

	err = c.DropPartitions(partitions)
	if err != nil {
		return err
	}

	err = c.AddPartitions(partitions)
	if err != nil {
		return err
	}

	return nil
}

// Enumerates all the existing Partitions on the configured Athena database and table
func (c Config) ShowPartitions() ([]map[string]string, error) {
	showPartitionsQuery := fmt.Sprintf("SHOW PARTITIONS `%s.%s`", c.DBName, c.TableName)
	partitionsResult, err := c.Query(showPartitionsQuery, Option{})
	if err != nil {
		return nil, err
	}

	return partitionsResult, nil
}

// Drops specified Partitions, if they exist, on the configured Athena database and table
func (c Config) DropPartitions(partitions []string) error {
	//	partitions format: "partitionName=\"value\""
	dropPartitionQuery := fmt.Sprintf("ALTER TABLE `%s.%s` DROP IF EXISTS ", c.DBName, c.TableName)
	for _, partition := range partitions {
		dropPartitionQuery = dropPartitionQuery + fmt.Sprintf("PARTITION (%s), ", partition)
	}
	dropPartitionQuery = strings.TrimSuffix(dropPartitionQuery, ", ")
	_, err := c.Query(dropPartitionQuery, Option{})
	if err != nil {
		return err
	}

	return nil
}

// Adds the specified partitions, if they don't already exist, on the configured Athena database and table
func (c Config) AddPartitions(partitions []string) error {
	//	partitions format: "partitionName=\"value\""
	addPartitionQuery := fmt.Sprintf("ALTER TABLE `%s.%s` ADD IF NOT EXISTS ", c.DBName, c.TableName)
	for _, partition := range partitions {
		addPartitionQuery = addPartitionQuery + fmt.Sprintf("PARTITION (%s), ", partition)
	}
	addPartitionQuery = strings.TrimSuffix(addPartitionQuery, ", ")
	_, err := c.Query(addPartitionQuery, Option{})
	if err != nil {
		return err
	}

	return nil
}

func (c Config) parseQueryExecutionOutput(out *athena.StartQueryExecutionOutput, option Option) ([]map[string]string, error) {
	//1-Wait for query completion
	//TODO: set timeout
	waitMultiplier := 1
	queryInProgress := true
	for queryInProgress {
		statusInput := &athena.GetQueryExecutionInput{
			QueryExecutionId: out.QueryExecutionId,
		}
		statusOut, err := c.Client.GetQueryExecution(statusInput)
		if err != nil {
			return []map[string]string{}, err
		}
		if *statusOut.QueryExecution.Status.State == "FAILED" {
			return []map[string]string{}, errors.New("Query executions failed witth error: " + *statusOut.QueryExecution.Status.StateChangeReason)
		}
		if *statusOut.QueryExecution.Status.State == "CANCELLED" {
			return []map[string]string{}, errors.New("Query execution was canceleld: " + *statusOut.QueryExecution.Status.StateChangeReason)
		}
		if *statusOut.QueryExecution.Status.State == "SUCCEEDED" {
			queryInProgress = false
		}
		time.Sleep(time.Duration(waitMultiplier) * time.Second)
		waitMultiplier = waitMultiplier + 1
	}

	//2-If query succeeded, get query result
	resultInput := &athena.GetQueryResultsInput{
		QueryExecutionId: out.QueryExecutionId,
	}
	resOut, err := c.Client.GetQueryResults(resultInput)
	if err != nil {
		return []map[string]string{}, err
	}
	allRows := resOut.ResultSet.Rows
	nPage := 0
	//if has more and we request more  additional pages (option.NPages > 0)
	for resOut.NextToken != nil && option.NAdditionalPages > nPage {
		resultInput.NextToken = resOut.NextToken
		resOut, err = c.Client.GetQueryResults(resultInput)
		if err != nil {
			return []map[string]string{}, err
		}
		allRows = append(allRows, resOut.ResultSet.Rows...)
		nPage = nPage + 1
	}
	c.Tx.Debug("Columns: %+v", resOut.ResultSet.ResultSetMetadata.ColumnInfo)
	res := []map[string]string{}
	if len(resOut.ResultSet.ResultSetMetadata.ColumnInfo) == 0 {
		c.Tx.Warn("No column return in result")
		return res, nil
	}
	for _, row := range allRows {
		item := map[string]string{}
		for i, datum := range row.Data {
			if datum.VarCharValue != nil {
				item[*resOut.ResultSet.ResultSetMetadata.ColumnInfo[i].Name] = *datum.VarCharValue
			}
		}
		res = append(res, item)
	}
	return res, nil
}
