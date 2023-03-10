package pinpoint

import "testing"

/**********
*PINPOINT
*************/
//TODO: dynamicallu create and delete pinpoint app
func TestPinpoint(t *testing.T) {
	/*
		hotelLocation, _ := time.LoadLocation("Europe/Paris")
		utc, _ := time.LoadLocation("UTC")
		checkIn, _ := time.ParseInLocation("20060102T15:04", "20200530T15:00", hotelLocation)
		checkIn = checkIn.In(utc)
		if checkIn.Format("2006-01-02T15:04:05Z") != "2020-05-30T13:00:00Z" {
			t.Errorf("[PINPOINT] Time formatting error for pinpoint journey activity raw date: %+v, formatted: %s", checkIn, checkIn.Format("2006-01-02T15:04:05Z"))
		}

		params := map[string]map[string]string{}
		params["dev"] = map[string]string{}
		params["dev"]["appId"] = "196a3a4319d54a9db70eaeecaa566bfc"
		params["dev"]["region"] = "eu-central-1"

		if params[env]["appId"] != "" {
			pinpointCfg := pinpointSvc.InitWithRegion(params[env]["appId"], params["dev"]["region"])
			tx := core.NewTransaction("TEST_AURORA", "")
			pinpointCfg.SetTx(tx)
			email := "test@gmail.com"
			loc := pinpointSvc.PinpointLocation{
				Country: "US",
			}
			attrs := map[string][]string{
				"hotel":    []string{"hotel-0"},
				"check-in": []string{"2020-05-30"},
			}
			userAttrs := map[string][]string{
				"firstName": []string{"joe"},
				"lastName":  []string{"biden"},
			}
			metrics := map[string]float64{}
			segName := "test-segment-name"
			subSegName := "test-segment-name"
			tags := map[string]string{
				"test-tag-key": "test-tag-value",
			}
			segAttributeValues := map[string]string{
				"hotel": "hotel-0",
			}
			subSegAttributeValues := map[string]string{
				"check-in": "2020-05-30",
			}
			edp, err := pinpointCfg.CreateUpdateEndpoint(email, "", loc, attrs, metrics, userAttrs)
			if err != nil {
				t.Errorf("[PINPOINT] CreateUpdateEndpoint error %v", err)
			}
			edp, err = pinpointCfg.GetEndpoint(email)
			if err != nil {
				t.Errorf("[PINPOINT] GetEndpoint error %v", err)
			}
			seg, err := pinpointCfg.CreateSegment(segName, tags, "", segAttributeValues)
			if err != nil {
				t.Errorf("[PINPOINT] CreateSegment error %v", err)
			}
			subSeg, err := pinpointCfg.CreateSegment(subSegName, tags, seg.Id, subSegAttributeValues)
			if err != nil {
				t.Errorf("[PINPOINT] CreateSegment error %v", err)
			}
			seg, err = pinpointCfg.GetSegment(seg.Id)
			if err != nil {
				t.Errorf("[PINPOINT] GetSegment error %v", err)
			}
			err = pinpointCfg.DeleteSegment(subSeg.Id)
			if err != nil {
				t.Errorf("[PINPOINT] DeleteSegment error %v", err)
			}
			err = pinpointCfg.DeleteSegment(seg.Id)
			if err != nil {
				t.Errorf("[PINPOINT] DeleteSegment error %v", err)
			}
			seg, err = pinpointCfg.GetSegment(seg.Id)
			if err == nil {
				t.Errorf("[PINPOINT] GetSegmentAfterDelete endpoint has not been deleted")
			}
			err = pinpointCfg.DeleteEndpoint(edp.Email)
			if err != nil {
				t.Errorf("[PINPOINT] DeleteEndpoint error %v", err)
			}
			edp, err = pinpointCfg.GetEndpoint(email)
			if err == nil {
				t.Errorf("[PINPOINT] GetEndpointAfterDelete endpoint has not been deleted")
			}
			//Journey
			journeyAttributeValues := map[string]string{
				"hotel": "hotel-0",
			}
			tplName := "testTemplateName"
			err = pinpointCfg.CreateTemplate(tplName, "testSeubject", "<h1>testHTML</h1>", "testText")
			if err != nil {
				t.Errorf("[PINPOINT] CreateTemplate create template fails")
			}
			_, err = pinpointCfg.GetTemplate(tplName)
			if err != nil {
				t.Errorf("[PINPOINT] GetTemplate create template fails")
			}
			_, err = pinpointCfg.GetTemplate(tplName + "notExist")
			if err == nil {
				t.Errorf("[PINPOINT] GetTemplate shoudl return error is template does not exist")
			}
			journeySeg, err := pinpointCfg.CreateSegment("journeySegment", tags, "", journeyAttributeValues)
			oneMin, _ := time.ParseDuration("1m")
			startDate := time.Now().Add(oneMin)
			endDate := startDate.AddDate(0, 0, 2)
			waitUntil := startDate.AddDate(0, 0, 1).Add(oneMin)
			//Removed Journey as test are failing until ilimit increase
				pinpointSvc.PinpointJourneyActivity{
					Id:          "a1",
					Description: "step1",
					WaitUntil:   &waitUntil,
				},
				pinpointSvc.PinpointJourneyActivity{
					Id:           "a2",
					Description:  "step1",
					TemplateName: tplName,
				},
			})
			if err != nil {
				t.Errorf("[PINPOINT] CreateJourney failed: %v", err)
			}
			err = pinpointCfg.DeleteJourney(journey.Id)
			if err != nil {
				t.Errorf("[PINPOINT] DeleteJourney journey has not been deleted")
			}
			pinpointCfg.DeleteSegment(journeySeg.Id)
			//Test activity validation
			_, valErr := pinpointCfg.CreateJourney(startDate, endDate, "testJourney", journeySeg.Id, []pinpointSvc.PinpointJourneyActivity{
				pinpointSvc.PinpointJourneyActivity{
					Id:          "a1-1",
					Description: "step1",
					WaitUntil:   &waitUntil,
				},
			})
			if valErr == nil {
				t.Errorf("[PINPOINT] CreateJourney should not be able to create journey with activities that have non-alphanumeric ID")
			}
			err = pinpointCfg.DeleteTemplate(tplName)
			if err != nil {
				t.Errorf("[PINPOINT] DeleteTemplate failed: %v", err)
			}
		}*/

}

//TODO: Dynaically Create ad delete pinpoint app
func TestPinpointSendEmail(t *testing.T) {
	/*params := map[string]map[string]string{}
	params["dev"] = map[string]string{}
	params["dev"]["appId"] = "196a3a4319d54a9db70eaeecaa566bfc"
	params["dev"]["region"] = "eu-central-1"
	if params[env]["appId"] != "" {
		pinpointCfg := pinpointSvc.InitWithRegion(params[env]["appId"], params["dev"]["region"])
		err := pinpointCfg.SendEmail("guest.cloudrack@gmail.com", "Cloudrack Lambda Core Tests", "<h1>test Passed!</h1>", "test Passed!")
		if err != nil {
			t.Errorf("[PINPOINT] SendEmail failed: %v", err)
		}
		err2 := pinpointCfg.SendEmailFromTemplate("guest.cloudrack@gmail.com", "cloudrack_dining_confirmation", map[string]string{"Attributes.time": "9pm"})
		if err2 != nil {
			t.Errorf("[PINPOINT] SendEmailFromTemplate failed: %v", err2)
		}
	}*/

}
