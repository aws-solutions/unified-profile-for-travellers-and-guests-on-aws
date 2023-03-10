package notif

import "testing"

/*******
* Notif
*****/
func TestNotif(t *testing.T) {
	endpoint := "http://endpoint.com"
	notifClient := Init(endpoint, "us-east-1")
	if notifClient.Endpoint != endpoint {
		t.Errorf("[core][notif] Endpoint should be %s", endpoint)
	}
}
