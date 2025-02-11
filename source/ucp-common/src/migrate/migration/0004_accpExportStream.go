// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"errors"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/feature"
)

type AccpExportStreamMigration struct {
	tx core.Transaction
}

var _ Migration = &AccpExportStreamMigration{}

func InitAccpExportStreamMigration(tx core.Transaction) (*AccpExportStreamMigration, error) {
	return &AccpExportStreamMigration{tx: tx}, nil
}

func (*AccpExportStreamMigration) VersionIntroduced() int {
	return feature.AccpExportStreamVersionIntroduced
}

func (m *AccpExportStreamMigration) Up(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	return feature.SimpleVersion(m.VersionIntroduced() - 1), errors.New("not yet implemented")
}

func (m *AccpExportStreamMigration) Down(ctx context.Context) (version feature.FeatureSetVersion, err error) {
	return feature.SimpleVersion(m.VersionIntroduced()), errors.New("not yet implemented")
}
