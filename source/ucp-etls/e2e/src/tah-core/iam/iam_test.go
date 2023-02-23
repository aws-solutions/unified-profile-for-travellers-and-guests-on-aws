package iam

import "testing"

/**********
* IAM
************/

func TestIam(t *testing.T) {
	acc := AccountFromRoleArn("arn:aws:iam::1234567890:role/tah-connector-hapi-bookings")
	if acc != "1234567890" {
		t.Errorf("[AcountFromRoleArn] error extracting account %v", acc)
	}
	arn := BuildPolicyArn(acc, "tah-connector-hapi-bookings")
	if arn != "arn:aws:iam::1234567890:policy/tah-connector-hapi-bookings" {
		t.Errorf("[AcountFromRoleArn] error creating policy %v", arn)
	}
	cfg := Init()
	roleName := "unit-test-role-3"
	roleArn, policyArn, err := cfg.CreateRoleWithActionResource(roleName, "events.amazonaws.com", "events:InvokeApiDestination", "arn:aws:events:eu-central-1:1234567890:api-destination/importBookings/*")
	if err != nil {
		t.Errorf("[TestIam] error %v", err)
	}
	arn = cfg.PolicyExists(AccountFromRoleArn(roleArn), roleName)
	if arn != policyArn {
		t.Errorf("[TestIam] invalid policy arn %v  returned by PolicyExists function. shoould be %v", arn, policyArn)
	}
	err = cfg.DetachRolePolicy(roleName, policyArn)
	if err != nil {
		t.Errorf("[TestIam] error detaching policy from role %v", err)
	}
	err = cfg.DeletePolicy(policyArn)
	if err != nil {
		t.Errorf("[TestIam] error deleting policy %v", err)
	}
	err = cfg.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestIam] error deleting role %v", err)
	}
	err = cfg.DeleteRoleIfExists("does-not-exist")
	if err != nil {
		t.Errorf("[TestIam] error deleting role %v", err)
	}
	roleName = "unit-test-role-4"
	_, _, err = cfg.CreateRoleWithActionResource(roleName, "events.amazonaws.com", "events:InvokeApiDestination", "arn:aws:events:eu-central-1:1234567890:api-destination/importBookings/*")
	if err != nil {
		t.Errorf("[TestIam] error %v", err)
	}
	err = cfg.DeleteRole(roleName)
	if err != nil {
		t.Errorf("[TestIam] error deleting role whith attached policies policies %v", err)
	}

}
