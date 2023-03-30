package main

import (
	"context"
	"os"
	core "tah/core/core"
	sqs "tah/core/sqs"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var UCP_REGION = getRegion()

func TestSqsErrors(t *testing.T) {
	svc := sqs.Init(UCP_REGION)
	qName := "Test-Queue-2" + core.GeneratUniqueId()
	_, err := svc.Create(qName)
	if err != nil {
		t.Errorf("[E2EDataStream] Could not create test queue")
	}
	req := events.KinesisEvent{Records: []events.KinesisEventRecord{
		events.KinesisEventRecord{
			Kinesis: events.KinesisRecord{
				Data: []byte("illformated record"),
			},
		},
	}}
	HandleRequestWithServices(context.Background(), req, svc)

	content, err1 := svc.Get()
	if err1 != nil {
		t.Errorf("[TestSQS] error getting msg from queue: %v", err1)
	}
	if len(content.Peek) == 0 {
		t.Errorf("[TestSQS] no record returned")
	}
	if content.Peek[0].Body != "illformated record" {
		t.Errorf("[TestSQS] message in queue should be %v", "illformated record")
	}

	err = svc.DeleteByName(qName)
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue by name: %v", err)
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
