// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	core "tah/upt/source/tah-core/core"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	//model "cloudrack-lambda-core/config/model"
)

func TestMain(t *testing.T) {
	stubReq := events.APIGatewayProxyRequest{
		Headers: map[string]string{core.TRANSACTION_ID_HEADER: "stub"},
	}
	stubContext := context.Background()

	HandleRequest(stubContext, stubReq)
}

func TestValidateRequest(t *testing.T) {
	err := ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"aws-tah-tx-id":            "1234567890123456789012345678901234567890",
			"customer-profiles-domain": "test_domain_domain",
		},
	})
	if err == nil {
		t.Errorf("Expected for Api gateway request with too long Tx header")
	}
	err = ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"aws-tah-tx-id":            "1234567890",
			"customer-profiles-domain": "test_domain_domain01test_domain_domain01test_domain_domain01test_domain_domain01",
		},
	})
	if err == nil {
		t.Errorf("Expected for Api gateway request with too domain name")
	}
	err = ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"aws-tah-tx-id":            "550e8400-e29b-41d4-a716-446655440000",
			"customer-profiles-domain": "test_domain_domain",
		},
	})
	if err != nil {
		t.Errorf("Valid Api gateway request should not return error but got: %v", err)
	}
	err = ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"aws-tah-tx-id": "550e8400-e29b-41d4-a716-446655440000",
		},
	})
	if err != nil {
		t.Errorf("Valid Api gateway request should not return error but got: %v", err)
	}
	err = ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"customer-profiles-domain": "test_domain_domain",
		},
	})
	if err != nil {
		t.Errorf("Valid Api gateway request should not return error but got: %v", err)
	}
	err = ValidateApiGatewayRequest(events.APIGatewayProxyRequest{
		Headers: map[string]string{},
	})
	if err != nil {
		t.Errorf("Valid Api gateway request should not return error but got: %v", err)
	}

}
