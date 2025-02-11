// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	aurora "tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/cloudwatch"
	core "tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DataHandler struct {
	AurSvc *aurora.PostgresDBConfig

	// Optional Aurora client that can be used for queries that don't need to run on the writer instance.
	// Currently, most of our workload is run directly on the writer. This is a pattern that we can continue to adopt and consider
	// initializing with every LCS handler.
	AurReaderSvc *aurora.PostgresDBConfig
	DynSvc       *db.DBConfig
	Tx           core.Transaction
	MetricLogger *cloudwatch.MetricLogger
	utils        LCSUtils
}

func (c *DataHandler) SetDynamoProfileTableProperties(tableName, pk, sk string) {
	c.DynSvc.TableName = tableName
	c.DynSvc.PrimaryKey = pk
	c.DynSvc.SortKey = sk
}

func (c *DataHandler) SetTx(tx core.Transaction) {
	tx.LogPrefix = "customerprofiles-lc-dh"
	c.Tx = tx
	c.AurSvc.SetTx(tx)
	c.DynSvc.SetTx(tx)
	c.utils.SetTx(tx)
}

func validateConnectIds(connectIDs []string) error {
	for _, id := range connectIDs {
		err := validateConnectID(id)
		if err != nil {
			return fmt.Errorf("the following connect ID is invalid: %v. must be a UUID", id)
		}
	}
	return nil
}

func validateConnectID(connectID string) error {
	//regexp to validate UUID
	regexp := regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
	if !regexp.MatchString(connectID) {
		return fmt.Errorf("invalid connect_id %v", connectID)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Profile Create, Update and Delete functions (incl Merge, cache update)
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Delete Profile in Dynamo
func (h DataHandler) DeleteProfileFromCache(connectID string) error {
	profileRecords := []DynamoProfileRecord{}
	forDelete := []map[string]string{}
	err := h.DynSvc.FindStartingWith(connectID, PROFILE_OBJECT_TYPE_PREFIX, &profileRecords)
	if err != nil {
		return err
	}
	for _, record := range profileRecords {
		forDelete = append(forDelete, map[string]string{"connect_id": record.Pk, "record_type": record.Sk})
	}
	return h.DynSvc.DeleteMany(forDelete)
}

// Builds a Profile object from interaction records with a given UPT ID and list of object types
func (h DataHandler) buildProfile(
	conn aurora.DBConnection,
	domain string,
	uptId string,
	objectTypeNameToId map[string]string,
	mappings []ObjectMapping,
	profileLevelRow map[string]interface{},
	objectPriority []string,
	pagination []PaginationOptions,
) (profilemodel.Profile, map[string]string, map[string]time.Time, error) {
	conn.Tx.Info("[buildProfile] Fetching interaction records for connect ID %v sorted ASC", uptId)
	startTime := time.Now()
	profileObjects, err := h.FindInteractionRecordsByConnectID(conn, domain, uptId, mappings, objectTypeNameToId, pagination)
	if err != nil {
		return profilemodel.Profile{}, map[string]string{}, map[string]time.Time{}, err
	}
	duration := time.Since(startTime)
	conn.Tx.Debug("FindInteractionRecordsByConnectID duration:  %v", duration)

	conn.Tx.Info("[buildProfile] Found %v interaction records", len(profileObjects))
	conn.Tx.Debug("[buildProfile] Creating profile object from interaction records for connectID %v", uptId)
	profile, lastUpdatedObjTypeMap, lastUpdatedTsMap := h.interactionsRecordsToProfile(
		uptId,
		profileObjects,
		mappings,
		profileLevelRow,
		objectPriority,
	)
	profile.Domain = domain
	profile.ProfileId = uptId
	return profile, lastUpdatedObjTypeMap, lastUpdatedTsMap, nil
}

func (h DataHandler) buildProfilePutProfileObject(
	conn aurora.DBConnection,
	domain string,
	connectId string,
	profileObjects profilemodel.ProfileObject,
	mappings []ObjectMapping,
	profileLevelRow map[string]interface{},
	objectPriority []string,
) (profilemodel.Profile, map[string]string, map[string]time.Time, error) {
	conn.Tx.Info("[buildProfile] Creating profile object from interaction records for connectID %v", connectId)
	profile, lastUpdatedObjTypeMap, lastUpdatedTsMap := h.interactionsRecordsToProfile(
		connectId,
		[]map[string]interface{}{},
		mappings,
		profileLevelRow,
		objectPriority,
	)
	profile.Domain = domain
	profile.ProfileId = connectId
	profile.ProfileObjects = []profilemodel.ProfileObject{profileObjects}
	conn.Tx.Debug("[buildProfile] Profile Object: %+v", profile)
	return profile, lastUpdatedObjTypeMap, lastUpdatedTsMap, nil
}

// TODO: Block empty strings from overriding dynamo fields
func (h DataHandler) UpdateProfileLowLatencyStore(
	profile profilemodel.Profile,
	domain string,
	objectIdToUpdate string,
	profObjects []profilemodel.ProfileObject,
) error {
	startTime := time.Now()
	h.Tx.Info(
		"[UpdateProfileLowLatencyStore] Updating profile %s in dynamo with objectId: '%s' (empty object id means all objects)",
		profile.ProfileId,
		SENSITIVE_LOG_MASK,
	)
	h.Tx.Debug("[UpdateProfileLowLatencyStore] Creating dynamoDB Records for profile with connectID %+v", profile.ProfileId)
	profileRecord, objectRecords := h.profileToDynamo(profile, objectIdToUpdate, profObjects)
	if objectIdToUpdate != "" && len(objectRecords) != 1 {
		return fmt.Errorf("objectIdToUpdate is not empty, there should be 1 object record. but we have %v", len(objectRecords))
	}

	h.Tx.Info("[UpdateProfileLowLatencyStore] Created %v dynamoDB object record. Saving to dynamo table", len(objectRecords))

	// Conditional Save for profileRecord; using h.DynSvc.DbService for v1.2
	if h.DynSvc.LambdaContext == nil {
		return errors.New("[UpdateProfileLowLatencyStore] LambdaContext is nil")
	}
	if h.DynSvc.PrimaryKey == "" {
		return errors.New("[UpdateProfileLowLatencyStore] PrimaryKey is empty")
	}
	av, err := attributevalue.MarshalMapWithOptions(profileRecord, func(encoderOptions *attributevalue.EncoderOptions) {
		encoderOptions.TagKey = "json"
	})
	if err != nil {
		return errors.New("[UpdateProfileLowLatencyStore] Error while marshalling profile record")
	}
	profileLastUpdated, err := attributevalue.Marshal(profile.LastUpdated.UnixMilli())
	if err != nil {
		return errors.New("[UpdateProfileLowLatencyStore] Error while marshalling profile lastUpdated field")
	}
	_, err = h.DynSvc.DbService.PutItem(h.DynSvc.LambdaContext, &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(h.DynSvc.TableName),
		ConditionExpression: aws.String("attribute_not_exists(lastUpdated) OR lastUpdated <= :newLastUpdated"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newLastUpdated": profileLastUpdated,
		},
	})

	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			//	Ignore Conditional Check Exception; this check ensures that the Dynamo Cache
			//	Always has the latest timestamp from Aurora
			h.Tx.Warn("[UpdateProfileLowLatencyStore] Conditional Check Failed Exception: %v", err)
		} else {
			h.Tx.Error("[UpdateProfileLowLatencyStore] Error while saving records to dynamo: %v", err)
			return err
		}
	}

	//	Save object records to Dynamo Cache
	err = h.DynSvc.SaveMany(objectRecords)
	if err != nil {
		h.Tx.Error("[UpdateProfileLowLatencyStore] Error while saving profile record to dynamo: %v", err)
		return err
	}
	duration := time.Since(startTime)
	//logging dynamo saveMany metric
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_DYNAMO_UPDATE_LATENCY,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	h.Tx.Debug("SaveMany duration:  %v", duration)
	return nil
}

func (h DataHandler) interactionsRecordsToProfile(
	connectID string,
	records []map[string]interface{},
	mappings []ObjectMapping,
	profileLevelRow map[string]interface{},
	objectPriority []string,
) (profilemodel.Profile, map[string]string, map[string]time.Time) {
	//we organize the mappings in a map[object_type_name][fieldname]Mapping
	//this allows a quick retrieve
	var profile profilemodel.Profile
	var lastUpdatedObjectType map[string]string
	var lastUpdatedTs map[string]time.Time

	if len(profileLevelRow) == 0 {
		h.Tx.Debug("No Search rows provided: Rebuilding profile level data from interaction")
		profile, lastUpdatedObjectType, lastUpdatedTs = h.buildProfileFromInteractions(connectID, records, mappings, objectPriority)
	} else {
		h.Tx.Debug("Search rows provided: Rebuilding profile level data from search")
		profile, lastUpdatedObjectType, lastUpdatedTs = h.utils.auroraToProfile(profileLevelRow)

	}
	h.Tx.Info("[interactionsRecordsToProfile] Adding %d objects to profile", len(records))
	profile.ProfileObjects = h.utils.auroraToProfileObjects(records, mappings)

	return profile, lastUpdatedObjectType, lastUpdatedTs
}

// builds profile levle data form interaction list
func (h DataHandler) buildProfileFromInteractions(
	connectID string,
	records []map[string]interface{},
	mappings []ObjectMapping,
	objectTypePriority []string,
) (profilemodel.Profile, map[string]string, map[string]time.Time) {
	h.Tx.Info("[interactionsRecordsToProfile] Creating profile object for connect ID %v", connectID)
	organizedMappings := map[string]map[string]FieldMapping{}
	objectUniqueKey := map[string]string{}
	//stores the object Type of the object that last updated a certain field. "fieldName" => object_type
	lastUpdatedObjectType := map[string]string{}
	//stores the timestamp of the object that last updated a certain field. "fieldName" => timestamp
	lastUpdatedTimestamp := map[string]time.Time{}
	for _, mapping := range mappings {
		organizedMappings[mapping.Name] = map[string]FieldMapping{}
		for _, field := range mapping.Fields {
			organizedMappings[mapping.Name][mappingToColumnName(field.Source)] = field
		}
	}
	h.Tx.Debug("[interactionsRecordsToProfile] objectUniqueKey map: %+v", objectUniqueKey)

	profile := profilemodel.Profile{
		ProfileId:  connectID,
		Attributes: map[string]string{},
	}

	h.Tx.Debug("[interactionsRecordsToProfile] Sort objects by priority and timestamp")
	records = h.sortByPriorityTimestamp(records, objectTypePriority)

	h.Tx.Debug("[interactionsRecordsToProfile] Creating profile level attributes")
	for _, record := range records {
		objectType, ok := record[OBJECT_TYPE_NAME_FIELD].(string)
		if !ok || objectType == "" {
			h.Tx.Warn("[interactionsRecordsToProfile] invalid record: %+v (no object type)", record)
			continue // skip empty/unusable values
		}
		//check if object is unique
		for key, value := range record {
			//	Special handling for timestamp field
			if key == RESERVED_FIELD_TIMESTAMP {
				valueTime, ok := value.(time.Time)
				if ok {
					profile.LastUpdated = h.utils.getLaterDate(profile.LastUpdated, valueTime)
				}
				continue
			}

			valueString, ok := value.(string)
			if !ok || valueString == "" {
				continue // skip empty/unusable values
			}
			if fieldMapping, ok := organizedMappings[objectType][key]; ok {
				if strings.HasPrefix(fieldMapping.Target, "_profile") {
					fieldName := strings.Replace(fieldMapping.Target, "_profile.", "", 1)
					lastUpdatedObjectType[fieldName] = objectType
					lastUpdatedTs, ok := record["timestamp"].(time.Time)
					if !ok {
						h.Tx.Warn("[interactionsRecordsToProfile] invalid record: %+v (no timestamp)", record)
					}
					lastUpdatedTimestamp[fieldMapping.Target] = lastUpdatedTs
					profile = SetProfileField(profile, fieldMapping.Target, valueString)
				}
			}
		}
	}
	h.Tx.Debug("[interactionsRecordsToProfile] LastUpdatedTimes: %+v", lastUpdatedTimestamp)
	h.Tx.Debug("[interactionsRecordsToProfile] LastUpdatedObjects: %+v", lastUpdatedObjectType)
	return profile, lastUpdatedObjectType, lastUpdatedTimestamp
}

func (h DataHandler) sortByPriorityTimestamp(records []map[string]interface{}, objectTypePriority []string) []map[string]interface{} {
	sort.Slice(records, func(i, j int) bool {
		objTypeI, objTypeIExists := records[i][OBJECT_TYPE_NAME_FIELD].(string)
		objTypeJ, objTypeJExists := records[j][OBJECT_TYPE_NAME_FIELD].(string)

		if objTypeIExists && objTypeJExists {
			// Sort by object type priority
			objTypePrioI := findPriority(objTypeI, objectTypePriority)
			objTypePrioJ := findPriority(objTypeJ, objectTypePriority)

			if objTypePrioI != objTypePrioJ {
				return objTypePrioI > objTypePrioJ
			}
		} else {
			h.Tx.Warn("[sortByPriorityTimestamp] invalid record: %+v (no object type),  %+v", records[i], records[j])
		}

		lastUpdatedIRawValue, lastUpdatedIExists := records[i][LAST_UPDATED_FIELD]
		lastUpdatedJRawValue, lastUpdatedJExists := records[j][LAST_UPDATED_FIELD]

		if lastUpdatedIExists && lastUpdatedJExists {
			lastUpdatedIStringValue, lastUpdatedIStringValueOk := lastUpdatedIRawValue.(string)
			lastUpdatedJStringValue, lastUpdatedJStringValueOk := lastUpdatedJRawValue.(string)
			if !lastUpdatedIStringValueOk || lastUpdatedIStringValue == "" {
				h.Tx.Warn("[sortByPriorityTimestamp] invalid record: %+v (invalid lastUpdated field)", records[i])
				return false
			}
			if !lastUpdatedJStringValueOk || lastUpdatedJStringValue == "" {
				h.Tx.Warn("[sortByPriorityTimestamp] invalid record: %+v (invalid lastUpdated field)", records[j])
				return false
			}
			lastUpdatedITimeValue, lastUpdatedIError := time.Parse(INGEST_TIMESTAMP_FORMAT, lastUpdatedIStringValue)
			lastUpdatedJTimeValue, lastUpdatedJError := time.Parse(INGEST_TIMESTAMP_FORMAT, lastUpdatedJStringValue)

			if lastUpdatedIError == nil && lastUpdatedJError == nil {
				return lastUpdatedITimeValue.Before(lastUpdatedJTimeValue)
			}
		}

		h.Tx.Warn("[sortByPriorityTimestamp] invalid record: %+v (invalid timestamp), %+v", records[i], records[j])
		return false
	})
	return records
}

func findPriority(objType string, objectTypePrio []string) int {
	for i, prio := range objectTypePrio {
		if objType == prio {
			return i
		}
	}
	return len(objectTypePrio)
}

/*
"ProfileId",
"AccountNumber",
"AdditionalInformation",
"PartyType",
"BusinessName",
"FirstName",
"MiddleName",
"LastName",
"BirthDate",
"Gender",
"PhoneNumber",
"MobilePhoneNumber",
"HomePhoneNumber",
"BusinessPhoneNumber",
"EmailAddress",
"BusinessEmailAddress",
"PersonalEmailAddress",
"Address",
"ShippingAddress",
"MailingAddress",
"BillingAddress",
"Attributes",
*/
func SetProfileField(profile profilemodel.Profile, mappingTarget string, value string) profilemodel.Profile {
	switch mappingTarget {
	case "_profile.FirstName":
		profile.FirstName = value
	case "_profile.LastName":
		profile.LastName = value
	case "_profile.MiddleName":
		profile.MiddleName = value
	case "_profile.BirthDate":
		profile.BirthDate = value
	case "_profile.Gender":
		profile.Gender = value
	case "_profile.PhoneNumber":
		profile.PhoneNumber = value
	case "_profile.MobilePhoneNumber":
		profile.MobilePhoneNumber = value
	case "_profile.HomePhoneNumber":
		profile.HomePhoneNumber = value
	case "_profile.BusinessPhoneNumber":
		profile.BusinessPhoneNumber = value
	case "_profile.EmailAddress":
		profile.EmailAddress = value
	case "_profile.BusinessEmailAddress":
		profile.BusinessEmailAddress = value
	case "_profile.PersonalEmailAddress":
		profile.PersonalEmailAddress = value
	case "_profile.BusinessName":
		profile.BusinessName = value
	}
	if strings.HasPrefix(mappingTarget, "_profile.Attributes") {
		fieldName := strings.ReplaceAll(mappingTarget, "_profile.Attributes.", "")
		profile.Attributes[fieldName] = value
	}
	if strings.HasPrefix(mappingTarget, "_profile.Address") {
		addrField := strings.ReplaceAll(mappingTarget, "_profile.Address.", "")
		profile.Address = SetAddress(profile.Address, addrField, value)
	}
	if strings.HasPrefix(mappingTarget, "_profile.ShippingAddress") {
		addrField := strings.ReplaceAll(mappingTarget, "_profile.ShippingAddress.", "")
		profile.ShippingAddress = SetAddress(profile.ShippingAddress, addrField, value)
	}
	if strings.HasPrefix(mappingTarget, "_profile.MailingAddress") {
		addrField := strings.ReplaceAll(mappingTarget, "_profile.MailingAddress.", "")
		profile.MailingAddress = SetAddress(profile.MailingAddress, addrField, value)
	}
	if strings.HasPrefix(mappingTarget, "_profile.BillingAddress") {
		addrField := strings.ReplaceAll(mappingTarget, "_profile.BillingAddress.", "")
		profile.BillingAddress = SetAddress(profile.BillingAddress, addrField, value)
	}

	return profile
}

