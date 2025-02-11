// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"

	"time"
)

var ACCP_OBJECTS_STRUCT_NAMES = []string{"AirBooking", "HotelBooking", "AirLoyalty", "HotelLoyalty", "EmailHistory", "PhoneHistory", "Clickstream", "HotelStay", "CustomerServiceInteraction", "LoyaltyTx", "AncillaryService"}
var ACCP_OBJECTS_STRUCTS = []model.AccpObject{model.AirBooking{}, model.HotelBooking{}, model.AirLoyalty{}, model.HotelLoyalty{}, model.EmailHistory{}, model.PhoneHistory{}, model.Clickstream{}, model.HotelStay{}, model.CustomerServiceInteraction{}, model.LoyaltyTx{}, model.AncillaryService{}}
var OPERATION_CREATE = "create"
var OPERATION_UPDATE = "update"
var OPERATION_DELETE = "delete"
var PROFILE_CONNECT_ID_FIELD_NAME = "profile_connect_id"

type operationFunc func(domain string, t model.Traveller) string
type operationObjectFunc func(domain string, profileConnectID string, data model.AccpObject) string

var operationMap = map[string]operationFunc{
	OPERATION_CREATE: GenerateProfileInsertStatement,
	OPERATION_UPDATE: GenerateProfileUpdateStatement,
	OPERATION_DELETE: GenerateProfileDeleteStatement,
}

var operationObjectMap = map[string]operationObjectFunc{
	OPERATION_CREATE: GenerateProfileObjectInsertStatement,
	OPERATION_UPDATE: GenerateProfileObjectUpdateStatement,
	OPERATION_DELETE: GenerateProfileObjectDeleteStatement,
}

func GenerateCreateTableStatements(domain string) []string {
	statements := []string{}
	statements = append(statements, GenerateProfileCreateTableStatement(domain))
	for _, obj := range ACCP_OBJECTS_STRUCTS {
		statements = append(statements, GenerateProfileObjectCreateTableStatement(domain, obj))
	}
	return statements
}

func GenerateDeleteTableStatements(domain string) []string {
	statements := []string{}
	statements = append(statements, GenerateProfileDeleteTableStatement(domain))
	for _, obj := range ACCP_OBJECTS_STRUCTS {
		statements = append(statements, GenerateProfileObjectDeleteTableStatement(domain, obj))
	}
	return statements
}

func TravellerToRedshift(domain string, t model.Traveller, withProfile bool, operation string) []string {
	statements := []string{}
	interfaceType := reflect.TypeOf((*model.AccpObject)(nil)).Elem()

	//connect ID can be nul in case the traveller struct is used to passe around an ACCP Object (see change processor)
	if withProfile {
		if operationFunc, found := operationMap[operation]; found {
			// Call the appropriate function and append the result to statements
			statements = append(statements, operationFunc(domain, t))
		} else {
			log.Printf("Unknown operation: %v", operation)
		}
	}
	reflectedValue := reflect.ValueOf(t)
	for i := 0; i < reflectedValue.NumField(); i++ {
		fieldValue := reflectedValue.Field(i)
		//log.Printf("Field type: %v", field.Type.Name())
		statements = addStatementsReflect(domain, t, statements, fieldValue, operation, interfaceType)
	}
	return statements
}

func addStatementsReflect(domain string, t model.Traveller, statements []string, fieldValue reflect.Value, operation string, interfaceType reflect.Type) []string {
	if fieldValue.Kind() == reflect.Slice {
		for j := 0; j < fieldValue.Len(); j++ {
			element := fieldValue.Index(j)

			if reflect.TypeOf(element.Interface()).Implements(interfaceType) {
				if operationObjectFunc, found := operationObjectMap[operation]; found {
					// Call the appropriate function and append the result to statements
					statements = append(statements, operationObjectFunc(domain, t.ConnectID, element.Interface().(model.AccpObject)))
				} else {
					log.Printf("Unknown operation: %v", operation)
				}
			}
		}
	}
	return statements
}

