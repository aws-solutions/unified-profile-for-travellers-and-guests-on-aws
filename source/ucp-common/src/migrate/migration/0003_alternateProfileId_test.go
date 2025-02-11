// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"tah/upt/source/tah-core/core"
	"testing"
)

func TestAlternateProfileIdMigration(t *testing.T) {
	t.Parallel()

	migration, err := InitAlternateProfileIdMigration(core.NewTransaction("", "", core.LogLevelDebug))
	if err != nil {
		t.Fatalf("error initializing migration: %v", err)
	}

	_, err = migration.Up(context.Background())
	if err == nil {
		t.Errorf("expected error running up migration")
	}

	_, err = migration.Down(context.Background())
	if err == nil {
		t.Errorf("expected error running down migration")
	}
}