func SetAddress(address profilemodel.Address, fieldName string, value string) profilemodel.Address {
	switch fieldName {
	case "Address1":
		address.Address1 = value
	case "Address2":
		address.Address2 = value
	case "Address3":
		address.Address3 = value
	case "Address4":
		address.Address4 = value
	case "City":
		address.City = value
	case "State":
		address.State = value
	case "PostalCode":
		address.PostalCode = value
	case "Province":
		address.Province = value
	case "Country":
		address.Country = value
	}
	return address
}

func dynamoToProfile(dynRec []DynamoProfileRecord) profilemodel.Profile {
	profile := profilemodel.Profile{}
	for _, dynRec := range dynRec {
		if dynRec.Sk == PROFILE_OBJECT_TYPE_PREFIX+"main" {
			profile.Domain = dynRec.Domain
			profile.ProfileId = dynRec.ProfileId
			profile.AccountNumber = dynRec.AccountNumber
			profile.FirstName = dynRec.FirstName
			profile.MiddleName = dynRec.MiddleName
			profile.LastName = dynRec.LastName
			profile.BirthDate = dynRec.BirthDate
			profile.Gender = dynRec.Gender
			profile.PhoneNumber = dynRec.PhoneNumber
			profile.MobilePhoneNumber = dynRec.MobilePhoneNumber
			profile.HomePhoneNumber = dynRec.HomePhoneNumber
			profile.BusinessPhoneNumber = dynRec.BusinessPhoneNumber
			profile.EmailAddress = dynRec.EmailAddress
			profile.BusinessEmailAddress = dynRec.BusinessEmailAddress
			profile.PersonalEmailAddress = dynRec.PersonalEmailAddress
			profile.Address = dynamoAddressToAddress(dynRec.Address)
			profile.MailingAddress = dynamoAddressToAddress(dynRec.MailingAddress)
			profile.BillingAddress = dynamoAddressToAddress(dynRec.BillingAddress)
			profile.ShippingAddress = dynamoAddressToAddress(dynRec.ShippingAddress)
			profile.BusinessName = dynRec.BusinessName
			profile.Attributes = dynRec.Attributes
		} else {
			profile.ProfileObjects = append(profile.ProfileObjects, profilemodel.ProfileObject{
				ID:         dynRec.ID,
				Type:       dynRec.Type,
				Attributes: dynRec.Attributes,
			})
		}
	}
	return profile
}

func dynamoAddressToAddress(dyn DynamoAddress) profilemodel.Address {
	return profilemodel.Address{
		Address1:   dyn.Address1,
		Address2:   dyn.Address2,
		Address3:   dyn.Address3,
		Address4:   dyn.Address4,
		City:       dyn.City,
		State:      dyn.State,
		PostalCode: dyn.PostalCode,
		Province:   dyn.Province,
		Country:    dyn.Country,
	}
}

func addressToDynamo(addr profilemodel.Address) DynamoAddress {
	return DynamoAddress{
		Address1:   addr.Address1,
		Address2:   addr.Address2,
		Address3:   addr.Address3,
		Address4:   addr.Address4,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.PostalCode,
		Province:   addr.Province,
		Country:    addr.Country,
	}
}

// if non empty objectID is provided, we will only return the dynamo profile record and the
func (h DataHandler) profileToDynamo(profile profilemodel.Profile, objectId string, profObjects []profilemodel.ProfileObject) (profileRecord DynamoProfileRecord, objRecords []DynamoProfileRecord) {
	profileRecord = DynamoProfileRecord{
		Pk:                            profile.ProfileId,
		Sk:                            PROFILE_OBJECT_TYPE_PREFIX + "main",
		Domain:                        profile.Domain,
		AccountNumber:                 profile.AccountNumber,
		FirstName:                     profile.FirstName,
		MiddleName:                    profile.MiddleName,
		LastName:                      profile.LastName,
		BirthDate:                     profile.BirthDate,
		Gender:                        profile.Gender,
		PhoneNumber:                   profile.PhoneNumber,
		MobilePhoneNumber:             profile.MobilePhoneNumber,
		HomePhoneNumber:               profile.HomePhoneNumber,
		BusinessPhoneNumber:           profile.BusinessPhoneNumber,
		PersonalEmailAddress:          profile.PersonalEmailAddress,
		BusinessEmailAddress:          profile.BusinessEmailAddress,
		EmailAddress:                  profile.EmailAddress,
		Address:                       addressToDynamo(profile.Address),
		MailingAddress:                addressToDynamo(profile.MailingAddress),
		BillingAddress:                addressToDynamo(profile.BillingAddress),
		ShippingAddress:               addressToDynamo(profile.ShippingAddress),
		BusinessName:                  profile.BusinessName,
		Attributes:                    profile.Attributes,
		LastUpdatedUnixMilliTimestamp: profile.LastUpdated.UnixMilli(),
	}

	for _, obj := range profObjects {
		if objectId != "" && obj.ID != objectId {
			h.Tx.Debug("object ID %v is provided. filtering object %v from dynamo write", objectId, obj.ID)
			continue
		}
		rec := DynamoProfileRecord{
			Pk:         profile.ProfileId,
			Sk:         PROFILE_OBJECT_TYPE_PREFIX + obj.Type + "_" + obj.ID,
			ProfileId:  obj.Attributes[LC_PROFILE_ID_KEY],
			Attributes: removeEmptyFields(obj.Attributes),
			ID:         obj.ID,
			Type:       obj.Type,
		}
		objRecords = append(objRecords, rec)
	}
	if objectId != "" && len(objRecords) != 1 {
		h.Tx.Warn("object ID %v is provided. There should only be 1 object record but found %v", objectId, len(objRecords))
	}
	return profileRecord, objRecords
}

func removeEmptyFields(m map[string]string) map[string]string {
	m2 := map[string]string{}
	for k, v := range m {
		if v != "" {
			m2[k] = m[k]
		}
	}
	return m
}

func (h DataHandler) validatePutProfileObjectInputs(
	profileID, objectTypeName, uniqueObjectKey string,
	objMap map[string]string,
	profMap map[string]string,
	object string,
) error {
	if profileID == "" {
		h.Tx.Error("Error inserting object: no profile ID provided in the request")
		return fmt.Errorf("no profile ID provided in the request")
	}
	if objectTypeName == "" {
		h.Tx.Error("Error inserting object: no object type name provided in the request")
		return fmt.Errorf("no object type name provided in the request")
	}
	if uniqueObjectKey == "" {
		h.Tx.Error("Error inserting object: no unique object key provided in the request")
		return fmt.Errorf("no unique object key provided in the request")
	}
	if len(objMap) == 0 {
		h.Tx.Error("Error inserting object: no object data provided in the request")
		return fmt.Errorf("no object data provided in the request")
	}
	if len(profMap) == 0 {
		h.Tx.Error("Error inserting object: no profile data provided in the request")
		return fmt.Errorf("no profile data provided in the request")
	}
	if len(objMap) > MAX_FIELDS_IN_OBJECT_TYPE {
		h.Tx.Error("Error inserting object: too many fields in object type. Maximum is %d", MAX_FIELDS_IN_OBJECT_TYPE)
		return fmt.Errorf("too many fields in object type. Maximum is %d", MAX_FIELDS_IN_OBJECT_TYPE)
	}
	return nil
}

func (h DataHandler) rollbackInsert(conn aurora.DBConnection, upsertType, connectID, logLine, domain string, insertError error) error {
	conn.Tx.Error(logLine)
	err := conn.RollbackTransaction()
	if err != nil {
		h.Tx.Error("[WARNING] Error rolling back transaction: %v", err)
	}
	h.MetricLogger.LogMetric(map[string]string{"domain": domain}, METRIC_AURORA_INSERT_ROLLBACK, cloudwatch.Count, 1.0)
	if upsertType == "insert" {
		conn.Tx.Info("rolling back a profile creation requires connect_id deletion to avoid shadow connect ID")
		err := h.deleteFromProfileSearchIndexTable(domain, connectID)
		if err != nil {
			h.Tx.Warn("[WARNING] Error deleting connect ID %s from master table: %v", connectID, err)
		}
	}
	return insertError
}

// used to rollback
func (h DataHandler) deleteFromProfileSearchIndexTable(domain, connectID string) error {
	_, err := h.AurSvc.Query(deleteProfileFromProfileSearchTableSql(domain), connectID)
	return err
}

// Returns a profile's Connect ID and EventType (EventTypeCreated or EventTypeUpdated)
func (h DataHandler) InsertProfileObject(
	domain, objectTypeName, uniqueObjectKey string,
	objMap map[string]string,
	profMap map[string]string,
	object string,
) (string, string, error) {
	tx := core.NewTransaction(h.Tx.TransactionID, "", h.Tx.LogLevel)
	tx.Info("[InsertProfileObject] Inserting object '%s' of type '%s' in domain %s", objMap["profile_id"], objectTypeName, domain)
	profileID := objMap["profile_id"]

	err := h.validatePutProfileObjectInputs(profileID, objectTypeName, uniqueObjectKey, objMap, profMap, object)
	if err != nil {
		tx.Error("Input validation failed: %v", err)
		return "", "", err
	}

	startTime := time.Now()
	//insert Profile ID in master table and create Connect ID if does not exists
	//this runs outside of the transaction to ensure atomicity
	connectID, upsertType, err := h.InsertProfileInMasterTable(tx, domain, profileID)
	duration := time.Since(startTime)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_UPDATE_MASTER_TABLE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	if err != nil {
		tx.Error("Error inserting profile in master table: %v", err)
		return "", "", err
	}
	tx.Info("Successfully inserted profile ID %s into master table in %s ", profileID, duration)

	//acquiring connection form pool
	conn, err := h.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Error("error acquiring DB connection from pool")
		return "", "", err
	}
	defer conn.Release()

	//Insert profile level data + profile count update + object and history
	tx.Debug("[InsertProfileObject] starting transaction")
	err = conn.StartTransaction()
	if err != nil {
		tx.Error("Error starting aurora transaction: %v", err)
		return "", "", fmt.Errorf("error starting aurora transaction: %s", err.Error())
	}
	tx.Debug(
		"[InsertProfileObject] operation: '%s' objectType: '%s' objectId: '%s' profileID: '%s' connectID: '%s' domain: '%s'",
		upsertType,
		objectTypeName,
		SENSITIVE_LOG_MASK,
		SENSITIVE_LOG_MASK,
		connectID,
		domain,
	)

	tx.Debug("[InsertProfileObject] upserting profile into search table")
	searchTableUpdateStartTime := time.Now()
	err = h.UpsertProfileInSearchTable(conn, domain, connectID, objMap, profMap, objectTypeName)
	duration = time.Since(searchTableUpdateStartTime)
	conn.Tx.Debug("[InsertProfileObject] update search table duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_UPDATE_SEARCH_TABLE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	duration = time.Since(startTime)
	tx.Debug("[InsertProfileObject] updateProfile (master + search table) duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_INSERT_PROFILE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	if err != nil {
		err = NewRetryable(fmt.Errorf("database level error occurred while inserting into profile search table %w", err))
		return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("The following database-level error occurred while inserting into profile search table: %v", err), domain, err)
	}

	var eventType string
	if upsertType == "insert" {
		eventType = EventTypeCreated
		startTime := time.Now()
		tx.Info("[InsertProfileObject] New Profile created, incrementing profile count in domain %s", domain)
		err = h.UpdateCount(conn, domain, countTableProfileObject, 1)
		if err != nil {
			return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("Error updating profile count: %v", err), domain, errors.New("Error updating profile count"))
		}
		duration := time.Since(startTime)
		tx.Debug("updateProfileCount duration:  %v", duration)
		h.MetricLogger.LogMetric(
			map[string]string{"domain": domain},
			METRIC_NAME_UPDATE_PROFILE_COUNT_INSERT,
			cloudwatch.Milliseconds,
			float64(duration.Milliseconds()),
		)

	} else {
		eventType = EventTypeUpdated
	}

	tx.Info("[InsertProfileObject] Inserting object '%s' in interaction table", profileID)
	startTime = time.Now()
	query, args := insertObjectSql(domain, connectID, objectTypeName, uniqueObjectKey, objMap)
	objRes, err := conn.Query(query, args...)
	if err != nil {
		return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("[InsertProfileObject] The following database-level error occurred while inserting into object table %v: %v ", objectTypeName, err), domain, errors.New("a database-level error occurred while inserting object-level data"))
	}
	duration = time.Since(startTime)
	tx.Debug("insertObjectSql duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_INSERT_OBJECT,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)

	if len(objRes) == 0 {
		// this case should only happen if the db connection was lost, which we should retry
		err = NewRetryable(errors.New("InsertProfileObject upsert statement return was empty"))
		return "", "", h.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObject] object upsert statement return was empty", domain, err)
	}
	if len(objRes) > 1 {
		return "", "", h.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObject] object upsert statement returned more than one record", domain, errors.New("object upsert statement returned more than one record"))
	}
	upsertType, ok := objRes[0]["upsert_type"].(string)
	if !ok || upsertType != "insert" && upsertType != "update" {
		return "", "", h.rollbackInsert(conn, upsertType, connectID, "[InsertProfileObject] object upsert statement returned invalid upsert_type", domain, errors.New("object upsert statement returned invalid upsert_type"))
	}
	if upsertType == "insert" {
		startTime := time.Now()
		tx.Info("[InsertProfileObject] New Object created, incrementing object counts in domain %s", domain)
		err = h.UpdateCount(conn, domain, objectTypeName, 1)
		if err != nil {
			return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("Error updating object count: %v", err), domain, errors.New("error updating object count"))
		}
		duration := time.Since(startTime)
		tx.Debug("updateObjectCount duration:  %v", duration)
		h.MetricLogger.LogMetric(
			map[string]string{"domain": domain},
			METRIC_NAME_UPDATE_OBJECT_COUNT_INSERT,
			cloudwatch.Milliseconds,
			float64(duration.Milliseconds()),
		)

	}
	tx.Info("[InsertProfileObject] Inserting object '%s' in interaction history table", profileID)
	startTime = time.Now()
	query, args = insertObjectHistorySql(domain, objectTypeName, profileID, objMap[uniqueObjectKey], object)
	_, err = conn.Query(query, args...)
	if err != nil {
		err = NewRetryable(fmt.Errorf("database level error occurred while inserting into object history table for object %w", err))
		return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("The following database-level error occurred while inserting into object history table for object %v: %v ", objectTypeName, err), domain, err)
	}
	duration = time.Since(startTime)
	tx.Debug("insertProfileHistorySql duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_INSERT_HISTORY,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	tx.Debug("Committing DB transaction")
	err = conn.CommitTransaction()
	if err != nil {
		return "", "", h.rollbackInsert(conn, upsertType, connectID, fmt.Sprintf("InsertProfileObject] error committing aurora transaction: %v ", err), domain, errors.New("error committing aurora transaction"))
	}
	tx.Info("[InsertProfileObject] operation successful: '%s' objectType: '%s' objectId: '%s' profileID: '%s' connectID: '%s' domain: '%s' pid: %d", upsertType, objectTypeName, SENSITIVE_LOG_MASK, SENSITIVE_LOG_MASK, connectID, domain)
	return connectID, eventType, err
}

func (h DataHandler) InsertProfileInMasterTable(tx core.Transaction, domain, profileID string) (string, string, error) {
	conn, err := h.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Error("error acquiring DB connection from pool")
		return "", "", err
	}
	defer conn.Release()
	tx.Debug("[InsertProfileObject] Inserting object '%s' in master table", profileID)
	query, args := h.insertProfileSql(domain, profileID)
	//we use a newly created connection for this query which is done outside of the main transaction
	res, err := conn.Query(query, args...)
	if err != nil {
		tx.Error("Error inserting profile in master table: %v", err)
		return "", "", fmt.Errorf("error executing query: %v", err.Error())
	}
	if len(res) == 0 {
		tx.Error("profile insert statement did not return connect_id")
		retryableErr := NewRetryable(errors.New("profile insert statement did not return connect_id"))
		return "", "", retryableErr
	}

	connectID, ok := res[0]["connect_id"].(string)

	if !ok || connectID == "" {
		tx.Error("[InsertProfileObject] profile upsert statement did not return connect_id")
		return "", "", fmt.Errorf("profile upsert statement did not return connect_id")
	}
	upsertType, ok := res[0]["upsert_type"].(string)
	if !ok || upsertType != "insert" && upsertType != "update" {
		// this case should only happen if the db connection was lost, which we should retry
		err = NewRetryable(errors.New("InsertProfileInMasterTable upsert statement return was invalid"))
		tx.Error("[InsertProfileObject] profile upsert statement returned invalid upsert_type")
		return "", "", err
	}
	return connectID, upsertType, nil
}

