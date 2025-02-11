// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildAlternateProfileIdObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_ALTERNATE_PROFILE_ID,
		Version: "1.0",
		Fields:  BuildAlternateProfileIdMapping(),
	}
}

func BuildAlternateProfileIdMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		//	Profile Object Unique Key
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  constants.PrependSourceField(constants.ACCP_OBJECT_ID_KEY),
			Target:  constants.ACCP_OBJECT_ID_KEY,
			Indexes: []string{customerprofiles.INDEX_UNIQUE},
			KeyOnly: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField("name"),
			Target: "name",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField("description"),
			Target: "description",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField("value"),
			Target: "value",
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField(customerprofiles.LAST_UPDATED_FIELD),
			Target: customerprofiles.LAST_UPDATED_FIELD,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField(constants.ACCP_OBJECT_TYPE_KEY),
			Target: constants.ACCP_OBJECT_TYPE_KEY,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: constants.PrependSourceField("model_version"),
			Target: "model_version",
		},
		{
			Type:    customerprofiles.MappingTypeString,
			Source:  constants.PrependSourceField("tx_id"),
			Target:  "tx_id",
			KeyOnly: true,
		},
		// Profile Data
		{
			Type:        customerprofiles.MappingTypeString,
			Source:      constants.PrependSourceField(constants.ACCP_OBJECT_TRAVELLER_ID_KEY),
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{customerprofiles.INDEX_PROFILE},
		},
	}
}
