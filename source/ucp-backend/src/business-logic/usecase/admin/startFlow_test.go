// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"log"
	"testing"

	"tah/upt/source/tah-core/appflow"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/mock"
)

func TestStartFlow(t *testing.T) {
	log.Printf("TestStartFlow: initialize test resources")
	appFlow := appflow.InitMock()
	appFlow.On("StartFlow", mock.Anything).Return(appflow.FlowStatusOutput{}, nil)
	tx := core.NewTransaction(t.Name(), "", core.LogLevelDebug)
	reg := registry.Registry{AppFlow: appFlow}
	uc := NewStartFlows()
	uc.Name()
	uc.SetTx(&tx)
	uc.Tx()
	uc.Registry()
	uc.SetRegistry(&reg)
	rq := events.APIGatewayProxyRequest{
		Body: `{"domain": {"integrations": [{"name":"test_integration"}]}}`,
	}
	tx.Debug("Api Gateway request", rq)
	wrapper, err0 := uc.CreateRequest(rq)
	if err0 != nil {
		t.Errorf("[%s] Error creating request %v", "StartFlow", err0)
	}
	err0 = uc.ValidateRequest(wrapper)
	if err0 != nil {
		t.Errorf("[%s] Error validating request request %v", "StartFlow", err0)
	}
	rs, err := uc.Run(wrapper)
	if err != nil {
		t.Errorf("[%s] Error running use case: %v", "StartFlow", err)
	}
	apiRes, err2 := uc.CreateResponse(rs)
	if err2 != nil {
		t.Errorf("[%s] Error creating response %v", "StartFlow", err2)
	}
	tx.Debug("Api Gateway response", apiRes)

	log.Printf("TestStartFlow: deleting test resources")
}