func (h DataHandler) UpsertProfileInSearchTable(conn aurora.DBConnection, domain string, connectID string, objMap map[string]string, profMap map[string]string, objectTypeName string) error {
	var query string
	var args []interface{}

	h.Tx.Debug("[UpsertProfileInSearchTable] Split profile map to limit query size")
	//we split the field into 2 part. run the upsert statement on the first N fields and th update statement on the rest
	upsertProfMap, updateProfMap := splitProfileMap(profMap, 20)
	h.Tx.Debug("[UpsertProfileInSearchTable] Running upsert statement on %d fields: %+v", len(upsertProfMap), upsertProfMap)
	query, args = h.upsertProfileForSearchSql(domain, connectID, objMap, upsertProfMap, objectTypeName)
	_, err := conn.Query(query, args...)
	if err != nil {
		conn.Tx.Error("Error upserting search record %v", err)
		return err
	}

	if len(updateProfMap) > 0 {
		h.Tx.Debug("[UpsertProfileInSearchTable] Running update statement on %d fields: %+v", len(updateProfMap), updateProfMap)
		query, args = h.updateProfileForSearchSql(domain, connectID, objMap, updateProfMap, objectTypeName)
		_, err = conn.Query(query, args...)
		//we do not rollback the change from the previous query if this fail since the calling function will
		// rollback the whole transaction (we know that because a connection object is passed to teh function)
		// which means the caller function owns the DB transaction
		if err != nil {
			conn.Tx.Error("[UpsertProfileInSearchTable] Error updating search record %v", err)
			return err
		}
	}
	conn.Tx.Info("[UpsertProfileInSearchTable] Successfully upserted profile in search table")
	return nil

}

// this function splits a map into 2 arbitrary maps with the first one of size 'maxMapSize'
// note that the outcome of this function can differ across executions given that
// ordering is not guaranteed when iterating on key/val. this does not matter in our case
func splitProfileMap(proMap map[string]string, maxMapSize int) (map[string]string, map[string]string) {
	upsertProfMap := make(map[string]string)
	updateProfMap := make(map[string]string)
	i := 0
	for k, v := range proMap {
		if v == "" {
			continue
		}
		if i < maxMapSize {
			upsertProfMap[k] = v
		} else {
			updateProfMap[k] = v
		}
		i++
	}
	return upsertProfMap, updateProfMap
}

func insertObjectSql(domain string, connectID, objectTypeName, uniqueObjectKey string, objMap map[string]string) (string, []interface{}) {
	var columns []string
	var values []string
	var args []interface{}
	var setStatements []string
	tableName := createInteractionTableName(domain, objectTypeName)

	columns = append(columns, "connect_id")
	values = append(values, "?")
	args = append(args, connectID)

	for k, v := range objMap {
		columns = append(columns, k)
		values = append(values, "?")
		args = append(args, v)
		setStatements = append(setStatements, fmt.Sprintf(`"%s" = EXCLUDED."%s"`, k, k))
	}

	columnList := `"` + strings.Join(columns, `", "`) + `"`
	valueList := strings.Join(values, `, `)

	query := fmt.Sprintf(`
		INSERT INTO "%s" (%v)
		VALUES (%v)
		ON CONFLICT ("profile_id","%s")
		DO UPDATE SET %v
		RETURNING txid_current() as "tx_id",pg_backend_pid() as "pid",  CASE WHEN NOT xmax = 0 THEN 'update' ELSE 'insert' END AS upsert_type;
	`, tableName, columnList, valueList, uniqueObjectKey, strings.Join(setStatements, ", "))

	return formatQuery(query), args
}

func (h DataHandler) DeleteProfile(conn aurora.DBConnection, domain string, connectId string) error {
	conn.Tx.Info("[DeleteProfile] Deleting connectId '%s' in domain %s", connectId, domain)

	conn.Tx.Debug("[DeleteProfile] Deleting '%s' from master table", connectId)
	startTime := time.Now()
	_, err := conn.Query(deleteProfileFromMasterTableSql(domain), connectId)
	if err != nil {
		conn.Tx.Error("The following database-level error occurred while deleting from master table: %v", err)
		return fmt.Errorf("a database-level error occurred while deleting profile data")
	}

	duration := time.Since(startTime)
	conn.Tx.Debug("deleteObjectSql duration:  %v", duration)
	h.MetricLogger.LogMetric(map[string]string{"domain": domain}, METRIC_NAME_AURORA_DELETE_PROFILE, cloudwatch.Milliseconds, float64(duration.Milliseconds()))

	conn.Tx.Info("[DeleteProfile] Profile deleted; decrementing profile Count")
	startTime = time.Now()
	err = h.UpdateCount(conn, domain, countTableProfileObject, -1)
	if err != nil {
		conn.Tx.Error("Error while updating profile count: %v", err)
		return fmt.Errorf("error updating profile count: %v", err)
	}

	duration = time.Since(startTime)
	conn.Tx.Debug("[DeleteProfile] updateProfileCount duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_UPDATE_PROFILE_COUNT_DELETE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)

	return nil
}

func deleteProfileFromMasterTableSql(domain string) string {
	return fmt.Sprintf(`DELETE FROM "%s" WHERE "connect_id" = ?`, domain)
}
func deleteProfileFromProfileSearchTableSql(domain string) string {
	return fmt.Sprintf(`DELETE FROM "%s" WHERE "connect_id" = ?`, searchIndexTableName(domain))
}

func (h DataHandler) rollbackMerge(conn aurora.DBConnection, domain, connectID string, profileIds []string) {
	err := conn.RollbackTransaction()
	if err != nil {
		h.Tx.Warn("[WARNING] Error rolling back merge transaction: %v", err)
	}
	h.MetricLogger.LogMetric(map[string]string{"domain": domain}, METRIC_AURORA_MERGE_ROLLBACK, cloudwatch.Count, 1.0)
	if len(profileIds) > 0 {
		conn.Tx.Info("rolling back a profile creation requires connect_id deletion to avoid shadow connect ID")
		err := h.rollbackMasterTableChanges(conn.Tx, domain, connectID, profileIds)
		if err != nil {
			h.Tx.Error("[WARNING] Error some profiles failed to rollback in master table %v", err)
		}
	}
}

func (h DataHandler) rollbackMasterTableChanges(tx core.Transaction, domain, connectID string, profileIds []string) error {
	batchSize := 5000
	tx.Info("[rollbackMasterTableChanges] rollback %d profile_ids in master table by batch of %d", len(profileIds), batchSize)
	batches := core.Chunk(core.InterfaceSlice(profileIds), batchSize)
	updateProfileIds := []string{}
	for batchID, batch := range batches {
		tx.Info("[rollbackMasterTableChanges]processing batch %d of length %d", batchID, len(batch))
		var paramPlaceholders []string
		var pidBatch []interface{}
		pidBatch = append(pidBatch, connectID)
		for _, val := range batch {
			paramPlaceholders = append(paramPlaceholders, "?")
			pidBatch = append(pidBatch, fmt.Sprintf("%s", val))
		}
		sql := fmt.Sprintf(`UPDATE "%s" SET "connect_id" = ?
		WHERE "profile_id" IN (%v)
		RETURNING profile_id`, domain, strings.Join(paramPlaceholders, ", "))
		res, err := h.AurSvc.Query(sql, pidBatch...)
		if err != nil {
			h.Tx.Error("[rollbackMasterTableChanges] error rolling back master table update for batch %d: %v", err, batchID)
			return fmt.Errorf("error rolling back master table update: %v", err)
		}
		for _, row := range res {
			pid, ok := row["profile_id"].(string)
			if !ok {
				h.Tx.Warn("[rollbackMasterTableChanges] invalid profile_id format: %+v", row["profile_id"])
			}
			if pid != "" {
				updateProfileIds = append(updateProfileIds, pid)
			}
		}
		h.Tx.Info("[rollbackMasterTableChanges] batch %d of %d profile_ids successfully rollbacked", batchID, batchSize)
	}
	if len(updateProfileIds) != len(profileIds) {
		h.Tx.Warn("[rollbackMasterTableChanges] not all profile_ids were successfully rolled back: %d of %d", len(updateProfileIds), len(profileIds))
		return fmt.Errorf("not all profile_ids were successfully rolled back")
	}
	h.Tx.Info("[rollbackMasterTableChanges]  successfully rolled back %d profile_ids in master table", len(updateProfileIds))
	return nil
}

// This function merges two profiles
// connectIdMerged - ID for the profile which is consuming the other profile
// connectIdToMerge - ID for the profile which is getting merged
func (h DataHandler) MergeProfiles(domain, connectIdMerged string, connectIdToMerge string, objectTypeNames []string, mergeContext ProfileMergeContext, mappings []ObjectMapping, mergeID string) error {
	if connectIdMerged == "" || connectIdToMerge == "" {
		return fmt.Errorf("connect ids cannot be empty")
	}
	if connectIdMerged == connectIdToMerge {
		return fmt.Errorf("connect ids cannot be the same")
	}
	tx := core.NewTransaction(h.Tx.TransactionID, "", h.Tx.LogLevel)

	// When initiating a transaction, the DB process creates a copy of the data at that point in time, during the duration of the transaction, the changes are hidden until commit.
	// Updating the connect_ids in the master table first, outside of the transaction, ensures an atomic cut-off
	tx.Info("[MergeProfiles][DataHandler][%s] Merging %v into %v for master table", mergeID, connectIdToMerge, connectIdMerged)
	profileIDs, err := h.MergeProfilesInMasterTable(tx, domain, connectIdMerged, connectIdToMerge)
	if err != nil {
		tx.Error("[MergeProfiles][DataHandler][%s] Error while updating master table to merge %v and %v: %v", mergeID, connectIdMerged, connectIdToMerge, err)
		return fmt.Errorf("error updating master table during merge: %v", err)
	}
	tx.Debug("[MergeProfiles] Successfully updated %d profile_ids in master table", len(profileIDs))

	conn, err := h.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Error("Error acquiring connection: %v", err)
		return fmt.Errorf("error acquiring connection: %s", err.Error())
	}
	defer conn.Release()

	conn.StartTransaction()
	err = h.MergeProfilesInSearchTable(conn, domain, connectIdMerged, connectIdToMerge, mergeID, mappings)
	if err != nil {
		tx.Error(
			"[MergeProfiles][DataHandler][%s] Error while updating search table to merge %v and %v: %v",
			mergeID,
			connectIdMerged,
			connectIdToMerge,
			err,
		)
		h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
		return fmt.Errorf("error merging profiles in search table: %v", err)
	}
	tx.Debug("[MergeProfiles][DataHandler][%s] updating search table (deleting old record)", mergeID)
	_, err = conn.Query(deleteOldSearchRecordSql(domain), connectIdToMerge)
	if err != nil {
		tx.Error(
			"[MergeProfiles][DataHandler][%s] Error while updating search table to merge %v and %v: %v",
			mergeID,
			connectIdMerged,
			connectIdToMerge,
			err,
		)
		h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
		return fmt.Errorf("error deleting old search record during merge: %v", err)
	}
	for _, objectTypeName := range objectTypeNames {
		tx.Debug("[MergeProfiles][DataHandler][%s] Merging %v and %v for object type name %v", mergeID, connectIdMerged, connectIdToMerge, objectTypeName)
		res, err := conn.Query(
			mergeProfileSql(domain, objectTypeName),
			connectIdMerged,
			mergeContext.ConfidenceUpdateFactor,
			connectIdToMerge,
		)
		if err != nil {
			tx.Error("[MergeProfiles][DataHandler][%s] Error while updating object table %v to merge %v and %v: %v", mergeID, objectTypeName, connectIdMerged, connectIdToMerge, err)
			h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
			return fmt.Errorf("error merging profiles in object table %v: %v", objectTypeName, err)
		}
		tx.Debug("object table %s update results: %+v", objectTypeNames, res)
	}
	tx.Debug("[MergeProfiles][DataHandler][%s] Saving merge history", mergeID)
	query, args := insertProfileHistorySql(domain, connectIdMerged, connectIdToMerge, mergeContext)
	_, err = conn.Query(query, args...)
	if err != nil {
		tx.Error("[MergeProfiles][DataHandler][%s] Error while saving merge history: %v", mergeID, err)
		h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
		return fmt.Errorf("error saving merge history: %v", err)
	}

	tx.Debug("[MergeProfiles][DataHandler][%s] Updating profile Count", mergeID)
	startTime := time.Now()
	err = h.UpdateCount(conn, domain, countTableProfileObject, -1)
	if err != nil {
		tx.Error("[MergeProfiles][DataHandler][%s] Error while updating profile count: %v", mergeID, err)
		h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
		return fmt.Errorf("error updating profile count: %v", err)
	}
	duration := time.Since(startTime)

	tx.Debug("[MergeProfiles][DataHandler][%s] updateProfileCount duration:  %v", mergeID, duration)
	h.MetricLogger.LogMetric(map[string]string{"domain": domain}, METRIC_NAME_UPDATE_PROFILE_COUNT_MERGE, cloudwatch.Milliseconds, float64(duration.Milliseconds()))
	tx.Info("[MergeProfiles][DataHandler][%s] Successfully merged profiles %s amd %s. committing transaction", mergeID, connectIdMerged, connectIdToMerge)
	err = conn.CommitTransaction()
	if err != nil {
		tx.Error("Error committing transaction: %v", err)
		h.rollbackMerge(conn, domain, connectIdToMerge, profileIDs)
		return err
	}
	return nil
}

func (h DataHandler) MergeProfilesInSearchTable(
	conn aurora.DBConnection,
	domain, connectIdMerged, connectIdToMerge, mergeID string,
	mappings []ObjectMapping,
) error {
	conn.Tx.Debug("[MergeProfilesInSearchTable][%s] updating search table (merging records)", mergeID)
	conn.Tx.Debug("[MergeProfilesInSearchTable][%s] creating column list from mappings", mergeID)
	dbColumns := h.utils.mappingsToProfileDbColumns(mappings)
	conn.Tx.Debug("[MergeProfilesInSearchTable][%s] found %d columns to update", mergeID, len(dbColumns))
	batchSize := 100
	//we have to batch here to avoid EOF error with too large queries
	batches := batchColumns(dbColumns, batchSize)
	conn.Tx.Debug("[MergeProfilesInSearchTable][%s] updating fields by Batch of %d", mergeID, batchSize)
	for i, columns := range batches {
		conn.Tx.Debug("[MergeProfilesInSearchTable][%s] processing batch %d of size %d", mergeID, i, len(columns))
		query, args := h.mergeSearchRecordsSql(domain, connectIdMerged, connectIdToMerge, columns)
		_, err := conn.Query(query, args...)
		if err != nil {
			conn.Tx.Error(
				"[MergeProfilesInSearchTable][%s] Error while updating search table to merge %v and %v: %v",
				mergeID,
				connectIdMerged,
				connectIdToMerge,
				err,
			)
			return err
		}
		conn.Tx.Debug("[MergeProfilesInSearchTable][%s] batch %d successfully process", mergeID, i, len(columns))
	}
	conn.Tx.Info("[MergeProfilesInSearchTable][%s] updating search table (merging records)", mergeID)
	return nil
}

func batchColumns(col []string, batchSize int) [][]string {
	batches := core.Chunk(core.InterfaceSlice(col), batchSize)
	merged := [][]string{}
	for _, batch := range batches {
		batchStr := []string{}
		for _, col := range batch {
			batchStr = append(batchStr, col.(string))
		}
		merged = append(merged, batchStr)
	}
	return merged
}

func (h DataHandler) mergeSearchRecordsSql(domain, connectIdMerged, connectIdToMerge string, dbColumns []string) (query string, args []interface{}) {
	tableName := searchIndexTableName(domain)
	filteredColumns := []string{}
	for _, dbCol := range dbColumns {
		if dbCol != "connect_id" && dbCol != "profile_id" && dbCol != "timestamp" {
			filteredColumns = append(filteredColumns, dbCol)
		}
	}
	setStatements := []string{UPDATE_SET_TIMESTAMP_LOCALTIMESTAMP}
	for _, fieldName := range filteredColumns {
		setStatements = append(setStatements, createMergeSetStatement(domain, fieldName)...)
	}

	setStatementsList := strings.Join(setStatements, ",\n")

	query = fmt.Sprintf(`UPDATE "%s" AS t1
	SET %s
	FROM "%s" AS t2
	WHERE t1."connect_id"=? AND t2.connect_id=?`, tableName, setStatementsList, tableName)
	args = append(args, connectIdMerged, connectIdToMerge)
	return query, args
}

