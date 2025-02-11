// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bedrock

import (
	"encoding/json"
	"log"

	core "tah/upt/source/tah-core/core"

	"github.com/aws/aws-sdk-go/aws"

	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	bedrock "github.com/aws/aws-sdk-go/service/bedrock"
	bedrockInterface "github.com/aws/aws-sdk-go/service/bedrockruntime"
)

type Config struct {
	modelClient     *bedrock.Bedrock
	interfaceClient *bedrockInterface.BedrockRuntime
	Region          string
}

type TitanBody struct {
	InputText            string          `json:"inputText"`
	TextGenerationConfig TitanTextConfig `json:"textGenerationConfig"`
}

type TitanTextConfig struct {
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"topP"`
	MaxTokenCount int      `json:"maxTokenCount"`
	StopSequences []string `json:"stopSequences"`
}

type ClaudeBodyV2_1 struct {
	Prompt            string   `json:"prompt"`
	Temperature       float64  `json:"temperature"`
	TopP              float64  `json:"top_p"`
	TopK              int      `json:"top_k"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences"`
}

type ClaudeBodyV3Sonnet struct {
	AnthropicVersion string     `json:"anthropic_version"`
	Temperature      float64    `json:"temperature"`
	TopP             float64    `json:"top_p"`
	TopK             int        `json:"top_k"`
	StopSequences    []string   `json:"stop_sequences"`
	MaxTokens        int        `json:"max_tokens"`
	Messages         []Messages `json:"messages"`
}

