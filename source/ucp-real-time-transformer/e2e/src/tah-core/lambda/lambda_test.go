package lambda

import (
	"encoding/json"
	core "tah/core/core"
	"testing"
)

/*******
* LAMBDA
*****/

type EnablementSession struct {
	ContactID string `json:"contactID"`
}

func (p EnablementSession) Decode(dec json.Decoder) (error, core.JSONObject) {
	return dec.Decode(&p), p
}
func TestLambda(t *testing.T) {
	lambdaSvc := Init("testFunction")
	res := EnablementSession{}
	_, err := lambdaSvc.Invoke(map[string]string{"key": "value"}, res, false)
	if err == nil {
		t.Errorf("[core][Lambda] Should return Error while invoking invalid lambda")
	}
	//t.Errorf("[core][Lambda] Error while invoking lambda %v", res2)
}
