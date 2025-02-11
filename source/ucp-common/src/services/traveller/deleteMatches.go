// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"tah/upt/source/tah-core/db"
)

func DeleteMatches(matchDb db.DBConfig, domainName string, inputId string) error {
	data, err := SearchMatches(matchDb, domainName, inputId)
	if err != nil {
		return err
	}
	toDelete := []map[string]interface{}{}
	for _, el := range data {
		toDelete = append(toDelete, map[string]interface{}{
			matchDb.PrimaryKey: el.Pk,
			matchDb.SortKey:    el.Sk,
		})
	}
	err = matchDb.DeleteMany(toDelete)
	if err != nil {
		return err
	}

	return nil
}

func DeleteMatch(matchDb db.DBConfig, domainName string, id string, matchID string) error {
	_, err := matchDb.Delete(map[string]interface{}{
		matchDb.PrimaryKey: domainName + "_" + id,
		matchDb.SortKey:    "match_" + matchID,
	})
	return err
}