type Messages struct {
	Role    string        `json:"role"`
	Content []TextContent `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type StableDiffusionBody struct {
	TextPrompts []StableDiffusionPrompt `json:"text_prompts"`
	CfgScale    int                     `json:"cfg_scale"`
	Seed        int                     `json:"seed"`
	Steps       int                     `json:"steps"`
	Width       int                     `json:"width"`
	Height      int                     `json:"height"`
}

type StableDiffusionPrompt struct {
	Text string `json:"text"`
}

const ANTHROPIC_MODEL_ID_V2_1 = "anthropic.claude-v2"
const ANTHROPIC_MODEL_ID_V3 = "anthropic.claude-3-sonnet-20240229-v1:0"
const STABLE_DIFFUSION_MODEL_ID = "stability.stable-diffusion-xl-v1"
const modelResponseKeyAnthropic = "content"
const modelStopReasonAnthropic = "stop_reason"

func Init(region, solutionId, solutionVersion string) Config {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := core.CreateClient(solutionId, solutionVersion)
	cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(client)
	return Config{
		modelClient:     bedrock.New(sess, cfg),
		interfaceClient: bedrockInterface.New(sess, cfg),
	}
}

func (c Config) InvokeModelWithBody(body []byte, modelId string) (map[string]interface{}, error) {
	MIMEType := "application/json"
	bedRockInput := bedrockInterface.InvokeModelInput{
		Accept:      &MIMEType,
		ContentType: &MIMEType,
		ModelId:     &modelId,
		Body:        body,
	}
	log.Printf("[bedrock] Request: %+v", bedRockInput)
	out, err := c.interfaceClient.InvokeModel(&bedRockInput)
	if err != nil {
		return map[string]interface{}{}, err
	}
	serializedResponse := out.Body
	log.Printf("[bedrock] Response: %s", string(serializedResponse))
	var v map[string]interface{}
	err = json.Unmarshal(serializedResponse, &v)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return v, nil
}

func (c Config) InvokeClaudeModel(promptFragment string, temperature float64, topP float64, topK int, maxTokensToSample int) (string, string, error) {
	// Create body
	body := c.GenerateClaudeBodyV3Sonnet(promptFragment, temperature, topP, topK, maxTokensToSample)
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("[InvokeClaudeModel] error creating claude body; error: %v", err)
	}

	// Invoke model
	response, err := c.InvokeModelWithBody(bodyBytes, ANTHROPIC_MODEL_ID_V3)
	if err != nil {
		return "", "", fmt.Errorf("[InvokeClaudeModel] error invoking claude; error: %v", err)
	}

	// Extract answer
	answerInterface, ok := response[modelResponseKeyAnthropic].([]interface{})
	if !ok {
		return "", "", fmt.Errorf("[InvokeClaudeModel] error parsing response")
	}

	if len(answerInterface) == 0 {
		return "", "", fmt.Errorf("[InvokeClaudeModel] claude did not respond with an answer")
	}

	content, ok := answerInterface[0].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("[InvokeClaudeModel] error parsing content from response")
	}

	answer, ok := content["text"].(string)
	if !ok {
		return "", "", fmt.Errorf("[InvokeClaudeModel] answer is not a string")
	}
	stopReasonInterface := response[modelStopReasonAnthropic]
	stopReason, ok := stopReasonInterface.(string)
	if !ok {
		return "", "", fmt.Errorf("[InvokeClaudeModel] stop reason is not a string")
	}

	return answer, stopReason, nil
}

func (c Config) GenerateClaudeBodyV2_1(promptFragment string, temperature float64, topP float64, topK int, maxTokensToSample int) ClaudeBodyV2_1 {
	prompt := "\n\nHuman:" + promptFragment + "\n\nAssistant:"
	stopSequence := []string{"\n\nHuman:"}
	return ClaudeBodyV2_1{
		Prompt:            prompt,
		Temperature:       temperature,
		TopP:              topP,
		MaxTokensToSample: maxTokensToSample,
		StopSequences:     stopSequence,
	}
}

func (c Config) GenerateClaudeBodyV3Sonnet(promptFragment string, temperature float64, topP float64, topK int, maxTokensToSample int) ClaudeBodyV3Sonnet {
	// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-anthropic-claude-messages.html#model-parameters-anthropic-claude-messages-request-response
	anthropicVersion := "bedrock-2023-05-31" // as specified in the above url
	messageRole := "user"                    // Valid values are user and assistant
	contentType := "text"                    // Valid values are image and text
	stopSequence := []string{"end_turn"}

	return ClaudeBodyV3Sonnet{
		AnthropicVersion: anthropicVersion,
		MaxTokens:        maxTokensToSample,
		Temperature:      temperature,
		TopP:             topP,
		TopK:             topK,
		StopSequences:    stopSequence,
		Messages: []Messages{
			{
				Role: messageRole,
				Content: []TextContent{{
					Type: contentType,
					Text: promptFragment,
				}},
			},
		},
	}
}

func (c Config) GenerateTitanBody(inputText string, temperature float64, topP float64, maxTokenCount int, stopSequences []string) TitanBody {
	return TitanBody{
		InputText:            inputText,
		TextGenerationConfig: TitanTextConfig{Temperature: temperature, TopP: topP, MaxTokenCount: maxTokenCount, StopSequences: stopSequences},
	}
}

func (c Config) InvokeStableDiffusionModel(textPrompts []string, cfgScale int, seed int, steps int, width int, height int) (string, string, float64, string, error) {
	if width == 0 {
		width = 512
	}
	if height == 0 {
		height = 512
	}
	body := c.GenerateStableDiffusionBody(textPrompts, cfgScale, seed, steps, width, height)
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", -1, "", err
	}
	log.Printf("[bedrock] StableDiffusion body: %+v", string(bodyBytes))

	responseInterface, err := c.InvokeModelWithBody(bodyBytes, STABLE_DIFFUSION_MODEL_ID)
	if err != nil {
		return "", "", -1, "", err
	}

	artifactsInterface, ok := responseInterface["artifacts"]
	if !ok {
		return "", "", -1, "", fmt.Errorf("artifacts is not in response")
	}
	artifacts, ok := artifactsInterface.([]interface{})
	if !ok {
		return "", "", -1, "", fmt.Errorf("artifacts is not a map string interface")
	}
	artifactInterface := artifacts[0]
	artifact, ok := artifactInterface.(map[string]interface{})
	if !ok {
		return "", "", -1, "", fmt.Errorf("artifact is not a map string interface")
	}
	photoEncodedInterface, ok := artifact["base64"]
	if !ok {
		return "", "", -1, "", fmt.Errorf("photo encoded is in artifacts")
	}
	photoEncoded, ok := photoEncodedInterface.(string)
	if !ok {
		return "", "", -1, "", fmt.Errorf("photo encoded is not a string")
	}
	finishReasonInterface, ok := artifact["finishReason"]
	if !ok {
		return "", "", -1, "", fmt.Errorf("finish reason is not in artifacts")
	}
	finishReason, ok := finishReasonInterface.(string)
	if !ok {
		return "", "", -1, "", fmt.Errorf("finish reason is not a string")
	}
	outputSeedInterface, ok := artifact["seed"]
	if !ok {
		return "", "", -1, "", fmt.Errorf("seed is not in artifacts")
	}
	outputSeed, ok := outputSeedInterface.(float64)
	if !ok {
		return "", "", -1, "", fmt.Errorf("seed is not an string")
	}
	resultInterface, ok := responseInterface["result"]
	if !ok {
		return "", "", -1, "", fmt.Errorf("result is not in response")
	}
	result, ok := resultInterface.(string)
	if !ok {
		return "", "", -1, "", fmt.Errorf("result is not a string")
	}
	return photoEncoded, finishReason, outputSeed, result, nil
}

func (c Config) GenerateStableDiffusionBody(textPrompts []string, cfgScale int, seed int, steps int, width int, height int) StableDiffusionBody {
	textPromptObjects := make([]StableDiffusionPrompt, len(textPrompts))
	for textPrompt := range textPrompts {
		textPromptObjects[textPrompt] = c.GenerateStableDiffusionPrompt(textPrompts[textPrompt])
	}
	return StableDiffusionBody{
		TextPrompts: textPromptObjects,
		CfgScale:    cfgScale,
		Seed:        seed,
		Steps:       steps,
		Width:       width,
		Height:      height,
	}
}

func (c Config) GenerateStableDiffusionPrompt(text string) StableDiffusionPrompt {
	return StableDiffusionPrompt{
		Text: text,
	}
}
