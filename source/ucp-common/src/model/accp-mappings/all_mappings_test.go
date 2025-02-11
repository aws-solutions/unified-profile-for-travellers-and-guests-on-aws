// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	constants "tah/upt/source/ucp-common/src/constant/admin"
	"testing"
)

func TestMappingMap(t *testing.T) {
	allMappingCount := len(ACCP_OBJECT_MAPPINGS)
	testedCount := 0
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_AIR_BOOKING]().Name != constants.ACCP_RECORD_AIR_BOOKING {
		t.Errorf("invalid object name for ACCP_RECORD_AIR_BOOKING")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_EMAIL_HISTORY]().Name != constants.ACCP_RECORD_EMAIL_HISTORY {
		t.Errorf("invalid object name for ACCP_RECORD_EMAIL_HISTORY")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_PHONE_HISTORY]().Name != constants.ACCP_RECORD_PHONE_HISTORY {
		t.Errorf("invalid object name for ACCP_RECORD_PHONE_HISTORY")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_AIR_LOYALTY]().Name != constants.ACCP_RECORD_AIR_LOYALTY {
		t.Errorf("invalid object name for ACCP_RECORD_AIR_LOYALTY")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_CLICKSTREAM]().Name != constants.ACCP_RECORD_CLICKSTREAM {
		t.Errorf("invalid object name for ACCP_RECORD_CLICKSTREAM")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_GUEST_PROFILE]().Name != constants.ACCP_RECORD_GUEST_PROFILE {
		t.Errorf("invalid object name for ACCP_RECORD_GUEST_PROFILE")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_HOTEL_LOYALTY]().Name != constants.ACCP_RECORD_HOTEL_LOYALTY {
		t.Errorf("invalid object name for ACCP_RECORD_HOTEL_LOYALTY")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_HOTEL_BOOKING]().Name != constants.ACCP_RECORD_HOTEL_BOOKING {
		t.Errorf("invalid object name for ACCP_RECORD_HOTEL_BOOKING")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_PAX_PROFILE]().Name != constants.ACCP_RECORD_PAX_PROFILE {
		t.Errorf("invalid object name for ACCP_RECORD_PAX_PROFILE")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_HOTEL_STAY_MAPPING]().Name != constants.ACCP_RECORD_HOTEL_STAY_MAPPING {
		t.Errorf("invalid object name for ACCP_RECORD_HOTEL_STAY_MAPPING")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_CSI]().Name != constants.ACCP_RECORD_CSI {
		t.Errorf("invalid object name for ACCP_RECORD_CSI")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_LOYALTY_TX]().Name != constants.ACCP_RECORD_LOYALTY_TX {
		t.Errorf("invalid object name for ACCP_RECORD_LOYALTY_TX")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_ANCILLARY]().Name != constants.ACCP_RECORD_ANCILLARY {
		t.Errorf("invalid object name for ACCP_RECORD_ANCILLARY")
	}
	testedCount++
	if ACCP_OBJECT_MAPPINGS[constants.ACCP_RECORD_ALTERNATE_PROFILE_ID]().Name != constants.ACCP_RECORD_ALTERNATE_PROFILE_ID {
		t.Errorf("invalid object name for ACCP_RECORD_ALTERNATE_PROFILE_ID")
	}
	testedCount++

	if allMappingCount != testedCount {
		t.Fatalf("Missing ACCP_OBJECT_MAPPING check")
	}
}
