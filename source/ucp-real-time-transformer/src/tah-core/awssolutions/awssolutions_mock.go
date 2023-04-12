package awssolutions

type MockConfig struct{}

func (c MockConfig) SendMetrics(data map[string]interface{}) (bool, error) {
	return true, nil
}
