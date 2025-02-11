// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"tah/upt/source/tah-core/cognito"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	travModel "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func Filter(traveller travModel.Traveller, permissionString string) *travModel.Traveller {
	/*
		Filter examples:
		all data => * / *
		all traveller => traveller/* => traveller/-dob
		all air_booking data => air_booking/*
		all email_history data => email_history/*
		air_booking only returning departureAirport and arrivalAirport => air_booking/departureAirport,arrivalAirport
		air_booking all fields where arrivalAirport=RDU => air_booking/*?arrivalAirport = RDU
		air_booking departure and arrival when arriving at RDU => air_booking/departureAirport,arrivalAirport?arrivalAirport = RDU

		Assumptions
		- admin permission MUST NOT include any other permissions, will result in error
		- by default user has zero permission
		- any biz object not included in permission statement will be excluded
		- you can either select to include OR exclude fields (include by default, include * to exclude listed fields)
		- fields to query on support operators =, >, <, AND, OR, IN
			- comparison operators =, >, >=, <, <=
			- logical operators AND, OR, IN
		- filter criteria is case sensitive and must be split by a space " "
			- example description => air_booking/*?arrivalAirport = RDU
			- criteria from description => arrivalAirport = RDU
			- field name is based on the object's json tag (`json:"arrivalAirport"` => arrivalAirport)
			- RDU is case sensitive and must be split by a space " ", if it cannot be parsed then it will return false and
			  the user will not have access to the object (zero permission default)
	*/

	isAdminUser, filters := parsePermission(permissionString)

	if isAdminUser {
		return &traveller
	}

	// No valid permissions, user has 0 access
	if len(filters) == 0 {
		return nil
	}

	// Create new traveller that will only have the fields that are allowed
	filteredTraveller := travModel.Traveller{}

	filteredTraveller, err := addFieldsToTraveller(traveller, filteredTraveller, filters)
	if err != nil {
		return nil
	}

	// Ensure required fields are present
	filteredTraveller.ModelVersion = traveller.ModelVersion
	filteredTraveller.Errors = traveller.Errors
	filteredTraveller.ParsingErrors = traveller.ParsingErrors

	// Filter each ACCP record type individually
	filteredTraveller = addAirBooking(traveller, filteredTraveller, filters)
	filteredTraveller = addAirLoyalty(traveller, filteredTraveller, filters)
	filteredTraveller = addClickstream(traveller, filteredTraveller, filters)
	filteredTraveller = addEmailHistory(traveller, filteredTraveller, filters)
	filteredTraveller = addHotelBooking(traveller, filteredTraveller, filters)
	filteredTraveller = addHotelLoyalty(traveller, filteredTraveller, filters)
	filteredTraveller = addHotelStay(traveller, filteredTraveller, filters)
	filteredTraveller = addPhoneHistory(traveller, filteredTraveller, filters)
	filteredTraveller = addCustomerServiceInteraction(traveller, filteredTraveller, filters)
	filteredTraveller = addLoyaltyTransaction(traveller, filteredTraveller, filters)
	filteredTraveller = addAncillaries(traveller, filteredTraveller, filters)

	return &filteredTraveller
}

func addFieldsToTraveller(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) (travModel.Traveller, error) {
	// Add fields to Traveller object
	filter, exists := filters["traveller"]
	if exists {
		shouldIncludeObject := shouldIncludeObject(filter.Condition, setContext(traveller))
		if shouldIncludeObject {
			original := reflect.ValueOf(traveller)
			for i := 0; i < reflect.TypeOf(traveller).NumField(); i++ {
				fieldName := original.Type().Field(i).Name
				fieldNameJson := original.Type().Field(i).Tag.Get("json")
				value := original.FieldByName(fieldName)
				shouldIncludeField := shouldIncludeField(fieldNameJson, filter.Fields, filter.ExcludeFields)
				if shouldIncludeField {
					reflect.ValueOf(&filteredTraveller).Elem().FieldByName(fieldName).Set(value)
				}
			}
		} else {
			return travModel.Traveller{}, errors.New("traveller object not included")
		}
	}
	return filteredTraveller, nil
}

func addAirBooking(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["air_booking"]
	// We might have added object records as part of traveller fields, need to clear it out
	filteredTraveller.AirBookingRecords = []travModel.AirBooking{}
	if exists {
		for _, rec := range traveller.AirBookingRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.AirBookingRecords = append(filteredTraveller.AirBookingRecords, *filtered.(*travModel.AirBooking))
			}
		}
	}
	return filteredTraveller
}

