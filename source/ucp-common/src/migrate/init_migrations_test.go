// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/customerprofiles"
	"tah/upt/source/ucp-common/src/feature"
	"testing"
)

func TestInitMigrations(t *testing.T) {
	t.Parallel()

	mockAccp := customerprofiles.InitMock()
	defer mockAccp.AssertExpectations(t)

	migrations, err := InitMigrations(context.Background(), core.NewTransaction("", "", core.LogLevelDebug), Services{Accp: mockAccp}, true)
	if err != nil {
		t.Fatalf("error initializing migrations: %v", err)
	}

	missing := getMissingMigrations(feature.INITIAL_FEATURE_SET_VERSION, feature.LATEST_FEATURE_SET_VERSION, migrations)
	if len(missing) > 0 {
		t.Errorf("expected no missing migrations between initial and latest version: %v", missing)
	}
}

func TestInitMigrationsInvalidAccp(t *testing.T) {
	t.Parallel()

	_, err := InitMigrations(context.Background(), core.NewTransaction("", "", core.LogLevelDebug), Services{Accp: nil}, true)
	if err == nil {
		t.Fatal("expected error initializing migrations")
	}
}
