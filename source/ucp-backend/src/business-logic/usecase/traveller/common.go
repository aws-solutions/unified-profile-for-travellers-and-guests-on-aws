package traveller

import (
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/traveller"
	"tah/ucp/src/business-logic/utils"
	"time"
)

// To be used while creating customer profile mappings for custom attributes
var ATTRIBUTE_KEY_HOTEL_BOOKING_HOTEL_CODE string = "hotelCode"
var ATTRIBUTE_KEY_HOTEL_BOOKING_PRODUCTS string = "products"
var ATTRIBUTE_KEY_HOTEL_BOOKING_N_NIGHTS string = "nNights"
var ATTRIBUTE_KEY_HOTEL_BOOKING_N_GUESTS string = "nGuests"
var ATTRIBUTE_KEY_HOTEL_BOOKING_CONFIRMATION_NUMBER string = "confirmationNumber"
var ATTRIBUTE_KEY_HOTEL_BOOKING_START_DATE string = "startDate"
var ATTRIBUTE_KEY_CLICKSTREAM_SESSION_ID string = "sessionId"
var ATTRIBUTE_KEY_CLICKSTREAM_START_DATE string = "startDate"
var ATTRIBUTE_KEY_CLICKSTREAM_ORIGIN string = "origin"
var ATTRIBUTE_KEY_CLICKSTREAM_LOCATION string = "location"
var ATTRIBUTE_KEY_CLICKSTREAM_HOTEL string = "hotelCode"
var ATTRIBUTE_KEY_CLICKSTREAM_SAILING string = "sailing"
var ATTRIBUTE_KEY_LOYALTY_PROFILE_LOYALTY_ID string = "loyaltyId"
var ATTRIBUTE_KEY_LOYALTY_PROFILE_STATUS string = "status"
var ATTRIBUTE_KEY_LOYALTY_PROFILE_POINTS string = "points"
var ATTRIBUTE_KEY_LOYALTY_PROFILE_PROGRAM string = "program"
var ATTRIBUTE_KEY_LOYALTY_PROFILE_JOINED string = "joined"

// Customer Profile object types
const (
	OBJECT_TYPE_AIR_BOOKING   string = "air_booking"
	OBJECT_TYPE_AIR_LOYALTY   string = "air_loyalty"
	OBJECT_TYPE_CLICKSTREAM   string = "clickstream_event"
	OBJECT_TYPE_EMAIL_HISTORY string = "email_history"
	OBJECT_TYPE_HOTEL_BOOKING string = "hotel_booking"
	OBJECT_TYPE_HOTEL_LOYALTY string = "hotel_loyalty"
	OBJECT_TYPE_HOTEL_STAY    string = "hotel_stay_revenue"
	OBJECT_TYPE_PHONE_HISTORY string = "phone_history"
)

func profilesToTravellers(profiles []customerprofiles.Profile) []model.Traveller {
	travellers := []model.Traveller{}
	for _, profile := range profiles {
		travellers = append(travellers, profileToTraveller(profile))
	}
	return travellers
}

