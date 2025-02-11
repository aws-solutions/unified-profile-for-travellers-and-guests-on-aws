// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package usecase

import (
	"context"
	"encoding/json"
	customerprofileslcs "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	"tah/upt/source/tah-core/ecs"
	model "tah/upt/source/ucp-common/src/model/async"
	"tah/upt/source/ucp-common/src/model/async/usecase"
	"tah/upt/source/ucp-common/src/services/admin"
)

type StartBatchIdResUsecase struct {
	name string
}

func InitStartBatchIdRes() *StartBatchIdResUsecase {
	return &StartBatchIdResUsecase{
		name: "StartBatchIdRes",
	}
}

func (u *StartBatchIdResUsecase) Name() string {
	return u.name
}

func (u *StartBatchIdResUsecase) Execute(payload model.AsyncInvokePayload, services model.Services, tx core.Transaction) error {
	payloadJson, err := json.Marshal(payload.Body)
	if err != nil {
		tx.Error("[%s] Error marshalling payload: %v", u.name, err)
		return err
	}

	body := usecase.StartBatchIdResBody{}
	err = json.Unmarshal(payloadJson, &body)
	if err != nil {
		tx.Error("[%s] Error unmarshalling payload body: %v", u.name, err)
		return err
	}

	type TaskConfigResult struct {
		TaskConfig ecs.TaskConfig
		Err        error
	}

	taskConfigChan := make(chan TaskConfigResult)
	go func(c chan TaskConfigResult) {
		namespace := services.Env["SSM_PARAM_NAMESPACE"]
		params, err := services.SsmConfig.GetParametersByPath(context.TODO(), namespace)
		if err != nil {
			tx.Error("error querying SSM parameters: %v", err)
			c <- TaskConfigResult{ecs.TaskConfig{}, err}
			return
		}
		idResTaskConfig, err := admin.GetTaskConfigForIdRes(params)
		if err != nil {
			tx.Error("Error loading ECS task config for batch ID res: %v", err)
			c <- TaskConfigResult{ecs.TaskConfig{}, err}
			return
		}
		c <- TaskConfigResult{idResTaskConfig, nil}
	}(taskConfigChan)

	services.AccpConfig.SetDomain(body.Domain)

	domain, err := services.AccpConfig.GetDomain()
	if err != nil {
		tx.Error("[%s] Error getting domain: %v", u.name, err)
		return err
	}

	// Invalidate cache to get the latest active rule set that might use an already running async lambda runtime
	err = services.AccpConfig.InvalidateStitchingRuleSetsCache()
	if err != nil {
		tx.Error("[%s] Error invalidating id res cache: %v", u.name, err)
		return err
	}

	activeRuleSet, err := services.AccpConfig.GetActiveRuleSet()
	if err != nil {
		tx.Error("[%s] Error getting active rule set %v", u.name, err)
		return err
	}

	taskConfigResult := <-taskConfigChan
	if taskConfigResult.Err != nil {
		tx.Error("[%s] Error initializing ECS client: %v", u.name, taskConfigResult.Err)
		return taskConfigResult.Err
	}

	err = customerprofileslcs.StartRuleBasedJob(
		context.TODO(),
		services.EcsConfig,
		taskConfigResult.TaskConfig,
		domain.Name,
		domain.AllCounts,
		activeRuleSet,
		payload.EventID,
	)
	if err != nil {
		tx.Error("[%s] Error starting rule-based job: %v", u.name, err)
		return err
	}

	return nil
}
