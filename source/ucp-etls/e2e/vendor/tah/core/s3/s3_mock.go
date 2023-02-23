package s3

import core "tah/core/core"

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

func (c MockS3Config) GetBucket() string {
	return c.Bucket
}
func (c MockS3Config) GetRegion() string {
	return c.Region
}