func profileToTraveller(profile customerprofiles.Profile) model.Traveller {
	errors := []string{}
	traveller := model.Traveller{
		// Metadata
		ModelVersion: "1",

		// Profile IDs
		ConnectID:   profile.ProfileId,
		TravellerID: profile.Attributes["profile_id"],
		PSSID:       profile.Attributes["pss_id"],
		GDSID:       profile.Attributes["gds_id"],
		PMSID:       profile.Attributes["pms_id"],
		CRSID:       profile.Attributes["crs_id"],

		// Profile Data
		Honorific:   profile.Attributes["honorific"],
		FirstName:   profile.FirstName,
		MiddleName:  profile.MiddleName,
		LastName:    profile.LastName,
		Gender:      profile.Gender,
		Pronoun:     profile.Attributes["pronoun"],
		BirthDate:   trySetDate(profile.BirthDate, &errors),
		JobTitle:    profile.Attributes["job_title"],
		CompanyName: profile.Attributes["company"],

		// Contact Info
		PhoneNumber:          profile.PhoneNumber,
		MobilePhoneNumber:    profile.Attributes["phone_mobile"],
		HomePhoneNumber:      profile.Attributes["phone_home"],
		BusinessPhoneNumber:  profile.Attributes["phone_business"],
		PersonalEmailAddress: profile.PersonalEmailAddress,
		BusinessEmailAddress: profile.Attributes["email_business"],
		NationalityCode:      profile.Attributes["nationality_code"],
		NationalityName:      profile.Attributes["nationality_name"],
		LanguageCode:         profile.Attributes["language_code"],
		LanguageName:         profile.Attributes["langauge_name"],

		// Addresses
		HomeAddress: model.Address{
			Address1:   profile.Address.Address1,
			Address2:   profile.Address.Address2,
			Address3:   profile.Address.Address3,
			Address4:   profile.Address.Address4,
			City:       profile.Address.City,
			State:      profile.Address.State,
			Province:   profile.Address.Province,
			PostalCode: profile.Address.PostalCode,
			Country:    profile.Address.Country,
		},
		// BusinessAddress: model.Address2{
		// 	Address1:   profile.BusinessAddress.Address1,
		// 	Address2:   profile.BusinessAddress.Address2,
		// 	Address3:   profile.BusinessAddress.Address3,
		// 	Address4:   profile.BusinessAddress.Address4,
		// 	City:       profile.BusinessAddress.City,
		// 	State:      profile.BusinessAddress.State,
		// 	Province:   profile.BusinessAddress.Province,
		// 	PostalCode: profile.BusinessAddress.PostalCode,
		// 	Country:    profile.BusinessAddress.Country,
		// },
		// MailingAddress: model.Address2{
		// 	Address1:   profile.MailingAddress.Address1,
		// 	Address2:   profile.MailingAddress.Address2,
		// 	Address3:   profile.MailingAddress.Address3,
		// 	Address4:   profile.MailingAddress.Address4,
		// 	City:       profile.MailingAddress.City,
		// 	State:      profile.MailingAddress.State,
		// 	Province:   profile.MailingAddress.Province,
		// 	PostalCode: profile.MailingAddress.PostalCode,
		// 	Country:    profile.MailingAddress.Country,
		// },
		// BillingAddress: model.Address2{
		// 	Address1:   profile.BillingAddress.Address1,
		// 	Address2:   profile.BillingAddress.Address2,
		// 	Address3:   profile.BillingAddress.Address3,
		// 	Address4:   profile.BillingAddress.Address4,
		// 	City:       profile.BillingAddress.City,
		// 	State:      profile.BillingAddress.State,
		// 	Province:   profile.BillingAddress.Province,
		// 	PostalCode: profile.BillingAddress.PostalCode,
		// 	Country:    profile.BillingAddress.Country,
		// },
	}

	// TODO: handle each object in one loop through object records
	traveller.AirBookingRecords = generateAirBookingRecords(&profile)
	traveller.AirLoyaltyRecords = generateAirLoyaltyRecords(&profile, &errors)
	traveller.ClickstreamRecords = generateClickstreamRecords(&profile, &errors)
	traveller.EmailHistoryRecords = generateEmailHistoryRecords(&profile, &errors)
	traveller.HotelBookingRecords = generateHotelBookingRecords(&profile, &errors)
	traveller.HotelLoyaltyRecords = generateHotelLoyaltyRecords(&profile, &errors)
	traveller.HotelStayRecords = generateHotelStayRecords(&profile, &errors)
	traveller.PhoneHistoryRecords = generatePhoneHistoryRecords(&profile, &errors)
	return traveller
}

func generateAirBookingRecords(profile *customerprofiles.Profile) []model.AirBooking {
	airBookingRecs := []model.AirBooking{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_AIR_BOOKING {
			continue
		}
		rec := model.AirBooking{
			BookingID:     order.Attributes["booking_id"],
			SegmentID:     order.Attributes["segment_id"],
			From:          order.Attributes["from"],
			To:            order.Attributes["to"],
			FlightNumber:  order.Attributes["flight_number"],
			DepartureDate: order.Attributes["departure_date"],
			DepartureTime: order.Attributes["departure_time"],
			ArrivalDate:   order.Attributes["arrival_date"],
			ArrivalTime:   order.Attributes["arrival_time"],
			Channel:       order.Attributes["channel"],
			Status:        order.Attributes["status"],
			Price:         order.Attributes["price"],
		}
		airBookingRecs = append(airBookingRecs, rec)
	}
	return airBookingRecs
}

