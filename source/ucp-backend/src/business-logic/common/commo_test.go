package common

import (
	"os"
	"tah/core/core"
	"testing"
)

var UCP_REGION = getRegion()

func TestCommon(t *testing.T) {
	tx := core.NewTransaction("test_validator", "")
	cx := Init(&tx, UCP_REGION)
	cx.Log("Testing common log function with '%v'", "test_value")
}

// TODO: move this somewhere centralized
func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
