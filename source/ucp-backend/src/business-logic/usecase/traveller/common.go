package traveller

import (
	"sort"
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/traveller"
	"tah/ucp/src/business-logic/utils"
	"time"
)

// Customer Profile object types
const (
	OBJECT_TYPE_AIR_BOOKING   string = "air_booking"
	OBJECT_TYPE_AIR_LOYALTY   string = "air_loyalty"
	OBJECT_TYPE_CLICKSTREAM   string = "clickstream"
	OBJECT_TYPE_EMAIL_HISTORY string = "email_history"
	OBJECT_TYPE_HOTEL_BOOKING string = "hotel_booking"
	OBJECT_TYPE_HOTEL_LOYALTY string = "hotel_loyalty"
	OBJECT_TYPE_HOTEL_STAY    string = "hotel_stay_revenue_items"
	OBJECT_TYPE_PHONE_HISTORY string = "phone_history"
)

var COMBINED_PROFILE_OBJECT_TYPES = []string{
	OBJECT_TYPE_AIR_BOOKING,
	OBJECT_TYPE_AIR_LOYALTY,
	OBJECT_TYPE_CLICKSTREAM,
	OBJECT_TYPE_EMAIL_HISTORY,
	OBJECT_TYPE_HOTEL_BOOKING,
	OBJECT_TYPE_HOTEL_LOYALTY,
	OBJECT_TYPE_HOTEL_STAY,
	OBJECT_TYPE_PHONE_HISTORY,
}

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
		/*BusinessAddress: model.Address{
			Address1:   profile.BusinessAddress.Address1,
			Address2:   profile.BusinessAddress.Address2,
			Address3:   profile.BusinessAddress.Address3,
			Address4:   profile.BusinessAddress.Address4,
			City:       profile.BusinessAddress.City,
			State:      profile.BusinessAddress.State,
			Province:   profile.BusinessAddress.Province,
			PostalCode: profile.BusinessAddress.PostalCode,
			Country:    profile.BusinessAddress.Country,
		},
		MailingAddress: model.Address{
			Address1:   profile.MailingAddress.Address1,
			Address2:   profile.MailingAddress.Address2,
			Address3:   profile.MailingAddress.Address3,
			Address4:   profile.MailingAddress.Address4,
			City:       profile.MailingAddress.City,
			State:      profile.MailingAddress.State,
			Province:   profile.MailingAddress.Province,
			PostalCode: profile.MailingAddress.PostalCode,
			Country:    profile.MailingAddress.Country,
		},
		BillingAddress: model.Address{
			Address1:   profile.BillingAddress.Address1,
			Address2:   profile.BillingAddress.Address2,
			Address3:   profile.BillingAddress.Address3,
			Address4:   profile.BillingAddress.Address4,
			City:       profile.BillingAddress.City,
			State:      profile.BillingAddress.State,
			Province:   profile.BillingAddress.Province,
			PostalCode: profile.BillingAddress.PostalCode,
			Country:    profile.BillingAddress.Country,
		},*/
	}

	// TODO: handle each object in one loop through object records
	traveller.AirBookingRecords = generateAirBookingRecords(&profile, &errors)
	traveller.AirLoyaltyRecords = generateAirLoyaltyRecords(&profile, &errors)
	traveller.ClickstreamRecords = generateClickstreamRecords(&profile, &errors)
	traveller.EmailHistoryRecords = generateEmailHistoryRecords(&profile, &errors)
	traveller.HotelBookingRecords = generateHotelBookingRecords(&profile, &errors)
	traveller.HotelLoyaltyRecords = generateHotelLoyaltyRecords(&profile, &errors)
	traveller.HotelStayRecords = generateHotelStayRecords(&profile, &errors)
	traveller.PhoneHistoryRecords = generatePhoneHistoryRecords(&profile, &errors)
	traveller.ParsingErrors = errors

	//sorting data
	//TODO: to parallelize
	sort.Sort(model.AirBookingByLastUpdated(traveller.AirBookingRecords))
	sort.Sort(model.AirLoyaltyByLastUpdated(traveller.AirLoyaltyRecords))
	sort.Sort(model.ClickstreamByLastUpdated(traveller.ClickstreamRecords))
	sort.Sort(model.EmailHistoryByLastUpdated(traveller.EmailHistoryRecords))
	sort.Sort(model.HotelBookingByLastUpdated(traveller.HotelBookingRecords))
	sort.Sort(model.HotelLoyaltyByLastUpdated(traveller.HotelLoyaltyRecords))
	sort.Sort(model.HotelStayByLastUpdated(traveller.HotelStayRecords))
	sort.Sort(model.PhoneHistoryByLastUpdated(traveller.PhoneHistoryRecords))

	return traveller
}

