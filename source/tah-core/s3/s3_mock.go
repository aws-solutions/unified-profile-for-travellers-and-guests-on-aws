// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package s3

import core "tah/upt/source/tah-core/core"

type MockS3Config struct {
	Region string
	Bucket string
}

func (c MockS3Config) SaveJson(path string, id string, data []byte) error {
	return nil
}

func (c MockS3Config) SetTx(tx core.Transaction) error {
	return nil
}

func (c MockS3Config) Delete(path string, id string) error {
	return nil
}

func (c MockS3Config) GetBucket() string {
	return c.Bucket
}
func (c MockS3Config) GetRegion() string {
	return c.Region
}
