// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import customerprofileslcs "tah/upt/source/storage"

type RebuildCacheBody struct {
	MatchBucketName string                        `json:"matchBucket"`
	Env             string                        `json:"env"`
	DomainName      string                        `json:"domainName"`
	CacheMode       customerprofileslcs.CacheMode `json:"cacheMode"`
}
