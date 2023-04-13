package admin

import (
	constant "tah/ucp-common/src/constant/admin"
	model "tah/ucp-common/src/model/admin"
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

}
