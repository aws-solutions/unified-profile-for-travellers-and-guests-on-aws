package athena

import (
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
)

var ATHENA_MAX_RESULTS = 1000

type Config struct {
	Client    *athena.Athena
	DBName    string
	TableName string
	Workgroup string
}

type Option struct {
	NAdditionalPages int //how many pages should be returned
}

func Init(dbName string, tableName string, workgroupName string) Config {
	mySession := session.Must(session.NewSession())
	// Create a Athena client from just a session.
	svc := athena.New(mySession)
	return Config{
		Client:    svc,
		DBName:    dbName,
		TableName: tableName,
		Workgroup: workgroupName,
	}
}

func (c Config) Query(sql string, option Option) ([]map[string]string, error) {
	//1-Start Athena Query Execution
	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(sql),
		WorkGroup:   aws.String(c.Workgroup),
	}
	out, err := c.Client.StartQueryExecution(input)
	if err != nil {
		return []map[string]string{}, err
	}

	//2-Wait for query completion
	//TODO: set timeout
	queryInProgress := true
	for queryInProgress {
		statusInput := &athena.GetQueryExecutionInput{
			QueryExecutionId: out.QueryExecutionId,
		}
		statusOut, err2 := c.Client.GetQueryExecution(statusInput)
		if err2 != nil {
			return []map[string]string{}, err2
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
		time.Sleep(1 * time.Second)
	}

	//2-If query succeeded, get query result
	resultInput := &athena.GetQueryResultsInput{
		QueryExecutionId: out.QueryExecutionId,
	}
	resOut, err3 := c.Client.GetQueryResults(resultInput)
	if err3 != nil {
		return []map[string]string{}, err3
	}
	allRows := resOut.ResultSet.Rows
	nPage := 0
	//if has more and we request more  additional pages (option.NPages > 0)
	for resOut.NextToken != nil && option.NAdditionalPages > nPage {
		resultInput.NextToken = resOut.NextToken
		resOut, err3 = c.Client.GetQueryResults(resultInput)
		if err3 != nil {
			return []map[string]string{}, err3
		}
		for _, row := range resOut.ResultSet.Rows {
			allRows = append(allRows, row)
		}
		nPage = nPage + 1
	}
	log.Printf("[ATHENA] Columns: %+v", resOut.ResultSet.ResultSetMetadata.ColumnInfo)
	res := []map[string]string{}
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
