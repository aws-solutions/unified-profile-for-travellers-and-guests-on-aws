// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package feature

type FeatureSetVersion struct {
	Version           int
	CompatibleVersion int
}

// A FeatureSetVersion with CompatibleVersion set to the same value as Version
func SimpleVersion(version int) FeatureSetVersion {
	return FeatureSetVersion{Version: version, CompatibleVersion: version}
}

const INITIAL_FEATURE_SET_VERSION = 1
const LATEST_FEATURE_SET_VERSION = 4

type FeatureSet struct {
	version FeatureSetVersion
}

func InitFeatureSetVersion(version FeatureSetVersion) (FeatureSet, error) {
	return FeatureSet{version: version}, nil
}

const AccpTravelerIdMappingIntroducedVersion = 2

func (fs *FeatureSet) SupportsCustomerProfilesTravelerIdMapping() bool {
	return fs.version.Version >= AccpTravelerIdMappingIntroducedVersion
}

const AlternateProfileIdIntroducedVersion = 3

func (fs *FeatureSet) SupportsAlternateProfileId() bool {
	return fs.version.Version >= AlternateProfileIdIntroducedVersion
}

const AccpExportStreamVersionIntroduced = 4

func (fs *FeatureSet) SupportsAccpExportStream() bool {
	return fs.version.Version >= AccpExportStreamVersionIntroduced
}
