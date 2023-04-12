package cognito

type MockCognitoConfig struct {
}

func (c MockCognitoConfig) SetTokenUrl(url string) {
}
func (c MockCognitoConfig) GetAccessTokenFromClientCredentials(clientID string, clientSecret string, scope string) (string, error) {
	return "mock_token", nil
}