func (h DataHandler) MergeProfilesInMasterTable(tx core.Transaction, domain, connectIdMerged, connectIdToMerge string) ([]string, error) {
	profileIDs := []string{}
	query, args := mergeProfileInMasterTableSql(domain, connectIdMerged, connectIdToMerge)
	res, err := h.AurSvc.Query(query, args...)
	if err != nil {
		tx.Error(
			"[MergeProfilesInMasterTable] Error while updating master table to merge %v and %v: %v",
			connectIdMerged,
			connectIdToMerge,
			err,
		)
		return profileIDs, fmt.Errorf("error merging profiles in master table: %v", err)
	}
	for _, row := range res {
		pid, ok := row["profile_id"].(string)
		if !ok {
			tx.Error(
				"[MergeProfilesInMasterTable] error merging profiles in master table: could not parse profile_id '%s' in response",
				row["profile_id"],
			)
			return profileIDs, fmt.Errorf(
				"error merging profiles in master table: could not parse profile_id '%s' in response",
				row["profile_id"],
			)
		}
		profileIDs = append(profileIDs, pid)
	}
	if len(profileIDs) == 0 {
		tx.Error(
			"[MergeProfilesInMasterTable] error merging profiles in master table: no profile IDs returned so no change in the master table",
		)
		return profileIDs, fmt.Errorf("error merging profiles in master table: no profile IDs returned so no change in the master table")
	}
	tx.Info("[MergeProfilesInMasterTable] Successfully updated '%d' rows in the master table", len(profileIDs))
	return profileIDs, err
}

func deleteOldSearchRecordSql(domain string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE connect_id=?", searchIndexTableName(domain))
}

// connectIdMerged - ID for the profile which is consuming the other profile (remaining)
// connectIdToMerge - ID for the profile which is getting merged (deleted)
func mergeProfileInMasterTableSql(masterTableName string, connectIdMerged string, connectIdToMerge string) (query string, args []interface{}) {
	query = fmt.Sprintf(`UPDATE %s SET connect_id=?, %s
	WHERE connect_id=?
	RETURNING profile_id`, masterTableName, UPDATE_SET_TIMESTAMP_LOCALTIMESTAMP)
	args = append(args, connectIdMerged, connectIdToMerge)
	return query, args
}

func mergeProfileSql(
	domain string,
	objectTypeName string,
) string {
	tableName := createInteractionTableName(domain, objectTypeName)
	//update request that changes the connect ID to the sourceConnectID for all rows with a connect_id set to targetConnectID and updates their confidence score
	return fmt.Sprintf(`UPDATE %s
						SET connect_id=?,
						overall_confidence_score=overall_confidence_score * ?
						WHERE connect_id=?
						RETURNING profile_id, connect_id`,
		tableName)
}

// Upsert a profile in a single query using Postgres' ON CONFLICT clause to handle inserts versus updates.
// https://www.postgresql.org/docs/current/sql-insert.html
//
// Returns the profile's Connect ID and 'insert' or 'update'.
// We use Postgres' xmax value to determine if insert vs update. xmax is a system column that tracks the transaction id (xid)
// for row deletion, which only occurs if an existing record is updated.
// (e.g. xmax = 0 is an insert, any other value indicates an update)
// https://www.postgresql.org/docs/9.1/ddl-system-columns.html
// https://stackoverflow.com/questions/34762732/how-to-find-out-if-an-upsert-was-an-update-with-postgresql-9-5-upsert
func (h DataHandler) insertProfileSql(domain string, profileID string) (query string, args []interface{}) {
	tableName := domain

	//we need to add SET profile_id = excluded.profile_id to allow returning to send the profile id in case of conflict
	query = fmt.Sprintf(`
		WITH uuid (val) as (
			values (uuid_generate_v4())
		)
		INSERT INTO "%s" (connect_id, original_connect_id, profile_id)
		VALUES((SELECT val FROM uuid),(SELECT val FROM uuid), ?)
		ON CONFLICT ("profile_id") DO UPDATE
		SET profile_id = excluded.profile_id
		RETURNING "connect_id", CASE WHEN NOT xmax = 0 THEN 'update' ELSE 'insert' END AS upsert_type;
		`, tableName)
	args = append(args, profileID)
	return query, args
}

// Upsert a profile in a single query using Postgres' ON CONFLICT clause to handle inserts versus updates.
// https://www.postgresql.org/docs/current/sql-insert.html
func (h DataHandler) upsertProfileForSearchSql(
	domain, connectID string,
	objMap, profMap map[string]string,
	objectTypeName string,
) (query string, args []interface{}) {
	tableName := searchIndexTableName(domain)
	var setStatements []string
	var columns []string
	var values []string

	ts := objMap[LAST_UPDATED_FIELD]
	if ts == "" {
		ts = "LOCALTIMESTAMP"
	} else {
		ts = `TIMESTAMP '` + ts + `'`
	}

	columns = append(columns, "connect_id")
	values = append(values, "?")
	args = append(args, connectID)
	for fieldName, fieldValue := range profMap {
		// prevent override with empty string
		if fieldValue == "" {
			continue
		}
		setStatements = append(setStatements, createUpsertSetStatement(domain, fieldName, ts, objectTypeName)...)
		// Append values (using parameterized query to handle user input values)
		columns = append(columns, `"`+fieldName+`"`)
		values = append(values, "?")
		args = append(args, fieldValue)

		// Append object type
		columns = append(columns, `"`+fieldName+"_obj_type"+`"`)
		values = append(values, `'`+objectTypeName+`'`)
		// Append timestamp
		// The field-attached timestamp is never automatically updated. We only update it when the field is actually changed.
		columns = append(columns, `"`+fieldName+"_ts"+`"`)
		values = append(values, ts)
	}

	//	Update _profile.timestamp with LOCALTIMESTAMP
	setStatements = append(setStatements, UPDATE_SET_TIMESTAMP_LOCALTIMESTAMP)

	setStatementsList := strings.Join(setStatements, `, `)
	//we need to add SET connect_id = excluded.connect_id to allow returning to send the connect id in case of conflict
	columnList := strings.Join(columns, `, `)
	valueList := strings.Join(values, `, `)

	query = fmt.Sprintf(`
	INSERT INTO "%s" (%s)
	VALUES(%s)
	ON CONFLICT (connect_id)
	DO UPDATE SET %s
	`, tableName, columnList, valueList, setStatementsList)
	return query, args
}

func (h DataHandler) updateProfileForSearchSql(
	domain, connectID string,
	objMap, profMap map[string]string,
	objectTypeName string,
) (query string, args []interface{}) {
	tableName := searchIndexTableName(domain)
	setStatements := []string{}

	ts := objMap[LAST_UPDATED_FIELD]
	if ts == "" {
		ts = "LOCALTIMESTAMP"
	} else {
		ts = `TIMESTAMP '` + ts + `'`
	}
	for fieldName, fieldValue := range profMap {
		// prevent override with empty string
		if fieldValue == "" {
			continue
		}
		statements, statementArg := createUpdateSetStatement(domain, fieldName, fieldValue, ts, objectTypeName)
		setStatements = append(setStatements, statements...)
		// Append values (using parameterized query to handle user input values)
		args = append(args, statementArg)
	}
	args = append(args, connectID)

	setStatements = append(setStatements, UPDATE_SET_TIMESTAMP_LOCALTIMESTAMP)

	setStatementsList := strings.Join(setStatements, `,
	`)

	query = fmt.Sprintf(`UPDATE %s SET %s WHERE connect_id=?`, tableName, setStatementsList)
	return query, args
}

func createUpsertSetStatement(domain, fieldName, timestamp string, objectTypeName string) []string {
	shouldUpdateFnName := shouldUpdateFnName(domain)
	tableName := searchIndexTableName(domain)
	statements := []string{}
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s"= CASE WHEN %s(%s."%s_ts",%s."%s_obj_type",%s,'%s') THEN EXCLUDED."%s" ELSE %s."%s" END`,
			fieldName,
			shouldUpdateFnName,
			tableName,
			fieldName,
			tableName,
			fieldName,
			timestamp,
			objectTypeName,
			fieldName,
			tableName,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_ts"= CASE WHEN %s(%s."%s_ts",%s."%s_obj_type",%s,'%s') THEN %s ELSE %s."%s_ts" END`,
			fieldName,
			shouldUpdateFnName,
			tableName,
			fieldName,
			tableName,
			fieldName,
			timestamp,
			objectTypeName,
			timestamp,
			tableName,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_obj_type"= CASE WHEN %s(%s."%s_ts",%s."%s_obj_type",%s,'%s') THEN '%s' ELSE %s."%s_obj_type" END`,
			fieldName,
			shouldUpdateFnName,
			tableName,
			fieldName,
			tableName,
			fieldName,
			timestamp,
			objectTypeName,
			objectTypeName,
			tableName,
			fieldName,
		),
	)
	return statements
}

func createUpdateSetStatement(domain, fieldName, fieldValue string, timestamp string, objectTypeName string) ([]string, interface{}) {
	shouldUpdateFnName := shouldUpdateFnName(domain)
	statements := []string{}
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s"= CASE WHEN %s("%s_ts","%s_obj_type",%s,'%s') THEN ? ELSE "%s" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			timestamp,
			objectTypeName,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_ts"= CASE WHEN %s("%s_ts","%s_obj_type",%s,'%s') THEN %s ELSE "%s_ts" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			timestamp,
			objectTypeName,
			timestamp,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_obj_type"= CASE WHEN %s("%s_ts","%s_obj_type",%s,'%s') THEN '%s' ELSE "%s_obj_type" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			timestamp,
			objectTypeName,
			objectTypeName,
			fieldName,
		),
	)
	return statements, fieldValue
}

func createMergeSetStatement(domain, fieldName string) []string {
	shouldUpdateFnName := shouldUpdateFnName(domain)
	statements := []string{}
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s"= CASE WHEN %s(t1."%s_ts",t1."%s_obj_type",t2."%s_ts",t2."%s_obj_type") THEN t2."%s" ELSE t1."%s" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_ts"= CASE WHEN %s(t1."%s_ts",t1."%s_obj_type",t2."%s_ts",t2."%s_obj_type") THEN t2."%s_ts" ELSE t1."%s_ts" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
		),
	)
	statements = append(
		statements,
		fmt.Sprintf(
			`"%s_obj_type"= CASE WHEN %s(t1."%s_ts",t1."%s_obj_type",t2."%s_ts",t2."%s_obj_type") THEN t2."%s_obj_type" ELSE t1."%s_obj_type" END`,
			fieldName,
			shouldUpdateFnName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
			fieldName,
		),
	)
	return statements
}

func shouldUpdateFnName(domain string) string {
	return fmt.Sprintf("su_%s", domain)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Profile Retrieve and Search functions
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h DataHandler) GetProfileByConnectIDDynamo(domain string, connectID string) (profilemodel.Profile, error) {
	h.Tx.Info("Fetching profile for connect id %v", connectID)
	startTime := time.Now()
	profiles, err := h.FetchProfiles(domain, []string{connectID})
	duration := time.Since(startTime)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_DYNAMO_FETCH_PROFILE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)

	if err != nil {
		return profilemodel.Profile{}, err
	}
	if len(profiles) == 0 {
		h.Tx.Error("No profile found. returning error")
		return profilemodel.Profile{}, ErrProfileNotFound
	}
	if len(profiles) > 1 {
		h.Tx.Warn("More than one profile found. returning error")
	}
	return profiles[0], nil
}

// fetches a list or profiles from DynamoDB profile table
func (h DataHandler) FetchProfiles(domain string, connectIds []string) ([]profilemodel.Profile, error) {
	if len(connectIds) > MAX_PROFILES_TO_FETCH {
		return nil, fmt.Errorf("max number of profiles to fetch exceeded")
	}
	if len(connectIds) == 0 {
		return []profilemodel.Profile{}, fmt.Errorf("at least one connect_id must be provided")
	}
	err := validateConnectIds(connectIds)
	if err != nil {
		return nil, fmt.Errorf("one or multiple connect IDs are invalid: %v", err)
	}
	batches := core.Chunk(core.InterfaceSlice(connectIds), PROFILE_BATCH_SIZE)
	profiles := []profilemodel.Profile{}
	var errs []error
	for _, ids := range batches {
		var wg sync.WaitGroup
		wg.Add(len(ids))
		var mu sync.Mutex
		for _, id := range ids {
			go func(connectID string) {
				profileRecords := []DynamoProfileRecord{}
				startTime := time.Now()
				err := h.DynSvc.FindStartingWith(connectID, PROFILE_OBJECT_TYPE_PREFIX, &profileRecords)
				duration := time.Since(startTime)
				h.MetricLogger.LogMetric(
					map[string]string{"domain": domain},
					METRIC_NAME_DYNAMO_FIND_STARTING_WITH_LATENCY,
					cloudwatch.Milliseconds,
					float64(duration.Milliseconds()),
				)
				h.MetricLogger.LogMetric(
					map[string]string{"domain": domain},
					METRIC_NAME_DYNAMO_FIND_STARTING_WITH_NUM_REC,
					cloudwatch.Count,
					float64(len(profileRecords)),
				)
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				if len(profileRecords) > 0 {
					mu.Lock()
					profiles = append(profiles, dynamoToProfile(profileRecords))
					mu.Unlock()
				}
				wg.Done()
			}(id.(string))
		}
		wg.Wait()
	}
	if len(profiles) != len(connectIds) {
		h.Tx.Error("Number of profiles fetched does not match the number of connect IDs provided. ")
		return profiles, fmt.Errorf("one of the requested profile could not be retreived (may not exist)")
	}
	return profiles, core.ErrorsToError(errs)
}

func (h DataHandler) GetConnectIdByTravelerIdDynamo(travelerId string) (string, error) {
	// Search on GSI to retrieve Traveler ID to Connect ID translation
	travelerIdToConnectId := []DynamoTravelerIndex{}
	// One travelerId may contain multiple object types but they all map to the same connectId.
	// We only need 1 of these objects to retrieve the mapping between travelerId and connectId
	err := h.DynSvc.FindByGsi(travelerId, DDB_GSI_NAME, DDB_GSI_PK, &travelerIdToConnectId, db.FindByGsiOptions{Limit: 1})
	if err != nil {
		h.Tx.Error("Error querying GSI %s", DDB_GSI_NAME)
		return "", fmt.Errorf("error querying gsi")
	}
	if len(travelerIdToConnectId) == 0 {
		h.Tx.Error("Traveler ID %s not found", travelerId)
		return "", fmt.Errorf("traveler id %s not found", travelerId)
	}

	return travelerIdToConnectId[0].ConnectId, nil
}

// return connect ids for a given search criteria
func (h DataHandler) FindConnectIDByObjectField(domain string, objectTypeName string, key string, values []string) ([]string, error) {
	if len(values) > MAX_VALUES_PER_SEARCH_QUERY {
		return nil, fmt.Errorf("max number of values per search query exceeded")
	}
	cids := []string{}
	query, args := findConnectIDByObjectFieldSql(domain, objectTypeName, key, values)
	res, err := h.AurSvc.Query(query, args...)
	for _, r := range res {
		cids = append(cids, r["connect_id"].(string))
	}
	return cids, err
}

func findConnectIDByObjectFieldSql(domain string, objectTypeName string, key string, values []string) (string, []interface{}) {
	// Build parameterized query using ? in place of each value, then pass values separately as args
	var whereInParams []string
	var args []interface{}
	for _, v := range values {
		whereInParams = append(whereInParams, "?")
		args = append(args, v)
	}

	tableName := createInteractionTableName(domain, objectTypeName)
	query := fmt.Sprintf(`
		SELECT DISTINCT connect_id
		FROM %s
		WHERE "%v" IN (%v);
	`, tableName, key, strings.Join(whereInParams, ", "))

	return formatQuery(query), args
}

func (h DataHandler) FindConnectIDByProfileID(domain, profileID string) (string, error) {
	query := fmt.Sprintf(`SELECT connect_id FROM %s WHERE "profile_id"=?`, domain)
	res, err := h.AurSvc.Query(query, profileID)
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return "", fmt.Errorf("no record found for Profile ID %s", profileID)
	}
	cid, ok := res[0]["connect_id"].(string)
	if !ok {
		return "", fmt.Errorf("missing connect_id in record: %v", res[0])
	}
	return cid, nil
}

