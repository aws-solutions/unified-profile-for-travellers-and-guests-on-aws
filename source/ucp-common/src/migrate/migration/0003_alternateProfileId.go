// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/feature"
)

type AlternateProfileIdMigration struct {
	tx core.Transaction
}

var _ Migration = &AlternateProfileIdMigration{}

func InitAlternateProfileIdMigration(tx core.Transaction) (*AlternateProfileIdMigration, error) {
	return &AlternateProfileIdMigration{tx: tx}, nil
}

func (*AlternateProfileIdMigration) VersionIntroduced() int {
	return feature.AlternateProfileIdIntroducedVersion
}

func (m *AlternateProfileIdMigration) Up(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	return feature.SimpleVersion(m.VersionIntroduced() - 1), errors.New("not yet implemented")
}

func (m *AlternateProfileIdMigration) Down(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	return feature.SimpleVersion(m.VersionIntroduced()), errors.New("not yet implemented")
}
