// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"log"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"time"
)

func (h UptHandler) GetAsyncEventStatus(asyncEventId, asyncEventUsecase string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	res, err := h.Client.HttpGet(map[string]string{
		"id":      asyncEventId,
		"usecase": asyncEventUsecase,
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "async", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) CreateDomain(name string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	res, err := h.Client.HttpPost(model.RequestWrapper{
		Domain: model.Domain{
			Name: name,
		},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "admin", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) GetDomainConfig(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "admin/" + domainName, Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) DeleteDomain(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	res, err := h.Client.HttpDelete(domainName, map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "admin", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) RunAllJobs(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		StartJobRq: model.StartJobRq{JobName: "run-all-jobs"},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "jobs", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) GetJobsStatus(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "jobs", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) SaveIdResRuleSet(domainName string, rules []customerprofileslcs.Rule) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		SaveRuleSet: model.SaveRuleSetRq{
			Rules: rules,
		},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "ruleSet", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) ActivateIdResRuleSet(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "ruleSet/activate", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) ActivateIdResRuleSetWithRetries(domainName string, attempts int) (model.ResponseWrapper, error) {
	log.Printf("[ActivateIdResRuleSetWithRetries] Running in retry mode")
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	var res interface{}
	var err error
	duration := 60 * time.Second
	for i := 0; i < attempts; i++ {
		log.Printf("[ActivateIdResRuleSetWithRetries] Attempt #%v", i)
		res, err = h.Client.HttpPost(model.RequestWrapper{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "ruleSet/activate", Headers: authHeader})
		if err == nil {
			log.Printf("[ActivateIdResRuleSetWithRetries] Successfully ran ActivateIdResRuleSetWithRetries")
			break
		}
		log.Printf("[ActivateIdResRuleSetWithRetries] Unable to activate rulesets, waiting for %s", duration)
		time.Sleep(duration)
	}

	return res.(model.ResponseWrapper), err
}

func (h UptHandler) RebuildCache(domainName string, cacheMode string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "cache?type=" + cacheMode, Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) SaveCacheRuleSet(domainName string, rules []customerprofileslcs.Rule) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		SaveRuleSet: model.SaveRuleSetRq{
			Rules: rules,
		},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "ruleSetCache", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) ActivateCacheRuleSet(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "ruleSetCache/activate", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) UpdatePromptConfig(domainName string, domainSetting model.DomainSetting) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		DomainSetting: domainSetting,
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "promptConfig", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}
