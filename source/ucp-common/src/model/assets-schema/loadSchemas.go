// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package assetsSchema

import (
	"io/ioutil"
	"log"
	"os"
	"tah/upt/source/tah-core/core"
	glue "tah/upt/source/tah-core/glue"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	commonModel "tah/upt/source/ucp-common/src/model/admin"
)

var SchemaPathMap = map[string]string{
	constant.BIZ_OBJECT_AIR_BOOKING:   "tah-common-glue-schemas/air_booking.glue.json",
	constant.BIZ_OBJECT_CLICKSTREAM:   "tah-common-glue-schemas/clickevent.glue.json",
	constant.BIZ_OBJECT_GUEST_PROFILE: "tah-common-glue-schemas/guest_profile.glue.json",
	constant.BIZ_OBJECT_HOTEL_BOOKING: "tah-common-glue-schemas/hotel_booking.glue.json",
	constant.BIZ_OBJECT_STAY_REVENUE:  "tah-common-glue-schemas/hotel_stay_revenue.glue.json",
	constant.BIZ_OBJECT_PAX_PROFILE:   "tah-common-glue-schemas/pax_profile.glue.json",
	constant.BIZ_OBJECT_CSI:           "tah-common-glue-schemas/customer_service_interaction.glue.json",
	"traveller":                       "tah-common-glue-schemas/traveller.glue.json",
}

func LoadSchema(tx *core.Transaction, schemaPath string, bizObject commonModel.BusinessObject) (glue.Schema, error) {
	// no schema path is provided we set teh default path to Lambda root folder for the task
	//this way we just need to set the env var during local unit test
	if schemaPath == "" {
		tx.Info("No schema path set in env var, using default path")
		schemaPath = "/var/task"
	} else {
		tx.Debug("Schema path set in env vars %s", schemaPath)
	}
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	tx.Debug("Current execution path: %s", path)
	schemaBytes, err := ioutil.ReadFile(schemaPath + "/" + SchemaPathMap[bizObject.Name])
	if err != nil {
		return glue.Schema{}, err
	}
	schema, err := glue.ParseSchema(string(schemaBytes))
	if err != nil {
		return glue.Schema{}, err
	}
	return schema, nil
}
