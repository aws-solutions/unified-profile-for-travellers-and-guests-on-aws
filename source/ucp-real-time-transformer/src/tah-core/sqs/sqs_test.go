package sqs

import "testing"

/**********
* Test SQS queue
******************/
func TestSQS(t *testing.T) {
	svc := Init("eu-central-1")
	_, err := svc.Create("Test-Queue")
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
	_, err = svc.Create("Test-Queue-2")
	if err != nil {
		t.Errorf("[TestSQS]error creating queue: %v", err)
	}
	err = svc.DeleteByName("Test-Queue-2")
	if err != nil {
		t.Errorf("[TestSQS]error deleting queue by name: %v", err)
	}
}