func (h DataHandler) FindConnectIdsMultiSearch(
	domain string,
	objectTypeNames []string,
	columnArray []string,
	valuesArray [][]string,
) ([]string, error) {
	if len(objectTypeNames) == 0 {
		return []string{}, errors.New("at least one object type name should be provided")
	}
	cids := []string{}
	whereStatements := []string{}
	joinStatements := []string{}
	args := []interface{}{}

	tableArray := []string{}
	for _, objectTypeName := range objectTypeNames {
		tableName := createInteractionTableName(domain, objectTypeName)
		tableArray = append(tableArray, tableName)
	}
	selectStatement := fmt.Sprintf(`%v."%v"`, tableArray[0], "connect_id")
	fromStatement := fmt.Sprintf(`FROM %v`, tableArray[0])

	for i, tableName := range tableArray {
		var whereInString []string
		for _, v := range valuesArray[i] {
			args = append(args, v)
			whereInString = append(whereInString, "?")
		}
		statement := fmt.Sprintf(`%v."%v" IN (%v)`, tableName, columnArray[i], strings.Join(whereInString, ", "))
		whereStatements = append(whereStatements, statement)
	}

	uniqueTableArray := uniqueElements[string](tableArray)
	for i := range uniqueTableArray {
		if i > 0 {
			joinStatements = append(
				joinStatements,
				fmt.Sprintf(`JOIN %v ON %v.connect_id=%v.connect_id`, uniqueTableArray[i], uniqueTableArray[i], uniqueTableArray[0]),
			)
		}
	}

	query := fmt.Sprintf(
		"SELECT %v %v %v WHERE %v",
		selectStatement,
		fromStatement,
		strings.Join(joinStatements, " "),
		strings.Join(whereStatements, " AND "),
	)
	res, err := h.AurSvc.Query(query, args...)
	for _, r := range res {
		cids = append(cids, r["connect_id"].(string))
	}
	return cids, err
}

// this function returns all interaction records sorted chronologically to be able to recompute the final result dynamically
// For most cases, objectTypeNameToId will be all object types mapped to an empty string
// For PutProfileObject, we only want to return a specific objectId of a specific object type which is where the objectTypeNameToId will be used
func (h DataHandler) FindInteractionRecordsByConnectID(
	conn aurora.DBConnection,
	domain string,
	connectID string,
	mappings []ObjectMapping,
	objectTypeNameToId map[string]string,
	pagination []PaginationOptions,
) ([]map[string]interface{}, error) {
	conn.Tx.Debug("fetching object types for connect_id %s", connectID)
	if len(objectTypeNameToId) > MAX_OBJECT_TYPES_PER_DOMAIN {
		return nil, fmt.Errorf("max number of object types per domain exceeded")
	}
	err := validatePagination(pagination)
	if err != nil {
		conn.Tx.Error("invalid pagination settings: %s", err)
		return nil, err
	}
	paginationByObjectType := organizePaginationByObjectType(pagination)
	res := [][]map[string]interface{}{}
	errs := []error{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(objectTypeNameToId))
	//cannot submit concurrent calls within a transaction. we choose speed over consistency here
	// we still use conn to log the trace ID
	for objectTypeName, objId := range objectTypeNameToId {
		conn.Tx.Debug("fetching object type %s (dedicated connection)", objectTypeName)
		go func(objectType, objectId string) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			paginationForObjType := paginationByObjectType[objectType]
			limit := MAX_INTERACTIONS_PER_PROFILE
			optionalOffset := ""
			if paginationForObjType.PageSize > 0 {
				conn.Tx.Debug("pagination provided for object type %s: %+v", objectType, paginationForObjType)
				limit = paginationForObjType.PageSize
				optionalOffset = fmt.Sprintf(" OFFSET %d", paginationForObjType.Page*paginationForObjType.PageSize)
			}
			var args []interface{}
			sql := fmt.Sprintf(
				`SELECT '%v' as %v, count(*) OVER() AS total_records, * FROM %s WHERE connect_id=? ORDER BY timestamp DESC LIMIT ?%s`,
				objectType,
				OBJECT_TYPE_NAME_FIELD,
				createInteractionTableName(domain, objectType),
				optionalOffset,
			)
			args = append(args, connectID)
			if objectId != "" {
				mapping, err := getObjectMapping(mappings, objectType)
				if err != nil {
					errs = append(errs, err)
					return
				}
				uniqueKey := getIndexKey(mapping, INDEX_UNIQUE)
				sql = fmt.Sprintf(
					`SELECT '%v' as %v, * FROM %s WHERE connect_id=? AND %v=? ORDER BY timestamp DESC LIMIT ?`,
					objectType,
					OBJECT_TYPE_NAME_FIELD,
					createInteractionTableName(domain, objectType),
					uniqueKey,
				)
				args = append(args, objectId)
			}
			args = append(args, limit)
			//we log the query here since the aurora log will use the master trace id
			conn.Tx.Debug("query: %s with arguments: %+v", sql, args)
			rows, err := h.AurSvc.Query(sql, args...)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if len(rows) > 0 {
				res = append(res, rows)
			}
		}(objectTypeName, objId)
	}
	wg.Wait()
	//merging interaction records in order
	return MergeInteractionRecords(res), core.ErrorsToError(errs)
}

func organizePaginationByObjectType(options []PaginationOptions) map[string]PaginationOptions {
	optionsByObjectType := map[string]PaginationOptions{}
	for _, opt := range options {
		optionsByObjectType[opt.ObjectType] = opt
	}
	return optionsByObjectType
}

// Merge takes N sorted slices and returns a single sorted slice
func MergeInteractionRecords(slices [][]map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	// Continue merging as long as there are still slices
	for len(slices) > 0 {

		// Take the first slice
		slice := slices[0]

		// Remove it from the slices array
		slices = slices[1:]

		// If result is empty, add the entire first slice
		if len(result) == 0 {
			result = append(result, slice...)
			continue
		}
		// Merge the first element of slice into result
		i := 0

		for j := 0; j < len(result); j++ {
			if i >= len(slice) {
				continue
			}
			t2 := result[j]["timestamp"].(time.Time)
			t1 := slice[i]["timestamp"].(time.Time)
			if t1.Before(t2) {
				//insert at possition
				result = append(result, map[string]interface{}{})
				copy(result[j+1:], result[j:])
				result[j] = slice[i]
				i++
			}
		}

		// Append any remaining elements of slice
		result = append(result, slice[i:]...)
	}

	return result
}

func (h DataHandler) GetSpecificProfileObject(
	domain string,
	objectKey string,
	objectId string,
	profileId string,
	objectTypeName string,
) (profilemodel.ProfileObject, error) {
	query, args := getSpecificProfileObjectSql(domain, objectTypeName, profileId, objectKey, objectId)
	res, err := h.AurSvc.Query(query, args...)
	if err != nil {
		return profilemodel.ProfileObject{}, err
	}
	if len(res) != 1 {
		return profilemodel.ProfileObject{}, fmt.Errorf("unexpected results for specific profile object")
	}
	return h.utils.auroraToProfileObject(objectTypeName, res[0], objectKey), nil
}

func getSpecificProfileObjectSql(
	domain string,
	objectTypeName string,
	profileId string,
	objectKey string,
	objectId string,
) (string, []interface{}) {
	tableName := createInteractionTableName(domain, objectTypeName)
	query := fmt.Sprintf(`SELECT * FROM %s WHERE "connect_id"= ? AND "%s" = ?`, tableName, objectKey)
	args := []interface{}{profileId, objectId}
	return formatQuery(query), args
}

func (h DataHandler) SearchProfilesByObjectLevelField(domain, objectTypeName, key string, value string) ([]map[string]interface{}, error) {
	tableName := createInteractionTableName(domain, objectTypeName)
	res, err := h.AurSvc.Query(
		fmt.Sprintf(
			`SELECT * FROM %s m WHERE "connect_id" IN (SELECT DISTINCT(o.connect_id) from %s o WHERE o."%s" = ? LIMIT ?)`,
			searchIndexTableName(domain),
			tableName,
			key,
		),
		value,
		MAX_SEARCH_RESPONSE,
	)
	return res, err
}

func (h DataHandler) SearchProfilesByProfileID(domain, profileID string) ([]map[string]interface{}, error) {
	if profileID == "" {
		return []map[string]interface{}{}, fmt.Errorf("no values provided")
	}
	res, err := h.AurSvc.Query(
		fmt.Sprintf(
			`SELECT * FROM %s WHERE "connect_id" = (SELECT connect_id from %s WHERE profile_id=?) LIMIT ?`,
			searchIndexTableName(domain),
			domain,
		),
		profileID,
		MAX_SEARCH_RESPONSE,
	)
	return res, err
}

func (h DataHandler) SearchProfiles(domain, key string, value string) ([]map[string]interface{}, error) {
	res, err := h.AurSvc.Query(
		fmt.Sprintf(`SELECT * FROM %s WHERE "%s" = ? LIMIT ?`, searchIndexTableName(domain), key),
		value,
		MAX_SEARCH_RESPONSE,
	)
	return res, err
}

func (h DataHandler) AdvancedSearchProfilesByProfileLevelField(
	domain string,
	searchCriteria []BaseCriteria,
) (rows []map[string]interface{}, err error) {
	if len(searchCriteria) == 0 {
		return []map[string]interface{}{}, fmt.Errorf("no search criteria provided")
	}

	// Build parameterized query using ? in place of each value, then pass values separately as args
	// whereClause, err := buildWhereClause(searchCriteria)
	whereClause, params, err := buildParameterizedWhereClause(searchCriteria)
	if err != nil {
		h.Tx.Error("error building the where clause: %v", err)
		return []map[string]interface{}{}, err
	}
	params = append(params, MAX_SEARCH_RESPONSE)

	// Build query
	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s LIMIT ?`, searchIndexTableName(domain), whereClause)

	rows, err = h.AurSvc.Query(query, params...)

	return rows, err
}

// get the profile level fields for a profile by upt ID
func (h DataHandler) GetProfileSearchRecord(domain, uptId string) (string, map[string]interface{}, error) {
	query, args := getProfileSearchRecordSql(domain, uptId)
	res, err := h.AurSvc.Query(query, args...)
	if err != nil {
		h.Tx.Error("[GetProfileSearchRecord] error searching profile with upt ID %s: %s", uptId, err)
		return "", map[string]interface{}{}, fmt.Errorf("error searching profile with upt ID %s: %s", uptId, err)
	}
	if len(res) == 0 {
		h.Tx.Error("[GetProfileSearchRecord] no profile exist in domain %s for upt_id %s", domain, uptId)
		return "", map[string]interface{}{}, ErrProfileNotFound
	}
	activeUptId, ok := res[0]["connect_id"].(string)
	if !ok {
		h.Tx.Error("[GetProfileSearchRecord] no connect_id found in search record")
		return "", map[string]interface{}{}, ErrProfileNotFound
	}
	return activeUptId, res[0], err
}

// Using the requested upt id, we check the master table to get the active id.
// This guarantees the profile will be found unless it has been deleted. Otherwise,
// there is an edge case when a profile is merged and the requested id no longer exists
// in the search table.
//
// Example:
// 1. [PutProfileObject phase 1] A profile object is inserted (done in a single transaction)
// 2. [MergeProfiles] The profile is merged into another profile by the Merger
// 3. [PutProfileObject phase 2] The updated profile is retrieved for id res, updating caches, and emitting events
// 4. Error occurs because the id used in PutProfileObject has changed in the middle of the function
func getProfileSearchRecordSql(domain, uptId string) (query string, args []interface{}) {
	query = fmt.Sprintf(`
		SELECT s.*
		FROM %s s
		JOIN %s m ON s.connect_id = m.connect_id
		WHERE m.original_connect_id = ?
		LIMIT 1
	`, searchIndexTableName(domain), domain)

	args = append(args, uptId)
	return formatQuery(query), args
}

////////////////////////////////////////////////////////////////////////
// General Aurora Setup
////////////////////////////////////////////////////////////////////////

// this function need only to be called once at initial setup of the cluster.
// since we don't have such as step, we call each at each domain creation and make it idempotent
func (h DataHandler) SetupAurora() error {
	h.Tx.Info("[SetupAurora] Setting up aurora cluster")
	//adding extension to support uuid generation
	h.Tx.Debug("[SetupAurora] Adding uuid extension")
	_, err := h.AurSvc.Query(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		h.Tx.Error("[SetupAurora] Error added uuid extension: %v", err)
		return err
	}
	return nil

}

////////////////////////////////////////////////////////////////////////
// Interaction  Table
////////////////////////////////////////////////////////////////////////

// creates the aurora DB table to store interaction records
func (h DataHandler) CreateInteractionTable(domain string, objMapping ObjectMapping) error {
	//creating master table
	_, err := h.AurSvc.Query(createInteractionTableSql(domain, objMapping.Name, objMapping.Fields))
	if err != nil {
		h.Tx.Error("[CreateInteractionTable] Error creating interaction table: %v", err)
		return err
	}
	_, err = h.AurSvc.Query(createInteractionTableConnectIdIndexSql(domain, objMapping.Name))
	if err != nil {
		h.Tx.Error("[CreateInteractionTable] Error creating interaction table connect_id index: %v", err)
	}
	return err
}

func createInteractionTableSql(domain string, name string, mappings []FieldMapping) string {
	fields := []string{
		`"connect_id" CHAR(36) NOT NULL`,
		`"profile_id" VARCHAR(255) NOT NULL`,
		`"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`"overall_confidence_score" FLOAT NOT NULL DEFAULT 1`,
	}
	objectUniqueKey := ""
	for _, field := range mappings {
		//we identify teh object unique key to create a primary key on it
		if core.ContainsString(field.Indexes, INDEX_UNIQUE) {
			objectUniqueKey = mappingToColumnName(field.Source)
		}
		columnName := mappingToColumnName(field.Source)
		fields = append(fields, fmt.Sprintf(`"%s" %s`, columnName, TYPE_MAPPING[field.Type]))
	}
	return fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS `+createInteractionTableName(domain, name)+` (%s, PRIMARY KEY(%s));`,
		strings.Join(fields, ","),
		strings.Join([]string{`"profile_id"`, `"` + objectUniqueKey + `"`}, ","),
	)

}
func createInteractionTableConnectIdIndexSql(domain string, name string) string {
	return fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS ` + createInteractionTableName(
			domain,
			name,
		) + `_connect_id_idx ON ` + createInteractionTableName(
			domain,
			name,
		) + ` ("connect_id");`,
	)
}

