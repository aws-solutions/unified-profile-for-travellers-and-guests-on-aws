// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package upt_sdk

import (
	"fmt"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	"tah/upt/source/ucp-common/src/model/admin"
	"tah/upt/source/ucp-common/src/utils/utils"
)

func (h UptHandler) SearchProfile(domainName string, key string, values []string) (model.ResponseWrapper, error) {
	if len(values) != 1 {
		return model.ResponseWrapper{}, fmt.Errorf("invalid number of values (only one value supported in upt sdk)") // we currently only support single value search
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{
		key: values[0],
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "profile", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) SearchProfileAdvanced(domainName string, keyValues map[string]string) (model.ResponseWrapper, error) {
	if len(keyValues) == 0 {
		return model.ResponseWrapper{}, fmt.Errorf("[upt-sdk] at least one value must be provided for advanced search") // we currently only support single value search
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(keyValues, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "profile", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) GetProfileId(domainName string, travelerId string) (string, error) {
	res, err := h.SearchProfile(domainName, "travellerId", []string{travelerId})
	if err != nil {
		return "", err
	}
	profiles := utils.SafeDereference(res.Profiles)
	if len(profiles) == 0 {
		return "", fmt.Errorf("no profile found for traveler id %s", travelerId)
	}
	return profiles[0].ConnectID, err
}

func (h UptHandler) RetreiveProfile(domainName, uptID string, objectTypeNames []string) (model.ResponseWrapper, error) {
	if uptID == "" {
		return model.ResponseWrapper{}, fmt.Errorf("upt id is empty")
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "profile/" + uptID, Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) DeleteProfile(domainName, uptID string) (model.ResponseWrapper, error) {
	if uptID == "" {
		return model.ResponseWrapper{}, fmt.Errorf("upt id is empty")
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpDelete(uptID, map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "profile", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) RetrieveInteractionHistory(domainName, objectType, interactionId, connectId string) ([]customerprofileslcs.ProfileMergeContext, error) {
	if connectId == "" {
		return []customerprofileslcs.ProfileMergeContext{}, fmt.Errorf("connect id is empty")
	}
	if interactionId == "" {
		return []customerprofileslcs.ProfileMergeContext{}, fmt.Errorf("interaction id is empty")
	}
	if objectType == "" {
		return []customerprofileslcs.ProfileMergeContext{}, fmt.Errorf("object type is empty")
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{
		"objectType": objectType,
		"connectId":  connectId,
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "interactionHistory/" + interactionId, Headers: authHeader})
	return *res.(model.ResponseWrapper).InteractionHistory, err
}

//Merge and Unmerge

func (h UptHandler) MergeProfiles(domainName, connectIdOne, connectIdTwo string) (model.ResponseWrapper, error) {
	if connectIdOne == "" || connectIdTwo == "" {
		return model.ResponseWrapper{}, fmt.Errorf("connect id is empty")
	}
	if connectIdOne == connectIdTwo {
		return model.ResponseWrapper{}, fmt.Errorf("cannot merge profile with itself")
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		MergeRq: []admin.MergeRq{
			{
				SourceProfileID: connectIdOne,
				TargetProfileID: connectIdTwo,
				MergeContext: customerprofileslcs.ProfileMergeContext{
					MergeType:              customerprofileslcs.MergeTypeManual,
					ConfidenceUpdateFactor: 1,
				},
			},
		},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "merge", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) UnmergeProfiles(domainName, mergedIntoConnectID, toUnmergeConnectID string) (model.ResponseWrapper, error) {
	if mergedIntoConnectID == "" || toUnmergeConnectID == "" {
		return model.ResponseWrapper{}, fmt.Errorf("one of the 2 connect ids is empty")
	}
	if mergedIntoConnectID == toUnmergeConnectID {
		return model.ResponseWrapper{}, fmt.Errorf("cannot unmerge profile from itself")
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		UnmergeRq: admin.UnmergeRq{

			MergedIntoConnectID: mergedIntoConnectID,
			ToUnmergeConnectID:  toUnmergeConnectID,
		},
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "unmerge", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

//Privacy Search

func (h UptHandler) CreatePrivacySearch(domainName string, connectIds []string) (model.ResponseWrapper, error) {
	if len(connectIds) == 0 {
		return model.ResponseWrapper{}, fmt.Errorf("need at least one connect ID for privacy search creation")
	}
	rq := []model.CreatePrivacySearchRq{}
	for _, connectId := range connectIds {
		rq = append(rq, model.CreatePrivacySearchRq{
			ConnectId: connectId,
		})
	}
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		CreatePrivacySearchRq: rq,
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "privacy", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) ListPrivacySearches(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "privacy", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) GetPrivacySearch(domainName string, connectId string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "privacy/" + connectId, Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) StartProfilePurge(domainName string, connectIds []string) (model.ResponseWrapper, error) {
	if len(connectIds) == 0 {
		return model.ResponseWrapper{}, fmt.Errorf("need at least one connect ID for privacy purge creation")
	}
	rq := model.CreatePrivacyPurgeRq{
		ConnectIds:     connectIds,
		AgentCognitoId: core.UUID(),
	}

	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpPost(model.RequestWrapper{
		PrivacyPurge: rq,
	}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "privacy/purge", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}

func (h UptHandler) GetIsPurgeRunning(domainName string) (model.ResponseWrapper, error) {
	authHeader := h.createAuthHeader()
	authHeader[DOMAIN_HEADER] = domainName
	res, err := h.Client.HttpGet(map[string]string{}, &model.ResponseWrapper{}, core.RestOptions{SubEndpoint: "privacy/purge", Headers: authHeader})
	return res.(model.ResponseWrapper), err
}
