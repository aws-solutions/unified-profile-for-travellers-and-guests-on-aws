package utils

import (
	"os"
	"testing"
)

var UCP_REGION = GetTestRegion()

func TestUtils(t *testing.T) {
}

//can't use the testutils one due to circular dependency
func GetTestRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
