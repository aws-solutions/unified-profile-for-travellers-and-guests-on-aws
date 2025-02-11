// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	controlPlaneCacheType_DomainConfig           = "domain_config"
	controlPlaneCacheType_Mappings               = "mappings"
	controlPlaneCacheType_ActiveStitchingRuleSet = "active_stitching_rule_set"
	controlPlaneCacheType_ActiveCacheRuleSet     = "active_cache_rule_set"
)

const cacheTtl = 60 * time.Second

func InitLcsCache() *cache.Cache {
	return cache.New(cacheTtl, cacheTtl)
}

func (c *ConfigHandler) storeInCache(domain string, cacheType string, data interface{}) error {
	c.Tx.Debug("[StoreInCache] saving %s for domain %s", cacheType, domain)
	err := c.cache.Add(cacheName(domain, cacheType), data, cacheTtl)
	if err != nil {
		c.Tx.Error("[StoreInCache] error saving %s for domain %s: %s", cacheType, domain, err)
	}
	return nil
}

func (c *ConfigHandler) invalidateCache(domain string, cacheType string) {
	c.Tx.Debug("[InvalidateCache] invalidating %s for domain %s", cacheType, domain)
	c.cache.Delete(cacheName(domain, cacheType))
}

func (c *ConfigHandler) retrieveFromCache(domain string, cacheType string) (data interface{}, found bool) {
	c.Tx.Debug("[RetrieveFromCache] retrieving %s for domain %s", cacheType, domain)
	data, found = c.cache.Get(cacheName(domain, cacheType))
	return data, found
}

func cacheName(domain, cacheType string) string {
	return domain + "-" + cacheType
}
