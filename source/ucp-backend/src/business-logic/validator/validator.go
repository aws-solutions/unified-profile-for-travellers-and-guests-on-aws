package validator

import (
	"errors"
	"fmt"
	"strings"
	"tah/core/customerprofiles"
	"tah/core/s3"
	common "tah/ucp/src/business-logic/common"
	model "tah/ucp/src/business-logic/model"
)

var NUMBER_FILES_TO_VALIDATE = 100
var MANDATORY_FIELDS = []string{"model_version", "object_type"}

type Usecase struct {
	Cx *common.Context
}

func (u Usecase) Validate(bucketName string, path string, mappings []customerprofiles.FieldMapping) ([]model.ValidationError, error) {
	u.Cx.Log("1-List all objects on the subpath")
	s3c := s3.Init(bucketName, "", u.Cx.Region)
	objects, err := s3c.Search(path, NUMBER_FILES_TO_VALIDATE)
	allValidationErrors := []model.ValidationError{}
	if err != nil {
		return []model.ValidationError{}, errors.New("Could not list files to validate on S3. Error: " + err.Error())
	}
	for _, object := range objects {
		data, err1 := s3c.ParseCsvFromS3(object)
		if err1 != nil {
			return []model.ValidationError{}, errors.New(fmt.Sprintf("Could not access file %v to validate on S3. Error: %v", object, err1.Error()))
		}
		if len(data) <= 2 {
			return []model.ValidationError{
				model.ValidationError{
					ErrType: model.ERR_TYPE_NO_HEADER,
					File:    object,
					Bucket:  bucketName,
					Row:     0,
					Col:     0,
					ColName: "",
					Msg:     "The CSV file should have at least one header and one row",
				},
			}, nil

		}
		valErrs, err2 := u.ValidateHeaderRow(object, bucketName, data[0], mappings)
		if err2 != nil {
			return []model.ValidationError{}, errors.New(fmt.Sprintf("Error validating header row %v", err2))
		}
		for _, valErr := range valErrs {
			allValidationErrors = append(allValidationErrors, valErr)
		}
		for rowNum, row := range data[1:] {
			valErrs, err3 := u.ValidateDataRow(object, bucketName, rowNum, row, data[0], mappings)
			if err2 != nil {
				return []model.ValidationError{}, errors.New(fmt.Sprintf("Error validating data row %v", err3))
			}
			for _, valErr := range valErrs {
				allValidationErrors = append(allValidationErrors, valErr)
			}
		}
	}
	return allValidationErrors, nil
}

func (u Usecase) ValidateHeaderRow(object string, bucketName string, fielNames []string, mappings []customerprofiles.FieldMapping) ([]model.ValidationError, error) {
	uniqueIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_UNIQUE, mappings)
	orderIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_ORDER, mappings)
	profileIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_PROFILE, mappings)
	cols := map[int]string{}
	fieldNames := map[string]bool{}
	valErrs := []model.ValidationError{}
	for colNum, fieldName := range fielNames {
		fieldNames[fieldName] = true
		cols[colNum] = fieldName
	}
	u.Cx.Log("Checking for source field in ACCP mappings")
	for _, mapping := range mappings {
		sourceFieldName := parseSourceFieldName(mapping.Source)
		if sourceFieldName == "" {
			return valErrs, errors.New("Invalid Mapping provided")
		}
		if !fieldNames[sourceFieldName] {
			valErrs = append(valErrs, model.ValidationError{
				ErrType: model.ERR_TYPE_MISSING_MAPPING_FIELD,
				File:    object,
				Bucket:  bucketName,
				Row:     0,
				Col:     0,
				ColName: sourceFieldName,
				Msg:     fmt.Sprintf("Field %v from ACCP mapping could not be found in CSV file header", sourceFieldName),
			})
		}
	}
	u.Cx.Log("Checking for mandatory fields %v", MANDATORY_FIELDS)
	for _, field := range MANDATORY_FIELDS {
		if !fieldNames[field] {
			valErrs = append(valErrs, model.ValidationError{
				ErrType: model.ERR_TYPE_MISSING_MANDATORY_FIELD,
				File:    object,
				Bucket:  bucketName,
				Row:     0,
				Col:     0,
				ColName: field,
				Msg:     fmt.Sprintf("Mandatory field %v could not be found in CSV file header", field),
			})
		}
	}
	u.Cx.Log("Checking for indexes %v", MANDATORY_FIELDS)
	if uniqueIndex != "" && !fieldNames[uniqueIndex] {
		valErrs = append(valErrs, model.ValidationError{
			ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD,
			File:    object,
			Bucket:  bucketName,
			Row:     0,
			Col:     0,
			ColName: uniqueIndex,
			Msg:     fmt.Sprintf("UNIQUE Index field %v could not be found in CSV file header", uniqueIndex),
		})
	}
	if orderIndex != "" && !fieldNames[orderIndex] {
		valErrs = append(valErrs, model.ValidationError{
			ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD,
			File:    object,
			Bucket:  bucketName,
			Row:     0,
			Col:     0,
			ColName: orderIndex,
			Msg:     fmt.Sprintf("ORDER Index field %v could not be found in CSV file header", orderIndex),
		})
	}
	if profileIndex != "" && !fieldNames[profileIndex] {
		valErrs = append(valErrs, model.ValidationError{
			ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD,
			File:    object,
			Bucket:  bucketName,
			Row:     0,
			Col:     0,
			ColName: profileIndex,
			Msg:     fmt.Sprintf("PROFILE Index field %v could not be found in CSV file header", profileIndex),
		})
	}
	return valErrs, nil
}

