// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"fmt"
	"log"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func TestCreateAuroraSecret(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	sm := InitWithRegion("", envCfg.Region, "", "")

	sn := "aurora_secret_name"

	_, err = sm.CreateAuroraSecret(sn, "username", "pwd")
	if err != nil {
		t.Error(fmt.Errorf("CreateAuroraSecret failed with error: %v", err))
	}
	if sm.Get("username") != "username" {
		t.Error("username not set")
	}
	if sm.Get("password") != "pwd" {
		t.Error("pwd not set")
	}

	log.Printf("Checking if secret: %v has been created successfully", sn)
	arn, err := sm.GetSecretArn(sn)
	if err != nil {
		log.Printf("Error accessing secret: %v", err)
		log.Printf("Secret: %v not created yet, waiting for 5 seconds", sn)
		time.Sleep(time.Second * 5)
		arn, err = sm.GetSecretArn(sn)
		if err != nil {
			t.Error(fmt.Errorf("GetSecretArn failed with error: %v", err))
		}
	}
	if arn == "" {
		t.Error("arn should not be empty")
	} else {
		log.Printf("Successfully retrieved secret arn: %v", arn)
	}

	log.Printf("Deleting secret: %v", sn)
	err = sm.DeleteSecret(sn, true)
	if err != nil {
		t.Error(err)
	}
}
