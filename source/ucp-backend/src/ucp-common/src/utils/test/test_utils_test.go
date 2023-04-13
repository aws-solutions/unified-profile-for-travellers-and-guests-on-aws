package test

import (
	"os"
	"testing"
)

var region = os.Getenv("UCP_REGION")

func TestGetTestRegion(t *testing.T) {
	tr := GetTestRegion()
	if tr != region {
		t.Errorf("Invalid test regions %v", tr)
	}
}
