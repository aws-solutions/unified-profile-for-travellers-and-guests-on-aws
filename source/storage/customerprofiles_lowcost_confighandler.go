// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"tah/upt/source/tah-core/cloudwatch"
	core "tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/ucp-common/src/feature"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type ConfigHandler struct {
	Svc          *db.DBConfig
	Tx           core.Transaction
	MetricLogger *cloudwatch.MetricLogger
	utils        LCSUtils
	cache        *cache.Cache
}

func (c *ConfigHandler) SetTx(tx core.Transaction) {
	tx.LogPrefix = "customerprofiles-lc-ch"
	c.Tx = tx
	c.Svc.SetTx(tx)
	c.utils.SetTx(tx)
}

// Domain Config
func (h ConfigHandler) CreateDomainConfig(cfg LcDomainConfig) error {
	dyn := DynamoLcDomainConfig{
		Pk:                   cfg.Name,
		Sk:                   DOMAIN_CONFIG_SK,
		Options:              cfg.Options,
		AiIdResolutionBucket: cfg.AiIdResolutionBucket,
		QueueUrl:             cfg.QueueUrl,
		KmsArn:               cfg.KmsArn,
		EnvName:              cfg.EnvName,
		CustomerProfileMode:  cfg.CustomerProfileMode,
		DynamoMode:           cfg.DynamoMode,
		DomainVersion:        cfg.Options.FeatureSetVersion,
		CompatibleVersion:    cfg.Options.FeatureSetVersion,
	}
	_, err := h.Svc.Save(dyn)
	return err
}

func (h ConfigHandler) DeleteDomainConfig(domainName string) error {
	_, err := h.Svc.Delete(map[string]string{configTablePk: domainName, configTableSk: DOMAIN_CONFIG_SK})
	h.invalidateCache(domainName, controlPlaneCacheType_DomainConfig)
	return err
}

func (h ConfigHandler) GetDomainConfig(domainName string) (LcDomainConfig, error) {
	// Check for cached config
	cachedConfig, found := h.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if found {
		h.Tx.Debug("[GetDomainConfig] cache hit")
		if domainConfig, ok := cachedConfig.(LcDomainConfig); ok {
			return domainConfig, nil
		}
		h.Tx.Warn("[GetDomainConfig] unable to read cached domain config")
	} else {
		h.Tx.Debug("[GetDomainConfig] cache miss")
	}

	// Retrieve domain config directly from DynamoDB
	h.Tx.Debug("[GetDomainConfig] retrieving domain config from DynamoDB")
	var dyn DynamoLcDomainConfig
	err := h.Svc.Get(domainName, DOMAIN_CONFIG_SK, &dyn)
	if err != nil {
		h.Tx.Error("[GetDomainConfig] unable to retrieve domain config from DynamoDB: %v", err)
		return LcDomainConfig{}, err
	}
	if dyn.Pk == "" {
		return LcDomainConfig{}, ErrDomainNotFound
	}
	cfg := LcDomainConfig{
		Name:                    dyn.Pk,
		Options:                 dyn.Options,
		AiIdResolutionBucket:    dyn.AiIdResolutionBucket,
		QueueUrl:                dyn.QueueUrl,
		KmsArn:                  dyn.KmsArn,
		EnvName:                 dyn.EnvName,
		CustomerProfileMode:     dyn.CustomerProfileMode,
		DynamoMode:              dyn.DynamoMode,
		AiIdResolutionGlueTable: dyn.AiIdResolutionGlueTable,
		DomainVersion:           dyn.DomainVersion,
		CompatibleVersion:       dyn.CompatibleVersion,
	}
	// for domains created before versioning was introduced, we set the version attributes to initial feature set version
	if dyn.DomainVersion == 0 {
		cfg.DomainVersion = feature.INITIAL_FEATURE_SET_VERSION
	}
	if dyn.CompatibleVersion == 0 {
		cfg.CompatibleVersion = feature.INITIAL_FEATURE_SET_VERSION
	}
	if dyn.CompatibleVersion > feature.LATEST_FEATURE_SET_VERSION {
		// this domain potentially has changes that we do not know how to handle
		return LcDomainConfig{}, fmt.Errorf(
			"domain version too new: %d, newest version is %d",
			dyn.CompatibleVersion,
			feature.LATEST_FEATURE_SET_VERSION,
		)
	}

	// Cache domain if it exists, GetDomainConfig() is used to check if a domain exists, and may be called
	if cfg.Name != "" {
		err = h.storeInCache(domainName, controlPlaneCacheType_DomainConfig, cfg)
		if err != nil {
			h.Tx.Warn("[GetDomainConfig] unable to store domain config in cache: %v", err)
		}
	}
	return cfg, nil
}

