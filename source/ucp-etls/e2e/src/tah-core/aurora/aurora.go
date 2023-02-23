package aurora

import (
	"fmt"
	"strconv"
	core "tah/core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
	rdsDs "github.com/aws/aws-sdk-go/service/rdsdataservice"
	_ "github.com/go-sql-driver/mysql"
	gorm "github.com/jinzhu/gorm"
)

var AURORA_MYSQL_DEFAULT_ENGINE string = "innodb"

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

//Config for gorm driver
type GormDBConfig struct {
	Db     *gorm.DB
	engine string
	conErr error            //connection error
	Tx     core.Transaction //service transaction ID

}

//Config for data services
type DataServiceDBConfig struct {
	Db               *rdsDs.RDSDataService
	SecretStorArn    string
	ClusterArn       string
	DbName           string
	CurTransactionId string           //for now we keep only one transaction at the teim for compatibility purpose with GORM implementation
	Tx               core.Transaction //service transaction ID
}

type Column struct {
	Field string `gorm:"column:Field"`
	Type  string `gorm:"column:Type"`
}

//////////////////////
// DATA SERVICES IMPLENTATION
////////////////////////

//Init FUnction For data Services
func InitDs(secretStoreArn string, clusterArn string, dbName string) *DataServiceDBConfig {
	mySession := session.Must(session.NewSession())
	svc := rdsdataservice.New(mySession)
	dbcfg := DataServiceDBConfig{
		Db:            svc,
		SecretStorArn: secretStoreArn,
		ClusterArn:    clusterArn,
		DbName:        dbName,
		Tx:            core.NewTransaction("AURORA", ""),
	}
	dbcfg.Tx.Log("RDS Data Services DB Config: %+v\n", dbcfg)
	return &dbcfg
}

func InitDsWithRegion(secretStoreArn string, clusterArn string, dbName string, region string) *DataServiceDBConfig {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)}))
	svc := rdsdataservice.New(mySession)
	dbcfg := DataServiceDBConfig{
		Db:            svc,
		SecretStorArn: secretStoreArn,
		ClusterArn:    clusterArn,
		DbName:        dbName,
		Tx:            core.NewTransaction("AURORA", ""),
	}
	dbcfg.Tx.Log("RDS Data Services DB Config: %+v\n", dbcfg)
	return &dbcfg
}

func (dbcfg *DataServiceDBConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "AURORA"
	dbcfg.Tx = tx
}

func (dbcfg *DataServiceDBConfig) HasTable(name string) bool {
	res, _ := dbcfg.Query("SHOW tables;")
	dbcfg.Tx.Log("[HasTable] RHas table result %+v\n", res)
	for _, rec := range res {
		if rec["TABLE_NAME"] == name {
			return true
		}
	}
	return false
}

func (dbcfg *DataServiceDBConfig) DropTable(tableName string) error {
	_, err := dbcfg.Query(fmt.Sprintf("DROP TABLE %s;", tableName))
	if err != nil {
		dbcfg.Tx.Log("[DropTable] Error while droping table: %+v\n", err)
	}
	return err
}

func (dbcfg *DataServiceDBConfig) Engine() string {
	return AURORA_MYSQL_DEFAULT_ENGINE
}

func (dbcfg *DataServiceDBConfig) ConErr() error {
	return nil
}

func (dbcfg *DataServiceDBConfig) GetColumns(tableName string) ([]Column, error) {
	res, err := dbcfg.Query(fmt.Sprintf("SHOW COLUMNS FROM %s;", tableName))
	//dbcfg.Tx.Log("[GetColumns] Get Columns result %+v\n", res)
	columns := []Column{}
	for _, rec := range res {
		columns = append(columns, Column{Field: rec["COLUMN_NAME"].(string), Type: rec["COLUMN_TYPE"].(string)})
	}
	return columns, err
}

func (dbcfg *DataServiceDBConfig) StartTransaction() error {
	input := &rdsDs.BeginTransactionInput{
		SecretArn:   aws.String(dbcfg.SecretStorArn),
		ResourceArn: aws.String(dbcfg.ClusterArn),
		Database:    aws.String(dbcfg.DbName),
	}
	out, err := dbcfg.Db.BeginTransaction(input)
	if err != nil {
		dbcfg.Tx.Log("Error while starting transaction %v\n", err)
		return err
	}
	dbcfg.CurTransactionId = *out.TransactionId
	return nil
}
func (dbcfg *DataServiceDBConfig) Commit() error {
	input := &rdsDs.CommitTransactionInput{
		SecretArn:     aws.String(dbcfg.SecretStorArn),
		ResourceArn:   aws.String(dbcfg.ClusterArn),
		TransactionId: aws.String(dbcfg.CurTransactionId),
	}
	_, err := dbcfg.Db.CommitTransaction(input)
	if err != nil {
		dbcfg.Tx.Log("Error while commiting transaction %v\n", err)
		return err
	}
	dbcfg.CurTransactionId = ""
	return nil
}
func (dbcfg *DataServiceDBConfig) Rollback() error {
	input := &rdsDs.RollbackTransactionInput{
		SecretArn:     aws.String(dbcfg.SecretStorArn),
		ResourceArn:   aws.String(dbcfg.ClusterArn),
		TransactionId: aws.String(dbcfg.CurTransactionId),
	}
	_, err := dbcfg.Db.RollbackTransaction(input)
	if err != nil {
		dbcfg.Tx.Log("Error during rollback transaction %v\n", err)
		return err
	}
	dbcfg.CurTransactionId = ""
	return nil
}

