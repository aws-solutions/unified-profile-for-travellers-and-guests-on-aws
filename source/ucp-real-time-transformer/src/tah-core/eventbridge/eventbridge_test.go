package eventbridge

import "testing"

/**********
* EventBridge
************/
type TestEventbridgeBody struct {
	Value string
}

//TODO: update these test too dynamically create cognito
func TestEventBridgeCreateHttpApiScheduledTrigger(t *testing.T) {
	/*cfg := eventbridge.Init("eu-central-1")
	body := TestEventbridgeBody{
		Value: "test",
	}
	endpoint := "https://wnb5esk1q8.execute-api.eu-central-1.amazonaws.com//api/hapi/searchBookings"
	clientID := "my_client_id"
	cognitoEndpoint := "https://tah-ind-conn-domain-1234567890-dev.auth.eu-central-1.amazoncognito.com/oauth2/token"
	clientSecret := "86t3d2839fgd928gd92"
	oAuthScope := "server/resource"

	err := cfg.CreateHttpApiScheduledTrigger(0, "api-test-rule", 7, "POST", endpoint, body, cognitoEndpoint, clientID, clientSecret, oAuthScope)
	if err == nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] shoudl return an error is 0 target provided")
	}
	err = cfg.CreateHttpApiScheduledTrigger(1, "api-test-rule", 7, "POST", endpoint, body, cognitoEndpoint, clientID, clientSecret, oAuthScope)
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] error creating secdule API event: %v", err)
	}
	err = cfg.CreateHttpApiScheduledTrigger(1, "api-test-rule", 7, "POST", endpoint, body, cognitoEndpoint, clientID, clientSecret, oAuthScope)
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] API should be idempotent (not return error if called twice)")
	}
	rule, err2 := cfg.RetreiveApiRule("api-test-rule")
	log.Printf("RetreiveApiRule: %+v", rule)
	_, _, err = cfg.GetErrors("api-test-rule")
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] %v", err)
	}
	if err2 != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] retreive api rule returns error: %v", err)
	}
	err = cfg.DeleteApiRule("api-test-rule")
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] error deleting sechdule API event: %v", err)
	}
	err = cfg.DeleteApiDestination(endpoint)
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] error deleting api destination %v", err)
	}
	err = cfg.DeleteCognitoConnection(cognitoEndpoint)
	if err != nil {
		t.Errorf("[TestEventBridgeCreateHttpApiScheduledTrigger] error deleting connection %v", err)
	}*/

}
