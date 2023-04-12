package appregistry

import (
	"os"
	"testing"
)

var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestAppRegistry(t *testing.T) {
	appName := "appregistry-test-app"
	client := Init(TAH_CORE_REGION)
	tags := make(map[string]string)
	tags["purpose"] = "test"
	app, err1 := client.CreateApplication(appName, "description", tags)
	if err1 != nil {
		t.Errorf("Error creating application %v", err1)
	}
	if app.Tags["purpose"] != "test" {
		t.Errorf("Error with created application")
	}
	list, err2 := client.ListApplications()
	if err2 != nil {
		t.Errorf("Error querying AppRegistry for applications: %v", err2)
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
	err3 := client.DeleteApplication(appName)
	if err3 != nil {
		t.Errorf("Error deleting application %v", err3)
	}
	list2, err4 := client.ListApplications()
	if err4 != nil {
		t.Errorf("Error querying AppRegistry for applications: %v", err4)
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
