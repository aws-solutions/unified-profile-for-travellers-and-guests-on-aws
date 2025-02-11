// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildAncillaryServiceObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_ANCILLARY,
		Version: "1.0",
		Fields:  BuildAncillaryServiceMapping(),
	}
}

func BuildAncillaryServiceMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		// Profile Object Unique Key
		{
			Type:    "STRING",
			Source:  "_source.accp_object_id",
			Target:  "accp_object_id",
			Indexes: []string{"UNIQUE"},
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_by",
			Target:  "last_updated_by",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
		{
			Type:   customerprofiles.MappingTypeString,
			Source: "_source.extended_data",
			Target: "extended_data",
		},
		// Profile Data
		{
			Type:        "STRING",
			Source:      "_source.traveller_id",
			Target:      "_profile.Attributes.profile_id",
			Searcheable: true,
			Indexes:     []string{"PROFILE"},
		},
		{
			Type:    "STRING",
			Source:  "_source.pax_index",
			Target:  "pax_index",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.ancillary_type",
			Target:  "ancillary_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.price",
			Target:  "price",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.currency",
			Target:  "currency",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.baggage_type",
			Target:  "baggage_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.quantity",
			Target:  "quantity",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.weight",
			Target:  "weight",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.dimentions_length",
			Target:  "dimentions_length",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.dimentions_width",
			Target:  "dimentions_width",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.dimentions_height",
			Target:  "dimentions_height",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.priority_bag_drop",
			Target:  "priority_bag_drop",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.priority_bag_return",
			Target:  "priority_bag_return",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.lot_bag_insurance",
			Target:  "lot_bag_insurance",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.valuable_baggage_insurance",
			Target:  "valuable_baggage_insurance",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.hands_free_baggage",
			Target:  "hands_free_baggage",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.change_type",
			Target:  "change_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.seat_number",
			Target:  "seat_number",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.seat_zone",
			Target:  "seat_zone",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.neighbor_free_seat",
			Target:  "neighbor_free_seat",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.upgrade_auction",
			Target:  "upgrade_auction",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.other_ancilliary_type",
			Target:  "other_ancilliary_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.priority_service_type",
			Target:  "priority_service_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.lounge_access",
			Target:  "lounge_access",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_id",
			Target:  "booking_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.flight_number",
			Target:  "flight_number",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.departure_date",
			Target:  "departure_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.tx_id",
			Target:  "tx_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.last_updated_partition",
			Target:  "last_updated_partition",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.model_version",
			Target:  "model_version",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.object_type",
			Target:  "object_type",
			KeyOnly: true,
		},
	}
}
