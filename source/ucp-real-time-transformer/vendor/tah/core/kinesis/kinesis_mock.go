package kinesis

type MockConfig struct {
}

func (c *MockConfig) Create(streamName string) error {
	return nil
}

func (c *MockConfig) Delete(streamName string) error {
	return nil
}

func (c *MockConfig) WaitForStreamCreation(timeoutSeconds int) (string, error) {
	return "", nil
}

func (c *MockConfig) Describe() (Stream, error) {
	return Stream{}, nil
}

func (c *MockConfig) PutRecords(recs []Record) (error, []IngestionError) {
	return nil, []IngestionError{}
}

func (c *MockConfig) InitConsumer(iteratorType string) error {
	return nil
}

func (c *MockConfig) FetchRecords(timeoutSeconds int) ([]Record, error) {
	return []Record{}, nil
}