func ProfileObjectToFields(profileConnectID string, data model.AccpObject) (string, []string, []string, []string) {
	reflectedValue := reflect.ValueOf(data)
	reflectedType := reflectedValue.Type()
	//log.Printf("Generating field names for Profile Object (type=%v)", reflectedType.Name())

	objectTypeName := utils.ToSnakeCase(reflectedType.Name())

	var fieldNames []string
	var fieldValues []string
	var fieldTypes []string

	for i := 0; i < reflectedValue.NumField(); i++ {
		field := reflectedType.Field(i)
		fieldValue := reflectedValue.Field(i)
		fieldNames = append(fieldNames, utils.ToSnakeCase(field.Name))
		fieldValues = append(fieldValues, formatField(fieldValue.Interface(), field.Type))
		fieldTypes = append(fieldTypes, getRedshiftColumnType(field.Type))
	}
	//adding connect_id field to allow joins
	fieldNames = append(fieldNames, PROFILE_CONNECT_ID_FIELD_NAME)
	fieldValues = append(fieldValues, profileConnectID)
	fieldTypes = append(fieldTypes, "VARCHAR(255)")
	return objectTypeName, fieldNames, fieldValues, fieldTypes
}

func ProfileToFields(data interface{}) ([]string, []string, []string) {
	reflectedValue := reflect.ValueOf(data)
	reflectedType := reflectedValue.Type()

	//log.Printf("Generating field names for Profile (type=%v)", reflectedType.Name())

	var fieldNames []string
	var fieldValues []string
	var fieldTypes []string

	for i := 0; i < reflectedValue.NumField(); i++ {
		field := reflectedType.Field(i)
		fieldValue := reflectedValue.Field(i)
		//log.Printf("Field type: %v", field.Type.Name())
		if field.Type.Name() == "string" || field.Type.Name() == "time.Time" {
			colName := utils.ToSnakeCase(field.Name)
			//log.Printf("Col name: %v", colName)
			fieldNames = append(fieldNames, colName)
			fieldValues = append(fieldValues, formatField(fieldValue.Interface(), field.Type))
			fieldTypes = append(fieldTypes, getRedshiftColumnType(field.Type))
		} else if field.Type.Name() == "Address" {
			//log.Printf("Generating address column names recursively for address: %v", field.Name)
			addrFields, addrValues, addrTypes := ProfileToFields(fieldValue.Interface())
			addrColName := utils.ToSnakeCase(field.Name)
			for i, f := range addrFields {
				fieldNames = append(fieldNames, addrColName+"_"+f)
				fieldValues = append(fieldValues, addrValues[i])
				fieldTypes = append(fieldTypes, addrTypes[i])

			}
		}
	}

	return fieldNames, fieldValues, fieldTypes
}

func Prefix(prefix string, fieldNames []string) []string {
	withPrefix := []string{}
	for _, f := range fieldNames {
		withPrefix = append(withPrefix, prefix+f)
	}
	return withPrefix
}

func GenerateProfileObjectInsertStatement(domainName string, profileConnectID string, data model.AccpObject) string {
	tableName := domainToTable(domainName)
	objectTypeName, fields, values, _ := ProfileObjectToFields(profileConnectID, data)
	return GenerateInsertStatement(tableName+"_"+objectTypeName, fields, values)
}
func GenerateProfileInsertStatement(domainName string, traveller model.Traveller) string {
	tableName := domainToTable(domainName)
	fields, values, _ := ProfileToFields(traveller)
	return GenerateInsertStatement(tableName, fields, values)
}

func GenerateProfileObjectUpdateStatement(domainName string, profileConnectID string, data model.AccpObject) string {
	tableName := domainToTable(domainName)
	objectTypeName, fields, values, _ := ProfileObjectToFields(profileConnectID, data)
	return GenerateUpdateStatement(tableName+"_"+objectTypeName, fields, values, "accp_object_id = '"+data.ID()+"' AND "+PROFILE_CONNECT_ID_FIELD_NAME+" = '"+profileConnectID+"'")
}
func GenerateProfileUpdateStatement(domainName string, traveller model.Traveller) string {
	tableName := domainToTable(domainName)
	fields, values, _ := ProfileToFields(traveller)
	return GenerateUpdateStatement(tableName, fields, values, "connect_id = '"+traveller.ConnectID+"'")
}