func addAirLoyalty(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["air_loyalty"]
	filteredTraveller.AirLoyaltyRecords = []travModel.AirLoyalty{}
	if exists {
		for _, rec := range traveller.AirLoyaltyRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.AirLoyaltyRecords = append(filteredTraveller.AirLoyaltyRecords, *filtered.(*travModel.AirLoyalty))
			}
		}
	}
	return filteredTraveller
}

func addClickstream(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["clickstream"]
	filteredTraveller.ClickstreamRecords = []travModel.Clickstream{}
	if exists {
		for _, rec := range traveller.ClickstreamRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.ClickstreamRecords = append(filteredTraveller.ClickstreamRecords, *filtered.(*travModel.Clickstream))
			}
		}
	}
	return filteredTraveller
}

func addEmailHistory(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["email_history"]
	filteredTraveller.EmailHistoryRecords = []travModel.EmailHistory{}
	if exists {
		for _, rec := range traveller.EmailHistoryRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.EmailHistoryRecords = append(filteredTraveller.EmailHistoryRecords, *filtered.(*travModel.EmailHistory))
			}
		}
	}
	return filteredTraveller
}

func addHotelBooking(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["hotel_booking"]
	filteredTraveller.HotelBookingRecords = []travModel.HotelBooking{}
	if exists {
		for _, rec := range traveller.HotelBookingRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.HotelBookingRecords = append(filteredTraveller.HotelBookingRecords, *filtered.(*travModel.HotelBooking))
			}
		}
	}
	return filteredTraveller
}

func addHotelLoyalty(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["hotel_loyalty"]
	filteredTraveller.HotelLoyaltyRecords = []travModel.HotelLoyalty{}
	if exists {
		for _, rec := range traveller.HotelLoyaltyRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.HotelLoyaltyRecords = append(filteredTraveller.HotelLoyaltyRecords, *filtered.(*travModel.HotelLoyalty))
			}
		}
	}
	return filteredTraveller
}

func addHotelStay(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["hotel_stay"]
	filteredTraveller.HotelStayRecords = []travModel.HotelStay{}
	if exists {
		for _, rec := range traveller.HotelStayRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.HotelStayRecords = append(filteredTraveller.HotelStayRecords, *filtered.(*travModel.HotelStay))
			}
		}
	}
	return filteredTraveller
}

func addPhoneHistory(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["phone_history"]
	filteredTraveller.PhoneHistoryRecords = []travModel.PhoneHistory{}
	if exists {
		for _, rec := range traveller.PhoneHistoryRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.PhoneHistoryRecords = append(filteredTraveller.PhoneHistoryRecords, *filtered.(*travModel.PhoneHistory))
			}
		}
	}
	return filteredTraveller
}

func addCustomerServiceInteraction(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["customer_service_interaction"]
	filteredTraveller.CustomerServiceInteractionRecords = []travModel.CustomerServiceInteraction{}
	if exists {
		for _, rec := range traveller.CustomerServiceInteractionRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.CustomerServiceInteractionRecords = append(filteredTraveller.CustomerServiceInteractionRecords, *filtered.(*travModel.CustomerServiceInteraction))
			}
		}
	}
	return filteredTraveller
}

func addLoyaltyTransaction(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["loyalty_transaction"]
	filteredTraveller.LoyaltyTxRecords = []travModel.LoyaltyTx{}
	if exists {
		for _, rec := range traveller.LoyaltyTxRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.LoyaltyTxRecords = append(filteredTraveller.LoyaltyTxRecords, *filtered.(*travModel.LoyaltyTx))
			}
		}
	}
	return filteredTraveller
}

func addAncillaries(traveller, filteredTraveller travModel.Traveller, filters map[string]travModel.Filter) travModel.Traveller {
	filter, exists := filters["loyalty_transaction"]
	filteredTraveller.AncillaryServiceRecords = []travModel.AncillaryService{}
	if exists {
		for _, rec := range traveller.AncillaryServiceRecords {
			filtered := filterRecord(filter, rec)
			if filtered != nil {
				filteredTraveller.AncillaryServiceRecords = append(filteredTraveller.AncillaryServiceRecords, *filtered.(*travModel.AncillaryService))
			}
		}
	}
	return filteredTraveller
}