func (h DataHandler) ShowInteractionTable(domain string, objectTypeName string, limit int) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ?`, createInteractionTableName(domain, objectTypeName)), limit)
}

func (h DataHandler) ShowInteractionTablePartitionPage(
	domain string,
	objectTypeName string,
	partition UuidPartition,
	lastId uuid.UUID,
	limit int,
) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}
	if lastId == uuid.Nil {
		query = fmt.Sprintf(
			`SELECT * FROM %s WHERE connect_id BETWEEN ? AND ? ORDER BY connect_id ASC LIMIT ?`,
			createInteractionTableName(domain, objectTypeName),
		)
		args = append(args, partition.LowerBound.String(), partition.UpperBound.String(), limit)
	} else {
		query = fmt.Sprintf(
			`SELECT * FROM %s WHERE connect_id BETWEEN ? AND ? AND connect_id > ? ORDER BY connect_id ASC LIMIT ?`,
			createInteractionTableName(domain, objectTypeName),
		)
		args = append(args, partition.LowerBound.String(), partition.UpperBound.String(), lastId.String(), limit)
	}
	return h.AurSvc.Query(query, args...)
}

func (h DataHandler) DeleteInteractionTable(domain, name string) error {
	_, err := h.AurSvc.Query(deleteInteractionTableSql(domain, name))
	return err
}

// must be un within the context of a transaction (so we pass a connection object)
func (h DataHandler) DeleteProfileObjectsFromInteractionTable(
	conn aurora.DBConnection,
	domain string,
	objectTypeName string,
	connectID string,
) error {
	conn.Tx.Info(
		"[DeleteProfileFromInteractionTable] deleting connectID: %s from interaction table: %s for domain: %s",
		connectID,
		objectTypeName,
		domain,
	)

	startTime := time.Now()
	_, err := conn.Query(deleteProfileInteractionSql(domain, objectTypeName), connectID)
	if err != nil {
		conn.Tx.Error("Error deleting profile from interaction table: %v", err)
		return fmt.Errorf("error deleting profile from interaction table: %s", err.Error())
	}

	duration := time.Since(startTime)
	conn.Tx.Debug("[DeleteProfileFromInteractionTable] duration: %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_DELETE_PROFILE_INTERACTION,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)

	startTime = time.Now()
	conn.Tx.Info("[DeleteProfileObjectsFromInteractionTable] Object deleted, decrementing object counts in domain %s", domain)
	err = h.UpdateCount(conn, domain, objectTypeName, -1)
	if err != nil {
		conn.Tx.Error("Error updating object count: %v", err)
		return fmt.Errorf("error updating object count: %s", err.Error())
	}
	duration = time.Since(startTime)
	conn.Tx.Debug("updateObjectCount duration:  %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_UPDATE_OBJECT_COUNT_DELETE,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	return nil
}

func deleteProfileInteractionSql(domain, objectTypeName string) string {
	return fmt.Sprintf(`DELETE FROM "%s" WHERE "connect_id" = ?`, createInteractionTableName(domain, objectTypeName))
}

func (h DataHandler) DeleteProfileObjectsFromObjectHistoryTable(
	conn aurora.DBConnection,
	domain string,
	name string,
	connectID string,
) error {
	conn.Tx.Info(
		"[DeleteProfileFromObjectHistoryTable] deleting connectID: %s from interaction table: %s for domain: %s",
		connectID,
		name,
		domain,
	)

	var err error

	profileIds := []string{}
	profiles, err := h.GetProfileIdRecords(conn, domain, connectID)
	if err != nil {
		conn.Tx.Error("Error getting profile by connectID: %v", err)
		return fmt.Errorf("error getting profile by connectID: %s", err.Error())
	}
	if len(profiles) == 0 {
		conn.Tx.Debug("No profiles found with connectID: %s, skipping object history deletion", connectID)
		return nil
	}
	for _, row := range profiles {
		if profileId, ok := row["profile_id"].(string); ok {
			profileIds = append(profileIds, profileId)
		}
	}
	startTime := time.Now()
	query, args := deleteProfileObjectHistorySql(domain, name, profileIds)
	_, err = h.AurSvc.Query(query, args...)
	if err != nil {
		conn.Tx.Error("Error deleting profile from object table: %v", err)
		return fmt.Errorf("error deleting profile from object table: %s", err.Error())
	}
	duration := time.Since(startTime)
	conn.Tx.Debug("[DeleteProfileFromObjectHistoryTable] duration: %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_DELETE_PROFILE_OBJECT_HISTORY,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	return nil
}

func deleteProfileObjectHistorySql(domain string, name string, profileIDs []string) (query string, args []interface{}) {
	var paramPlaceholders []string
	for _, profileId := range profileIDs {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, profileId)
	}
	query = fmt.Sprintf(
		`DELETE FROM "%s" WHERE "profile_id" in (%v)`,
		createObjectHistoryTableName(domain, name),
		strings.Join(paramPlaceholders, ", "),
	)
	return query, args
}

//////////////////////////////////////////////////////////////////////
// Master Table
////////////////////////////////////////////////////////

func (h DataHandler) CreateProfileSearchIndexTable(domain string, mappings []ObjectMapping, domainOptions DomainOptions) error {
	h.Tx.Debug("[CreateProfileSearchIndexTable] one time setup")
	err := h.SetupAurora()
	if err != nil {
		h.Tx.Error("[CreateProfileSearchIndexTable] error setting up aurora for LCS: %v", err)
		return err
	}
	h.Tx.Debug("[CreateProfileSearchIndexTable] creating object type priority function")
	err = h.CreateObjectTypeMappingFunction(domain, domainOptions.ObjectTypePriority)
	if err != nil {
		h.Tx.Error("[CreateProfileSearchIndexTable] error creating object type priority function: %v", err)
		return err
	}
	err = h.CreateShouldUpdateFunction(domain)
	if err != nil {
		h.Tx.Error("[CreateProfileSearchIndexTable] error creating should update function: %v", err)
		return err
	}
	h.Tx.Info("[CreateProfileSearchIndexTable] creating profile search table")
	_, err = h.AurSvc.Query(h.createProfileSearchIndexTableSql(domain, mappings))
	if err != nil {
		h.Tx.Error("[CreateProfileSearchIndexTable] error creating profile search table")
		return err
	}
	return err
}

func (h DataHandler) CreateProfileMasterTable(domain string) error {
	h.Tx.Info("[CreateProfileMasterTable] creating profile master table")
	_, err := h.AurSvc.Query(h.createProfileMasterTableSql(domain))
	if err != nil {
		h.Tx.Error("[CreateProfileMasterTable] error creating profile master table")
		return err
	}
	h.Tx.Info("[CreateProfileMasterTable] creating index for connect_id")
	_, err = h.AurSvc.Query(h.createProfileMasterTableConnectIdIndexSql(domain))
	if err != nil {
		h.Tx.Error("[CreateProfileMasterTable] error creating profile master table connect_id index")
		return err
	}
	h.Tx.Info("[CreateProfileMasterTable] creating index for original_connect_id")
	_, err = h.AurSvc.Query(h.createProfileMasterTableOriginalConnectIdIndexSql(domain))
	if err != nil {
		h.Tx.Error("[CreateProfileMasterTable] error creating profile master table original_connect_id index")
		return err
	}
	return nil
}

func (h DataHandler) createProfileMasterTableOriginalConnectIdIndexSql(domain string) string {
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_idx_original_connect_id ON %s (original_connect_id)", domain, domain)
}

func (h DataHandler) AddCustomAttributesFieldToSearchTable(domain string, mapping ObjectMapping) error {
	_, err := h.AurSvc.Query(addCustomAttributesFieldToSearchTableSql(domain, mapping))
	if err != nil {
		h.Tx.Error("Error while adding custom attributes field to master table: %v", err)
		return err
	}
	return nil
}

func (h DataHandler) GetProfileMasterColumns(domain string) ([]string, error) {
	resp, err := h.AurSvc.Query(getProfileMasterColumnsSql(domain))
	columnNames := []string{}
	for _, r := range resp {
		columnNames = append(columnNames, r["column_name"].(string))
	}
	return columnNames, err
}

func (h DataHandler) ShowMasterTable(domain string, limit int) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ?`, domain), limit)
}

func (h DataHandler) ShowProfileSearchTable(domain string, limit int) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ?`, searchIndexTableName(domain)), limit)
}

func (h DataHandler) createProfileSearchIndexTableSql(domain string, mappings []ObjectMapping) string {
	fields := []string{
		`"connect_id" VARCHAR(255) NOT NULL`,
		`"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
	}

	for _, field := range h.utils.mappingsToProfileDbColumns(mappings) {
		if field == RESERVED_FIELD_TIMESTAMP || field == RESERVED_FIELD_PROFILE_ID || field == RESERVED_FIELD_CONNECT ||
			field == RESERVED_FIELD_ORIGINAL_CONNECT_ID {
			continue
		}
		//for each field we store the value, the timestamp of the interaction and the object type
		//this allows to decide whether this field is updated by new interaction based on timestamp and object_type priority
		fields = append(fields, fmt.Sprintf(`"%s" VARCHAR(255)`, field))
		fields = append(fields, fmt.Sprintf(`"%s_ts"  TIMESTAMP`, field))
		fields = append(fields, fmt.Sprintf(`"%s_obj_type" VARCHAR(255)`, field))
	}
	fields = append(fields, `PRIMARY KEY ("connect_id")`)

	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS `+searchIndexTableName(domain)+` (%s);`, strings.Join(fields, ","))
}

func (h DataHandler) createProfileMasterTableSql(domain string) string {
	fields := []string{
		`"profile_id" VARCHAR(255) NOT NULL`,
		`"connect_id" VARCHAR(255) NOT NULL`,
		`"original_connect_id" VARCHAR(255) NOT NULL`,
		`"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`PRIMARY KEY ("profile_id")`,
	}
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS `+domain+` (%s);`, strings.Join(fields, ","))
}

func (h DataHandler) createProfileMasterTableConnectIdIndexSql(domain string) string {
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_idx_connect_id ON %s (connect_id)", domain, domain)
}

func (h DataHandler) DeleteProfileMasterTable(domain string) error {
	_, err := h.AurSvc.Query(deleteProfileMasterTableSql(domain))
	if err != nil {
		h.Tx.Error("[DeleteProfileMasterTable] error deleting profile master table: %v", err)
	}
	return err
}

func (h DataHandler) DeleteProfileSearchIndexTable(domain string) error {
	err := h.DeleteObjectTypeMappingFunction(domain)
	if err != nil {
		h.Tx.Error("[DeleteProfileMasterTable] error deleting object type mapping function: %v. continuing deletion", err)
	}
	err = h.DeleteShouldUpdateFunction(domain)
	if err != nil {
		h.Tx.Error("[DeleteProfileMasterTable] error deleting should update function: %v. continuing deletion", err)
	}
	_, err = h.AurSvc.Query(deleteProfileSearchIndexTableSql(domain))
	if err != nil {
		h.Tx.Error("[DeleteProfileMasterTable] error deleting profile master table: %v", err)
	}
	return err
}

func searchIndexTableName(domain string) string {
	return domain + "_search"
}

func getProfileMasterColumnsSql(domain string) string {
	return fmt.Sprintf(`SELECT column_name
	FROM information_schema.columns
	WHERE table_schema = 'public'
	AND table_name = '%s';`, searchIndexTableName(domain))
}

func addCustomAttributesFieldToSearchTableSql(domain string, mapping ObjectMapping) string {
	addColStatements := []string{}
	fields := customAttributeFieldsNameFromMapping(mapping)
	for _, field := range fields {
		addColStatements = append(addColStatements, fmt.Sprintf(`ADD COLUMN IF NOT EXISTS "%s" VARCHAR(255)`, field))
		addColStatements = append(addColStatements, fmt.Sprintf(`ADD COLUMN IF NOT EXISTS "%s" TIMESTAMP `, field+"_ts"))
		addColStatements = append(addColStatements, fmt.Sprintf(`ADD COLUMN IF NOT EXISTS "%s" VARCHAR(255)`, field+"_obj_type"))

	}
	return fmt.Sprintf(`ALTER TABLE %s %s;`, searchIndexTableName(domain), strings.Join(addColStatements, ", "))
}

func deleteInteractionTableSql(domain string, name string) string {
	return dropTableSql(createInteractionTableName(domain, name))
}

func deleteProfileMasterTableSql(domain string) string {
	return dropTableSql(domain)
}
func deleteProfileSearchIndexTableSql(domain string) string {
	return dropTableSql(searchIndexTableName(domain))
}

// added to avoid sonarcube duplication error
func dropTableSql(tableName string) string {
	return fmt.Sprintf("DROP TABLE %v;", tableName)
}

func createInteractionTableName(domain string, name string) string {
	if name == PROFILE_OBJECT_TYPE_PREFIX {
		return searchIndexTableName(domain)
	}
	return domain + "_" + name
}

func (h DataHandler) CreateObjectHistoryTable(domain string, object string) error {
	_, err := h.AurSvc.Query(createObjectHistorySql(domain, object))
	return err
}

func createObjectHistorySql(domain, object string) string {
	tableName := createObjectHistoryTableName(domain, object)
	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			"profile_id" VARCHAR(255) NOT NULL,
			"accp_object_id" VARCHAR(255) NOT NULL,
			"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"data" JSONB NOT NULL
		);
	`, tableName)
}

func createObjectHistoryTableName(domain string, object string) string {
	return fmt.Sprintf("%s_%s_history", domain, object)
}

func insertObjectHistorySql(
	domain string,
	objectTypeName string,
	profileID string,
	objectID string,
	rawObject string,
) (query string, args []interface{}) {
	tableName := createObjectHistoryTableName(domain, objectTypeName)
	query = fmt.Sprintf(`
		INSERT INTO %s (profile_id, accp_object_id, data)
		VALUES (?, ?, ?)
	`, tableName)
	args = []interface{}{
		profileID,
		objectID,
		rawObject,
	}

	return query, args
}

func (h DataHandler) DeleteObjectHistoryTable(domain string, object string) error {
	_, err := h.AurSvc.Query(deleteObjectHistorySql(domain, object))
	return err
}

func deleteObjectHistorySql(domain, object string) string {
	tableName := createObjectHistoryTableName(domain, object)
	return fmt.Sprintf(`DROP TABLE %s;`, tableName)
}

func (h DataHandler) retrieveObjectHistorySql(domain string, objectTypeName string) string {
	tableName := createObjectHistoryTableName(domain, objectTypeName)
	return fmt.Sprintf(`SELECT timestamp FROM %s WHERE accp_object_id = ? AND profile_id = ?`, tableName)
}

func (h DataHandler) CreateProfileHistoryTable(domain string) error {
	_, err := h.AurSvc.Query(createProfileHistorySql(domain))
	return err
}

