package kms

import (
	"os"
	"testing"
)

var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestCreateDelete(t *testing.T) {
	cfg := Init(TAH_CORE_REGION)
	keyArn, err := cfg.CreateKey("tah core unit test key")
	if err != nil {
		t.Errorf("Error creating KMS key: %v", err)
	}
	err = cfg.DeleteKey(keyArn)
	if err != nil {
		t.Errorf("Error deleting KMS key: %v", err)
	}
}