func filterRecord(filter travModel.Filter, rec interface{}) interface{} {
	shouldIncludeObject := shouldIncludeObject(filter.Condition, setContext(rec))
	if shouldIncludeObject {
		originalRec := reflect.ValueOf(rec)
		newRec := reflect.New(originalRec.Type()).Interface()
		count := reflect.TypeOf(rec).NumField()
		for i := 0; i < count; i++ {
			fieldName := originalRec.Type().Field(i).Name
			fieldNameJson := originalRec.Type().Field(i).Tag.Get("json")
			value := originalRec.FieldByName(fieldName)
			shouldIncludeField := shouldIncludeField(fieldNameJson, filter.Fields, filter.ExcludeFields)
			if shouldIncludeField {
				reflect.ValueOf(newRec).Elem().FieldByName(fieldName).Set(value)
			} else {
				reflect.ValueOf(newRec).Elem().FieldByName(fieldName).Set(reflect.Zero(value.Type()))
			}
		}
		return newRec
	} else {
		return nil
	}
}

func parsePermission(permissionString string) (bool, map[string]travModel.Filter) {
	filters := make(map[string]travModel.Filter)
	bizObjects := make(map[string]bool)

	permissions := strings.Split(permissionString, "\n")
	if len(permissions) == 0 {
		return false, nil
	}

	// Check for admin user
	if permissions[0] == "*/*" {
		if len(permissions) == 1 {
			// Valid admin user
			return true, nil
		} else {
			// Invalid admin user, other permissions present
			return false, nil
		}
	}

	for _, p := range permissions {
		boo, filter, obj := nonAdminUser(bizObjects, p)
		if !boo {
			// Invalid permission, treat as user with zero permission
			return false, nil
		}
		filters[obj] = filter
	}

	return false, filters
}

func nonAdminUser(bizObjects map[string]bool, p string) (bool, travModel.Filter, string) {
	parts := strings.Split(p, "/")
	if len(parts) != 2 {
		// There is an error with permission, treat as user with zero permission
		return false, travModel.Filter{}, ""
	}

	obj, fieldData := parts[0], parts[1]
	// Check that we do not already have a permission string for this object
	// A wildcard operator is also invalid as an object type
	if bizObjects[obj] || obj == "*" {
		return false, travModel.Filter{}, obj
	} else {
		bizObjects[obj] = true
	}

	// Create filter
	parts = strings.Split(fieldData, "?")
	if len(parts) == 0 || len(parts) > 2 {
		return false, travModel.Filter{}, obj
	}
	fieldsString := parts[0]
	fields := strings.Split(fieldsString, ",")
	// "*" indicates access to all fields (vs default access to zero fields),
	// so presence of * means include all and EXCLUDE any fields present in field list
	excludeFields := strings.HasPrefix(fieldsString, "*")
	condition := ""
	if len(parts) == 2 {
		condition = parts[1]
	}

	filter := travModel.Filter{
		ObjectName:    obj,
		Fields:        fields,
		ExcludeFields: excludeFields,
		Condition:     condition,
	}
	return true, filter, obj
}

func setContext(record interface{}) map[string]interface{} {
	context := make(map[string]interface{})
	t := reflect.ValueOf(record)

	for i := 0; i < t.NumField(); i++ {
		field := t.Type().Field(i)
		fieldJson := field.Tag.Get("json")
		fieldValue := t.Field(i).Interface()
		context[string(fieldJson)] = fieldValue
	}

	return context
}

// Helper function to evaluate a single operation
func evaluateOperation(leftKey, right, op string, context map[string]interface{}) bool {
	left := context[leftKey]
	switch op {
	case "and":
		return shouldIncludeObject(leftKey, context) && shouldIncludeObject(right, context)
	case "or":
		return shouldIncludeObject(leftKey, context) || shouldIncludeObject(right, context)
	default:
		isMatch, err := compareByType(left, right, op)
		if err != nil {
			log.Printf("[evaluateOperation] Error: %v", err)
			return false
		}
		return isMatch
	}
}

// Permission string can include * that indicates all fields should be included, except for named fields.
// Otherwise, named fields are the only fields to include.
//
// Include named fields (to, from) => clickstream/to,from
// Exclude named fields (birthDate) => clickstream/*,birthDate
func shouldIncludeField(fieldName string, fields []string, excludeNamedFields bool) bool {
	if excludeNamedFields {
		// field name should not be in list of excluded fields
		return !utils.ContainsString(fields, fieldName)
	} else {
		// field name should be in list of included fields
		return utils.ContainsString(fields, fieldName)
	}
}