func createProfileHistorySql(domain string) string {
	tableName := createProfileHistoryTableName(domain)
	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			operation_id VARCHAR(255) NOT NULL PRIMARY KEY,
			to_merge VARCHAR(255) NOT NULL,
			merge_type VARCHAR(255) NOT NULL,
			merged_into VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			confidence_update_factor FLOAT,
			rule_id VARCHAR(255),
			ruleset_version VARCHAR(255),
			operator_id VARCHAR(255)
		);
	`, tableName)
}

func createProfileHistoryTableName(domain string) string {
	return domain + "_profile_history"
}

func insertProfileHistorySql(domain, connectIdMerged string, connectIdToMerge string, mergeContext ProfileMergeContext) (query string, args []interface{}) {
	tableName := createProfileHistoryTableName(domain)
	query = fmt.Sprintf(`
		INSERT INTO %s (operation_id, to_merge, merged_into, merge_type, confidence_update_factor, rule_id, ruleset_version, operator_id)
		VALUES (uuid_generate_v4()::text, ?, ?, ?, ?, ?, ?, ?);
	`, tableName)
	args = append(args, connectIdToMerge, connectIdMerged, mergeContext.MergeType, mergeContext.ConfidenceUpdateFactor, mergeContext.RuleID, mergeContext.RuleSetVersion, mergeContext.OperatorID)
	return query, args
}

// For a profile, recursively gets all the profiles merged into it (directly/indirectly)
func (h DataHandler) RetrieveProfileHistory(conn aurora.DBConnection, profileID, domain string, depth int) ([]ProfileHistory, error) {
	tableName := createProfileHistoryTableName(domain)
	var args []interface{}
	query := fmt.Sprintf(`
		WITH RECURSIVE profile_history AS (
			SELECT to_merge, timestamp, merged_into, merge_type, 0 AS depth
			FROM %s
			WHERE merged_into = ? AND merge_type != ?

			UNION ALL

			SELECT source.to_merge, source.timestamp, source.merged_into, source.merge_type, target.depth + 1
			FROM %s source
			JOIN profile_history target ON source.merged_into = target.to_merge
			WHERE target.depth < ? AND target.merge_type != ?
		)
		SELECT
		    to_merge,
			timestamp,
			merged_into,
			depth
		FROM profile_history;
	`, tableName, tableName)
	args = append(args, profileID, MergeTypeUnmerge, depth, MergeTypeUnmerge)
	res, err := conn.Query(query, args...)
	if err != nil {
		conn.Tx.Error("Error retrieving profile history %s", err)
		return nil, err
	}

	profileHistory := []ProfileHistory{}
	for _, row := range res {
		toMergeUPTID, ok := row["to_merge"].(string)
		if !ok {
			conn.Tx.Error("Error retrieving profile history: could not cast 'to_merge' (%v) into string: ", row["to_merge"])
			return nil, fmt.Errorf("error retrieving profile history: invalid 'to_merge' field")
		}
		mergedIntoUPTID, ok := row["merged_into"].(string)
		if !ok {
			conn.Tx.Error("Error retrieving profile history: could not cast 'merged_into' (%v) into string: ", row["merged_into"])
			return nil, fmt.Errorf("error retrieving profile history: invalid 'merged_into' field")
		}
		depth, ok := row["depth"].(int32)
		if !ok {
			conn.Tx.Error("Error retrieving profile history: could not cast 'depth' (%v) into int64: ", row["depth"])
			return nil, fmt.Errorf("error retrieving profile history: invalid 'depth' field")
		}
		timestamp, ok := row["timestamp"].(time.Time)
		if !ok {
			conn.Tx.Error("Error retrieving profile history: could not cast 'timestamp' (%v) into time.Time: ", row["timestamp"])
			return nil, fmt.Errorf("error retrieving profile history: invalid 'timestamp' field")
		}

		ph := ProfileHistory{
			toMergeUPTID:    toMergeUPTID,
			mergedIntoUPTID: mergedIntoUPTID,
			Depth:           int64(depth),
			Timestamp:       timestamp,
		}
		profileHistory = append(profileHistory, ph)
	}
	return profileHistory, nil
}

func (h DataHandler) DeleteProfileHistoryTable(domain string) error {
	_, err := h.AurSvc.Query(deleteProfileHistorySql(domain))
	return err
}

func deleteProfileHistorySql(domain string) string {
	tableName := createProfileHistoryTableName(domain)
	return fmt.Sprintf(`DROP TABLE %s;`, tableName)
}

func (h DataHandler) AddDeletionRecordToProfileHistoryTable(
	conn aurora.DBConnection,
	domain string,
	connectID string,
	operatorId string,
) error {
	conn.Tx.Info(
		"[AddDeletionRecordToProfileHistoryTable] adding deletion record for connectID: %s for domain %s to profile history table by operatorId: %s",
		connectID,
		domain,
		operatorId,
	)

	startTime := time.Now()
	query, args := insertProfileHistorySql(domain, core.EmptyUUID(), connectID, ProfileMergeContext{
		MergeType:              OperationTypeDelete,
		ConfidenceUpdateFactor: 0,
		RuleID:                 "",
		RuleSetVersion:         "",
		OperatorID:             operatorId,
	})
	_, err := conn.Query(query, args...)
	if err != nil {
		conn.Tx.Error("Error inserting profile history record: %v", err)
		return err
	}
	duration := time.Since(startTime)
	conn.Tx.Debug("[DeleteProfileFromObjectHistoryTable] duration: %v", duration)
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_AURORA_DELETE_PROFILE_OBJECT_HISTORY,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	return nil
}

////////////////////////////////////////////////////////////////////////////
// Master table update logic
////////////////////////////////////////////////////////////////////

func (h DataHandler) CreateObjectTypeMappingFunction(domain string, objectTypePrio []string) error {
	_, err := h.AurSvc.Query(createObjectTypePrioFunctionSql(domain, objectTypePrio))
	if err != nil {
		h.Tx.Error("Error while creating objectTypePriority function: %v", err)
		return err
	}

	return nil
}
func (h DataHandler) DeleteObjectTypeMappingFunction(domain string) error {
	_, err := h.AurSvc.Query(deleteObjectTypePrioFunctionSql(domain))
	if err != nil {
		h.Tx.Error("Error while deleting objectTypePriority function: %v", err)
		return err
	}

	return nil
}

func createObjectTypePrioFunctionSql(domain string, objectTypePrio []string) string {
	statements := []string{}
	if len(objectTypePrio) == 0 {
		statements = append(statements, "value := 0;")
	} else {
		statements = append(statements, fmt.Sprintf("IF key = '%s' THEN value := %d;", objectTypePrio[0], 1))
		for i := 1; i < len(objectTypePrio); i++ {
			statements = append(statements, fmt.Sprintf("	  ELSIF key = '%s' THEN value := %d;", objectTypePrio[i], i+1))
		}
		statements = append(statements, "	  ELSE value := 0;")
		statements = append(statements, "	  END IF;")
	}

	funName := fmt.Sprintf(`CREATE OR REPLACE FUNCTION %s(key text) RETURNS integer AS $$
	DECLARE
	  value integer;
	BEGIN
	  -- Define a hardcoded map
	  %s

	  RETURN value;
	END;
	$$ LANGUAGE plpgsql;`, prioFunctionName(domain), strings.Join(statements, "\n"))
	return funName
}

func (h DataHandler) CreateShouldUpdateFunction(domain string) error {
	_, err := h.AurSvc.Query(createShouldUpdateFunctionSql(domain))
	if err != nil {
		h.Tx.Error("Error while creating objectTypePriority function: %v", err)
		return err
	}

	return nil
}
func (h DataHandler) DeleteShouldUpdateFunction(domain string) error {
	_, err := h.AurSvc.Query(deleteShouldUpdateFunctionSql(domain))
	if err != nil {
		h.Tx.Error("Error while deleting objectTypePriority function: %v", err)
		return err
	}

	return nil
}

func createShouldUpdateFunctionSql(domain string) string {
	prioFn := prioFunctionName(domain)
	name := shouldUpdateFnName(domain)
	return fmt.Sprintf(`CREATE OR REPLACE FUNCTION %s(
		fn_ts TIMESTAMP,
		fn_obj_type TEXT,
		cur_ts TIMESTAMP,
		cur_obj_type TEXT
	)
	RETURNS BOOLEAN
	AS $$
	DECLARE
		is_higher_priority BOOLEAN;
		is_same_priority BOOLEAN;
	BEGIN
		-- Check if the current object type has a higher priority
		is_higher_priority := %s(fn_obj_type) < %s(cur_obj_type);

		-- Check if the priorities are the same
		is_same_priority := %s(fn_obj_type) = %s(cur_obj_type);

		-- Return true if:
		-- 1. The current object type has a higher priority, or
		-- 2. The priorities are the same, the timestamp is older than the current time
		-- 3. The existing timestamp is NULL (field never updated)
		RETURN (
			is_higher_priority
			OR (
				is_same_priority
				AND (fn_ts <= cur_ts OR fn_ts IS NULL)
			)
		);
	END;
	$$ LANGUAGE plpgsql;`, name, prioFn, prioFn, prioFn, prioFn)
}

func deleteShouldUpdateFunctionSql(domain string) string {
	return fmt.Sprintf(`DROP FUNCTION IF EXISTS %s(
		fn_ts TIMESTAMP,
		fn_obj_type TEXT,
		cur_ts TIMESTAMP,
		cur_obj_type TEXT
	);`, shouldUpdateFnName(domain))
}

func deleteObjectTypePrioFunctionSql(domain string) string {
	return fmt.Sprintf(`DROP FUNCTION %s(key text)`, prioFunctionName(domain))
}

// max identifier name in aurora is 63 bytes. Given that domains can be only alphanumerics
// this is 63 char
func prioFunctionName(domain string) string {
	domTruncated := core.TruncateString(domain, 63-len(PRIO_FUNCTION_PREFIX)-1)
	return fmt.Sprintf("%s_%s", PRIO_FUNCTION_PREFIX, domTruncated)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Data Lineage and Unmerge
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// test that
func (h DataHandler) HaveBeenMergedAndNotUnmerged(domain, connectIdMerged, connectIdToMerge string) (bool, error) {

	mergedHistoryRecords, err := h.AurSvc.Query(haveBeenMergedSql(domain), MergeTypeUnmerge, connectIdToMerge, connectIdMerged)
	if err != nil {
		h.Tx.Error("[HaveBeenMergedAndNotUnmerged] Error querying for merged history records: %v", err)
		return false, err
	}
	unmergedHistoryRecords, err := h.AurSvc.Query(haveBeenUnmergedSql(domain), MergeTypeUnmerge, connectIdToMerge, connectIdMerged)
	if err != nil {
		h.Tx.Error("[HaveBeenMergedAndNotUnmerged] Error querying for unmerged history records: %v", err)
		return false, err
	}
	h.Tx.Debug(
		"[HaveBeenMergedAndNotUnmerged]Merge history has %d record with 'MERGE' operation and %d records with 'UNMERGE' operation",
		len(mergedHistoryRecords),
		len(unmergedHistoryRecords),
	)
	if len(mergedHistoryRecords) > len(unmergedHistoryRecords) {
		h.Tx.Debug("[HaveBeenMergedAndNotUnmerged] record have been merge more times than unmerged")
		return true, nil
	}
	h.Tx.Debug("[HaveBeenMergedAndNotUnmerged] record have never been merged or have been merged and unmerged")
	return false, nil
}

func haveBeenMergedSql(domain string) string {
	tableName := createProfileHistoryTableName(domain)
	return fmt.Sprintf(
		`SELECT to_merge, merged_into FROM %s WHERE (merge_type!=? AND to_merge=? AND merged_into=?)`,
		tableName,
	)
}
func haveBeenUnmergedSql(domain string) string {
	tableName := createProfileHistoryTableName(domain)
	return fmt.Sprintf(
		`SELECT to_merge, merged_into FROM %s WHERE (merge_type=? AND to_merge=? AND merged_into=?)`,
		tableName,
	)
}

func unmergeProfileInMasterTableSql(masterTableName string, uptIdToUnmerge string, profileIdsToUnmerge []string) (query string, args []interface{}) {
	var paramPlaceholders []string
	args = append(args, uptIdToUnmerge)
	for _, profId := range profileIdsToUnmerge {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, profId)
	}
	query = fmt.Sprintf(
		`UPDATE %s SET connect_id=? WHERE profile_id IN (%v)`,
		masterTableName,
		strings.Join(paramPlaceholders, ", "),
	)
	return query, args
}

// Traverses the 'mergeHistoryMap' starting at root - 'uptId' and gets all its descendants top-down
// mergeHistoryMap holds the merge and unmerge events as a map of map of int. Breaking this down -
// Consider a merge event where Profile A gets merged into Profile B
// - The outer map is for the mergedInto profile i.e. Profile B. Key for the outer map is profile ID and value is the inner map.
// - For every outer map, its inner map holds the profile that got merged i.e. Profile A. Key for the inner map in profile ID and value is an integer
func buildDescendantsFromMap(uptId string, mergeHistoryMap map[string]map[string]int) []string {
	filteredProfileSet := make(map[string]bool)
	queue := []string{uptId}

	// Breadth first search
	for len(queue) > 0 {
		profileId := queue[0]
		queue = queue[1:]

		if !filteredProfileSet[profileId] {
			filteredProfileSet[profileId] = true
			directChildren, ok := mergeHistoryMap[profileId]
			if ok {
				for key, value := range directChildren {
					if !filteredProfileSet[key] && value > 0 {
						queue = append(queue, key)
					}
				}
			}
		}
	}

	// Convert set to []string
	var filteredChildren []string
	for key := range filteredProfileSet {
		filteredChildren = append(filteredChildren, key)
	}
	return filteredChildren
}

// For a profile present in the originalConnectId list, this function queries the master table to get their profileId or connectId depending on the idType.
func (h DataHandler) getIdsForOriginalConnectIds(
	conn aurora.DBConnection,
	masterTableName string,
	idType string,
	originalConnectIds []string,
) ([]string, error) {
	query, args := GetIdFromMasterTableByOriginalConnectIdSql(masterTableName, originalConnectIds)
	profileIds, err := conn.Query(query, args...)
	if err != nil {
		return []string{}, fmt.Errorf("[getIdsForOriginalConnectIds] unable to get the associated id, error: %s", err)
	}
	var idList []string
	for _, row := range profileIds {
		currentId, ok := row[idType].(string)
		if ok {
			idList = append(idList, currentId)
		} else {
			return []string{}, fmt.Errorf("[getIdsForOriginalConnectIds] unable to get the associated id, error: %s", err)
		}
	}
	return idList, nil
}

// Returns the profilesIds for all the children of 'uptId'
func (h DataHandler) buildLineageForProfile(conn aurora.DBConnection, domain, uptId string) ([]string, float64, error) {

	// Step 1: Get all the merges for profile 'uptId' and recursively get the merge of its children (direct/indirect descendants)
	listOfMerge, err := h.RetrieveProfileHistory(conn, uptId, domain, PROFILE_HISTORY_DEPTH)

	if err != nil {
		conn.Tx.Error("[buildLineageForProfile] unable to fetch merge history for profile %s; error - %v", uptId, err)
		return []string{}, 0, fmt.Errorf("[buildLineageForProfile] unable to fetch merge history for profile %s; error - %v", uptId, err)
	}
	// List of all children for uptId (direct/indirect descendants)
	lineageList := []string{uptId}

	// The merge operation is stored in two maps
	// childrenMap is a Map of Sets which stores the direct children for each profile
	// Outerkey is the mergedIntoUPTID/parentUptId, InnerKey is the toMergeUPTID/childUptId
	childrenMap := make(map[string]map[string]int)

	// parentMap is a Map of Sets to store the direct parent for each profile
	// Outerkey is the toMergeUPTID/childUptId, InnerKey is the mergedIntoUPTID/parentUptId
	parentMap := make(map[string]map[string]int)

	// Build the maps
	// For both maps, value for the inner key in an integer which hold count. See example -
	// Step 1: Merge   profile 0 into profile 1
	// Step 2: Unmerge profile 0 from profile 1
	// Step 3: Merge   profile 0 into profile 1
	// While building the maps, we may come across duplicate merge events (step 1 & 3)
	// We increment the map[1][0] when we see steps 1 and 3 and decrement on step 2
	for _, merge := range listOfMerge {
		lineageList = append(lineageList, merge.toMergeUPTID)
		if childrenMap[merge.mergedIntoUPTID] == nil {
			childrenMap[merge.mergedIntoUPTID] = make(map[string]int)
		}
		childrenMap[merge.mergedIntoUPTID][merge.toMergeUPTID] += 1

		if parentMap[merge.toMergeUPTID] == nil {
			parentMap[merge.toMergeUPTID] = make(map[string]int)
		}
		parentMap[merge.toMergeUPTID][merge.mergedIntoUPTID] += 1
	}

	// Step 2: Get all unmerges for the intial set of profiles
	listOfUnmerges, err := h.getUnmerges(conn, domain, lineageList)
	if err != nil {
		conn.Tx.Error("[buildLineageForProfile] unable to fetch unmerge history; error - %v", err)
		return []string{}, 0, fmt.Errorf("[buildLineageForProfile] unable to fetch unmerge history; error - %v", err)
	}

	// For every unmerge, gets the original parent for a child as per the merge operation (from parentMap)
	// Then it updates the childrenMap to indicate the unmerge occurred
	for _, unmerge := range listOfUnmerges {
		parentSet, ok := parentMap[unmerge.toUnmergeUPTID]
		if ok {
			for originalParent := range parentSet {
				h.Tx.Debug("Unmerging " + unmerge.toUnmergeUPTID + " from its actual parent " + originalParent)
				childrenMap[originalParent][unmerge.toUnmergeUPTID] -= 1
			}
		}
	}

	// Performs a Breadth first search on the children map to get the children of uptId without the unmerged profiles
	filteredUptIds := buildDescendantsFromMap(uptId, childrenMap)

	profileIds, err := h.getIdsForOriginalConnectIds(conn, domain, "profile_id", filteredUptIds)
	if err != nil {
		conn.Tx.Error("[buildLineageForProfile] unable to fetch profile ids; error - %v", err)
		return []string{}, 0, fmt.Errorf("[buildLineageForProfile] unable to fetch profile ids; error - %v", err)
	}

	// When unmerging, the confidence score should be reverted to a value close to what it would have been if the merge had never happened
	// For example, consider this merge scenario 0 -> 1 -> 2 -> 3 -> 4
	// If unmerging 2 from 4, we need to revert confidence score associated with merge 2 -> 3 and merge 3 -> 4

	// For the profile to unmerge, get the events leading to current parent from merge history table
	history, err := h.getMergesLeadingToParent(conn, domain, uptId)
	if err != nil {
		conn.Tx.Error("error getMergesLeadingToParent: %v", err)
		return []string{}, 0, fmt.Errorf("[buildLineageForProfile] unable to recompute confidence; error - %v", err)
	}

	var confidenceRevertFactor float64 = 1
	for _, merge := range history {
		if merge.ConfidenceUpdateFactor != 0 {
			if merge.MergeType == MergeTypeUnmerge {
				confidenceRevertFactor = confidenceRevertFactor / merge.ConfidenceUpdateFactor
			} else {
				confidenceRevertFactor = confidenceRevertFactor * merge.ConfidenceUpdateFactor
			}
		}
	}
	return profileIds, confidenceRevertFactor, nil
}

/*
This function unmerges two profiles.
uptIdMerged - this is the profile which contains the other profile as a child
uptIdToUnmerge - this is the profile to unmerge out of 'uptIdMerged'
objectTypeNames - list of all interaction names to update
*/
func (h DataHandler) UnmergeProfiles(
	domain, uptIdMerged string,
	uptIdToUnmerge string,
	interactionTypeNames []string,
	mappings []ObjectMapping,
	objectPriority []string,
) error {
	tx := core.NewTransaction(h.Tx.TransactionID, "", h.Tx.LogLevel)
	if uptIdMerged == "" || uptIdToUnmerge == "" {
		return fmt.Errorf("[UnmergeProfiles] connect ids cannot be empty")
	}

	if uptIdMerged == uptIdToUnmerge {
		return fmt.Errorf("[UnmergeProfiles] connect ids cannot be the same")
	}

	conn, err := h.AurSvc.AcquireConnection(tx)
	if err != nil {
		tx.Error("Error aquiring connection from pool: %v", err)
		return err
	}
	defer conn.Release()

	// starting the transaction before we start building the lineage to guarantee the correctness of the profiles to unmerge
	err = conn.StartTransaction()
	if err != nil {
		tx.Error("Error starting transcation: %v", err)
	}

	// For the profile to unmerge, get all of its children (direct and indirect)
	profileIdsToUnmerge, confidenceRevertFactor, err := h.buildLineageForProfile(conn, domain, uptIdToUnmerge)
	if err != nil {
		tx.Error("Error building lineage for profile %v: %v", uptIdToUnmerge, err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error finding children profile; error - %v", err)
	}

	tx.Info("[UnmergeProfiles] Unmerging %v from %v", uptIdToUnmerge, uptIdMerged)

	// In the master table, update uptID = uptIdToUnmerge for the profile 'uptIdToUnmerge' and all its children
	query, args := unmergeProfileInMasterTableSql(domain, uptIdToUnmerge, profileIdsToUnmerge)
	_, err = conn.Query(query, args...)
	if err != nil {
		tx.Error("Error unmergeProfileInMasterTableSql: %v", err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error while updating master table to unmerge %v from %v: %v", uptIdToUnmerge, uptIdMerged, err)
	}

	// For interaction types listed, update uptID = uptIdToUnmerge for all interactions belonging to the profile 'uptIdToUnmerge' and all its children
	for _, objectTypeName := range interactionTypeNames {
		query, args = unmergeInteractionsSql(domain, uptIdToUnmerge, profileIdsToUnmerge, objectTypeName, confidenceRevertFactor)
		_, err := conn.Query(query, args...)
		if err != nil {
			tx.Error("Error unmergeInteractionsSql: %v", err)
			conn.RollbackTransaction()
			return fmt.Errorf("[UnmergeProfiles] error unmerging profiles in object table %v: %v", objectTypeName, err)
		}
	}

	// Record the unmerge operation in the Profile history table
	query, args = insertProfileHistorySql(
		domain,
		uptIdMerged,
		uptIdToUnmerge,
		ProfileMergeContext{MergeType: MergeTypeUnmerge, ConfidenceUpdateFactor: confidenceRevertFactor},
	)
	_, err = conn.Query(query, args...)
	if err != nil {
		tx.Error("Error insertProfileHistorySql: %v", err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error persisting unmerge operation: %v", err)
	}

	//rebuild profile Search record
	err = h.recreatedMasterRecordFromInteractions(conn, domain, uptIdToUnmerge, interactionTypeNames, mappings, objectPriority)
	if err != nil {
		tx.Error("Error recreatedMasterRecordFromInteractions: %v", err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error recreating master record from interactions: %v", err)
	}

	err = h.recreatedMasterRecordFromInteractions(
		conn,
		domain,
		uptIdMerged,
		interactionTypeNames,
		mappings,
		objectPriority,
	)
	if err != nil {
		tx.Error("Error recreatedMasterRecordFromInteractions: %v", err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error recreating master record from interactions: %v", err)
	}

	// Add unmerged profile back into count table
	startUnmergeCountUpdate := time.Now()
	err = h.UpdateCount(conn, domain, countTableProfileObject, 1)
	if err != nil {
		tx.Error("Error UpdateCount: %v", err)
		conn.RollbackTransaction()
		return fmt.Errorf("[UnmergeProfiles] error updating count %s", err)
	}
	h.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_UPDATE_PROFILE_COUNT_UNMERGE,
		cloudwatch.Milliseconds,
		float64(time.Since(startUnmergeCountUpdate).Milliseconds()),
	)

	err = conn.CommitTransaction()
	if err != nil {
		tx.Error("Error commiting transaction: %v", err)
		return fmt.Errorf("[UnmergeProfiles] error commiting transaction: %v", err)
	}
	tx.Info("[UnmergeProfiles] Successfully unmerged %v from %v", uptIdToUnmerge, uptIdMerged)
	return nil
}

func (h DataHandler) recreatedMasterRecordFromInteractions(
	conn aurora.DBConnection,
	domain string,
	uptId string,
	objectTypeNames []string,
	mappings []ObjectMapping,
	objectPriority []string,
) error {
	conn.Tx.Info("[recreatedMasterRecordFromInteractions] building profile from interactions")
	objectTypeNameToId := make(map[string]string, len(objectTypeNames))
	for _, objectType := range objectTypeNames {
		objectTypeNameToId[objectType] = ""
	}
	profile, lastOpdatedObjectTypes, lastUpdatedTimestamps, err := h.buildProfile(
		conn,
		domain,
		uptId,
		objectTypeNameToId,
		mappings,
		map[string]interface{}{},
		objectPriority,
		[]PaginationOptions{},
	)
	if err != nil {
		conn.Tx.Error("[recreatedMasterRecordFromInteractions] error building profile: %v", err)
		return fmt.Errorf("[recreatedMasterRecordFromInteractions] error building profile: %v", err)
	}

	conn.Tx.Debug("[recreatedMasterRecordFromInteractions]  Deleting existing record")
	_, err = conn.Query(h.buildDeleteSearchRecordSql(domain), profile.ProfileId)
	if err != nil {
		conn.Tx.Error("[recreatedMasterRecordFromInteractions] error deleting search profile: %v", err)
		return fmt.Errorf("[recreatedMasterRecordFromInteractions] error deleting search profile: %v", err)
	}

	conn.Tx.Debug("[recreatedMasterRecordFromInteractions] creating SQL query fr search row update")
	sql, args := h.buildSearchProfileInsertRowFromProfileSql(domain, profile, lastOpdatedObjectTypes, lastUpdatedTimestamps, mappings)
	_, err = conn.Query(sql, args...)
	if err != nil {
		conn.Tx.Error("[recreatedMasterRecordFromInteractions] error inserting search profile: %v", err)
		return fmt.Errorf("[recreatedMasterRecordFromInteractions] error inserting search profile: %v", err)
	}
	conn.Tx.Info("[recreatedMasterRecordFromInteractions] Successfully recreated search profile for %v", uptId)
	return nil
}

func (h DataHandler) buildDeleteSearchRecordSql(domain string) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE connect_id = ?`, searchIndexTableName(domain))
}

