package model

import "time"

type ResponseWrapper struct {
}

type PartitionUpdateStatus struct {
	Pk          string    `json:"item_id"`
	Sk          string    `json:"item_type"`
	LastUpdated time.Time `json:"lastUpdated"`
	TableName   string    `json:"tableName"`
	Error       string    `json:"error"`
	Status      string    `json:"status"`
}
