// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"encoding/json"
	"errors"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/kinesis"
	changeEvent "tah/upt/source/ucp-common/src/model/change-event"
	"time"
)

/*
Create Event Example
{
	"SchemaVersion": 0,
	"EventId": "f2a97405-4ed6-4c09-90a7-b37eb9966d22",
	"EventTimestamp": "2023-04-28 14:42:18.761 +0000 UTC",
	"EventType": "CREATED",
	"DomainName": "example_domain",
	"ObjectTypeName": "_profile",
	"Object": {
		"Address":{},
		"Attributes":{
			"company":"CompanyName",
			"email_business":"email@example.com",
			"honorific":"Ms",
			"job_title":"Architect",
			"payment_type":"voucher",
			"phone_business":"2459719737",
			"profile_id":"4443810144"
		},
		"BillingAddress":{},
		"BirthDate":"2022-03-15",
		"FirstName":"Giovanni",
		"Gender":"other",
		"GenderString":"other",
		"LastName":"Walsh",
		"MailingAddress":{
			"Address1":"86331 Manors port",
			"Address2":"more content line 2",
			"Address3":"more content line 3",
			"Address4":"more content line 4",
			"City":"St. Petersburg",
			"Country":"IQ",
			"PostalCode":"78847"
			},
		"MiddleName":"Marietta",
		"ProfileId":"ae5e36d15d9641e8978ecce97c542cca"
	},
	"IsMessageRealTime": true
}

Update Event Example
{
	"SchemaVersion": 0,
	"EventId": "f2a97405-4ed6-4c09-90a7-b37eb9966d22",
	"EventTimestamp": "2023-04-28 14:42:18.761 +0000 UTC",
	"EventType": "UPDATED",
	"DomainName": "ucp-test-sprint-9-4",
	"ObjectTypeName": "air_booking",
	"AssociatedProfileId":"6831e372b0d149bf9a0e2f6772c9c0a1",
	"ProfileObjectUniqueKey":"w7rb3HISZsELqNtJuEC/oh7flFAT2YJclNS1joYceJY=",
	"Object": {
		"address_billing_city":"",
		"address_billing_country":"",
		"address_billing_line1":"",
		"address_billing_line2":"",
		"address_billing_line3":"",
		"address_billing_line4":"",
		"address_billing_postal_code":"",
		"address_billing_state_province":"",
		"address_business_city":"",
		"address_business_country":"",
		"address_business_line1":"",
		"address_business_line2":"",
		"address_business_line3":"",
		"address_business_line4":"",
		"address_business_postal_code":"",
		"address_business_state_province":"",
		"address_city":"Henderson",
		"address_country":"CX",
		"address_line1":"65845 Ridge borough",
		"address_line2":"more content line 2",
		"address_line3":"more content line 3",
		"address_line4":"more content line 4",
		"address_mailing_city":"",
		"address_mailing_country":"",
		"address_mailing_line1":"",
		"address_mailing_line2":"",
		"address_mailing_line3":"",
		"address_mailing_line4":"",
		"address_mailing_postal_code":"",
		"address_mailing_state_province":"",
		"address_postal_code":"32941",
		"address_state_province":"",
		"address_type":"",
		"arrival_date":"2021-11-03",
		"arrival_time":"16:21",
		"booking_id":"YQ3P9B",
		"cc_cvv":"",
		"cc_exp":"",
		"cc_name":"",
		"cc_token":"",
		"cc_type":"",
		"channel":"mobile",
		"company":"Health Grades",
		"date_of_birth":"1919-06-01",
		"departure_date":"2021-11-03",
		"departure_time":"12:12",
		"email":"",
		"email_business":"email@example.com",
		"email_type":"business", "first_name":"Lacy",
		"flight_number":"UA3833",
		"from":"SFO",
		"gds_id":"",
		"gender":"male",
		"honorific":"Ms",
		"job_title":"Analyst",
		"language_code":"",
		"language_name":"",
		"last_name":"Keller",
		"last_updated":"2021-12-31T025607.912662687Z",
		"last_updated_by":"Martina Upton",
		"middle_name":"Earnest",
		"model_version":"1.0",
		"nationality_code":"",
		"nationality_name":"",
		"object_type":"air_booking",
		"payment_type":"bank_account",
		"phone":"",
		"phone_business":"351.804.6565",
		"phone_home":"",
		"phone_mobile":"",
		"phone_type":"business",
		"price":"0.0",
		"pronoun":"he",
		"pss_id":"",
		"segment_id":"YQ3P9B-SFO-LAS",
		"status":"canceled",
		"to":"LAS",
		"traveller_id":"2661022692"
	},
	"IsMessageRealTime": true
}

Merge Event Example
{
	"SchemaVersion": 0,
	"EventId": "b88d5339-69a3-4ef0-ac0e-208be1823a55",
	"EventTimestamp": "2023-08-24T02:43:20.000Z",
	"EventType": "MERGED",
	"DomainName": "IrvinDomain",
	"ObjectTypeName": "_profile",
	"Object": {
	  "FirstName": "FirstName",
	  "LastName": "LastName",
	  "AdditionalInformation": "Main",
	  "ProfileId": "111",
	  "Attributes": {
		"profile_id": "tid_111"
	}
	},
	"MergeId": "045d69ae-42c2-11ee-be56-0242ac120002",
	"DuplicateProfiles": [
	  {
		"FirstName": "FirstName",
		"LastName": "LastName",
		"AdditionalInformation": "Duplicate1",
		"ProfileId": "222",
		"Attributes": {
			"profile_id": "tid_222"
		}
	  },
	  {
		"FirstName": "FirstName",
		"LastName": "LastName",
		"AdditionalInformation": "Duplicate2",
		"ProfileId": "333",
		"Attributes": {
			"profile_id": "tid_333"
		}
	  }
	],
	"IsMessageRealTime": true
}

Delete Event Example
n/a: there is no existing example since Customer Profiles does not send delete events.
*/