func compareByType(a interface{}, b, op string) (bool, error) {
	expectedType := reflect.TypeOf(a)
	t := time.Now()
	timeType := reflect.TypeOf(t)

	log.Printf("Comparing %v and %v as %v", a, b, expectedType)

	if op == "in" {
		boo, err := opIn(a, b)
		return boo, err
	}

	// We need to handle types for each object type that is included in traveller profile
	switch expectedType.Kind() {
	case reflect.String:
		switch op {
		case "=", "==":
			return a == b, nil
		case "!=":
			return a != b, nil
		default:
			return false, errors.New(constant.INVALID_OPERATOR_ERROR)
		}
	case reflect.Int:
		boo, err := caseInt(a, b, op)
		return boo, err
	case reflect.Float64:
		boo, err := caseFloat(a, b, op)
		return boo, err
	case timeType.Kind():
		boo, err := typeKind(a, b, op)
		return boo, err
	default:
		// Do nothing and return an error
	}
	msg := fmt.Sprintf("Error when trying to compare %v to %v. Type %v is not supported.", a, b, expectedType)
	return false, errors.New(msg)
}

func opIn(a interface{}, b string) (bool, error) {
	values := strings.Split(strings.Trim(b, "[]"), ",")
	for _, v := range values {
		match, err := compareByType(a, v, "=")
		if err != nil {
			log.Printf("Error trying to compare types: %v", err)
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func caseInt(a interface{}, b, op string) (bool, error) {
	a2 := a.(int)
	b2, err := strconv.Atoi(b)
	if err != nil {
		return false, err
	}
	// Operator must be evaluated by individual type. Operators like > or <
	// are not available on interface{} types.
	switch op {
	case "=", "==":
		return a2 == b2, nil
	case ">":
		return a2 > b2, nil
	case ">=":
		return a2 >= b2, nil
	case "<":
		return a2 < b2, nil
	case "<=":
		return a2 <= b2, nil
	case "!=":
		return a2 != b2, nil
	default:
		return false, errors.New(constant.INVALID_OPERATOR_ERROR)
	}
}

func caseFloat(a interface{}, b, op string) (bool, error) {
	a2 := a.(float64)
	b2, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return false, err
	}
	switch op {
	case "=", "==":
		return a2 == b2, nil
	case ">":
		return a2 > b2, nil
	case ">=":
		return a2 >= b2, nil
	case "<":
		return a2 < b2, nil
	case "<=":
		return a2 <= b2, nil
	case "!=":
		return a2 != b2, nil
	default:
		return false, errors.New(constant.INVALID_OPERATOR_ERROR)
	}
}

func typeKind(a interface{}, b, op string) (bool, error) {
	// truncate to compare dates by day, ignoring hh/mm/ss etc
	a2 := a.(time.Time).Truncate(24 * time.Hour)
	b2, err := time.Parse("2006/01/02", b)
	b2 = b2.Truncate(24 * time.Hour)
	if err != nil {
		return false, err
	}
	switch op {
	case "=", "==":
		return a2.Equal(b2), nil
	case ">":
		return a2.After(b2), nil
	case ">=":
		return a2.After(b2) || a2.Equal(b2), nil
	case "<":
		return a2.Before(b2), nil
	case "<=":
		return a2.Before(b2) || a2.Equal(b2), nil
	case "!=":
		return !a2.Equal(b2), nil
	default:
		return false, errors.New(constant.INVALID_OPERATOR_ERROR)
	}
}

func GetUserPermission(isPermissionSystemEnabled bool, cognitoClient cognito.ICognitoConfig, req events.APIGatewayProxyRequest, domain string) (dataPermission string, appPermission constant.AppPermission, err error) {
	if !isPermissionSystemEnabled {
		// Granular permission is disabled, grant full access
		return "*/*", constant.AdminPermission, nil
	}

	// Get username from request
	username := cognito.ParseUserFromLambdaRq(req)

	// Get groups for user
	groups, err := cognitoClient.ListGroupsForUser(username)
	if err != nil {
		log.Printf("Error getting user's groups: %v", err)
		return "", constant.AppPermission(0), err
	}

	// Find matching group name
	dataGroupPrefix := "ucp-" + domain + "-"
	appGlobalPrefix := constant.AppAccessPrefix + "-global"
	appDomainPrefix := constant.AppAccessPrefix + "-" + domain

	for _, g := range groups {
		// Check data group
		if strings.HasPrefix(g.Name, dataGroupPrefix+constant.DataAccessPrefix) {
			dataPermission = g.Description
		}
		//Check global app group
		if strings.HasPrefix(g.Name, appGlobalPrefix) || strings.HasPrefix(g.Name, appDomainPrefix) {
			permissionStringParts := strings.Split(g.Name, "/")
			if len(permissionStringParts) == 2 {
				val, err := strconv.ParseUint(permissionStringParts[1], 16, constant.AppPermissionSize)
				if err != nil {
					log.Printf("Unable to parse group %s; error: %v\nResuming with the next group", g.Name, err)
				}
				appPermission = appPermission | constant.AppPermission(val)
			}
		}
	}

	return dataPermission, appPermission, nil
}
