// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package accpmappings

import (
	customerprofiles "tah/upt/source/storage"
	constants "tah/upt/source/ucp-common/src/constant/admin"
)

func BuildLoyaltyTransactionObjectMapping() customerprofiles.ObjectMapping {
	return customerprofiles.ObjectMapping{
		Name:    constants.ACCP_RECORD_LOYALTY_TX,
		Version: "1.0",
		Fields:  BuildLoyaltyTransactionMapping(),
	}
}

func BuildLoyaltyTransactionMapping() customerprofiles.FieldMappings {
	return []customerprofiles.FieldMapping{
		{
			Type:    "STRING",
			Source:  "_source.object_type",
			Target:  "object_type",
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
			Source:  "_source.last_updated",
			Target:  "last_updated",
			KeyOnly: true,
		},
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
			Source:  "_source.category",
			Target:  "category",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.points_offset",
			Target:  "points_offset",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.points_unit",
			Target:  "points_unit",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.origin_points_offset",
			Target:  "origin_points_offset",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.qualifying_point_offset",
			Target:  "qualifying_point_offset",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.source",
			Target:  "source",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.booking_date",
			Target:  "booking_date",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.order_number",
			Target:  "order_number",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.product_id",
			Target:  "product_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.expire_in_days",
			Target:  "expire_in_days",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.amount",
			Target:  "amount",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.amount_type",
			Target:  "amount_type",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.voucher_quantity",
			Target:  "voucher_quantity",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.corporate_reference_number",
			Target:  "corporate_reference_number",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.promotions",
			Target:  "promotions",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.location",
			Target:  "location",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.activity_day",
			Target:  "activity_day",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.to_loyalty_id",
			Target:  "to_loyalty_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.from_loyalty_id",
			Target:  "from_loyalty_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.organization_code",
			Target:  "organization_code",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.event_name",
			Target:  "event_name",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.document_number",
			Target:  "document_number",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.corporate_id",
			Target:  "corporate_id",
			KeyOnly: true,
		},
		{
			Type:    "STRING",
			Source:  "_source.program_name",
			Target:  "program_name",
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
	}
}
