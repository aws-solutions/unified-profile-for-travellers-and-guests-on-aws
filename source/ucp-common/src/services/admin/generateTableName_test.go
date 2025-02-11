// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	constant "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/admin"
	"testing"
)

func TestTableName(t *testing.T) {

	businessObjectHotel := model.BusinessObject{Name: constant.BIZ_OBJECT_HOTEL_BOOKING}
	tableName := BuildTableName("dev", businessObjectHotel, "test_domain")
	if tableName != "ucp_dev_hotel_booking_test_domain" {
		t.Errorf("Table Name incorrect, name produced is %v", tableName)
	}

	businessObjectClick := model.BusinessObject{Name: constant.BIZ_OBJECT_CLICKSTREAM}
	tableNameClick := BuildTableName("dev", businessObjectClick, "test_domain")
	if tableNameClick != "ucp_dev_clickstream_test_domain" {
		t.Errorf("Table Name incorrect, name produced is %v", tableNameClick)
	}

	fakeS3BucketName := "ucpHotel"
	glueDestinationHotel := BuildGlueDestination(fakeS3BucketName, "test_domain")
	if glueDestinationHotel != "ucpHotel/test_domain" {
		t.Errorf("Glue Destination incorrect, name produced is %v", glueDestinationHotel)
	}

}
