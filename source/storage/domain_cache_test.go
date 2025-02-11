// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package customerprofileslcs

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Parallel()
	handler := ConfigHandler{cache: InitLcsCache()}
	domainName := "test_cache"

	_, found := handler.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if found {
		t.Errorf("Cache should not have found the domain config")
	}

	err := handler.storeInCache(domainName, controlPlaneCacheType_DomainConfig, LcDomainConfig{Name: domainName})
	if err != nil {
		t.Errorf("Error storing in cache: %v", err)
	}

	cachedDomainConfig, found := handler.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if !found {
		t.Errorf("Cache should have found the domain config")
	}
	if config, ok := cachedDomainConfig.(LcDomainConfig); ok {
		if config.Name != domainName {
			t.Errorf("Domain config name should be %s, but got %s", domainName, config.Name)
		}
	} else {
		t.Errorf("Cached domain config should be of type LcDomainConfig")
	}

	handler.invalidateCache(domainName, controlPlaneCacheType_DomainConfig)

	_, found = handler.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if found {
		t.Errorf("Cache should not have found the domain config after invalidation")
	}
}

func TestCacheExpiration(t *testing.T) {
	t.Parallel()
	handler := ConfigHandler{cache: InitLcsCache()}
	domainName := "test_cache_expiration"

	err := handler.storeInCache(domainName, controlPlaneCacheType_DomainConfig, LcDomainConfig{Name: domainName})
	if err != nil {
		t.Errorf("Error storing in cache: %v", err)
	}
	_, found := handler.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if !found {
		t.Errorf("Cache should have found the domain config")
	}

	time.Sleep(cacheTtl + 1*time.Second)

	_, found = handler.retrieveFromCache(domainName, controlPlaneCacheType_DomainConfig)
	if found {
		t.Errorf("Cache should not have found the domain config after expiration")
	}
}
