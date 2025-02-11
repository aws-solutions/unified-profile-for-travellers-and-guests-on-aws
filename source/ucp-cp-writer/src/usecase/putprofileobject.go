// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func PutProfileObject(msg events.SQSMessage, sr ServiceRegistry) error {
	var rq PutProfileObjectRq
	err := json.Unmarshal([]byte(msg.Body), &rq)
	if err != nil {
		sr.Tx.Error(fmt.Sprintf("Error unmarshalling PutProfileObjectRq: %v", err))
		return err
	}

	sr.Tx.Info("Inserting object %s for profile %s", rq.ProfileObject.ID, rq.ProfileID)
	err = sr.CPConfig.PutProfileObjectFromLcs(rq.ProfileObject, rq.ProfileID)
	if err != nil {
		sr.Tx.Error(fmt.Sprintf("Error while inserting object: %v", err))
		return err
	}

	return nil
}
