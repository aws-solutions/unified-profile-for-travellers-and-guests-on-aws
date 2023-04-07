package kinesis

import (
	"log"
	"os"
	"testing"
)

var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")
var TAH_CORE_ACCOUNT = os.Getenv("TAH_CORE_ACCOUNT")

func InterfaceTestFn(i IConfig) {

}

func TestInterface(t *testing.T) {
	InterfaceTestFn(&MockConfig{})
	InterfaceTestFn(&Config{})
}

func TestStreamCrud(t *testing.T) {
	streamName := "tah_core_test_stream"
	testRec1 := "{\"key\":\"test record 1\"}"
	testRec2 := "{\"key\":\"test record 2\"}"
	log.Printf("0-Creating stream %v", streamName)
	cfg, err := InitAndCreate(streamName, TAH_CORE_REGION)
	if err != nil {
		t.Errorf("[TestKinesis] error creating stream: %v", err)
	}
	cfg.WaitForStreamCreation(10)
	log.Printf("1-Describing Stream %v", streamName)
	stream, err3 := cfg.Describe()
	if err3 != nil {
		t.Errorf("[TestKinesis] error describing stream: %v", err3)
	}
	log.Printf("Stream: %+v", stream)

	err3 = cfg.InitConsumer(ITERATOR_TYPE_LATEST)
	if err3 != nil {
		t.Errorf("[TestKinesis] error initializing consumer %v", err3)
	}
	log.Printf("2-Sending 2 records to stream %v", streamName)
	err, errs := cfg.PutRecords([]Record{
		Record{
			Pk:   "1",
			Data: testRec1,
		},
		Record{
			Pk:   "2",
			Data: testRec2,
		}})
	if err != nil {
		t.Errorf("[TestKinesis] error sending data to stream: %v. %+v", err, errs)
	}
	log.Printf("3-Fetching records from stream %v", streamName)
	recs, err2 := cfg.FetchRecords(5)
	log.Printf("...Fetched: %v", recs)
	if err2 != nil {
		t.Errorf("[TestKinesis] error fetching data from stream: %v", err2)
	}
	if len(recs) != 2 {
		t.Errorf("[TestKinesis] Stream should have 2 records")
	} else {
		if recs[0].Data != testRec1 {
			t.Errorf("[TestKinesis] firt record should be %v", testRec1)
		}
		if recs[1].Data != testRec2 {
			t.Errorf("[TestKinesis] firt record should be %v", testRec2)
		}
	}

	log.Printf("4-Deleting stream%v", streamName)
	err = cfg.Delete(streamName)
	if err != nil {
		t.Errorf("[TestKinesis] error deleting stream: %v", err)
	}
}