func (h ConfigHandler) SaveDomainConfig(domainConfig LcDomainConfig) error {
	dynamoConfig := domainConfigToDynamo(domainConfig)
	_, err := h.Svc.Save(dynamoConfig)
	if err != nil {
		return err
	}
	h.invalidateCache(domainConfig.Name, controlPlaneCacheType_DomainConfig)
	return nil
}

func domainConfigToDynamo(cfg LcDomainConfig) DynamoLcDomainConfig {
	return DynamoLcDomainConfig{
		Pk:                      cfg.Name,
		Sk:                      DOMAIN_CONFIG_SK,
		Options:                 cfg.Options,
		AiIdResolutionBucket:    cfg.AiIdResolutionBucket,
		QueueUrl:                cfg.QueueUrl,
		KmsArn:                  cfg.KmsArn,
		EnvName:                 cfg.EnvName,
		AiIdResolutionGlueTable: cfg.AiIdResolutionGlueTable,
		CustomerProfileMode:     cfg.CustomerProfileMode,
		DynamoMode:              cfg.DynamoMode,
		DomainVersion:           cfg.DomainVersion,
		CompatibleVersion:       cfg.CompatibleVersion,
	}
}

func (h ConfigHandler) ListDomainConfigs() ([]LcDomainConfig, error) {
	var items []DynamoLcDomainConfig
	_, err := h.Svc.FindAll(&items, nil)
	if err != nil {
		return nil, err
	}

	domainConfigs := []LcDomainConfig{}
	for _, item := range items {
		if item.Sk == DOMAIN_CONFIG_SK {
			cfg := LcDomainConfig{
				Name:                 item.Pk,
				Options:              item.Options,
				AiIdResolutionBucket: item.AiIdResolutionBucket,
				QueueUrl:             item.QueueUrl,
				KmsArn:               item.KmsArn,
				EnvName:              item.EnvName,
			}
			domainConfigs = append(domainConfigs, cfg)
		}
	}

	return domainConfigs, nil
}

func (h ConfigHandler) DomainConfigExists(domainName string) (LcDomainConfig, bool) {
	res, err := h.GetDomainConfig(domainName)
	return res, err == nil && res.Name != ""
}

// Mappings
func (h ConfigHandler) SaveMapping(domain string, objectMapping ObjectMapping) error {
	dynMappings := mappingToDynamo(domain, objectMapping)
	_, err := h.Svc.Save(dynMappings)
	if err != nil {
		return err
	}
	h.invalidateCache(domain, controlPlaneCacheType_Mappings)
	return nil
}

func mappingToDynamo(domain string, objectMapping ObjectMapping) DynamoObjectMapping {
	return DynamoObjectMapping{
		Pk:            domain,
		Sk:            DOMAIN_MAPPING_PREFIX + objectMapping.Name,
		Name:          objectMapping.Name,
		Description:   objectMapping.Description,
		FieldMappings: objectMapping.Fields,
	}
}

func (h ConfigHandler) ListMappings(domain string) ([]ObjectMapping, error) {
	// Check for cached config
	cachedMappings, found := h.retrieveFromCache(domain, controlPlaneCacheType_Mappings)
	if found {
		h.Tx.Debug("[ListMappings] cache hit")
		if mappings, ok := cachedMappings.([]ObjectMapping); ok {
			return mappings, nil
		}
		h.Tx.Warn("[ListMappings] unable to read cached mappings")
	} else {
		h.Tx.Debug("[ListMappings] cache miss")
	}

	// Retrieve mappings directly from DynamoDB
	h.Tx.Debug("[ListMappings] retrieving mappings from DynamoDB")
	var dynMappings []DynamoObjectMapping
	var objMappings []ObjectMapping
	err := h.Svc.FindStartingWith(domain, DOMAIN_MAPPING_PREFIX, &dynMappings)
	for _, dynMapping := range dynMappings {
		objMappings = append(objMappings, ObjectMapping{
			Name:        dynMapping.Name,
			Fields:      dynMapping.FieldMappings,
			Description: dynMapping.Description,
		})
	}
	if err != nil && errors.Is(err, net.ErrClosed) {
		err = NewRetryable(err)
	}
	if err != nil {
		h.Tx.Error("[ListMappings] unable to retrieve mappings from DynamoDB: %v", err)
		return nil, err
	}
	err = h.storeInCache(domain, controlPlaneCacheType_Mappings, objMappings)
	if err != nil {
		h.Tx.Warn("[ListMappings] unable to store mappings in cache: %v", err)
	}
	return objMappings, nil
}

func (h ConfigHandler) DeleteMapping(domain string, mappingName string) error {
	_, err := h.Svc.Delete(map[string]string{configTablePk: domain, configTableSk: DOMAIN_MAPPING_PREFIX + mappingName})
	h.invalidateCache(domain, controlPlaneCacheType_Mappings)
	return err
}

