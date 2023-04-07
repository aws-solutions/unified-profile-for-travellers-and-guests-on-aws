package traveller

import (
	"log"
	"tah/core/core"
	"tah/core/customerprofiles"
	"tah/core/kms"
	"tah/core/s3"
	model "tah/ucp/src/business-logic/model/common"
	testutils "tah/ucp/src/business-logic/testutils"
	admin "tah/ucp/src/business-logic/usecase/admin"
	"tah/ucp/src/business-logic/usecase/registry"
	"testing"
	"time"
)

var UCP_REGION = testutils.GetTestRegion()

var guestProfileId = "123456"
var guestEmail = "howelltoy@leannon.org"
var guest = `{"traveller_id":"` + guestProfileId + `","first_name":"Thomas","last_name":"Johnson","email":"` + guestEmail + `","booking_id":"BOOKING-123","hotel_code":"ABC","n_nights":"3"}`

func TestTraveller(t *testing.T) {
	// Set up resources
	s3Client := s3.InitRegion(UCP_REGION)
	bucketName, err := s3Client.CreateRandomBucket("travellertest")
	if err != nil {
		t.Errorf("[TestTraveller] Error creating test bucket: %v", err)
	}
	s3Client.Bucket = bucketName
	resource := []string{"arn:aws:s3:::" + bucketName, "arn:aws:s3:::" + bucketName + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": {"appflow.amazonaws.com"}}
	err = s3Client.AddPolicy(bucketName, resource, actions, principal)
	if err != nil {
		t.Errorf("[TestTraveller] Error adding policy to test bucket: %v", err)
	}
	kmsClient := kms.Init(UCP_REGION)
	keyArn, err := kmsClient.CreateKey("ucp-traveller360-test-key")
	if err != nil {
		t.Errorf("[TestTraveller] Error creating test kms key: %v", err)
	}
	profileClient := customerprofiles.Init(UCP_REGION)
	domainName := "traveller360-test-domain" + time.Now().Format("2006-01-02-15-04-05")

	// Set up Customer Profile domain
	tx := core.NewTransaction("ucp_test", "")
	reg := registry.NewRegistry(UCP_REGION, nil, nil, nil, &profileClient, nil)
	reg.AddEnv("KMS_KEY_PROFILE_DOMAIN", keyArn)
	reg.AddEnv("LAMBDA_ENV", "dev_test")
	reg.AddEnv("CONNECT_PROFILE_SOURCE_BUCKET", s3Client.Bucket)
	createUc := admin.NewCreateDomain()
	createUc.SetRegistry(&reg)
	createUc.SetTx(&tx)
	req := model.RequestWrapper{
		Domain: model.Domain{Name: domainName},
	}
	_, err = createUc.Run(req)
	if err != nil {
		t.Errorf("[TestTraveller] Error creating Customer Profile domain: %v", err)
	}
	// Wait for creation
	i, max, duration := 0, 12, 5*time.Second
	for ; i < max; i++ {
		integrations, err := profileClient.GetIntegrations()
		if len(integrations) > 0 && err == nil {
			break
		}
		log.Printf("[TestTraveller] Domain not ready, waiting %v seconds", duration.Seconds())
		time.Sleep(duration)
	}
	if i == max {
		t.Errorf("[TestTraveller] Error: unable to retrieve domain integration.")
	}

	// Populate domain with test data
	err = profileClient.PutProfileObject(guest, OBJECT_TYPE_HOTEL_BOOKING)
	if err != nil {
		t.Errorf("[TestTraveller] Error adding test profile: %v", err)
	}
	i, max, duration = 0, 12, 5*time.Second
	if err != nil {
		i = 12 // if there was an error creating profile, skip searching
	}
	for ; i < max; i++ {
		integrations, err := profileClient.GetProfileId(guestProfileId)
		if len(integrations) > 0 && err == nil {
			break
		}
		log.Printf("[TestTraveller] Profile is not available yet, waiting %v seconds", duration.Seconds())
		time.Sleep(duration)
	}
	if i == max {
		t.Errorf("[TestTraveller] Error: unable to retrieve customer profile id")
	}

	// Test Traveller functions
	search := NewSearchProfile()
	search.SetRegistry(&reg)
	req.SearchRq.Email = guestEmail
	log.Printf("[TestTraveller] Searching for user by email: %v", req.SearchRq.Email)
	res, err := search.Run(req)
	if err != nil {
		t.Errorf("[TestTraveller] Error searching for user: %v", err)
	}
	// Get profile if one was returned, otherwise skip and show error. Prevents out of range panic on Profiles[0].
	if len(res.Profiles) > 0 {
		profile, err := profileClient.GetProfile(res.Profiles[0].ConnectID, COMBINED_PROFILE_OBJECT_TYPES)
		if err != nil {
			t.Errorf("[TestTraveller] Error getting orders for user: %v", err)
		}
		traveller360 := profileToTraveller(profile)
		log.Printf("[TestTraveller] Traveller 360: %v", traveller360)
	} else {
		t.Errorf("[TestTraveller] Error: unable to retrieve customer profile")
	}

	// Clean up resources
	s3Client.EmptyAndDelete()
	kmsClient.DeleteKey(keyArn)
	profileClient.DeleteDomain()
}
