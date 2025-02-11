// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/ucp-sync/src/business-logic/model"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestMain(t *testing.T) {
	mockAccp := customerprofiles.InitMockV2()
	mockAccp.On("ListDomains").Return([]customerprofiles.Domain{}, nil)
	_, err := HandleRequestWithParams(context.Background(), events.CloudWatchEvent{}, model.AWSConfig{
		Accp: mockAccp,
	})
	if err != nil {
		t.Errorf("[TestMain] Should not have error")
	}
}

func TestBuildConfig(t *testing.T) {
	cfg := buildConfig(events.CloudWatchEvent{
		DetailType: "ucp_manual_trigger",
		Resources:  []string{"domain:test_domain"},
	})
	if len(cfg.Domains) != 1 {
		t.Errorf("[TestBuildConfig] Should have 1 domain")
	}
	if len(cfg.Domains) > 0 && cfg.Domains[0].Name != "test_domain" {
		t.Errorf("[TestBuildConfig] Should have domain test_domain")
	}
	if cfg.Env["METHOD_OF_INVOCATION"] != "ucp_manual_trigger" {
		t.Errorf("[TestBuildConfig] METHOD_OF_INVOCATION should be ucp_manual_trigger")
	}
	cfg = buildConfig(events.CloudWatchEvent{
		DetailType: "ucp_manual_trigger",
		Resources:  []string{"job:test_job"},
	})
	if cfg.RequestedJob != "test_job" {
		t.Errorf("[TestBuildConfig] RequestedJob shoudl be %v and not %v", "test_job", cfg.RequestedJob)
	}
}

func TestParseDomain(t *testing.T) {
	domains := parseDomain([]string{"domain:test_domain", "job:test_job", "something:else"})
	if len(domains) != 1 {
		t.Errorf("[TestParseDomain] Should have 1 domain")
		return
	}
	if domains[0].Name != "test_domain" {
		t.Errorf("[TestParseDomain] Should have domain test_domain")
	}
}

func TestParseJob(t *testing.T) {
	job := parseJob([]string{"domain:test_domain", "job:test_job", "something:else"})
	if job != "test_job" {
		t.Errorf("[TestParseDomain] Should have domain test_job")
	}
}

func TestConfig(t *testing.T) {
	cfg := buildConfig(events.CloudWatchEvent{})
	if cfg.Tx.TransactionID == "" {
		t.Error("Expected non-empty transaction ID")
	}
	if cfg.Env["METHOD_OF_INVOCATION"] == "ucp_manual_trigger" {
		t.Errorf("[TestBuildConfig] METHOD_OF_INVOCATION should not be ucp_manual_trigger")
	}
}
