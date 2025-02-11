// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package usecase

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func PutProfile(msg events.SQSMessage, sr ServiceRegistry) error {
	var rq PutProfileRq
	err := json.Unmarshal([]byte(msg.Body), &rq)
	if err != nil {
		sr.Tx.Error("Error unmarshalling PutProfileRq: %v", err)
		return err
	}

	sr.Tx.Info("Inserting profile %s into domain %s", rq.Profile.ProfileId, rq.Profile.Domain)
	err = sr.CPConfig.PutProfileAsObject(rq.Profile)
	if err != nil {
		sr.Tx.Error("Error inserting profile %s into domain %s: %v", rq.Profile.ProfileId, rq.Profile.Domain, err)
		return err
	}

	return nil
}
