// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"tah/upt/source/tah-core/cloudwatch"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/ucp-common/src/utils/utils"
	"tah/upt/source/ucp-cp-writer/src/usecase"
	"time"
)

//Create domain

func (c *CustomerProfileConfigLowCost) CreateCPCache(
	domainName string,
	tags map[string]string,
	matchS3Name string,
	cpDomainParams CpDomainParams,
) error {
	c.Tx.Info("CP mode is on: Creating domain in CP mode")
	err := c.CustomerProfileConfig.CreateDomainWithQueue(
		domainName,
		cpDomainParams.kmsArn,
		tags,
		cpDomainParams.queueUrl,
		matchS3Name,
		customerprofiles.DomainOptions{},
	)
	if err != nil {
		c.Tx.Error("[CreateCPDomain] error creating domain: %v", err)
		return err
	}
	c.Tx.Debug("CP mode is on: Creating event stream %v", cpDomainParams.streamArn)
	err = c.CustomerProfileConfig.CreateEventStream(domainName, domainName, cpDomainParams.streamArn)
	// create event stream
	if err != nil {
		c.Tx.Error("[CreateCPDomain] error creating event stream: %v", err)
		return err
	}
	profileFields := customerprofiles.BuildProfileFieldMapping()
	_, err = utils.Retry(func() (interface{}, error) {
		err = c.CustomerProfileConfig.CreateMapping(
			customerprofiles.PROFILE_FIELD_OBJECT_TYPE,
			"Object where all fields map to profile level fields",
			profileFields,
		)
		return nil, err
	}, CP_CREATE_MAPPING_MAX_RETRY, CP_CREATE_MAPPING_INTERVAL, CP_CREATE_MAPPING_MAX_INTERVAL)
	if err != nil {
		c.Tx.Error("[CreateCPDomain] error creating cp mapping: %v", err)
		return err
	}
	// wait for mapping to exist before mappings can use it
	delay := time.Second
	maxDelay := 60 * time.Second
	maxIterations := 10
	var mapping customerprofiles.ObjectMapping
	for i := 0; i < maxIterations; i++ {
		time.Sleep(delay)
		mapping, err = c.CustomerProfileConfig.GetMapping(customerprofiles.PROFILE_FIELD_OBJECT_TYPE)
		if err == nil && mapping.Name == customerprofiles.PROFILE_FIELD_OBJECT_TYPE {
			break
		}
		delay = min(2*delay+(time.Duration(rand.Intn(1000))*time.Millisecond), maxDelay)
	}
	if err != nil {
		c.Tx.Error("[CreateCPDomain] error waiting for profile field mapping to become available: %v", err)
		return errors.Join(errors.New("error waiting for profile field mapping to become available"), err)
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) CreateDynamoCache(domainName string, tags map[string]string) error {
	c.Tx.Info("Dynamo Mode is on. Creating profile table")
	err := c.Data.CreateProfileTable(domainName, tags)
	if err != nil {
		c.Tx.Error("Error creating profile table: %v. cleaning up domain data", err)
		return err
	}
	return nil
}

//Delete domain

func (c *CustomerProfileConfigLowCost) DeleteCPCacheByName(domainName string) error {
	c.Tx.Info("Deleting CP domain")
	return c.CustomerProfileConfig.DeleteDomainByName(domainName)
}

func (c *CustomerProfileConfigLowCost) DeleteDynamoCacheByName(domainName string) error {
	c.Tx.Info("Deleting dynamo domain")
	err := c.Data.DeleteProfileTable(domainName)
	if err != nil {
		c.Tx.Error("Error while deleting profile table: %v", err)
		return err
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) DeleteDomainProfileStore(domainName string, cpMode, dynamoMode bool) error {
	var errs error
	if cpMode {
		err := c.DeleteCPCacheByName(domainName)
		if err != nil {
			c.Tx.Warn("Error deleting domain on CP side: %v. continuing domain deletion", err)
			errs = errors.Join(err, errs)
		}
	}
	if dynamoMode {
		err := c.DeleteDynamoCacheByName(domainName)
		if err != nil {
			c.Tx.Warn("Error deleting domain on Dynamo side: %v", err)
			errs = errors.Join(err, errs)
		}
	}
	if errs != nil {
		return fmt.Errorf("error deleting, %v", errs)
	}
	return nil
}

//Mapping in cp domain for profile objects

func (c *CustomerProfileConfigLowCost) CreateCPDomainMapping(domainName string, objectMapping ObjectMapping) error {
	if c.DomainConfig.CustomerProfileMode {
		c.CustomerProfileConfig.DomainName = domainName

		var fieldMappings customerprofiles.FieldMappings
		if c.features.SupportsCustomerProfilesTravelerIdMapping() {
			fieldMappings = c.BuildObjectFieldMappingWithTravellerId(objectMapping)
		} else {
			fieldMappings = customerprofiles.BuildGenericIntegrationFieldMapping()
		}
		err := c.CustomerProfileConfig.CreateMapping(objectMapping.Name, objectMapping.Description, fieldMappings)
		if err != nil {
			c.Tx.Error("[CreateCPDomainMapping] error putting mappings: %v", err)
			return err
		}

		fields := customAttributeFieldsNameFromMapping(objectMapping)
		if len(fields) > 0 {
			err := c.CustomerProfileConfig.UpdateProfileAttributesFieldMapping(
				customerprofiles.PROFILE_FIELD_OBJECT_TYPE,
				"Object where all fields map to profile level fields",
				fields,
			)
			if err != nil {
				c.Tx.Error("[CreateCPDomainMapping] error updating profile attributes field mapping: %v", err)
				return err
			}
		}
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) BuildObjectFieldMappingWithTravellerId(objectMapping ObjectMapping) customerprofiles.FieldMappings {
	startingMapping := customerprofiles.BuildGenericIntegrationFieldMappingWithTravelerId()
	for _, fieldMapping := range objectMapping.Fields {
		if len(fieldMapping.Indexes) == 0 && !strings.Contains(fieldMapping.Target, "_profile.") {
			fieldMapping.Indexes = nil
			newField := customerprofiles.FieldMapping{
				Type:    fieldMapping.Type,
				Source:  fieldMapping.Source,
				Target:  fieldMapping.Target,
				Indexes: fieldMapping.Indexes,
				KeyOnly: true,
			}
			startingMapping = append(startingMapping, newField)
		}
	}
	return startingMapping
}

func (c *CustomerProfileConfigLowCost) UpdateProfileInCustomerProfilesCache(
	profile profilemodel.Profile,
	domain string,
	profObjects []profilemodel.ProfileObject,
) error {
	c.Tx.Info("[UpdateCustomerProfileDomain] Updating profile %s in customer profile domain", profile.ProfileId)
	c.Tx.Debug("[UpdateCustomerProfileDomain] Updating customer profile records for profile with connectID %+v", profile.ProfileId)

	startTime := time.Now()

	// Send trimmed profile to CP Writer
	profile.ProfileObjects = []profilemodel.ProfileObject{}
	putProfileRq := usecase.PutProfileRq{
		Profile: profile,
	}
	msg, err := json.Marshal(putProfileRq)
	if err != nil {
		c.Tx.Error("[UpdateCustomerProfileDomain] Error marshaling profile: %v", err)
		return err
	}
	c.CPWriterClient.SendWithStringAttributes(string(msg), map[string]string{
		"request_type": usecase.CPWriterRequestType_PutProfile,
		"domain":       domain,
		"txid":         c.Tx.TransactionID,
	})

	// Send profile objects to CP Writer
	var allErrors error
	for _, obj := range profObjects {
		prepForCP(&obj)
		putProfileObjectRq := usecase.PutProfileObjectRq{
			ProfileID:     profile.ProfileId,
			ProfileObject: obj,
		}
		msg, err := json.Marshal(putProfileObjectRq)
		if err != nil {
			c.Tx.Error("[UpdateCustomerProfileDomain] Error marshaling profile object: %v", err)
			allErrors = errors.Join(allErrors, err)
			continue
		}
		err = c.CPWriterClient.SendWithStringAttributes(string(msg), map[string]string{
			"request_type": usecase.CPWriterRequestType_PutProfileObject,
			"domain":       domain,
			"txid":         c.Tx.TransactionID,
		})
		if err != nil {
			c.Tx.Error("[UpdateCustomerProfileDomain] Error sending profile object to CP Writer: %v", err)
			allErrors = errors.Join(allErrors, err)
		}
	}

	duration := time.Since(startTime)
	c.MetricLogger.LogMetric(
		map[string]string{"domain": domain},
		METRIC_NAME_CP_UPDATE_LATENCY,
		cloudwatch.Milliseconds,
		float64(duration.Milliseconds()),
	)
	c.Tx.Debug("SaveMany duration:  %v", duration)

	return allErrors
}

// Apply changes for objects before being sent to Amazon Connect Customer Profiles
func prepForCP(obj *profilemodel.ProfileObject) {
	// Base64 decode the conversation object
	decodeConversation(obj)
}

// Decode the base64 encoded conversation object, so it can be displayed in a meaningful way in Amazon Connect.
func decodeConversation(obj *profilemodel.ProfileObject) {
	if obj.Type != "customer_service_interaction" {
		return
	}
	conversation, ok := obj.Attributes["conversation"]
	if !ok || conversation == "" {
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(conversation)
	if err != nil {
		return
	}
	obj.AttributesInterface["conversation"] = string(decoded)
}

// generic function to update downstream cache, customer profiles or dynamo
func (c *CustomerProfileConfigLowCost) UpdateDownstreamCache(
	profile profilemodel.Profile,
	objectIdToUpdate string,
	cacheMode CacheMode,
	profObjects []profilemodel.ProfileObject,
) error {
	var errs error
	tx := core.NewTransaction(c.Tx.TransactionID, "", c.Tx.LogLevel)
	tx.Info("[UpdateDownstreamCache] Getting cache skip rules from dynamo")
	activeRuleSet, err := c.Config.getRuleSet(c.DomainName, RULE_SET_TYPE_ACTIVE, RULE_SK_CACHE_PREFIX)
	if err != nil {
		c.Tx.Warn("[UpdateDownstreamCache] Error while retrieving active cache skip rule set: %v", err)
		return err
	}
	if len(activeRuleSet.Rules) > 0 {
		tx.Debug("[UpdateDownstreamCache] found %d cache skip rules", len(activeRuleSet.Rules))
		for _, rule := range activeRuleSet.Rules {
			tx.Debug("[UpdateDownstreamCache] checking rule %s (id:%s)", rule.Name, rule.Index)
			allConditions := true
			if len(rule.Conditions) == 0 {
				continue
			}
			for _, condition := range rule.Conditions {
				tx.Debug("[UpdateDownstreamCache] checking condition %+v on profile %s", condition, profile.ProfileId)
				pdm := prepareProfile(profile)
				tx.Debug("[UpdateDownstreamCache] profile data map: %+v", pdm)
				skip := c.IR.skipConditionApplies(condition, pdm, PROFILE_OBJECT_TYPE_NAME, map[string]string{})
				if skip {
					tx.Debug("[UpdateDownstreamCache] condition %s matches on profile data", condition.Index)
				}
				allConditions = allConditions && skip
			}
			if allConditions {
				c.Tx.Debug(
					"[UpdateDownstreamCache] Skipping cache update (rule '%s' (id: %s)) as all conditions are met to skip profile id %s",
					rule.Name,
					rule.Index,
					profile.ProfileId,
				)
				return nil
			} else {
				tx.Info("[UpdateDownstreamCache] conditions not met for rule '%s' (id: %s) on profile %s continuing", rule.Name, rule.Index, profile.ProfileId)
			}
		}
	} else {
		tx.Debug("[UpdateDownstreamCache] no skip rule activated fo profile %s. continuing cache ingestion", profile.ProfileId)
	}
	dynamoMode := cacheMode&DYNAMO_MODE == DYNAMO_MODE
	cpMode := cacheMode&CUSTOMER_PROFILES_MODE == CUSTOMER_PROFILES_MODE
	var wg sync.WaitGroup
	var mu sync.Mutex
	if cpMode {
		tx.Info("[UpdateDownstreamCache] CP mode is enabled. Updating profile %s in CP Cache", profile.ProfileId)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.UpdateProfileInCustomerProfilesCache(profile, c.DomainName, profObjects)
			if err != nil {
				c.Tx.Warn("Error while updating profile object in customer profiles: %v", err)
				mu.Lock()
				defer mu.Unlock()
				errs = errors.Join(err, errs)
			}
		}()
	}
	if dynamoMode {
		tx.Info("[UpdateDownstreamCache] DynamoDB mode is enabled. Updating profile %s DynamoDB Cache", profile.ProfileId)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.Data.UpdateProfileLowLatencyStore(profile, c.DomainName, objectIdToUpdate, profObjects)
			if err != nil {
				c.Tx.Warn("Error while updating profile object in dynamoDB: %v", err)
				mu.Lock()
				defer mu.Unlock()
				errs = errors.Join(err, errs)
			}
		}()
	}
	wg.Wait()
	if errs != nil {
		return fmt.Errorf("[UpdateDownstreamCache] error updating profile %s in one of the caches, %v", profile.ProfileId, errs)
	}
	return nil
}

func (c *CustomerProfileConfigLowCost) DeleteProfileFromCustomerProfileCache(uptId string) error {
	deleteProfileRq := usecase.DeleteProfileRq{
		ProfileID:  uptId,
		NumRetries: 2,
	}
	msg, err := json.Marshal(deleteProfileRq)
	if err != nil {
		c.Tx.Error("[DeleteProfileFromCustomerProfileCache] Error marshalling DeleteProfileRq: %v", err)
		return err
	}

	err = c.CPWriterClient.SendWithStringAttributes(string(msg), map[string]string{
		"request_type": usecase.CPWriterRequestType_DeleteProfile,
		"domain":       c.DomainName,
		"txid":         c.Tx.TransactionID,
	})
	if err != nil {
		c.Tx.Error("[DeleteProfileFromCustomerProfileCache] Error sending DeleteProfileRq to CP Writer: %v", err)
		return err
	}

	return nil
}