func generateAirBookingRecords(profile *customerprofiles.Profile, errors *[]string) []model.AirBooking {
	airBookingRecs := []model.AirBooking{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_AIR_BOOKING {
			continue
		}
		rec := model.AirBooking{
			BookingID:     obj.Attributes["booking_id"],
			SegmentID:     obj.Attributes["segment_id"],
			From:          obj.Attributes["from"],
			To:            obj.Attributes["to"],
			FlightNumber:  obj.Attributes["flight_number"],
			DepartureDate: obj.Attributes["departure_date"],
			DepartureTime: obj.Attributes["departure_time"],
			ArrivalDate:   obj.Attributes["arrival_date"],
			ArrivalTime:   obj.Attributes["arrival_time"],
			Channel:       obj.Attributes["channel"],
			Status:        obj.Attributes["status"],
			Price:         obj.Attributes["price"],
			LastUpdated:   trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		airBookingRecs = append(airBookingRecs, rec)
	}
	return airBookingRecs
}

func generateAirLoyaltyRecords(profile *customerprofiles.Profile, errors *[]string) []model.AirLoyalty {
	airLoyaltyRecs := []model.AirLoyalty{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_AIR_LOYALTY {
			continue
		}
		rec := model.AirLoyalty{
			LoyaltyID:        obj.Attributes["loyalty_id"],
			ProgramName:      obj.Attributes["program_name"],
			Miles:            obj.Attributes["miles"],
			MilesToNextLevel: obj.Attributes["miles_to_next_level"],
			Level:            obj.Attributes["level"],
			LastUpdated:      trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		joined := obj.Attributes["joined"]
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
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_CLICKSTREAM {
			continue
		}
		rec := model.Clickstream{
			SessionID:                       obj.Attributes["session_id"],
			EventTimestamp:                  trySetDate(obj.Attributes["event_timestamp"], errors),
			EventType:                       obj.Attributes["event_type"],
			EventVersion:                    obj.Attributes["event_version"],
			ArrivalTimestamp:                trySetDate(obj.Attributes["arrival_timestamp"], errors),
			UserAgent:                       obj.Attributes["user_agent"],
			Products:                        obj.Attributes["products"],
			FareClass:                       obj.Attributes["fare_class"],
			FareType:                        obj.Attributes["fare_type"],
			FlightSegmentsDepartureDateTime: trySetDate(obj.Attributes["flight_segments_departure_date_time"], errors),
			FlightNumbers:                   obj.Attributes["flight_numbers"],
			FlightMarket:                    obj.Attributes["flight_market"],
			FlightType:                      obj.Attributes["flight_type"],
			OriginDate:                      obj.Attributes["origin_date"],
			OriginDateTime:                  trySetDate(obj.Attributes["origin_date_time"], errors),
			ReturnDate:                      obj.Attributes["return_date"],
			ReturnDateTime:                  trySetDate(obj.Attributes["return_date_time"], errors),
			ReturnFlightRoute:               obj.Attributes["return_flight_route"],
			NumPaxAdults:                    trySetInt(obj.Attributes["num_pax_adults"], errors),
			NumPaxInf:                       trySetInt(obj.Attributes["num_pax_inf"], errors),
			NumPaxChildren:                  trySetInt(obj.Attributes["num_pax_children"], errors),
			PaxType:                         obj.Attributes["pax_type"],
			TotalPassengers:                 trySetInt(obj.Attributes["total_passengers"], errors),
			LastUpdated:                     trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		clickstreamRecs = append(clickstreamRecs, rec)
	}
	return clickstreamRecs
}

func generateEmailHistoryRecords(profile *customerprofiles.Profile, errors *[]string) []model.EmailHistory {
	emailRecs := []model.EmailHistory{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_EMAIL_HISTORY {
			continue
		}
		rec := model.EmailHistory{
			Address:       obj.Attributes["address"],
			Type:          obj.Attributes["type"],
			LastUpdated:   trySetTimestamp(obj.Attributes["last_updated"], errors),
			LastUpdatedBy: obj.Attributes["last_updated_by"],
		}
		emailRecs = append(emailRecs, rec)
	}
	return emailRecs
}

func generateHotelBookingRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelBooking {
	hotelBookingRecs := []model.HotelBooking{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_HOTEL_BOOKING {
			continue
		}
		rec := model.HotelBooking{

			BookingID:             obj.Attributes["booking_id"],
			HotelCode:             obj.Attributes["hotel_code"],
			NumNights:             trySetInt(obj.Attributes["n_nights"], errors),
			NumGuests:             trySetInt(obj.Attributes["n_guests"], errors),
			ProductID:             obj.Attributes["product_id"],
			CheckInDate:           trySetDate(obj.Attributes["check_in_date"], errors),
			RoomTypeCode:          obj.Attributes["room_type_code"],
			RoomTypeName:          obj.Attributes["room_type_name"],
			RoomTypeDescription:   obj.Attributes["room_type_description"],
			AttributeCodes:        obj.Attributes["attribute_codes"],
			AttributeNames:        obj.Attributes["attribute_names"],
			AttributeDescriptions: obj.Attributes["attribute_descriptions"],
			LastUpdated:           trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		hotelBookingRecs = append(hotelBookingRecs, rec)
	}
	return hotelBookingRecs
}

func generateHotelLoyaltyRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelLoyalty {
	hotelLoyaltyRecs := []model.HotelLoyalty{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_HOTEL_LOYALTY {
			continue
		}
		rec := model.HotelLoyalty{
			LoyaltyID:         obj.Attributes["loyalty_id"],
			ProgramName:       obj.Attributes["program_name"],
			Points:            obj.Attributes["points"],
			Units:             obj.Attributes["units"],
			PointsToNextLevel: obj.Attributes["points_to_next_level"],
			Level:             obj.Attributes["level"],
			Joined:            trySetDate(obj.Attributes["joined"], errors),
			LastUpdated:       trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		hotelLoyaltyRecs = append(hotelLoyaltyRecs, rec)
	}
	return hotelLoyaltyRecs
}

func generateHotelStayRecords(profile *customerprofiles.Profile, errors *[]string) []model.HotelStay {
	hotelStayRecs := []model.HotelStay{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_HOTEL_STAY {
			continue
		}
		rec := model.HotelStay{
			ID:             obj.Attributes["id"],
			BookingID:      obj.Attributes["booking_id"],
			CurrencyCode:   obj.Attributes["currency_code"],
			CurrencyName:   obj.Attributes["currency_name"],
			CurrencySymbol: obj.Attributes["currency_symbol"],
			FirstName:      obj.Attributes["first_name"],
			LastName:       obj.Attributes["last_name"],
			Email:          obj.Attributes["email"],
			Phone:          obj.Attributes["phone"],
			StartDate:      trySetDate(obj.Attributes["start_date"], errors),
			HotelCode:      obj.Attributes["hotel_code"],
			Type:           obj.Attributes["type"],
			Description:    obj.Attributes["description"],
			Amount:         obj.Attributes["amount"],
			Date:           trySetDate(obj.Attributes["date"], errors),
			LastUpdated:    trySetTimestamp(obj.Attributes["last_updated"], errors),
		}
		hotelStayRecs = append(hotelStayRecs, rec)
	}
	return hotelStayRecs
}

func generatePhoneHistoryRecords(profile *customerprofiles.Profile, errors *[]string) []model.PhoneHistory {
	phoneHistoryRecs := []model.PhoneHistory{}
	for _, obj := range profile.ProfileObjects {
		if obj.Type != OBJECT_TYPE_PHONE_HISTORY {
			continue
		}
		rec := model.PhoneHistory{
			Number:        obj.Attributes["number"],
			CountryCode:   obj.Attributes["country_code"],
			Type:          obj.Attributes["type"],
			LastUpdated:   trySetTimestamp(obj.Attributes["last_updated"], errors),
			LastUpdatedBy: obj.Attributes["last_updated_by"],
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

func trySetTimestamp(val string, errors *[]string) time.Time {
	parsed, err := utils.TryParseTime(val, "2006-01-02T15:04:05.999999Z")
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