func domainCfgToDomain(cfg LcDomainConfig) Domain {
	tags := map[string]string{
		DOMAIN_CONFIG_ENV_KEY: cfg.EnvName,
	}
	var mode CacheMode
	if cfg.CustomerProfileMode {
		mode = mode | CUSTOMER_PROFILES_MODE
	}
	if cfg.DynamoMode {
		mode = mode | DYNAMO_MODE
	}
	return Domain{
		Name:                cfg.Name,
		MatchingEnabled:     cfg.Options.AiIdResolutionOn,
		RuleBaseStitchingOn: cfg.Options.RuleBaseIdResolutionOn,
		Tags:                tags,
		IsLowCostEnabled:    true,
		CacheMode:           mode,
	}
}

// this function retrieves the column names from the domain mapping.
// it is subsequently used to create the Aurora table and the entity resolution schema
func (h ConfigHandler) GetMasterColumns(domain string) ([]string, error) {
	h.Tx.Debug("[GetMasterColumns] fetching mappings for domain %s", domain)
	mappings, err := h.ListMappings(domain)
	if err != nil {
		h.Tx.Error("[GetMasterColumns] error while fetching mappings: %v", err)
		return nil, err
	}
	h.Tx.Debug("[GetMasterColumns] found %d mappings", len(mappings))
	return h.utils.mappingsToProfileDbColumns(mappings), err
}

func (h ConfigHandler) GetObjectLevelColumns(domain string) (map[string][]string, error) {
	h.Tx.Debug("[GetMasterColumns] fetching mappings for domain %s", domain)
	mappings, err := h.ListMappings(domain)
	if err != nil {
		h.Tx.Error("[GetMasterColumns] error while fetching mappings: %v", err)
		return nil, err
	}
	h.Tx.Debug("[GetMasterColumns] found %d mappings", len(mappings))
	_, objectMappings := h.utils.MappingsToDbColumns(mappings)
	return objectMappings, err
}

// RuleSet
func (h ConfigHandler) createRuleSet(domainName string, ruleSet RuleSet, prefix string) error {
	dynamoRules := ruleSetToDynamoRules(domainName, ruleSet, prefix)
	return h.Svc.SaveMany(dynamoRules)
}

func (h ConfigHandler) deleteRuleSet(domainName, ruleSetType string, prefix string) error {
	dynamoRules := []DynamoRule{}
	forDelete := []map[string]string{}
	err := h.Svc.FindStartingWith(domainName, prefix+ruleSetType, &dynamoRules)
	if err != nil {
		return err
	}
	for _, rule := range dynamoRules {
		forDelete = append(forDelete, map[string]string{"pk": rule.Pk, "sk": rule.Sk})
	}
	return h.Svc.DeleteMany(forDelete)
}

func (h ConfigHandler) DeleteAllRuleSetsByDomain(domainName string) error {
	// Remove from caches
	h.invalidateCache(domainName, controlPlaneCacheType_ActiveStitchingRuleSet)
	h.invalidateCache(domainName, controlPlaneCacheType_ActiveCacheRuleSet)

	// Remove from DynamoDB
	dynamoRules := []DynamoRule{}
	dynamoCacheRules := []DynamoRule{}
	forDelete := []map[string]string{}
	err := h.Svc.FindStartingWith(domainName, RULE_SK_PREFIX, &dynamoRules)
	if err != nil {
		return err
	}
	for _, rule := range dynamoRules {
		forDelete = append(forDelete, map[string]string{"pk": rule.Pk, "sk": rule.Sk})
	}
	err = h.Svc.FindStartingWith(domainName, RULE_SK_CACHE_PREFIX, &dynamoCacheRules)
	if err != nil {
		return err
	}
	for _, rule := range dynamoCacheRules {
		forDelete = append(forDelete, map[string]string{"pk": rule.Pk, "sk": rule.Sk})
	}
	return h.Svc.DeleteMany(forDelete)
}

