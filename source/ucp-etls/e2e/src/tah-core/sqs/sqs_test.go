package sqs

import (
	"os"
	core "tah/core/core"
	"testing"
)

/**********
* Test SQS queue
******************/

var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

func TestSQSCrud(t *testing.T) {
	svc := Init(TAH_CORE_REGION)
	_, err := svc.Create("Test-Queue" + core.GeneratUniqueId())
	if err != nil {
		t.Errorf("[TestSQS]error creating queue: %v", err)
	}
	err = svc.SetPolicy("Service", "profile.amazonaws.com", []string{"SQS:*"})
	if err != nil {
		t.Errorf("[TestSQS]error setting queue policy queue: %v", err)
	}
	_, err2 := svc.GetAttributes()
	if err2 != nil {
		t.Errorf("[TestSQS] error getting queue atttribute: %v", err2)
	}
	err = svc.Delete()
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue: %v", err)
	}
}

func TestSQSMessages(t *testing.T) {
	svc := Init(TAH_CORE_REGION)
	qName := "Test-Queue-2" + core.GeneratUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[TestSQS]error creating queue: %v", err)
	}
	err = svc.Send("test msg")
	if err != nil {
		t.Errorf("[TestSQS]error sending msgd to  queue: %v", err)
	}
	content, err1 := svc.Get()
	if err1 != nil {
		t.Errorf("[TestSQS]error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "test msg" {
		t.Errorf("[TestSQS] message in queue should be %v", "test msg")
	}
	err = svc.SendWithStringAttributes("test msg 2", map[string]string{"key": "testVal"})
	if err != nil {
		t.Errorf("[TestSQS]error sending msgd to  queue: %v", err)
	}
	content, err1 = svc.Get()
	if err1 != nil {
		t.Errorf("[TestSQS] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "test msg 2" {
		t.Errorf("[TestSQS] message in queue should be %v", "test msg 2")
	}
	if content.Peek[0].MessageAttributes["key"] != "testVal" {
		t.Errorf("[TestSQS] message in queue should have message attribute %v", map[string]string{"key": "testVal"})
	}
	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue by name: %v", err)
	}
}