func generateAirLoyaltyRecords(profile *customerprofiles.Profile, errors *[]string) []model.AirLoyalty {
	airLoyaltyRecs := []model.AirLoyalty{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_AIR_LOYALTY {
			continue
		}
		rec := model.AirLoyalty{
			LoyaltyID:        order.Attributes["loyalty_id"],
			ProgramName:      order.Attributes["program_name"],
			Miles:            order.Attributes["miles"],
			MilesToNextLevel: order.Attributes["miles_to_next_level"],
			Level:            order.Attributes["level"],
		}
		joined := order.Attributes["joined"]
		if joined != "" {
			parsed, err := utils.TryParseTime(profile.BirthDate, "2006-01-02")
			if err == nil {
				rec.Joined = parsed
			} else {
				// TODO: decide on standard error messaging
				*errors = append(*errors, "error")
			}
		}
		airLoyaltyRecs = append(airLoyaltyRecs, rec)
	}
	return airLoyaltyRecs
}

func generateClickstreamRecords(profile *customerprofiles.Profile, errors *[]string) []model.Clickstream {
	clickstreamRecs := []model.Clickstream{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_CLICKSTREAM {
			continue
		}
		rec := model.Clickstream{
			SessionID:                       order.Attributes["session_id"],
			EventTimestamp:                  trySetDate(order.Attributes["event_timestamp"], errors),
			EventType:                       order.Attributes["event_type"],
			EventVersion:                    order.Attributes["event_version"],
			ArrivalTimestamp:                trySetDate(order.Attributes["arrival_timestamp"], errors),
			UserAgent:                       order.Attributes["user_agent"],
			Products:                        order.Attributes["products"],
			FareClass:                       order.Attributes["fare_class"],
			FareType:                        order.Attributes["fare_type"],
			FlightSegmentsDepartureDateTime: trySetDate(order.Attributes["flight_segments_departure_date_time"], errors),
			FlightNumbers:                   order.Attributes["flight_numbers"],
			FlightMarket:                    order.Attributes["flight_market"],
			FlightType:                      order.Attributes["flight_type"],
			OriginDate:                      order.Attributes["origin_date"],
			OriginDateTime:                  trySetDate(order.Attributes["origin_date_time"], errors),
			ReturnDate:                      order.Attributes["return_date"],
			ReturnDateTime:                  trySetDate(order.Attributes["return_date_time"], errors),
			ReturnFlightRoute:               order.Attributes["return_flight_route"],
			NumPaxAdults:                    trySetInt(order.Attributes["num_pax_adults"], errors),
			NumPaxInf:                       trySetInt(order.Attributes["num_pax_inf"], errors),
			NumPaxChildren:                  trySetInt(order.Attributes["num_pax_children"], errors),
			PaxType:                         order.Attributes["pax_type"],
			TotalPassengers:                 trySetInt(order.Attributes["total_passengers"], errors),
		}
		clickstreamRecs = append(clickstreamRecs, rec)
	}
	return clickstreamRecs
}

func generateEmailHistoryRecords(profile *customerprofiles.Profile, errors *[]string) []model.EmailHistory {
	emailRecs := []model.EmailHistory{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_EMAIL_HISTORY {
			continue
		}
		rec := model.EmailHistory{
			Address:       order.Attributes["address"],
			Type:          order.Attributes["type"],
			LastUpdated:   trySetDate(order.Attributes["last_updated"], errors),
			LastUpdatedBy: order.Attributes["last_updated_by"],
		}
		emailRecs = append(emailRecs, rec)
	}
	return emailRecs
}

func generateHotelBookingRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelBooking {
	hotelBookingRecs := []model.HotelBooking{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_HOTEL_BOOKING {
			continue
		}
		rec := model.HotelBooking{
			BookingID:             order.Attributes["booking_id"],
			HotelCode:             order.Attributes["hotel_code"],
			NumNights:             trySetInt(order.Attributes["n_nights"], errors),
			NumGuests:             trySetInt(order.Attributes["n_guests"], errors),
			ProductID:             order.Attributes["product_id"],
			CheckInDate:           trySetDate(order.Attributes["check_in_date"], errors),
			RoomTypeCode:          order.Attributes["room_type_code"],
			RoomTypeName:          order.Attributes["room_type_name"],
			RoomTypeDescription:   order.Attributes["room_type_description"],
			AttributeCodes:        order.Attributes["attribute_codes"],
			AttributeNames:        order.Attributes["attribute_names"],
			AttributeDescriptions: order.Attributes["attribute_descriptions"],
		}
		hotelBookingRecs = append(hotelBookingRecs, rec)
	}
	return hotelBookingRecs
}

func generateHotelLoyaltyRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelLoyalty {
	hotelLoyaltyRecs := []model.HotelLoyalty{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_HOTEL_LOYALTY {
			continue
		}
		rec := model.HotelLoyalty{
			LoyaltyID:         order.Attributes["loyalty_id"],
			ProgramName:       order.Attributes["program_name"],
			Points:            order.Attributes["points"],
			Units:             order.Attributes["units"],
			PointsToNextLevel: order.Attributes["points_to_next_level"],
			Level:             order.Attributes["level"],
			Joined:            trySetDate(order.Attributes["joined"], errors),
		}
		hotelLoyaltyRecs = append(hotelLoyaltyRecs, rec)
	}
	return hotelLoyaltyRecs
}

func generateHotelStayRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelStay {
	hotelStayRecs := []model.HotelStay{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_HOTEL_STAY {
			continue
		}
		rec := model.HotelStay{
			ID:             order.Attributes["id"],
			BookingID:      order.Attributes["booking_id"],
			CurrencyCode:   order.Attributes["currency_code"],
			CurrencyName:   order.Attributes["currency_name"],
			CurrencySymbol: order.Attributes["currency_symbol"],
			FirstName:      order.Attributes["first_name"],
			LastName:       order.Attributes["last_name"],
			Email:          order.Attributes["email"],
			Phone:          order.Attributes["phone"],
			StartDate:      trySetDate(order.Attributes["start_date"], errors),
			HotelCode:      order.Attributes["hotel_code"],
			Type:           order.Attributes["type"],
			Description:    order.Attributes["description"],
			Amount:         order.Attributes["amount"],
			Date:           trySetDate(order.Attributes["date"], errors),
		}
		hotelStayRecs = append(hotelStayRecs, rec)
	}
	return hotelStayRecs
}

func generatePhoneHistoryRecords(profile *customerprofiles.Profile, errors *[]string) []model.PhoneHistory {
	phoneHistoryRecs := []model.PhoneHistory{}
	for _, order := range profile.Orders {
		if order.Attributes["object_type"] != OBJECT_TYPE_PHONE_HISTORY {
			continue
		}
		rec := model.PhoneHistory{
			Number:        order.Attributes["number"],
			CountryCode:   order.Attributes["country_code"],
			Type:          order.Attributes["type"],
			LastUpdated:   trySetDate(order.Attributes["last_updated"], errors),
			LastUpdatedBy: order.Attributes["last_updated_by"],
		}
		phoneHistoryRecs = append(phoneHistoryRecs, rec)
	}
	return phoneHistoryRecs
}

func trySetDate(val string, errors *[]string) time.Time {
	parsed, err := utils.TryParseTime(val, "2006-01-02")
	if err != nil {
		// TODO: decide on standard error messaging
		*errors = append(*errors, err.Error())
	}
	return parsed
}

// TODO: decide what default value should be if we have an empty string
// 0 could be misleading if we don't have the data
func trySetInt(val string, errors *[]string) int {
	parsed, err := utils.TryParseInt(val)
	if err != nil {
		// TODO: decide on standard error messaging
		*errors = append(*errors, err.Error())
	}
	return parsed
}