func (u Usecase) ValidateDataRow(object string, bucketName string, rowNum int, row []string, fieldNames []string, mappings []customerprofiles.FieldMapping) ([]model.ValidationError, error) {
	uniqueIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_UNIQUE, mappings)
	orderIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_ORDER, mappings)
	profileIndex := fieldNameForIndex(customerprofiles.STANDARD_IDENTIFIER_PROFILE, mappings)
	valErrs := []model.ValidationError{}
	for colNum, fieldValue := range row {
		if uniqueIndex != "" && fieldNames[colNum] == uniqueIndex && fieldValue == "" {
			valErrs = append(valErrs, model.ValidationError{
				ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD_VALUE,
				File:    object,
				Bucket:  bucketName,
				Row:     rowNum,
				Col:     colNum,
				ColName: fieldNames[colNum],
				Msg:     fmt.Sprintf("Indexed field %v is empty (UNIQUE)", fieldNames[colNum]),
			})
		}
		if orderIndex != "" && fieldNames[colNum] == orderIndex && fieldValue == "" {
			valErrs = append(valErrs, model.ValidationError{
				ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD_VALUE,
				File:    object,
				Bucket:  bucketName,
				Row:     rowNum,
				Col:     colNum,
				ColName: fieldNames[colNum],
				Msg:     fmt.Sprintf("Indexed field %v is empty (ORDER)", fieldNames[colNum]),
			})
		}
		if profileIndex != "" && fieldNames[colNum] == profileIndex && fieldValue == "" {
			valErrs = append(valErrs, model.ValidationError{
				ErrType: model.ERR_TYPE_MISSING_INDEX_FIELD_VALUE,
				File:    object,
				Bucket:  bucketName,
				Row:     rowNum,
				Col:     colNum,
				ColName: fieldNames[colNum],
				Msg:     fmt.Sprintf("Indexed field %v is empty (PROFILE)", fieldNames[colNum]),
			})
		}
		for _, field := range MANDATORY_FIELDS {
			if fieldNames[colNum] == field && fieldValue == "" {
				valErrs = append(valErrs, model.ValidationError{
					ErrType: model.ERR_TYPE_MISSING_MANDATORY_FIELD_VALUE,
					File:    object,
					Bucket:  bucketName,
					Row:     0,
					Col:     0,
					ColName: field,
					Msg:     fmt.Sprintf("Mandatory field %v is empty", field),
				})
			}
		}
	}
	return valErrs, nil
}

func fieldNameForIndex(indexName string, mappings []customerprofiles.FieldMapping) string {
	for _, mapping := range mappings {
		for _, index := range mapping.Indexes {
			if index == indexName {
				return parseSourceFieldName(mapping.Source)
			}
		}
	}
	return ""
}

func parseSourceFieldName(source string) string {
	parts := strings.Split(source, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
