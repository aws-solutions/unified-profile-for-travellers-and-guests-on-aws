package notif

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

type WebSocketConfig struct {
	Endpoint string
	Region   string
}

func Init(endpoint string, region string) WebSocketConfig {
	log.Printf("[NOTIF] Init Websocket for endpoint  %s and region %s", endpoint, region)
	return WebSocketConfig{Endpoint: endpoint, Region: region}
}

func (ws WebSocketConfig) Notify(connectionID string, notif interface{}) error {

	log.Printf("[NOTIF] Sending notification to connectionID %+v", connectionID)
	//escapedId := strings.Replace(conn.ConnetionId,"=","%3D",-1)
	rqBody, err := json.Marshal(notif)
	if err != nil {
		log.Printf("[NOTIF] error while marshalling notif:  %+v", err)
		return err
	}
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), os.Getenv("AWS_SESSION_TOKEN"))
	signer := v4.NewSigner(creds)
	bodyPayload := strings.NewReader(string(rqBody))
	endpoint := strings.Replace(ws.Endpoint, "wss", "https", -1) + "/@connections/" + connectionID
	req, _ := http.NewRequest("POST", endpoint, bodyPayload)
	req.Header.Set("Content-Type", "application/json")
	signer.Sign(req, bodyPayload, "execute-api", ws.Region, time.Now())
	log.Printf("[NOTIF] signed request to POST:  %+v", req)
	//req.Body = ioutil.NopCloser(bodyPayload)
	//res, err2 := sigv4.Post(WEB_SOCKET_ENDPOINT+escapedId , "application/json", data)
	dump, _ := httputil.DumpRequest(req, true)
	log.Printf("[NOTIF] REQUEST DUMP:  %s", dump)
	res, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		log.Printf("[NOTIF] response error:  %+v", err2)
	}
	log.Printf("[NOTIF] response from web socket POST:  %+v", res)
	return err2
}
