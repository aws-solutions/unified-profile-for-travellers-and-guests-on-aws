package kinesis

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	kinesis "tah/core/kinesis"
	"testing"
)

var KINESIS_NAME_REAL_TIME = os.Getenv("KINESIS_NAME_REAL_TIME")
var LAMBDA_NAME_REAL_TIME = os.Getenv("LAMBDA_NAME_REAL_TIME")
var KINESIS_NAME_OUTPUT_REAL_TIME = os.Getenv("KINESIS_NAME_OUTPUT_REAL_TIME")
var UCP_REGION = getRegion()

type businessObjectRecord struct {
	ObjectType   string      `json:"objectType"`
	ModelVersion string      `json:"modelVersion"`
	Data         interface{} `json:"data"`
}

func TestKinesis(t *testing.T) {
	log.Printf("E2EDataStream] Running E2E test for data stream")
	incfg := kinesis.Init(KINESIS_NAME_REAL_TIME, UCP_REGION)
	outcfg := kinesis.Init(KINESIS_NAME_OUTPUT_REAL_TIME, UCP_REGION)

	outcfg.InitConsumer(kinesis.ITERATOR_TYPE_LATEST)
	log.Printf("E2EDataStream] 1- sending records to kinesis stream %v", KINESIS_NAME_REAL_TIME)

	clickstreamContent, err := os.ReadFile("../../../../test_data/clickstream/data1.json")
	log.Printf("%s", string(clickstreamContent))
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return
	}

	var v interface{}
	json.Unmarshal(clickstreamContent, &v)

	clickstreamRecord := businessObjectRecord{
		ObjectType:   "clickstream",
		ModelVersion: "1",
		Data:         v,
	}

	clickstreamSerialized, _ := json.Marshal(clickstreamRecord)
	clickstreamString := string(clickstreamSerialized)
	log.Printf("%s", clickstreamString)

	err, errs := incfg.PutRecords([]kinesis.Record{
		kinesis.Record{
			Pk:   clickstreamRecord.ObjectType + "-" + clickstreamRecord.ModelVersion + "-1",
			Data: clickstreamString,
		},
	})
	if err != nil {
		t.Errorf("[E2EDataStream] error sending data to stream: %v. %+v", err, errs)
	}

	clickstreamContentOutput, err := os.ReadFile("../../../../test_data/clickstream/data1_expected.json")
	if err != nil {
		log.Printf("Error reading JSON file: %s", err)
		return
	}

	var vout map[string]interface{}
	var vexpected map[string]interface{}
	json.Unmarshal(clickstreamContentOutput, &vout)

	log.Printf("E2EDataStream] 3-Fetching records from output stream %v", KINESIS_NAME_OUTPUT_REAL_TIME)
	recs, err2 := outcfg.FetchRecords(10)
	log.Printf("...Fetched: %v", recs)
	if err2 != nil {
		t.Errorf("[E2EDataStream] error fetching data from stream: %v", err2)
	}

	if len(recs) != 1 {
		t.Errorf("[E2EDataStream] Output Stream should have 1 record")
	} else {
		kinesisData := []byte(recs[0].Data)
		json.Unmarshal(kinesisData, &vexpected)
		if !reflect.DeepEqual(vout, vexpected) {
			first, _ := json.Marshal(vout)
			t.Errorf("[E2EDataStream] first record should be %v and second %v", string(first), recs[0].Data)
		}
	}
}

func getRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