// Retrieve a rule set by combination of prefix + type
// e.g. rule_ + draft or cache_ + active
func (h ConfigHandler) getRuleSet(domainName, ruleSetType string, prefix string) (RuleSet, error) {
	// Active rule sets (both stitching and caching) are on the critical path, we want to cache them for best performance.
	// For draft rule sets, we purposely do not cache. It could have unexpected consequences since ActivateRuleSet() reads the latest draft
	var cacheType string
	if ruleSetType == RULE_SET_TYPE_ACTIVE {
		if prefix == RULE_SK_PREFIX {
			cacheType = controlPlaneCacheType_ActiveStitchingRuleSet
		} else if prefix == RULE_SK_CACHE_PREFIX {
			cacheType = controlPlaneCacheType_ActiveCacheRuleSet
		} else {
			return RuleSet{}, fmt.Errorf("invalid prefix: %s", prefix)
		}
		h.Tx.Debug("[getRuleSet] fetching rule set: %s", cacheType)
		cachedRuleSet, found := h.retrieveFromCache(domainName, cacheType)
		if found {
			h.Tx.Debug("[getRuleSet] cache hit")
			if ruleSet, ok := cachedRuleSet.(RuleSet); ok {
				return ruleSet, nil
			}
			h.Tx.Warn("[getRuleSet] unable to read cached rule set")
		} else {
			h.Tx.Debug("[getRuleSet] cache miss")
		}
	}

	var dynamoRules []DynamoRule
	ruleSet := RuleSet{
		Name:  ruleSetType,
		Rules: []Rule{},
	}
	err := h.Svc.FindStartingWith(domainName, prefix+ruleSetType, &dynamoRules)
	if err != nil {
		return ruleSet, err
	}
	if len(dynamoRules) == 0 {
		h.Tx.Debug("[getRuleSet] rule set has 0 rules, caching empty rule set: %s", cacheType)
		err = h.storeInCache(domainName, cacheType, ruleSet)
		if err != nil {
			h.Tx.Warn("[getRuleSet] unable to store rule set %s in cache: %v", cacheType, err)
		}
		return ruleSet, nil
	}

	rules := dynamoRulesToRules(dynamoRules)
	ruleSet = RuleSet{
		Name:                    ruleSetType,
		Rules:                   rules,
		LatestVersion:           dynamoRules[0].Latest,
		LastActivationTimestamp: dynamoRules[0].LastActivationTimestamp,
	}

	if ruleSetType == RULE_SET_TYPE_ACTIVE {
		h.Tx.Debug("[getRuleSet] caching active rule set: %s", cacheType)
		err = h.storeInCache(domainName, cacheType, ruleSet)
		if err != nil {
			h.Tx.Warn("[getRuleSet] unable to store rule set %s in cache: %v", cacheType, err)
		}
	}
	return ruleSet, nil
}

func (h ConfigHandler) GetAllRuleSets(domainName string, prefix string) ([]RuleSet, error) {
	dynamoRules := []DynamoRule{}
	ruleSets := []RuleSet{}
	err := h.Svc.FindStartingWith(domainName, prefix, &dynamoRules)
	if err != nil {
		return ruleSets, err
	}

	ruleMap := make(map[string]RuleSet)
	for _, dynamoRule := range dynamoRules {
		skSplit := strings.Split(dynamoRule.Sk, "_")
		ruleSetName := skSplit[1]
		ruleSet := ruleMap[ruleSetName]
		if ruleSet.Rules == nil {
			newRuleSet := RuleSet{
				Name:                    ruleSetName,
				LatestVersion:           dynamoRule.Latest,
				LastActivationTimestamp: dynamoRule.LastActivationTimestamp,
				Rules:                   []Rule{dynamoRule.Rule},
			}
			ruleMap[ruleSetName] = newRuleSet
		} else {
			ruleSet := ruleMap[ruleSetName]
			ruleSet.Rules = append(ruleSet.Rules, dynamoRule.Rule)
			ruleMap[ruleSetName] = ruleSet
		}
	}
	keys := make([]string, 0, len(ruleMap))
	for key := range ruleMap {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return ruleMap[keys[i]].LastActivationTimestamp.Before(ruleMap[keys[j]].LastActivationTimestamp)
	})

	for _, k := range keys {
		ruleSets = append(ruleSets, ruleMap[k])
	}
	return ruleSets, nil
}

func ruleSetToDynamoRules(domainName string, ruleSet RuleSet, prefix string) []DynamoRule {
	dynamoRules := []DynamoRule{}
	for _, rule := range ruleSet.Rules {
		dynamoRule := DynamoRule{
			Pk:                      domainName,
			Sk:                      prefix + ruleSet.Name + "_" + fmt.Sprintf("%05d", rule.Index),
			Rule:                    rule,
			LastActivationTimestamp: ruleSet.LastActivationTimestamp,
		}
		if ruleSet.Name == RULE_SET_TYPE_ACTIVE {
			dynamoRule.Latest = ruleSet.LatestVersion
			dynamoRule.LastActivationTimestamp = time.Now().UTC()
		}
		dynamoRules = append(dynamoRules, dynamoRule)
	}
	return dynamoRules
}

func dynamoRulesToRules(dynamoRules []DynamoRule) []Rule {
	rules := []Rule{}
	for _, rule := range dynamoRules {
		rules = append(rules, rule.Rule)
	}
	return rules
}
