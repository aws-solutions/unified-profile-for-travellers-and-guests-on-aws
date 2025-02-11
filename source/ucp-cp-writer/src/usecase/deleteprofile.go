// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"

	model "tah/upt/source/ucp-common/src/model/dynamo-schema"

	"github.com/aws/aws-lambda-go/events"
)

func DeleteProfile(msg events.SQSMessage, domainName string, sr ServiceRegistry) error {
	var rq DeleteProfileRq
	var cpIdMap model.CpIdMap
	err := json.Unmarshal([]byte(msg.Body), &rq)
	if err != nil {
		sr.Tx.Error("Error unmarshalling DeleteProfileRq: %v", err)
		return err
	}

	err = sr.DbConfig.ConsistentSafeGet(domainName, "lcs_"+rq.ProfileID, &cpIdMap)
	if err != nil {
		sr.Tx.Error("Error while getting Connect ID for profile %s: %v", rq.ProfileID, err)
		return err
	}
	if cpIdMap.DomainName == "" && cpIdMap.ConnectId == "" {
		retriesLeft := rq.NumRetries
		if retriesLeft > 0 {
			sr.Tx.Info("No Connect ID found for profile %s, retrying record in 5 seconds", rq.ProfileID)
			sr.Tx.Debug("Retries left: %d", retriesLeft)
			// Retry the message
			newDeleteRq := DeleteProfileRq{
				ProfileID:  rq.ProfileID,
				NumRetries: retriesLeft - 1,
			}
			var newBody string
			byteBody, err := json.Marshal(newDeleteRq)
			if err != nil {
				sr.Tx.Error("Error marshalling DeleteProfileRq: %v", err)
				return err
			}
			newBody = string(byteBody)
			msgAttr := convertAttributes(msg.MessageAttributes)
			err = sr.CpQueue.SendWithStringAttributesAndDelay(newBody, msgAttr, 5)
			return err
		} else {
			sr.Tx.Error("No Connect ID found for profile %s, giving up and returning error", rq.ProfileID)
			return err
		}
	}

	cpId := cpIdMap.CpId

	sr.Tx.Info("Deleting UPT profile %s with Connect ID %s", rq.ProfileID, cpId)
	err = sr.CPConfig.DeleteProfile(cpId)
	if err != nil {
		sr.Tx.Error("Error deleting profile %s: %v", cpId, err)
		return err
	}

	return nil
}

func convertAttributes(attributes map[string]events.SQSMessageAttribute) map[string]string {
	result := make(map[string]string)
	for key, attr := range attributes {
		if attr.StringValue != nil {
			result[key] = *attr.StringValue
		} else if attr.BinaryValue != nil {
			result[key] = string(attr.BinaryValue)
		}
	}
	return result
}