func (dbcfg *DataServiceDBConfig) Query(query string) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 0)
	input := &rdsDs.ExecuteStatementInput{
		SecretArn:             aws.String(dbcfg.SecretStorArn),
		ResourceArn:           aws.String(dbcfg.ClusterArn),
		Sql:                   aws.String(query),
		Database:              aws.String(dbcfg.DbName),
		IncludeResultMetadata: aws.Bool(true),
	}
	if dbcfg.CurTransactionId != "" {
		input.TransactionId = aws.String(dbcfg.CurTransactionId)
	}
	dbcfg.Tx.Log("[DataApi] Executing Query %s\n", query)
	output, err := dbcfg.Db.ExecuteStatement(input) // Note: Ignoring errors for brevity
	if err != nil {
		dbcfg.Tx.Log("Error while runing query %s => %+v\n", query, err)
		return results, err
	}
	//dbcfg.Tx.Log("raw response  %+v\n", output)
	for _, rec := range output.Records {
		value := make(map[string]interface{})
		for i, field := range rec {
			var key string
			if output.ColumnMetadata != nil {
				key = *output.ColumnMetadata[i].Name
			} else {
				key = strconv.Itoa(i)
			}

			if field.BooleanValue != nil {
				value[key] = *field.BooleanValue
			} else if field.DoubleValue != nil {
				value[key] = *field.DoubleValue
			} else if field.LongValue != nil {
				value[key] = *field.LongValue
			} else if field.StringValue != nil {
				value[key] = *field.StringValue
			} else if field.IsNull != nil {
				//temp fix to type nullable fields correctly
				if *output.ColumnMetadata[i].TypeName == "int(11)" {
					value[key] = 0
				} else {
					value[key] = ""
				}
			} else {
				fmt.Printf("unsupport data type for value '%+v' now\n", field)
				// TODO remember add other data type
			}
		}
		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		dbcfg.Tx.Log("Fetched row %+v\n", value)
		results = append(results, value)
	}
	return results, nil
}

//////////////////////
// GORM IMPLEMNATTION
////////////////////////

//Init function for mysql driver
func Init(url string, port string, dbName string, user string, passwod string) *GormDBConfig {
	tx := core.NewTransaction("AURORA", "")
	tx.Log("Attenpting connect MYSQL cluster on Aurora to %v:%v with dbName=%v and user  %v \n", url, port, dbName, user)
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, passwod, url, port, dbName)
	tx.Log("Connection string: %v \n", args)
	db, err := gorm.Open("mysql", args)
	if err != nil {
		tx.Log("ERROR: Canct connect to connect MYSQL cluster on Aurora: %+v \n", err)
		return &GormDBConfig{conErr: err}
	}
	dbcfg := GormDBConfig{
		Db:     db,
		engine: AURORA_MYSQL_DEFAULT_ENGINE,
	}
	tx.Log("Successfully connected to MYSQL cluster on Aurora. DB CONfiG: %+v\n", dbcfg)
	return &dbcfg
}

func BuildStartTransactionSql() string {
	return "START TRANSACTION;"
}
func BuildCommitSql() string {
	return "COMMIT;"
}
func BuildRollbackSql() string {
	return "ROLLBACK;"
}
func (dbcfg *GormDBConfig) Engine() string {
	return dbcfg.engine
}
func (dbcfg *GormDBConfig) ConErr() error {
	return dbcfg.conErr
}

func (dbcfg *GormDBConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "AURORA"
	dbcfg.Tx = tx
}

func (dbcfg *GormDBConfig) StartTransaction() error {
	_, err := dbcfg.Query(BuildStartTransactionSql())
	return err
}
func (dbcfg *GormDBConfig) Commit() error {
	_, err := dbcfg.Query(BuildCommitSql())
	return err
}
func (dbcfg *GormDBConfig) Rollback() error {
	_, err := dbcfg.Query(BuildRollbackSql())
	return err
}

func (dbcfg *GormDBConfig) HasTable(name string) bool {
	return dbcfg.Db.HasTable(name)
}

func (dbcfg *GormDBConfig) GetColumns(tableName string) ([]Column, error) {
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

func (dbcfg *GormDBConfig) Query(query string) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0, 0)
	//the query functino only exist on the go sql package not in GORM
	dbcfg.Tx.Log("[GORM] Executing Query %s\n", query)
	rows, err := dbcfg.Db.DB().Query(query) // Note: Ignoring errors for brevity
	if err != nil {
		dbcfg.Tx.Log("Error while runing query %s => %+v\n", query, err)
		return results, err
	}

	columns, _ := rows.Columns()
	length := len(columns)
	//dbcfg.Tx.Log("raw response  %v\n", rows)

	for rows.Next() {
		current := makeResultReceiver(length)
		if err := rows.Scan(current...); err != nil {
			panic(err)
		}
		value := make(map[string]interface{})
		for i := 0; i < length; i++ {
			key := columns[i]
			val := *(current[i]).(*interface{})
			if val == nil {
				value[key] = nil
				continue
			}
			switch val.(type) {
			case int64:
				value[key] = val.(int64)
			case string:
				value[key] = val.(string)
			case time.Time:
				value[key] = val.(time.Time)
			case []uint8:
				value[key] = string(val.([]uint8))
			default:
				fmt.Printf("unsupport data type for value '%+v' now\n", val)
				// TODO remember add other data type
			}
		}
		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		dbcfg.Tx.Log("Fetched row %+v\n", value)
		results = append(results, value)
	}
	return results, nil
}

func makeResultReceiver(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{}
		current = struct{}{}
		result = append(result, &current)
	}
	return result
}

func (dbcfg *GormDBConfig) Close() {
	dbcfg.Db.Close()
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
