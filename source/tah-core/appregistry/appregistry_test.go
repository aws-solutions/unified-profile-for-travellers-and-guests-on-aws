// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package appregistry

import (
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

func TestAppRegistry(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	appName := "appregistry-test-app"
	client := Init(envCfg.Region, core.TEST_SOLUTION_ID, core.TEST_SOLUTION_VERSION)
	tags := make(map[string]string)
	tags["purpose"] = "test"
	app, err := client.CreateApplication(appName, "description", tags)
	if err != nil {
		t.Errorf("Error creating application %v", err)
	}
	if app.Tags["purpose"] != "test" {
		t.Errorf("Error with created application")
	}
	list, err := client.ListApplications()
	if err != nil {
		t.Errorf("Error querying AppRegistry for applications: %v", err)
	}
	appInList := false
	for i := 0; i < len(list); i++ {
		if list[i].Name == appName {
			appInList = true
		}
	}
	if !appInList {
		t.Errorf("Error searching AppRegistry for %v", appName)
	}
	err = client.DeleteApplication(appName)
	if err != nil {
		t.Errorf("Error deleting application %v", err)
	}
	list2, err := client.ListApplications()
	if err != nil {
		t.Errorf("Error querying AppRegistry for applications: %v", err)
	}
	appInList2 := false
	for i := 0; i < len(list2); i++ {
		if list2[i].Name == appName {
			appInList = true
		}
	}
	if appInList2 {
		t.Errorf("Error, deleted application still in AppRegistry for %v", appName)
	}
}
