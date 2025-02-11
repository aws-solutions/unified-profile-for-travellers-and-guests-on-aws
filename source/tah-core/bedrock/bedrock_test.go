// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bedrock

import (
	"log"
	"os"
	"strings"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
)

var TAH_CORE_BEDROCK_REGION = os.Getenv("TAH_CORE_BEDROCK_REGION")

/************************************
** Read before running the tests
* you need to go to the bedrock console and enable the following models:
- all Anthropic models
- all Amazon (Titan) models
- all Stability AI models
**************************************/

func TestClaude(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	region := TAH_CORE_BEDROCK_REGION
	if TAH_CORE_BEDROCK_REGION == "" || TAH_CORE_BEDROCK_REGION == "null" {
		log.Printf("TAH_CORE_BEDROCK_REGION not set, using %s", envCfg.Region)
		region = envCfg.Region
	}
	log.Printf("Configuring bedrock Using region %v", region)
	bedrockConfig := Init(region, "TEST_TAH_SOLUTION_ID", "1.0")

	answer, stop_reason, err := bedrockConfig.InvokeClaudeModel("Hello Claude, what is one plus one?", 0.5, 1, 250, 200)
	if err != nil {
		t.Fatal(err)
	}
	if stop_reason != "end_turn" {
		t.Fatal("Stop Reason is not end_turn")
	}
	if answer == "" {
		t.Fatal("Answer is empty")
	}
	// this is consistently the answer I get, if you don't get it exactly comment out this test
	const expectedVal = "one plus one equals two"
	var ans = strings.ToLower(answer)
	if strings.Contains(expectedVal, ans) {
		t.Fatalf("Answer is not correct; expected - %s, got %s", expectedVal, ans)
	}

	log.Printf("Claude Response: %s", answer)
}

func TestStableDiffusion(t *testing.T) {
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		t.Fatalf("error loading env config: %v", err)
	}
	region := TAH_CORE_BEDROCK_REGION
	if TAH_CORE_BEDROCK_REGION == "" || TAH_CORE_BEDROCK_REGION == "null" {
		log.Printf("TAH_CORE_BEDROCK_REGION not set, using TAH_CORE_REGION=%v", envCfg.Region)
		region = envCfg.Region
	}
	bedrockConfig := Init(region, "TEST_TAH_SOLUTION_ID", "1.0")
	inputSeed := 0
	stablePrompt := "Generate a profile picture of a 50 year old male with the city of Dallas as a background"
	img, finishReason, outputSeed, result, err := bedrockConfig.InvokeStableDiffusionModel([]string{stablePrompt}, 10, inputSeed, 100, 512, 512)
	if err != nil {
		t.Error(err)
	}
	if finishReason != "SUCCESS" {
		t.Error("finish reason not success")
	}
	if result != "success" {
		t.Error("result not success")
	}
	if outputSeed != float64(inputSeed) {
		t.Error("outputSeed does not match inputSeed")
	}

	log.Printf("Generated Image: %+v", img)

	// f, err := os.Create("image.txt")

	// if err != nil {
	// 	t.Error(err)
	// }

	// defer f.Close()

	// _, err2 := f.WriteString(photoEncoded)

	// if err2 != nil {
	// 	t.Error(err2)
	// }
}