// We send ProfileChangeEvent records for every profile and object level change.
// For example, when we call PutProfileObject, we may be creating or updating an existing profile and object.
// Two events will be emitted. Ex: _profile and air_booking
// When we merge or delete a profile, we are updating the profile.
// One event will be emitted. Ex: _profile
func SendChangeEvents(kinesisCfg kinesis.IConfig, params ...ChangeEventParams) error {
	var profileEvents []changeEvent.ProfileChangeRecordWithProfile
	var objectEvents []changeEvent.ProfileChangeRecordWithProfileObject
	for _, param := range params {
		profile := trimProfile(param)
		profileChangeEvent := changeEvent.ProfileChangeRecordWithProfile{
			SchemaVersion:       SchemaVersion,
			EventId:             param.EventID,
			EventTimestamp:      param.Time.UTC().Format(time.RFC3339),
			EventType:           param.EventType,
			DomainName:          param.DomainName,
			IsMessageRealTime:   param.IsRealTime,
			Object:              profile,
			ObjectTypeName:      PROFILE_OBJECT_TYPE_NAME,
			AssociatedProfileId: profile.ProfileId,
		}
		if param.EventType == EventTypeMerged || param.EventType == EventTypeUnmerged {
			profileChangeEvent.MergeId = param.MergeID
			profileChangeEvent.DuplicateProfiles = param.MergedProfiles
		}
		profileEvents = append(profileEvents, profileChangeEvent)

		if param.ObjectTypeName != "" && len(profile.ProfileObjects) == 1 {
			objectChangeEvent := changeEvent.ProfileChangeRecordWithProfileObject{
				SchemaVersion:       SchemaVersion,
				EventId:             param.EventID,
				EventTimestamp:      param.Time.UTC().Format(time.RFC3339),
				EventType:           param.EventType,
				DomainName:          param.DomainName,
				ObjectTypeName:      param.ObjectTypeName,
				Object:              profile.ProfileObjects[0].AttributesInterface,
				IsMessageRealTime:   param.IsRealTime,
				AssociatedProfileId: profile.ProfileId,
			}
			objectEvents = append(objectEvents, objectChangeEvent)
		}
	}

	var recs []kinesis.Record
	for _, event := range profileEvents {
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		recs = append(recs, kinesis.Record{
			Pk:   event.Object.ProfileId,
			Data: string(data),
		})
	}
	for _, event := range objectEvents {
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		recs = append(recs, kinesis.Record{
			Pk:   event.Object["profile_id"].(string),
			Data: string(data),
		})
	}

	err, errs := kinesisCfg.PutRecords(recs)
	if err != nil || len(errs) != 0 {
		var allErrors error
		allErrors = errors.Join(allErrors, err)
		for _, e := range errs {
			allErrors = errors.Join(allErrors, errors.New(e.ErrorMessage))
		}
		return allErrors
	}
	return nil
}

func kinesisRecToChangeEvent(rec kinesis.Record) (profileEvent changeEvent.ProfileChangeRecordWithProfile, objectEvent changeEvent.ProfileChangeRecordWithProfileObject, err error) {
	var event map[string]interface{}
	err = json.Unmarshal([]byte(rec.Data), &event)
	if err != nil {
		return profileEvent, objectEvent, err
	}
	if event["ObjectTypeName"] != nil {
		objectType := event["ObjectTypeName"]
		if objectType == PROFILE_OBJECT_TYPE_NAME {
			var newEvent changeEvent.ProfileChangeRecordWithProfile
			err = json.Unmarshal([]byte(rec.Data), &newEvent)
			if err != nil {
				return profileEvent, objectEvent, err
			}
			profileEvent = newEvent
		} else if objectType != "" {
			var newEvent changeEvent.ProfileChangeRecordWithProfileObject
			err = json.Unmarshal([]byte(rec.Data), &newEvent)
			if err != nil {
				return profileEvent, objectEvent, err
			}
			objectEvent = newEvent
		}
	}
	return profileEvent, objectEvent, err
}

// Profile may be sent with full list of profile objects, but we are only interested in the
// object being created/updated. We need to remove all other objects so the one being changed
// is clear, and to reduce event size before sending to Kinesis (large profile could exceed 1MB limit).
func trimProfile(params ChangeEventParams) profilemodel.Profile {
	profile := params.Profile
	if params.EventType == EventTypeCreated || params.EventType == EventTypeUpdated {
		// Send event with ONLY the object that is being created/updated
		obj := findObject(profile.ProfileObjects, params.ObjectID)
		profile.ProfileObjects = []profilemodel.ProfileObject{obj}
	} else {
		// Remove any profile objects, not relevant for merge/delete events
		profile.ProfileObjects = nil                            // serializes to null
		profile.ProfileObjects = []profilemodel.ProfileObject{} // serializes to empty array
	}

	return profile
}

func findObject(objects []profilemodel.ProfileObject, objectId string) profilemodel.ProfileObject {
	for _, obj := range objects {
		if obj.ID == objectId {
			return obj
		}
	}

	return profilemodel.ProfileObject{}
}
