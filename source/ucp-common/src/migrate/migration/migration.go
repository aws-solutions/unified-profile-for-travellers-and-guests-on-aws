// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"context"
	"tah/upt/source/ucp-common/src/feature"
)

// A migration implementation must perform all the actions required to bring a domain from the previous version to the
// version associated with the migration. It must fail gracefully, rolling back incremental operations if any fail,
// restoring the domain the the state it was in before the migration began.
//
// The feature set version returned by the migration function must reflect the version of the domain after the migration
// completes, whether the migration succeeded or failed and was rolled back.
type Migration interface {
	// The domain version that first introduced this change. Performing an Up migration on a domain of the prior version
	// will bring the domain to this verison.
	//
	// The implementation of `VersionIntroduced` should not reference the receiver (i.e. it should be "static"). The
	// following call should succeed on an implementation of this interface:
	//  (*Migration)(nil).VersionIntroduced()
	VersionIntroduced() int

	// Perform all the actions required to bring the UPT domain to the version advertised by this migration.
	Up(context.Context) (feature.FeatureSetVersion, error)

	// Undo all the actions required to bring the UPT domain to the version advertised by this migration, restoring the
	// domain to the previous version.
	Down(context.Context) (feature.FeatureSetVersion, error)
}
