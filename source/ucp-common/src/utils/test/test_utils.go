// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/aurora"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	"tah/upt/source/tah-core/db"
	"tah/upt/source/tah-core/kinesis"
	secrets "tah/upt/source/tah-core/secret"
	"tah/upt/source/tah-core/sqs"
	constants "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/config"
	"testing"
	"time"
)

func GetTestRegion() string {
	//getting region for local testing
	region := os.Getenv("UCP_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}

func GetTestName(i any) string {
	return strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")[1]
}

func FormatErrorMessage(msg string, expected interface{}, actual interface{}) string {
	return fmt.Sprintf("%s\nExpected: %v\nActual: %v", msg, expected, actual)
}

type ProfileStorageParams struct {
	LcsConfigTable      *db.DBConfig
	LcsKinesisStream    kinesis.IConfig
	MergeQueueClient    sqs.IConfig
	CPWriterQueueClient sqs.IConfig
	MetricNamespace     string
	DomainName          *string
	Tx                  *core.Transaction
}

type ExpectedField struct {
	ObjectType string
	ObjectID   string
	FieldName  string
	FieldValue string
}

type ProfileObjectRef struct {
	ObjectType string
	ObjectID   string
}

// Initialize Profile Storage in either ACCP or LCS mode. If provided, it will also set domain and tx.
//
// If running in LCS, we will look at local 'source/env.json' file to load config for the Aurora cluster to run on.
func InitProfileStorage(params ProfileStorageParams) (customerprofiles.ICustomerProfileLowCostConfig, error) {
	// Validate required LCS resources

	if params.LcsConfigTable == nil || params.LcsKinesisStream == nil {
		return nil, fmt.Errorf("missing required resources to initialize in LCS mode")
	}

	// Load config
	ec, err := config.LoadEnvConfig()
	if err != nil {
		return nil, err
	}
	if ec.ArtifactBucket == "" {
		return nil, fmt.Errorf("artifact bucket name not provided by local config")
	}

	var profileStorage customerprofiles.ICustomerProfileLowCostConfig
	smClient := secrets.InitWithRegion(ec.AuroraSecretName, ec.Region, "", "")
	user := smClient.Get("username")
	pwd := smClient.Get("password")
	auroraClient := aurora.InitPostgresql(ec.AuroraEndpoint, ec.AuroraDbName, user, pwd, 50, core.LogLevelDebug)
	if err := auroraClient.Ping(); err != nil {
		log.Printf("Error connecting to Aurora.")
		log.Printf("Local development: validate you are connected to Amazon CORP VPN.")
		log.Printf("Pipeline: validate the pipeline is located in the same VPC as the RDS cluster.")
		return nil, fmt.Errorf("cannot ping Aurora: %v", err)
	}
	options := customerprofiles.CustomerProfileInitOptions{
		MergeQueueClient: params.MergeQueueClient,
	}
	profileStorage = customerprofiles.InitLowCost(
		ec.Region,
		auroraClient,
		params.LcsConfigTable,
		params.LcsKinesisStream,
		params.CPWriterQueueClient,
		params.MetricNamespace,
		"",
		"",
		core.LogLevelDebug,
		&options,
		customerprofiles.InitLcsCache(),
	)

	if params.DomainName != nil {
		profileStorage.SetDomain(*params.DomainName)
	}
	if params.Tx != nil {
		profileStorage.SetTx(*params.Tx)
	}

	return profileStorage, nil
}

func WaitForProfile(reqHelper EndToEndRequest, profileID string, timeout int) (model.Traveller, error) {
	log.Printf("[WaitForProfile] Waiting %v seconds for profile ID %v creation", timeout, profileID)
	it := 0
	for it*5 < timeout {
		profiles, err := SearchProfiles(reqHelper, "travellerId", []string{profileID})
		if len(profiles) > 0 {
			log.Printf("[WaitForProfile] Found profile ID: %v. (ConnectID: %v)", profiles[0].TravellerID, profiles[0].ConnectID)
			fullProfile, err := GetProfile(
				reqHelper,
				profiles[0].ConnectID,
				constants.AccpRecordsNames(),
				[]customerprofiles.PaginationOptions{},
			)
			if err != nil {
				log.Printf("[WaitForProfile] SearchProfiles error: %v", err)
			}
			return fullProfile, nil
		} else if err != nil {
			log.Printf("[WaitForProfile] SearchProfiles error: %v", err)
		}
		log.Printf("[WaitForProfile] Not found. Waiting 5 seconds")
		time.Sleep(5 * time.Second)
		it += 1
	}
	return model.Traveller{}, fmt.Errorf("could not find profile ID %s: timout expired", profileID)
}

func WaitForProfileObject(
	reqHelper EndToEndRequest,
	objectId string,
	profileId string,
	objectType string,
	timeout int,
) (profilemodel.ProfileObject, error) {
	log.Printf("[WaitForProfileObject] Waiting %v seconds for profile object ID %v creation", timeout, objectId)
	it := 0
	var err error
	for it*5 < timeout {
		profObject, err := GetSpecificProfileObject(reqHelper, objectId, profileId, objectType)
		if err != nil {
			log.Printf("[WaitForProfileObject] GetSpecificProfileObject error: %v", err)
			log.Printf("[WaitForProfileObject] Not found. Waiting 5 seconds")
			_, _ = WaitForProfile(reqHelper, profileId, 5)
			time.Sleep(5 * time.Second)
			it += 1
		} else {
			log.Printf("[WaitForProfileObject] Profile object: %+v", profObject.Attributes)
			return profObject, nil
		}
	}
	return profilemodel.ProfileObject{}, fmt.Errorf("could not find profile object ID: timeout expired last error: %v", err)
}

func GetFieldValue(traveler model.Traveller, objectType string, fieldName string) string {
	if objectType == "_profile" {
		if fieldName == "ConnectId" {
			return traveler.ConnectID
		}
		if fieldName == "travellerId" {
			return traveler.TravellerID
		}
		if fieldName == "firstName" {
			return traveler.FirstName
		}
		if fieldName == "lastName" {
			return traveler.LastName
		}
		if fieldName == "middleName" {
			return traveler.MiddleName
		}
		if fieldName == "birthDate" {
			return traveler.BirthDate.Format("2006-01-02")
		}
		if fieldName == "gender" {
			return traveler.Gender
		}
		if fieldName == "jobTitle" {
			return traveler.JobTitle
		}
		if fieldName == "companyName" {
			return traveler.CompanyName
		}
		if fieldName == "honorific" {
			return traveler.Honorific
		}
		if fieldName == "phoneNumber" {
			return traveler.PhoneNumber
		}
		if fieldName == "homePhoneNumber" {
			return traveler.HomePhoneNumber
		}
		if fieldName == "mobilePhoneNumber" {
			return traveler.MobilePhoneNumber
		}
		if fieldName == "businessPhoneNumber" {
			return traveler.BusinessPhoneNumber
		}
		if fieldName == "personalEmailAddress" {
			return traveler.PersonalEmailAddress
		}
		if fieldName == "businessEmailAddress" {
			return traveler.BusinessEmailAddress
		}
		if fieldName == "addressLine1" {
			return traveler.HomeAddress.Address1
		}
		if fieldName == "addressLine2" {
			return traveler.HomeAddress.Address2
		}
		if fieldName == "city" {
			return traveler.HomeAddress.City
		}
		if fieldName == "postalCode" {
			return traveler.HomeAddress.PostalCode
		}
		if fieldName == "stateCode" {
			return traveler.HomeAddress.State
		}
		if fieldName == "countryCode" {
			return traveler.HomeAddress.Country
		}
		if fieldName == "countryName" {
			return traveler.HomeAddress.Country
		}
		if fieldName == "bizAddressLine1" {
			return traveler.BusinessAddress.Address1
		}
		if fieldName == "bizAddressLine2" {
			return traveler.BusinessAddress.Address2
		}
		if fieldName == "bizCity" {
			return traveler.BusinessAddress.City
		}
		if fieldName == "bizPostalCode" {
			return traveler.BusinessAddress.PostalCode
		}
		if fieldName == "bizStateCode" {
			return traveler.BusinessAddress.State
		}
		if fieldName == "bizProvinceCode" {
			return traveler.BusinessAddress.Province
		}
		if fieldName == "bizCountryCode" {
			return traveler.BusinessAddress.Country
		}
		if fieldName == "bizCountryName" {
			return traveler.BusinessAddress.Country
		}
		if fieldName == "languageCode" {
			return traveler.LanguageCode
		}

	}
	return "unknown field " + fieldName
}

func assert(t *testing.T, profileId, fieldName, fieldvalue, expected string) {
	if fieldvalue != expected {
		t.Errorf("Profile %v should have %s=%s and not %s", profileId, fieldName, expected, fieldvalue)
	}
}

func assertObject(t *testing.T, profileId, objectType, objectId, fieldName, fieldvalue, expected string) {
	if fieldvalue != expected {
		t.Errorf(
			"[Traveller: %s] Profile Object %v with id=%v should have %s=%s and not %s",
			profileId,
			objectType,
			objectId,
			fieldName,
			expected,
			fieldvalue,
		)
	}
}

func GetProfileObject(
	t *testing.T,
	traveler model.Traveller,
	objectType string,
	objectId string,
	expected string,
) (model.AccpObject, error) {
	objs := []model.AccpObject{}
	if objectType == "hotel_loyalty" {
		for _, obj := range traveler.HotelLoyaltyRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "hotel_booking" {
		for _, obj := range traveler.HotelBookingRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "clickstream" {
		for _, obj := range traveler.ClickstreamRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "air_booking" {
		for _, obj := range traveler.AirBookingRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "air_loyalty" {
		for _, obj := range traveler.AirLoyaltyRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "ancillary_service" {
		for _, obj := range traveler.AncillaryServiceRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "loyalty_transaction" {
		for _, obj := range traveler.LoyaltyTxRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "customer_service_interaction" {
		for _, obj := range traveler.CustomerServiceInteractionRecords {
			objs = append(objs, obj)
		}
	} else if objectType == "alternate_profile_id" {
		for _, obj := range traveler.AlternateProfileIDs {
			objs = append(objs, obj)
		}
	}
	for _, obj := range objs {
		if obj.ID() == objectId {
			return obj, nil
		}
	}
	log.Printf("unknown object type: %s", objectType)
	return model.HotelLoyalty{}, errors.New("unknown object type: " + objectType)
}

func GetFieldValueFromObject(obj model.AccpObject, objectType string, fieldName string) string {
	if objectType == "hotel_loyalty" {
		loy := obj.(model.HotelLoyalty)
		if fieldName == "travellerId" {
			return loy.TravellerID
		}
		if fieldName == "loyaltyId" {
			return loy.LoyaltyID
		}
		if fieldName == "programName" {
			return loy.ProgramName
		}
		if fieldName == "points" {
			return loy.Points
		}
		if fieldName == "units" {
			return loy.Units
		}
		if fieldName == "pointsToNextLevel" {
			return loy.PointsToNextLevel
		}
		if fieldName == "level" {
			return loy.Level
		}
		if fieldName == "joined" {
			return loy.Joined.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "lastUpdated" {
			return loy.LastUpdated.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "lastUpdatedBy" {
			return loy.LastUpdatedBy
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "air_loyalty" {
		loy := obj.(model.AirLoyalty)
		if fieldName == "travellerId" {
			return loy.TravellerID
		}
		if fieldName == "loyaltyId" {
			return loy.LoyaltyID
		}
		if fieldName == "programName" {
			return loy.ProgramName
		}
		if fieldName == "miles" {
			return loy.Miles
		}
		if fieldName == "pointsToNextLevel" {
			return loy.MilesToNextLevel
		}
		if fieldName == "level" {
			return loy.Level
		}
		if fieldName == "joined" {
			return loy.Joined.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "lastUpdated" {
			return loy.LastUpdated.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "lastUpdatedBy" {
			return loy.LastUpdatedBy
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "hotel_booking" {
		book := obj.(model.HotelBooking)
		if fieldName == "travellerId" {
			return book.TravellerID
		}
		if fieldName == "roomTypeCode" {
			return book.RoomTypeCode
		}
		if fieldName == "ratePlanCode" {
			return book.RatePlanCode
		}
		if fieldName == "status" {
			return book.Status
		}
		if fieldName == "totalBeforeTax" {
			return fmt.Sprintf("%v", book.TotalSegmentBeforeTax)
		}
		if fieldName == "totalAfterTax" {
			return fmt.Sprintf("%v", book.TotalSegmentAfterTax)
		}
		if fieldName == "nNights" {
			return fmt.Sprintf("%v", book.NumNights)
		}
		if fieldName == "nGuests" {
			return fmt.Sprintf("%v", book.NumGuests)
		}
		if fieldName == "checkInDate" {
			return book.CheckInDate.Format("2006-01-02")
		}
		if fieldName == "lastUpdated" {
			return book.LastUpdated.Format("2006-01-02T15:04:05.999Z")
		}
		if fieldName == "lastUpdatedBy" {
			return book.LastUpdatedBy
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "ancillary_service" {
		ans := obj.(model.AncillaryService)
		if fieldName == "travellerId" {
			return ans.TravellerID
		}
		if fieldName == "ancillaryType" {
			return ans.AncillaryType
		}
		if fieldName == "bookingId" {
			return ans.BookingID
		}
		if fieldName == "flightNumber" {
			return ans.FlightNumber
		}
		if fieldName == "baggageType" {
			return ans.BaggageType
		}
		if fieldName == "weight" {
			return fmt.Sprintf("%v", ans.Weight)
		}
		if fieldName == "dimentionsWidth" {
			return fmt.Sprintf("%v", ans.DimentionsWidth)
		}
		if fieldName == "dimentionsLength" {
			return fmt.Sprintf("%v", ans.DimentionsLength)
		}
		if fieldName == "dimentionsHeight" {
			return fmt.Sprintf("%v", ans.DimentionsHeight)
		}
		if fieldName == "priorityBagDrop" {
			return fmt.Sprintf("%v", ans.PriorityBagDrop)
		}
		if fieldName == "priorityBagReturn" {
			return fmt.Sprintf("%v", ans.PriorityBagReturn)
		}
		if fieldName == "lotBagInsurance" {
			return fmt.Sprintf("%v", ans.LotBagInsurance)
		}
		if fieldName == "valuableBaggageInsurance" {
			return fmt.Sprintf("%v", ans.ValuableBaggageInsurance)
		}
		if fieldName == "handsFreeBaggage" {
			return fmt.Sprintf("%v", ans.HandsFreeBaggage)
		}
		if fieldName == "seatNumber" {
			return ans.SeatNumber
		}
		if fieldName == "seatZone" {
			return ans.SeatZone
		}
		if fieldName == "neighborFreeSeat" {
			return fmt.Sprintf("%v", ans.NeighborFreeSeat)
		}
		if fieldName == "upgradeAuction" {
			return fmt.Sprintf("%v", ans.UpgradeAuction)
		}
		if fieldName == "changeType" {
			return ans.ChangeType
		}
		if fieldName == "otherAncilliaryType" {
			return fmt.Sprintf("%v", ans.OtherAncilliaryType)
		}
		if fieldName == "priorityServiceType" {
			return fmt.Sprintf("%v", ans.PriorityServiceType)
		}
		if fieldName == "loungeAccess" {
			return fmt.Sprintf("%v", ans.LoungeAccess)
		}
		if fieldName == "upgradeAuction" {
			return fmt.Sprintf("%v", ans.UpgradeAuction)
		}
		if fieldName == "paxIndex" {
			return fmt.Sprintf("%v", ans.PaxIndex)
		}
		if fieldName == "quantity" {
			return fmt.Sprintf("%v", ans.Quantity)
		}
		if fieldName == "total" {
			return fmt.Sprintf("%v", ans.Price)
		}
		if fieldName == "currency" {
			return ans.Currency
		}
		if fieldName == "lastUpdated" {
			return ans.LastUpdated.Format("2006-01-02T15:04:05.999Z")
		}
		if fieldName == "lastUpdatedBy" {
			return ans.LastUpdatedBy
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "clickstream" {
		cs := obj.(model.Clickstream)
		if fieldName == "travellerId" {
			return cs.TravellerID
		}
		if fieldName == "sessionId" {
			return cs.SessionID
		}
		if fieldName == "eventTimestamp" {
			return cs.EventTimestamp.Format("2006-01-02T15:04:05.999Z")
		}
		if fieldName == "eventType" {
			return cs.EventType
		}
		if fieldName == "eventVersion" {
			return cs.EventVersion
		}
		if fieldName == "arrivalTimestamp" {
			return cs.ArrivalTimestamp.Format("2006-01-02T15:04:05.999Z")
		}
		if fieldName == "userAgent" {
			return cs.UserAgent
		}
		if fieldName == "error" {
			return cs.Error
		}
		if fieldName == "customEventName" {
			return cs.CustomEventName
		}
		if fieldName == "customerBirthdate" {
			return cs.CustomerBirthdate.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "customerCountry" {
			return cs.CustomerCountry
		}
		if fieldName == "customerEmail" {
			return cs.CustomerEmail
		}
		if fieldName == "customerFirstName" {
			return cs.CustomerFirstName
		}
		if fieldName == "customerGender" {
			return cs.CustomerGender
		}
		if fieldName == "customerId" {
			return cs.CustomerID
		}
		if fieldName == "customerLastName" {
			return cs.CustomerLastName
		}
		if fieldName == "customerNationality" {
			return cs.CustomerNationality
		}
		if fieldName == "customerPhone" {
			return cs.CustomerPhone
		}
		if fieldName == "customerType" {
			return cs.CustomerType
		}
		if fieldName == "customerLoyaltyID" {
			return cs.CustomerLoyaltyID
		}
		if fieldName == "languageCode" {
			return cs.LanguageCode
		}
		if fieldName == "currency" {
			return cs.Currency
		}
		if fieldName == "products" {
			return cs.Products
		}
		if fieldName == "quantities" {
			return cs.Quantities
		}
		if fieldName == "productsPrices" {
			return cs.ProductsPrices
		}
		if fieldName == "ecommerceAction" {
			return cs.EcommerceAction
		}
		if fieldName == "orderPaymentType" {
			return cs.OrderPaymentType
		}
		if fieldName == "orderPromoCode" {
			return cs.OrderPromoCode
		}
		if fieldName == "pageName" {
			return cs.PageName
		}
		if fieldName == "pageTypeEnvironment" {
			return cs.PageTypeEnvironment
		}
		if fieldName == "transactionId" {
			return cs.TransactionID
		}
		if fieldName == "bookingId" {
			return cs.BookingID
		}
		if fieldName == "geofenceName" {
			return cs.GeofenceName
		}
		if fieldName == "geofenceID" {
			return cs.GeofenceID
		}
		if fieldName == "geofenceLatitude" {
			return cs.GeofenceLatitude
		}
		if fieldName == "geofenceLongitude" {
			return cs.GeofenceLongitude
		}
		if fieldName == "poiId" {
			return cs.PoiID
		}
		if fieldName == "url" {
			return cs.URL
		}
		if fieldName == "custom" {
			return cs.Custom
		}
		if fieldName == "fareClass" {
			return cs.FareClass
		}
		if fieldName == "fareType" {
			return cs.FareType
		}
		if fieldName == "flightSegmentsDepartureDateTime" {
			return cs.FlightSegmentsDepartureDateTime
		}
		if fieldName == "flightSegmentsArrivalDateTime" {
			return cs.FlightSegmentsArrivalDateTime
		}
		if fieldName == "flightSegments" {
			return cs.FlightSegments
		}
		if fieldName == "flightSegmentSku" {
			return cs.FlightSegmentSku
		}
		if fieldName == "flightRoute" {
			return cs.FlightRoute
		}
		if fieldName == "flightNumbers" {
			return cs.FlightNumbers
		}
		if fieldName == "flightMarket" {
			return cs.FlightMarket
		}
		if fieldName == "flightType" {
			return cs.FlightType
		}
		if fieldName == "originDate" {
			return cs.OriginDate
		}
		if fieldName == "originDateTime" {
			return cs.OriginDateTime.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "returnDate" {
			return cs.ReturnDate
		}
		if fieldName == "returnDateTime" {
			return cs.ReturnDateTime.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "returnFlightRoute" {
			return cs.ReturnFlightRoute
		}
		if fieldName == "lengthOfStay" {
			return fmt.Sprintf("%v", cs.LengthOfStay)
		}
		if fieldName == "origin" {
			return cs.Origin
		}
		if fieldName == "selectedSeats" {
			return cs.SelectedSeats
		}
		if fieldName == "numPaxAdults" {
			return fmt.Sprintf("%v", cs.NumPaxAdults)
		}
		if fieldName == "numPaxInf" {
			return fmt.Sprintf("%v", cs.NumPaxInf)
		}
		if fieldName == "numPaxChildren" {
			return fmt.Sprintf("%v", cs.NumPaxChildren)
		}
		if fieldName == "totalPassengers" {
			return fmt.Sprintf("%v", cs.TotalPassengers)
		}
		if fieldName == "fareClass" {
			return cs.FareClass
		}
		if fieldName == "paxType" {
			return cs.PaxType
		}
		if fieldName == "checkinDate" {
			return cs.CheckinDate.Format("2006-01-02")
		}
		if fieldName == "checkoutDate" {
			return cs.CheckoutDate.Format("2006-01-02")
		}
		if fieldName == "numNights" {
			return fmt.Sprintf("%v", cs.NumNights)
		}
		if fieldName == "numGuests" {
			return fmt.Sprintf("%v", cs.NumGuests)
		}
		if fieldName == "numGuestsAdult" {
			return fmt.Sprintf("%v", cs.NumGuestsAdult)
		}
		if fieldName == "numGuestsChildren" {
			return fmt.Sprintf("%v", cs.NumGuestsChildren)
		}
		if fieldName == "hotelCode" {
			return cs.HotelCode
		}
		if fieldName == "ratePlanCode" {
			return cs.RatePlan
		}
		if fieldName == "roomTypeCode" {
			return cs.RoomType
		}
		if fieldName == "hotelName" {
			return cs.HotelName
		}
		if fieldName == "hotelCodeList" {
			return cs.HotelCodeList
		}
		if fieldName == "destination" {
			return cs.Destination
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "loyalty_transaction" {
		loy := obj.(model.LoyaltyTx)
		if fieldName == "travellerId" {
			return loy.TravellerID
		}
		if fieldName == "accpObjectID" {
			return loy.AccpObjectID
		}
		if fieldName == "overallConfidenceScore" {
			return fmt.Sprintf("%v", loy.OverallConfidenceScore)
		}
		if fieldName == "lastUpdated" {
			return loy.LastUpdated.Format("2006-01-02T15:04:05.999Z")
		}
		if fieldName == "lastUpdatedBy" {
			return loy.LastUpdatedBy
		}
		if fieldName == "pointsOffset" {
			return fmt.Sprintf("%v", loy.PointsOffset)
		}
		if fieldName == "pointUnit" {
			return loy.PointUnit
		}
		if fieldName == "originPointsOffset" {
			return fmt.Sprintf("%v", loy.OriginPointsOffset)
		}
		if fieldName == "qualifyingPointsOffset" {
			return fmt.Sprintf("%v", loy.QualifyingPointsOffset)
		}
		if fieldName == "source" {
			return loy.Source
		}
		if fieldName == "category" {
			return loy.Category
		}
		if fieldName == "bookingDate" {
			return loy.BookingDate.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "orderNumber" {
			return loy.OrderNumber
		}
		if fieldName == "productId" {
			return loy.ProductId
		}
		if fieldName == "expireInDays" {
			return fmt.Sprintf("%v", loy.ExpireInDays)
		}
		if fieldName == "amount" {
			return fmt.Sprintf("%v", loy.Amount)
		}
		if fieldName == "amountType" {
			return loy.AmountType
		}
		if fieldName == "voucherQuantity" {
			return fmt.Sprintf("%v", loy.VoucherQuantity)
		}
		if fieldName == "corporateReferenceNumber" {
			return loy.CorporateReferenceNumber
		}
		if fieldName == "promotions" {
			return loy.Promotions
		}
		if fieldName == "activityDay" {
			return loy.ActivityDay.Format("2006-01-02T15:04:05Z")
		}
		if fieldName == "location" {
			return loy.Location
		}
		if fieldName == "toLoyaltyId" {
			return loy.ToLoyaltyId
		}
		if fieldName == "fromLoyaltyId" {
			return loy.FromLoyaltyId
		}
		if fieldName == "organizationCode" {
			return loy.OrganizationCode
		}
		if fieldName == "eventName" {
			return loy.EventName
		}
		if fieldName == "documentNumber" {
			return loy.DocumentNumber
		}
		if fieldName == "corporateId" {
			return loy.CorporateId
		}
		if fieldName == "programName" {
			return loy.ProgramName
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "customer_service_interaction" {
		csi := obj.(model.CustomerServiceInteraction)
		if fieldName == "conversationSummary" {
			return csi.ConversationSummary
		}
		if fieldName == "cartID" {
			return csi.CartID
		}
		if fieldName == "channel" {
			return csi.Channel
		}
		if fieldName == "conversation" {
			return csi.Conversation
		}
		if fieldName == "interactionType" {
			return csi.InteractionType
		}
		if fieldName == "loyaltyId" {
			return csi.LoyaltyID
		}
		if fieldName == "bookingId" {
			return csi.BookingID
		}
		if fieldName == "duration" {
			return fmt.Sprintf("%v", csi.Duration)
		}
		if fieldName == "endTime" {
			return fmt.Sprintf("%v", csi.EndTime)
		}
		if fieldName == "startTime" {
			return fmt.Sprintf("%v", csi.StartTime)
		}
		if fieldName == "sentimentScore" {
			return fmt.Sprintf("%v", csi.SentimentScore)
		}
		return "unknown field '" + fieldName + "' in object " + objectType
	}
	if objectType == "alternate_profile_id" {
		altProfId := obj.(model.AlternateProfileID)
		if fieldName == "description" {
			return altProfId.Description
		}
		if fieldName == "value" {
			return altProfId.Value
		}
		if fieldName == "name" {
			return altProfId.Name
		}
	}
	return "unknown object type: " + objectType
}

func CheckExpected(t *testing.T, traveler model.Traveller, expected ExpectedField) {
	if expected.ObjectType == "_profile" {
		val := GetFieldValue(traveler, expected.ObjectType, expected.FieldName)
		if strings.HasPrefix(val, "unknown field") {
			t.Errorf(val)
		} else {
			assert(t, traveler.TravellerID, expected.FieldName, val, expected.FieldValue)
		}
	} else {
		obj, err := GetProfileObject(t, traveler, expected.ObjectType, expected.ObjectID, expected.FieldValue)
		if err != nil {
			t.Errorf("Could not find profile object of type '%v' with id '%s' on traveller '%s'", expected.ObjectType, expected.ObjectID, traveler.TravellerID)
			return
		}
		val := GetFieldValueFromObject(obj, expected.ObjectType, expected.FieldName)
		assertObject(t, traveler.TravellerID, expected.ObjectType, expected.ObjectID, expected.FieldName, val, expected.FieldValue)
	}

}