func GenerateProfileObjectDeleteStatement(domainName string, profileConnectID string, data model.AccpObject) string {
	tableName := domainToTable(domainName)
	objectTypeName, _, _, _ := ProfileObjectToFields(profileConnectID, data)
	return GenerateDeleteStatement(tableName+"_"+objectTypeName, "accp_object_id = '"+data.ID()+"' AND "+PROFILE_CONNECT_ID_FIELD_NAME+" = '"+profileConnectID+"'")
}
func GenerateProfileDeleteStatement(domainName string, traveller model.Traveller) string {
	tableName := domainToTable(domainName)
	return GenerateDeleteStatement(tableName, "connect_id = '"+traveller.ConnectID+"'")
}

func GenerateUpdateStatement(tableName string, fields []string, values []string, condition string) string {
	setClauses := make([]string, len(fields))
	fields = formatFieldNames(fields)
	values = formatFieldValues(values)
	for i := 0; i < len(fields); i++ {
		setClauses[i] = fmt.Sprintf("%s = %s", fields[i], values[i])
	}

	updateStmt := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClauses, ", "))

	if condition != "" {
		updateStmt += fmt.Sprintf(" WHERE %s", condition)
	}

	return updateStmt
}

func GenerateDeleteStatement(tableName string, condition string) string {
	stm := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, condition)
	return stm
}

func GenerateInsertStatement(tableName string, fieldNames, fieldValues []string) string {
	fieldNames = formatFieldNames(fieldNames)
	fieldValues = formatFieldValues(fieldValues)
	insertStatement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);",
		tableName,
		strings.Join(fieldNames, ", "),
		strings.Join(fieldValues, ", "))

	return insertStatement
}

func GenerateProfileObjectCreateTableStatement(domainName string, data model.AccpObject) string {
	tableName := domainToTable(domainName)
	objectTypeName, fieldNames, _, fieldTypes := ProfileObjectToFields("", data)
	return GenerateCreateTableStatement(tableName+"_"+objectTypeName, fieldNames, fieldTypes)
}
func GenerateProfileCreateTableStatement(domainName string) string {
	tableName := domainToTable(domainName)
	fieldNames, _, fieldTypes := ProfileToFields(model.Traveller{})
	return GenerateCreateTableStatement(tableName, fieldNames, fieldTypes)
}

func GenerateProfileObjectDeleteTableStatement(domainName string, data interface{}) string {
	tableName := domainToTable(domainName)
	reflectedValue := reflect.ValueOf(data)
	reflectedType := reflectedValue.Type()
	objectTypeName := utils.ToSnakeCase(reflectedType.Name())
	return GenerateDeleteTableStatement(tableName + "_" + objectTypeName)
}
func GenerateProfileDeleteTableStatement(domainName string) string {
	tableName := domainToTable(domainName)
	return GenerateDeleteTableStatement(tableName)
}

func GenerateDeleteTableStatement(tableName string) string {
	return fmt.Sprintf("DROP TABLE %s;", tableName)
}

func domainToTable(domainName string) string {
	return utils.ToSnakeCase(domainName)
}

func GenerateCreateTableStatement(tableName string, fieldNames, fieldTypes []string) string {
	query := fmt.Sprintf("CREATE TABLE %s (", tableName)
	fieldNames = formatFieldNames(fieldNames)
	for i, fieldName := range fieldNames {
		query += fmt.Sprintf("%s %s,", fieldName, fieldTypes[i])
	}
	query = query[:len(query)-1] + ")"
	return query
}

func formatField(fieldVal interface{}, field reflect.Type) string {
	if field == reflect.TypeOf(time.Time{}) {
		//we format date at a redshift compatible format
		return fmt.Sprintf("%v", fieldVal.(time.Time).Format("2006-01-02 15:04:05"))
	} else {
		return fmt.Sprintf("%v", fieldVal)
	}
}

func formatFieldValues(values []string) []string {
	res := []string{}
	for _, fn := range values {
		fn = strings.Replace(fn, "'", "\\'", -1)
		res = append(res, "'"+fn+"'")
	}
	return res
}

func formatFieldNames(fNames []string) []string {
	res := []string{}
	for _, fn := range fNames {
		res = append(res, "\""+fn+"\"")
	}
	return res
}

func getRedshiftColumnType(field reflect.Type) string {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "DECIMAL(10,2)"
	case reflect.String:
		return "VARCHAR(255)"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Struct:
		if field == reflect.TypeOf(time.Time{}) {
			return "TIMESTAMP"
		}
	default:
		log.Fatalf("Unsupported field type: %v", field)
		return ""
	}
	return ""
}