func (h DataHandler) buildSearchProfileInsertRowFromProfileSql(
	domain string,
	profile profilemodel.Profile,
	lastUpdatedObjectTypes map[string]string,
	lastUpdatedTimestamps map[string]time.Time,
	mappings []ObjectMapping,
) (string, []interface{}) {
	tableName := searchIndexTableName(domain)
	fields := []string{}
	params := []string{}
	args := []interface{}{}
	profileColumns := h.utils.mappingsToProfileDbColumns(mappings)

	fields = append(fields, `"connect_id"`)
	params = append(params, "?")
	args = append(args, profile.ProfileId)
	fields = append(fields, `"timestamp"`)
	params = append(params, "?")
	if profile.LastUpdated.IsZero() {
		args = append(args, time.Now())
	} else {
		args = append(args, profile.LastUpdated)
	}

	for _, fieldName := range profileColumns {
		// prevent override with empty string
		fieldValue := profile.Get(fieldName)
		if fieldValue == "" {
			//field empty. skiping it
			continue
		}

		// Append values (using parameterized query to handle user input values)
		fields = append(fields, `"`+fieldName+`"`)
		params = append(params, "?")
		args = append(args, fieldValue)
		fields = append(fields, `"`+fieldName+`_ts"`)
		params = append(params, "?")
		args = append(args, lastUpdatedTimestamps[fieldName])
		fields = append(fields, `"`+fieldName+`_obj_type"`)
		params = append(params, "?")
		args = append(args, lastUpdatedObjectTypes[fieldName])
	}
	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, tableName, strings.Join(fields, ","), strings.Join(params, ","))
	return sql, args
}

// Retrieves the unmerge history for profiles with ids - uptIds
func (h DataHandler) getUnmerges(conn aurora.DBConnection, domain string, uptIds []string) ([]ProfileHistory, error) {
	tableName := createProfileHistoryTableName(domain)

	// TODO: evaluate the need for timestamp in where clause
	var args []interface{}
	var paramPlaceholders []string
	for _, uptId := range uptIds {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, uptId)
	}
	args = append(args, args...)
	args = append(args, MergeTypeUnmerge)
	res, err := conn.Query(
		fmt.Sprintf(
			`SELECT to_merge, timestamp, merged_into FROM %s WHERE (merged_into IN (%s) OR to_merge IN (%s)) AND merge_type=?`,
			tableName,
			strings.Join(paramPlaceholders, ", "),
			strings.Join(paramPlaceholders, ", "),
		),
		args...,
	)
	if err != nil {
		h.Tx.Error("[getUnmerges] Unable to fetch unmerge history %s", err)
		return nil, err
	}

	profileHistory := []ProfileHistory{}
	for _, row := range res {
		profileHistory = append(profileHistory, ProfileHistory{
			toUnmergeUPTID:    row["to_merge"].(string),
			unmergedFromUPTID: row["merged_into"].(string),
			Timestamp:         row["timestamp"].(time.Time),
		})
	}
	return profileHistory, nil
}

// Gets the profileIds for profiles with originalConnectIds - 'uptIds' from the Master table
func GetIdFromMasterTableByOriginalConnectIdSql(masterTableName string, uptIds []string) (query string, args []interface{}) {
	var paramPlaceholders []string
	for _, uptId := range uptIds {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, uptId)
	}
	query = fmt.Sprintf(
		`SELECT profile_id, connect_id FROM %s WHERE original_connect_id IN (%v)`,
		masterTableName,
		strings.Join(paramPlaceholders, ", "),
	)
	return query, args
}

// Gets all profiles with connectId = uptId
func (h DataHandler) GetProfileIdRecords(conn aurora.DBConnection, domain string, uptId string) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(
		fmt.Sprintf(`SELECT profile_id FROM %s WHERE connect_id=? LIMIT ?`, domain),
		uptId,
		MAX_PROFILE_ID_PER_CONNECT_ID,
	)
}

// Gets all profiles with connectId = uptId
func (h DataHandler) GetAllFromMasterTable(masterTableName string, uptId string) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(fmt.Sprintf(`SELECT * FROM %s WHERE connect_id=?`, masterTableName), uptId)
}

// For all 'profileIds', updates the connect = uptIdToUnmerge in the interaction table
func unmergeInteractionsSql(
	domain string,
	uptIdToUnmerge string,
	profileIds []string,
	objectTypeName string,
	confidenceRevertFactor float64,
) (query string, args []interface{}) {
	var paramPlaceholders []string
	args = append(args, uptIdToUnmerge)
	args = append(args, confidenceRevertFactor)
	for _, profId := range profileIds {
		paramPlaceholders = append(paramPlaceholders, "?")
		args = append(args, profId)
	}
	query = fmt.Sprintf(`UPDATE %s
						SET connect_id=?,
						overall_confidence_score=overall_confidence_score / ?
						WHERE profile_id IN (%s)`,
		createInteractionTableName(domain, objectTypeName), strings.Join(paramPlaceholders, ", "))
	return query, args
}

// Given an interactionId, finds the profile it was originally created for
// (Optional) currentConnectId: id of the profile which currently has the interaction
// if currentConnectId is provided, it is used to narrow down the scope while searching for an interaction in the interaction table
// BuildInteractionHistory sends the currentConnectId
func (h DataHandler) getOriginalConnectIdForInteraction(
	conn aurora.DBConnection,
	domain, interactionId, objectTypeName string, currentConnectId string,
) (string, string, error) {
	// Get the profileId for the interactionId
	var query string
	var args []interface{}

	query = fmt.Sprintf(`
            SELECT profile_id, timestamp
            FROM %s
            WHERE accp_object_id=?`,
		createInteractionTableName(domain, objectTypeName))
	args = []interface{}{interactionId}

	if currentConnectId != "" {
		query += " AND connect_id=?"
		args = append(args, currentConnectId)
	}

	// sort by timestamp to get the latest one
	query += " ORDER BY timestamp DESC LIMIT 1"

	profileIdResult, err := conn.Query(query, args...)
	if err != nil {
		conn.Tx.Error("[getOriginalConnectIdForInteraction] Unable to find the interaction's original profileId. db error: %v", err)
		return "", "", err
	}

	if len(profileIdResult) == 0 {
		conn.Tx.Error("[getOriginalConnectIdForInteraction] Unable to find the interaction's profile")
		return "", "", errors.New("cannot find profile for the interaction")
	}

	profileId, ok := profileIdResult[0]["profile_id"].(string)

	if !ok {
		conn.Tx.Error("[getOriginalConnectIdForInteraction] Unable to find the interaction's profile: invalid 'profile_id' field")
		return "", "", errors.New("cannot find profile for the interaction: invalid 'profile_id' field")
	}

	// get the original connectId using profileId
	connectIdResult, err := h.AurSvc.Query(fmt.Sprintf(`
		SELECT original_connect_id
		FROM %s
		WHERE profile_id=?`,
		domain),
		profileId,
	)

	if err != nil || len(connectIdResult) != 1 {
		h.Tx.Error("[getOriginalConnectIdForInteraction] Unable to find the interaction's original connect_id. db error: %v", err)
		return "", "", err
	}

	connectId, ok := connectIdResult[0]["original_connect_id"].(string)

	if !ok {
		h.Tx.Error("[getOriginalConnectIdForInteraction] Unable to find the interaction's original connect_id. invalid 'original_connect_id' field")
		return "", "", err
	}
	return connectId, profileId, nil
}

// Starting at a profile (connectId), get all merges leading to its final parent
func (h DataHandler) getMergesLeadingToParent(conn aurora.DBConnection, domain, connectId string) ([]ProfileMergeContext, error) {
	tableName := createProfileHistoryTableName(domain)

	mergeHistory, err := conn.Query(fmt.Sprintf(`
		WITH RECURSIVE profile_history AS (
			SELECT timestamp, to_merge, merged_into, merge_type, confidence_update_factor, rule_id, ruleset_version, operator_id
			FROM %s WHERE to_merge = ?

			UNION ALL

			SELECT source.timestamp, source.to_merge, source.merged_into, source.merge_type, source.confidence_update_factor, source.rule_id, source.ruleset_version, source.operator_id
			FROM %s source
			JOIN profile_history target ON source.to_merge = target.merged_into
		)
		SELECT * FROM profile_history;
	`, tableName, tableName),
		connectId,
	)

	if err != nil {
		conn.Tx.Error("Error retrieving profile history %s", err)
		return nil, err
	}

	interactionHistory := []ProfileMergeContext{}
	for _, row := range mergeHistory {
		confidence, ok := row["confidence_update_factor"].(float64)
		if !ok {
			conn.Tx.Error("error casting confidence factors %v into float64", row["confidence_update_factor"])
			return nil, fmt.Errorf("error while computing profile lienage")
		}

		stringAttributes := []string{"merge_type", "rule_id", "ruleset_version", "operator_id", "to_merge", "merged_into"}
		attrbutesMap := map[string]string{}
		for _, attribute := range stringAttributes {
			if value, ok := row[attribute].(string); ok {
				attrbutesMap[attribute] = value
			} else {
				return interactionHistory, fmt.Errorf("[getMergesLeadingToParent] error asserting attribute: %s", attribute)
			}
		}
		timestamp, ok := row["timestamp"].(time.Time)
		if !ok {
			return interactionHistory, fmt.Errorf("[getMergesLeadingToParent] error asserting attribute: timestamp")
		}

		interactionHistory = append(interactionHistory, ProfileMergeContext{
			Timestamp:              timestamp,
			MergeType:              attrbutesMap["merge_type"],
			ConfidenceUpdateFactor: confidence,
			RuleID:                 attrbutesMap["rule_id"],
			RuleSetVersion:         attrbutesMap["ruleset_version"],
			OperatorID:             attrbutesMap["operator_id"],
			ToMergeConnectId:       attrbutesMap["to_merge"],
			MergeIntoConnectId:     attrbutesMap["merged_into"],
		})
	}
	return interactionHistory, nil
}

func (h DataHandler) retrieveInteractionCreation(
	conn aurora.DBConnection,
	domain string,
	objectTypeName string,
	interactionId string,
	profileId string,
) (ProfileMergeContext, error) {
	objectCreation, err := conn.Query(h.retrieveObjectHistorySql(domain, objectTypeName), interactionId, profileId)
	if err != nil || len(objectCreation) == 0 {
		h.Tx.Error("[retrieveObjectHistorySql] Unable to find the interaction %s", err)
		return ProfileMergeContext{}, err
	}

	return ProfileMergeContext{
		Timestamp:              objectCreation[0]["timestamp"].(time.Time),
		MergeType:              "INTERACTION_CREATED",
		ConfidenceUpdateFactor: 1,
	}, nil
}

func (h DataHandler) GetProfilePartition(domain, partitionLowerBound, partitionUpperBound string) ([]map[string]interface{}, error) {
	return h.AurSvc.Query(profilePartitionSql(domain), partitionLowerBound, partitionUpperBound)
}

func profilePartitionSql(domain string) string {
	return fmt.Sprintf(
		"SELECT DISTINCT connect_id FROM %v WHERE connect_id BETWEEN ? AND ?",
		domain,
	)
}

////////////////////////////////////////////////////////////////////////
// Dynamo Profile Table
////////////////////////////////////////////////////////////////////////

// creates the dynamoDB table to store profile for low latency access
func (h DataHandler) CreateProfileTable(domainName string, tags map[string]string) error {
	h.Tx.Info("Creating profile table in dynamoDB for %s", domainName)
	tName, pk, sk := buildDynamoProfileName(domainName)
	travellerGsi := db.GSI{
		IndexName:      DDB_GSI_NAME,
		PkName:         DDB_GSI_PK,
		PkType:         "S",
		ProjectionType: types.ProjectionTypeKeysOnly,
	}
	err := h.DynSvc.CreateTableWithOptions(tName, "connect_id", "record_type", db.TableOptions{
		Tags: tags,
		GSIs: []db.GSI{travellerGsi},
	})
	if err != nil {
		return err
	}
	h.DynSvc.TableName = tName
	h.DynSvc.PrimaryKey = pk
	h.DynSvc.SortKey = sk
	h.Tx.Debug("Waiting for dynamoDB table creation for %s", domainName)
	err = h.DynSvc.WaitForTableCreation()
	return err
}

func buildDynamoProfileName(domain string) (string, string, string) {
	name := "ucp_domain_" + domain
	pk := "connect_id"
	sk := "record_type"
	return name, pk, sk
}

func (h DataHandler) DeleteProfileTable(domainName string) error {
	tName, _, _ := buildDynamoProfileName(domainName)
	err := h.DynSvc.DeleteTable(tName)
	if err != nil {
		return err
	}
	h.DynSvc.TableName = tName
	err = h.DynSvc.WaitForTableDeletion()
	if err != nil {
		return err
	}
	return err
}
