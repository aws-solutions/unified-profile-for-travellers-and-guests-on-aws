// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"fmt"
	"math/rand"
	"strings"
	"tah/upt/source/tah-core/aurora"
)

const countTablePartitions int = 10000
const countTableProfileObject string = "profile"

func (h DataHandler) CreateRecordCountTable(domain string) error {
	h.Tx.Info("Creating Record Count table for domain: %s", domain)
	query := h.createProfileCountTableSql(domain)
	_, err := h.AurSvc.Query(query)
	if err != nil {
		return err
	}

	return nil
}

func (h DataHandler) createProfileCountTableSql(domain string) string {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			object_type VARCHAR(255) NOT NULL,
			partition INT NOT NULL,
			dimension VARCHAR(255) DEFAULT 'all' NOT NULL,
			count BIGINT NOT NULL,
			last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (object_type, partition, dimension)
		);
	`, counterTableName(domain))
	return formatQuery(query)
}

// For each object type, we initialize n number of partitions to distribute
// count columns (n being const countTablePartitions). This prevents lock
// contention on individual count rows.
//
// partition is 1 based from 1 to countTablePartitions
func (h DataHandler) InitializeCount(domain string, objectTypeNames []string) error {
	h.Tx.Debug("Initializing count table for domain:  %s", domain)
	h.Tx.Debug("Object types being added: %s", strings.Join(objectTypeNames, ", "))
	queries := h.populateInitialCountsSql(domain, objectTypeNames)
	for _, query := range queries {
		_, err := h.AurSvc.Query(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h DataHandler) populateInitialCountsSql(domain string, objectTypeNames []string) []string {
	var queries []string

	for _, objName := range objectTypeNames {
		query := fmt.Sprintf(`
			INSERT INTO %s (object_type, dimension, count, partition)
			SELECT '%s', 'all', 0, generate_series(1, %d);
		`, counterTableName(domain), objName, countTablePartitions)
		queries = append(queries, formatQuery(query))
	}

	return queries
}

func (h DataHandler) DeleteRecordCountTable(domain string) error {
	h.Tx.Info("Deleting Record Count table for domain: %s", domain)
	query := h.deleteRecordCountTableSql(domain)
	_, err := h.AurSvc.Query(query)

	return err
}

func (h DataHandler) deleteRecordCountTableSql(domain string) string {
	query := fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
	`, counterTableName(domain))

	return formatQuery(query)
}

// UpdateCount keeps the `CreateProfileCountTable` in sync with actual counts as we perform create/delete operations. This allows us to keep a running total of records without an expensive count(*) query.
// 1. Should always be part of the same transaction as the query(s) it is updating the count for.
// 2. Must be called every time we create/delete an object or profile (incl Merge, Unmerge)
// we pass a connection struct to ensure the count update is done within the appropriate transaction
// even in the context of concurrent callers
func (h DataHandler) UpdateCount(conn aurora.DBConnection, domain string, objectTypeName string, n int) error {
	h.Tx.Debug("Incrementing %s", objectTypeName)
	partition := rand.Intn(countTablePartitions) + 1 // rand int from [0, n), then converted to [1 to n]
	query, args := h.updateCountSql(domain, objectTypeName, partition, n)
	_, err := conn.Query(query, args...)
	return err
}

func (h DataHandler) updateCountSql(domain string, objectTypeName string, partition int, n int) (query string, args []interface{}) {
	query = fmt.Sprintf(`
		UPDATE %s
		SET count = count + %d
		WHERE object_type = ?
		AND partition = ?
		AND dimension = 'all';
	`, counterTableName(domain), n)
	args = append(args, objectTypeName, partition)
	return formatQuery(query), args
}

func (h DataHandler) GetCount(domain string, objectTypeName string) (int64, error) {
	h.Tx.Info("Getting count for %s", objectTypeName)
	query := h.getCountSql(domain)
	res, err := h.AurSvc.Query(query, objectTypeName)
	if err != nil {
		return 0, err
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("expected one row to be returned")
	}
	count, ok := res[0]["count"].(int64)
	if !ok {
		return 0, fmt.Errorf("expected count to be int64")
	}
	return count, nil
}

func (h DataHandler) getCountSql(domain string) string {
	query := fmt.Sprintf(`
		SELECT SUM(count)::bigint AS count
		FROM %s
		WHERE object_type = ?
		AND dimension = 'all';
	`, counterTableName(domain))

	return formatQuery(query)
}

func (h DataHandler) GetAllObjectCounts(domain string) (map[string]int64, error) {
	h.Tx.Info("Getting object counts")
	query := h.getAllObjectCountsSql(domain)
	res, err := h.AurSvc.Query(query)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range res {
		count, ok := row["count"].(int64)
		if !ok {
			return nil, fmt.Errorf("expected count to be int64")
		}
		objectType, ok := row["object_type"].(string)
		if !ok {
			return nil, fmt.Errorf("expected object_type to be a string")
		}
		counts[objectType] = count
	}

	return counts, nil
}

func (h DataHandler) getAllObjectCountsSql(domain string) string {
	query := fmt.Sprintf(`
		SELECT object_type, SUM(count)::bigint AS count
		FROM %s
		WHERE dimension = 'all'
		GROUP BY object_type;
	`, counterTableName(domain))

	return formatQuery(query)
}

func counterTableName(domain string) string {
	return domain + "_counts"
}
