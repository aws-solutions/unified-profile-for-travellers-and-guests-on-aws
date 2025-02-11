// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"strings"
	"tah/upt/source/tah-core/core"
	profilemodel "tah/upt/source/tah-core/customerprofilesmodel"
	constant "tah/upt/source/ucp-common/src/constant/admin"
	model "tah/upt/source/ucp-common/src/model/traveler"
	"tah/upt/source/ucp-common/src/utils/utils"
	"time"
)

func ProfilesToTravellers(tx core.Transaction, profiles []profilemodel.Profile) []model.Traveller {
	travellers := []model.Traveller{}
	for _, profile := range profiles {
		travellers = append(travellers, ProfileToTraveller(tx, profile))
	}
	return travellers
}

func ProfileToTraveller(tx core.Transaction, profile profilemodel.Profile) model.Traveller {
	tx.Debug("ProfileToTraveller")
	errors := []string{}
	traveller := model.Traveller{
		// Metadata
		ModelVersion: "1",
		LastUpdated:  trySetTimestamp("profile.last_updated", profile.Attributes["last_updated_on"], &errors),

		// Profile IDs
		ConnectID:   profile.ProfileId,
		Domain:      profile.Domain,
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
		BirthDate:   trySetDate("profile.BirthDate", profile.BirthDate, &errors),
		JobTitle:    profile.Attributes["job_title"],
		CompanyName: profile.BusinessName,

		// Contact Info
		PhoneNumber:          profile.PhoneNumber,
		MobilePhoneNumber:    profile.MobilePhoneNumber,
		HomePhoneNumber:      profile.HomePhoneNumber,
		BusinessPhoneNumber:  profile.BusinessPhoneNumber,
		PersonalEmailAddress: profile.PersonalEmailAddress,
		BusinessEmailAddress: profile.BusinessEmailAddress,
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
		//we use the shipping address as business email address since it is not in use for T&H
		BusinessAddress: model.Address{
			Address1:   profile.ShippingAddress.Address1,
			Address2:   profile.ShippingAddress.Address2,
			Address3:   profile.ShippingAddress.Address3,
			Address4:   profile.ShippingAddress.Address4,
			City:       profile.ShippingAddress.City,
			State:      profile.ShippingAddress.State,
			Province:   profile.ShippingAddress.Province,
			PostalCode: profile.ShippingAddress.PostalCode,
			Country:    profile.ShippingAddress.Country,
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
		},
	}

	// handle each object in one loop through object records
	traveller.AirBookingRecords, traveller.NAirBookingRecords = generateAirBookingRecords(&profile, &errors)
	traveller.AirLoyaltyRecords, traveller.NAirLoyaltyRecords = generateAirLoyaltyRecords(&profile, &errors)
	traveller.ClickstreamRecords, traveller.NClickstreamRecords = generateClickstreamRecords(&profile, &errors)
	traveller.EmailHistoryRecords, traveller.NEmailHistoryRecords = generateEmailHistoryRecords(&profile, &errors)
	traveller.HotelBookingRecords, traveller.NHotelBookingRecords = generateHotelBookingRecords(&profile, &errors)
	traveller.HotelLoyaltyRecords, traveller.NHotelLoyaltyRecords = generateHotelLoyaltyRecords(&profile, &errors)
	traveller.HotelStayRecords, traveller.NHotelStayRecords = generateHotelStayRecords(&profile, &errors)
	traveller.CustomerServiceInteractionRecords, traveller.NCustomerServiceInteractionRecords = generateCustomerServiceInteractionRecords(&profile, &errors)
	traveller.LoyaltyTxRecords, traveller.NLoyaltyTxRecords = generateLoyaltyTxRecords(&profile, &errors)
	traveller.PhoneHistoryRecords, traveller.NPhoneHistoryRecords = generatePhoneHistoryRecords(&profile, &errors)
	traveller.AncillaryServiceRecords, traveller.NAncillaryServiceRecords = generateAncillaryRecords(&profile, &errors)
	traveller.AlternateProfileIDs, traveller.NAlternateProfileIdRecords = generateAlternateProfileIdRecords(&profile, &errors)
	traveller.ParsingErrors = errors

	//sorting data
	//to parallelize
	/*sort.Sort(model.AirBookingByLastUpdated(traveller.AirBookingRecords))
	sort.Sort(model.AirLoyaltyByLastUpdated(traveller.AirLoyaltyRecords))
	sort.Sort(model.ClickstreamByLastUpdated(traveller.ClickstreamRecords))
	sort.Sort(model.EmailHistoryByLastUpdated(traveller.EmailHistoryRecords))
	sort.Sort(model.HotelBookingByLastUpdated(traveller.HotelBookingRecords))
	sort.Sort(model.HotelLoyaltyByLastUpdated(traveller.HotelLoyaltyRecords))
	sort.Sort(model.HotelStayByLastUpdated(traveller.HotelStayRecords))
	sort.Sort(model.PhoneHistoryByLastUpdated(traveller.PhoneHistoryRecords))
	sort.Sort(model.LoyaltyTxByLastUpdated(traveller.LoyaltyTxRecords))
	sort.Sort(model.CustomerServiceInteractionByLastUpdated(traveller.CustomerServiceInteractionRecords))*/

	return traveller
}

func generateCustomerServiceInteractionRecords(profile *profilemodel.Profile, errors *[]string) ([]model.CustomerServiceInteraction, int64) {
	recs := []model.CustomerServiceInteraction{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_CSI {
			continue
		}
		rec := model.CustomerServiceInteraction{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			TravellerID:            obj.Attributes["traveller_id"],
			LastUpdated:            trySetTimestamp("CustomerServiceInteraction.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			StartTime:              trySetTimestamp("CustomerServiceInteraction.start_time", obj.Attributes["start_time"], errors),
			EndTime:                trySetTimestamp("CustomerServiceInteraction.end_time", obj.Attributes["end_time"], errors),
			Duration:               trySetInt(obj.Attributes["duration"], errors),
			SessionId:              obj.Attributes["traveller_id"],
			Channel:                obj.Attributes["channel"],
			InteractionType:        obj.Attributes["interaction_type"],
			Status:                 obj.Attributes["status"],
			LanguageCode:           obj.Attributes["language_code"],
			LanguageName:           obj.Attributes["language_name"],
			Conversation:           obj.Attributes["conversation"],
			ConversationSummary:    obj.Attributes["summary"],
			BookingID:              obj.Attributes["booking_id"],
			LoyaltyID:              obj.Attributes["loyalty_id"],
			CartID:                 obj.Attributes["cart_id"],
			SentimentScore:         trySetFloat(obj.Attributes["sentiment_score"], errors),
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			CampaignJobID:          obj.Attributes["campaign_job_id"],
			CampaignStrategy:       obj.Attributes["campaign_strategy"],
			CampaignProgram:        obj.Attributes["campaign_program"],
			CampaignProduct:        obj.Attributes["campaign_product"],
			CampaignName:           obj.Attributes["campaign_name"],
			Category:               obj.Attributes["category"],
			Subject:                obj.Attributes["subject"],
			LoyaltyProgramName:     obj.Attributes["loyalty_program_name"],
			IsVoiceOtp:             trySetBool(obj.Attributes["is_voice_otp"]),
		}
		nRecs = obj.TotalCount
		recs = append(recs, rec)
	}
	return recs, nRecs
}

func generateLoyaltyTxRecords(profile *profilemodel.Profile, errors *[]string) ([]model.LoyaltyTx, int64) {
	recs := []model.LoyaltyTx{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_LOYALTY_TX {
			continue
		}
		rec := model.LoyaltyTx{
			AccpObjectID:             obj.Attributes["accp_object_id"],
			TravellerID:              obj.Attributes["traveller_id"],
			LastUpdated:              trySetTimestamp("LoyaltyTx.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:            obj.Attributes["last_updated_by"],
			Category:                 obj.Attributes["category"],
			PointsOffset:             trySetFloat(obj.Attributes["points_offset"], errors),
			PointUnit:                obj.Attributes["points_unit"],
			OriginPointsOffset:       trySetFloat(obj.Attributes["origin_points_offset"], errors),
			QualifyingPointsOffset:   trySetFloat(obj.Attributes["qualifying_point_offset"], errors),
			Source:                   obj.Attributes["source"],
			BookingDate:              trySetTimestamp("LoyaltyTx.booking_date", obj.Attributes["booking_date"], errors),
			OrderNumber:              obj.Attributes["order_number"],
			ProductId:                obj.Attributes["product_id"],
			ExpireInDays:             trySetInt(obj.Attributes["expire_in_days"], errors),
			Amount:                   trySetFloat(obj.Attributes["amount"], errors),
			AmountType:               obj.Attributes["amount_type"],
			VoucherQuantity:          trySetInt(obj.Attributes["voucher_quantity"], errors),
			CorporateReferenceNumber: obj.Attributes["corporate_reference_number"],
			Promotions:               obj.Attributes["promotions"],
			ActivityDay:              trySetTimestamp("LoyaltyTx.activity_day", obj.Attributes["activity_day"], errors),
			Location:                 obj.Attributes["location"],
			ToLoyaltyId:              obj.Attributes["to_loyalty_id"],
			FromLoyaltyId:            obj.Attributes["from_loyalty_id"],
			OrganizationCode:         obj.Attributes["organization_code"],
			EventName:                obj.Attributes["event_name"],
			DocumentNumber:           obj.Attributes["document_number"],
			CorporateId:              obj.Attributes["corporate_id"],
			ProgramName:              obj.Attributes["program_name"],
			OverallConfidenceScore:   trySetFloat(obj.Attributes["overall_confidence_score"], errors),
		}
		nRecs = obj.TotalCount
		recs = append(recs, rec)
	}
	return recs, nRecs

}

func generateAirBookingRecords(profile *profilemodel.Profile, errors *[]string) ([]model.AirBooking, int64) {
	airBookingRecs := []model.AirBooking{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_AIR_BOOKING {
			continue
		}
		rec := model.AirBooking{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			BookingID:              obj.Attributes["booking_id"],
			SegmentID:              obj.Attributes["segment_id"],
			From:                   obj.Attributes["from"],
			To:                     obj.Attributes["to"],
			FlightNumber:           obj.Attributes["flight_number"],
			DepartureDate:          obj.Attributes["departure_date"],
			DepartureTime:          obj.Attributes["departure_time"],
			ArrivalDate:            obj.Attributes["arrival_date"],
			ArrivalTime:            obj.Attributes["arrival_time"],
			Channel:                obj.Attributes["channel"],
			Status:                 obj.Attributes["status"],
			TotalPrice:             trySetFloat(obj.Attributes["price"], errors),
			TravellerPrice:         trySetFloat(obj.Attributes["traveller_price"], errors),
			LastUpdated:            trySetTimestamp("AirBooking.joined", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			CreationChannelID:      obj.Attributes["creation_channel_id"],
			LastUpdateChannelID:    obj.Attributes["last_update_channel_id"],
			BookerID:               obj.Attributes["booker_id"],
			DayOfTravelEmail:       obj.Attributes["day_of_travel_email"],
			DayOfTravelPhone:       obj.Attributes["day_of_travel_phone"],
			FirstName:              obj.Attributes["first_name"],
			LastName:               obj.Attributes["last_name"],
		}
		nRecs = obj.TotalCount
		airBookingRecs = append(airBookingRecs, rec)
	}
	return airBookingRecs, nRecs
}

func generateAncillaryRecords(profile *profilemodel.Profile, errors *[]string) ([]model.AncillaryService, int64) {
	ancillaryRecs := []model.AncillaryService{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_ANCILLARY {
			continue
		}
		rec := model.AncillaryService{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			LastUpdated:            trySetTimestamp("AncillaryService.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			PaxIndex:               int64(trySetInt(obj.Attributes["pax_index"], errors)),
			AncillaryType:          obj.Attributes["ancillary_type"],
			BookingID:              obj.Attributes["booking_id"],
			FlightNumber:           obj.Attributes["flight_number"],
			DepartureDate:          obj.Attributes["departure_date"],
			Price:                  trySetFloat(obj.Attributes["price"], errors),
			Currency:               obj.Attributes["currency"],
			//baggage
			BaggageType:              obj.Attributes["baggage_type"],
			Quantity:                 int64(trySetInt(obj.Attributes["quantity"], errors)),
			Weight:                   trySetFloat(obj.Attributes["weight"], errors),
			DimentionsLength:         trySetFloat(obj.Attributes["dimentions_length"], errors),
			DimentionsWidth:          trySetFloat(obj.Attributes["dimentions_width"], errors),
			DimentionsHeight:         trySetFloat(obj.Attributes["dimentions_height"], errors),
			PriorityBagDrop:          trySetBool(obj.Attributes["priority_bag_drop"]),
			PriorityBagReturn:        trySetBool(obj.Attributes["priority_bag_return"]),
			LotBagInsurance:          trySetBool(obj.Attributes["lot_bag_insurance"]),
			ValuableBaggageInsurance: trySetBool(obj.Attributes["valuable_baggage_insurance"]),
			HandsFreeBaggage:         trySetBool(obj.Attributes["hands_free_baggage"]),
			//change
			ChangeType: obj.Attributes["change_type"],
			//seat
			SeatNumber:       obj.Attributes["seat_number"],
			SeatZone:         obj.Attributes["seat_zone"],
			NeighborFreeSeat: trySetBool(obj.Attributes["neighbor_free_seat"]),
			UpgradeAuction:   trySetBool(obj.Attributes["upgrade_auction"]),
			//Other
			OtherAncilliaryType: obj.Attributes["other_ancilliary_type"],
			//priority
			PriorityServiceType: obj.Attributes["priority_service_type"],
			LoungeAccess:        trySetBool(obj.Attributes["lounge_access"]),
		}
		nRecs = obj.TotalCount
		ancillaryRecs = append(ancillaryRecs, rec)
	}
	return ancillaryRecs, nRecs
}

func generateAirLoyaltyRecords(profile *profilemodel.Profile, errors *[]string) ([]model.AirLoyalty, int64) {
	airLoyaltyRecs := []model.AirLoyalty{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_AIR_LOYALTY {
			continue
		}
		rec := model.AirLoyalty{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			LoyaltyID:              obj.Attributes["id"],
			ProgramName:            obj.Attributes["program_name"],
			Miles:                  obj.Attributes["miles"],
			MilesToNextLevel:       obj.Attributes["miles_to_next_level"],
			Level:                  obj.Attributes["level"],
			Joined:                 trySetTimestamp("AirLoyalty.joined", obj.Attributes["joined"], errors),
			LastUpdated:            trySetTimestamp("AirLoyalty.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			EnrollmentSource:       obj.Attributes["enrollment_source"],
			Currency:               obj.Attributes["currency"],
			Amount:                 trySetFloat(obj.Attributes["amount"], errors),
			AccountStatus:          obj.Attributes["account_status"],
			ReasonForClose:         obj.Attributes["reason_for_close"],
			LanguagePreference:     obj.Attributes["language_preference"],
			DisplayPreference:      obj.Attributes["display_preference"],
			MealPreference:         obj.Attributes["meal_preference"],
			SeatPreference:         obj.Attributes["seat_preference"],
			HomeAirport:            obj.Attributes["home_airport"],
			DateTimeFormatPref:     obj.Attributes["date_time_format_preference"],
			CabinPreference:        obj.Attributes["cabin_preference"],
			FareTypePreference:     obj.Attributes["fare_type_preference"],
			ExpertMode:             trySetBool(obj.Attributes["expert_mode"]),
			PrivacyIndicator:       obj.Attributes["privacy_indicator"],
			CarPreferenceVendor:    obj.Attributes["car_preference_vendor"],
			CarPreferenceType:      obj.Attributes["car_preference_type"],
			SpecialAccommodation1:  obj.Attributes["special_accommodation_1"],
			SpecialAccommodation2:  obj.Attributes["special_accommodation_2"],
			SpecialAccommodation3:  obj.Attributes["special_accommodation_3"],
			SpecialAccommodation4:  obj.Attributes["special_accommodation_4"],
			SpecialAccommodation5:  obj.Attributes["special_accommodation_5"],
			MarketingOptIns:        obj.Attributes["marketing_opt_ins"],
			RenewDate:              obj.Attributes["renew_date"],
			NextBillAmount:         trySetFloat(obj.Attributes["next_bill_amount"], errors),
			ClearEnrollDate:        obj.Attributes["clear_enroll_date"],
			ClearRenewDate:         obj.Attributes["clear_renew_date"],
			ClearTierLevel:         obj.Attributes["clear_tier_level"],
			ClearNextBillAmount:    trySetFloat(obj.Attributes["clear_next_bill_amount"], errors),
			ClearIsActive:          trySetBool(obj.Attributes["clear_is_active"]),
			ClearAutoRenew:         trySetBool(obj.Attributes["clear_auto_renew"]),
			ClearHasBiometrics:     trySetBool(obj.Attributes["clear_has_biometrics"]),
			ClearHasPartnerPricing: trySetBool(obj.Attributes["clear_has_partner_pricing"]),
			TsaType:                obj.Attributes["tsa_type"],
			TsaSeqNum:              obj.Attributes["tsa_seq_num"],
			TsaNumber:              obj.Attributes["tsa_number"],
		}
		nRecs = obj.TotalCount
		airLoyaltyRecs = append(airLoyaltyRecs, rec)
	}
	return airLoyaltyRecs, nRecs
}

func generateClickstreamRecords(profile *profilemodel.Profile, errors *[]string) ([]model.Clickstream, int64) {
	clickstreamRecs := []model.Clickstream{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_CLICKSTREAM {
			continue
		}
		rec := model.Clickstream{
			//Accp fields
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			LastUpdated:            trySetTimestamp("Clickstream.last_updated", obj.Attributes["last_updated"], errors),
			//General Clickstream
			SessionID:        obj.Attributes["session_id"],
			EventTimestamp:   trySetTimestamp("Clickstream.event_timestamp", obj.Attributes["event_timestamp"], errors),
			EventType:        obj.Attributes["event_type"],
			EventVersion:     obj.Attributes["event_version"],
			ArrivalTimestamp: trySetTimestamp("Clickstream.arrival_timestamp", obj.Attributes["arrival_timestamp"], errors),
			UserAgent:        obj.Attributes["user_agent"],
			CustomEventName:  obj.Attributes["custom_event_name"],
			Error:            obj.Attributes["error"],

			// Customer
			CustomerBirthdate:   trySetDate("Clickstream.customer_birthdate", obj.Attributes["customer_birthdate"], errors),
			CustomerCountry:     obj.Attributes["customer_country"],
			CustomerEmail:       obj.Attributes["customer_email"],
			CustomerFirstName:   obj.Attributes["customer_first_name"],
			CustomerGender:      obj.Attributes["customer_gender"],
			CustomerID:          obj.Attributes["customer_id"],
			CustomerLoyaltyID:   obj.Attributes["customer_loyalty_id"],
			CustomerLastName:    obj.Attributes["customer_last_name"],
			CustomerNationality: obj.Attributes["customer_nationality"],
			CustomerPhone:       obj.Attributes["customer_phone"],
			LanguageCode:        obj.Attributes["language_code"],
			CustomerType:        obj.Attributes["customer_type"],

			// Cross industry
			Currency:            obj.Attributes["currency"],
			Products:            obj.Attributes["products"],
			ProductsPrices:      obj.Attributes["products_prices"],
			Quantities:          obj.Attributes["quantities"],
			EcommerceAction:     obj.Attributes["ecommerce_action"],
			OrderPaymentType:    obj.Attributes["order_payment_type"],
			OrderPromoCode:      obj.Attributes["order_promo_code"],
			PageName:            obj.Attributes["page_name"],
			PageTypeEnvironment: obj.Attributes["page_type_environment"],
			TransactionID:       obj.Attributes["transaction_id"],
			BookingID:           obj.Attributes["booking_id"],
			GeofenceLatitude:    obj.Attributes["geofence_latitude"],
			GeofenceLongitude:   obj.Attributes["geofence_longitude"],
			GeofenceID:          obj.Attributes["geofence_id"],
			GeofenceName:        obj.Attributes["geofence_name"],
			PoiID:               obj.Attributes["poi_id"],
			Custom:              obj.Attributes["custom"],
			URL:                 obj.Attributes["url"],

			//Air
			FareClass:                       obj.Attributes["fare_class"],
			FareType:                        obj.Attributes["fare_type"],
			FlightSegmentsDepartureDateTime: obj.Attributes["flight_segments_departure_date_time"],
			FlightNumbers:                   obj.Attributes["flight_numbers"],
			FlightMarket:                    obj.Attributes["flight_market"],
			FlightType:                      obj.Attributes["flight_type"],
			OriginDate:                      obj.Attributes["origin_date"],
			OriginDateTime:                  trySetDate("Clickstream.origin_date_time", obj.Attributes["origin_date_time"], errors),
			ReturnDate:                      obj.Attributes["return_date"],
			ReturnDateTime:                  trySetDate("Clickstream.return_date_time", obj.Attributes["return_date_time"], errors),
			ReturnFlightRoute:               obj.Attributes["return_flight_route"],
			NumPaxAdults:                    trySetInt(obj.Attributes["num_pax_adults"], errors),
			NumPaxInf:                       trySetInt(obj.Attributes["num_pax_inf"], errors),
			NumPaxChildren:                  trySetInt(obj.Attributes["num_pax_children"], errors),
			PaxType:                         obj.Attributes["pax_type"],
			TotalPassengers:                 trySetInt(obj.Attributes["total_passengers"], errors),
			LengthOfStay:                    trySetInt(obj.Attributes["length_of_stay"], errors),
			Origin:                          obj.Attributes["origin"],
			SelectedSeats:                   obj.Attributes["selected_seats"],

			// Hotels
			RoomType:          obj.Attributes["room_type"],
			RatePlan:          obj.Attributes["rate_plan"],
			CheckinDate:       trySetDate("Clickstream.checkin_date", obj.Attributes["checkin_date"], errors),
			CheckoutDate:      trySetDate("Clickstream.checkout_date", obj.Attributes["checkout_date"], errors),
			NumNights:         trySetInt(obj.Attributes["num_nights"], errors),
			NumGuests:         trySetInt(obj.Attributes["num_guests"], errors),
			NumGuestsAdult:    trySetInt(obj.Attributes["num_guest_adults"], errors),
			NumGuestsChildren: trySetInt(obj.Attributes["num_guest_children"], errors),
			HotelCode:         obj.Attributes["hotel_code"],
			HotelCodeList:     obj.Attributes["hotel_code_list"],
			HotelName:         obj.Attributes["hotel_name"],
			Destination:       obj.Attributes["destination"],
		}
		nRecs = obj.TotalCount
		clickstreamRecs = append(clickstreamRecs, rec)
	}
	return clickstreamRecs, nRecs
}

func generateEmailHistoryRecords(profile *profilemodel.Profile, errors *[]string) ([]model.EmailHistory, int64) {
	emailRecs := []model.EmailHistory{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_EMAIL_HISTORY {
			continue
		}
		rec := model.EmailHistory{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			Address:                obj.Attributes["address"],
			Type:                   obj.Attributes["type"],
			LastUpdated:            trySetTimestamp("EmailHistory.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			IsVerified:             trySetBool(obj.Attributes["is_verified"]),
		}
		nRecs = obj.TotalCount
		emailRecs = append(emailRecs, rec)
	}
	return emailRecs, nRecs
}

func generateHotelBookingRecords(profile *profilemodel.Profile, errors *[]string) ([]model.HotelBooking, int64) {
	hotelBookingRecs := []model.HotelBooking{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_HOTEL_BOOKING {
			continue
		}
		rec := model.HotelBooking{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			BookingID:              obj.Attributes["booking_id"],
			HotelCode:              obj.Attributes["hotel_code"],
			NumNights:              trySetInt(obj.Attributes["n_nights"], errors),
			NumGuests:              trySetInt(obj.Attributes["n_guests"], errors),
			ProductID:              obj.Attributes["product_id"],
			CheckInDate:            trySetDate("HotelBooking.check_in_date", obj.Attributes["check_in_date"], errors),
			RoomTypeCode:           obj.Attributes["room_type_code"],
			RoomTypeName:           obj.Attributes["room_type_name"],
			RoomTypeDescription:    obj.Attributes["room_type_description"],
			RatePlanCode:           obj.Attributes["rate_plan_code"],
			RatePlanName:           obj.Attributes["rate_plan_name"],
			RatePlanDescription:    obj.Attributes["rate_plan_description"],
			AttributeCodes:         obj.Attributes["attribute_codes"],
			AttributeNames:         obj.Attributes["attribute_names"],
			AttributeDescriptions:  obj.Attributes["attribute_descriptions"],
			AddOnCodes:             obj.Attributes["add_on_codes"],
			AddOnNames:             obj.Attributes["add_on_names"],
			AddOnDescriptions:      obj.Attributes["add_on_descriptions"],
			LastUpdated:            trySetTimestamp("HotelBooking.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			BookerID:               obj.Attributes["booker_id"],
			Status:                 obj.Attributes["status"],
			CreationChannelID:      obj.Attributes["creation_channel_id"],
			LastUpdateChannelID:    obj.Attributes["last_update_channel_id"],
			TotalSegmentAfterTax:   trySetFloat(obj.Attributes["total_segment_after_tax"], errors),
			TotalSegmentBeforeTax:  trySetFloat(obj.Attributes["total_segment_before_tax"], errors),
			FirstName:              obj.Attributes["first_name"],
			LastName:               obj.Attributes["last_name"],
		}
		nRecs = obj.TotalCount
		hotelBookingRecs = append(hotelBookingRecs, rec)
	}
	return hotelBookingRecs, nRecs
}

func generateHotelLoyaltyRecords(profile *profilemodel.Profile, errors *[]string) ([]model.HotelLoyalty, int64) {
	hotelLoyaltyRecs := []model.HotelLoyalty{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_HOTEL_LOYALTY {
			continue
		}
		rec := model.HotelLoyalty{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			LoyaltyID:              obj.Attributes["id"],
			ProgramName:            obj.Attributes["program_name"],
			Points:                 obj.Attributes["points"],
			Units:                  obj.Attributes["units"],
			PointsToNextLevel:      obj.Attributes["points_to_next_level"],
			Level:                  obj.Attributes["level"],
			Joined:                 trySetTimestamp("HotelLoyalty.joined", obj.Attributes["joined"], errors),
			LastUpdated:            trySetTimestamp("HotelLoyalty.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
		}
		nRecs = obj.TotalCount
		hotelLoyaltyRecs = append(hotelLoyaltyRecs, rec)
	}
	return hotelLoyaltyRecs, nRecs
}

func generateHotelStayRecords(profile *profilemodel.Profile, errors *[]string) ([]model.HotelStay, int64) {
	hotelStayRecs := []model.HotelStay{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_HOTEL_STAY_MAPPING {
			continue
		}
		rec := model.HotelStay{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			StayID:                 obj.Attributes["id"],
			BookingID:              obj.Attributes["booking_id"],
			CurrencyCode:           obj.Attributes["currency_code"],
			CurrencyName:           obj.Attributes["currency_name"],
			CurrencySymbol:         obj.Attributes["currency_symbol"],
			FirstName:              obj.Attributes["first_name"],
			LastName:               obj.Attributes["last_name"],
			Email:                  obj.Attributes["email"],
			Phone:                  obj.Attributes["phone"],
			StartDate:              trySetDate("hotel_stay.start_date", obj.Attributes["start_date"], errors),
			HotelCode:              obj.Attributes["hotel_code"],
			Type:                   obj.Attributes["type"],
			Description:            obj.Attributes["description"],
			Amount:                 obj.Attributes["amount"],
			Date:                   trySetTimestamp("hotel_stay.date", obj.Attributes["date"], errors),
			LastUpdated:            trySetTimestamp("hotel_stay.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
		}
		nRecs = obj.TotalCount
		hotelStayRecs = append(hotelStayRecs, rec)
	}
	return hotelStayRecs, nRecs
}

func generatePhoneHistoryRecords(profile *profilemodel.Profile, errors *[]string) ([]model.PhoneHistory, int64) {
	phoneHistoryRecs := []model.PhoneHistory{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_PHONE_HISTORY {
			continue
		}
		rec := model.PhoneHistory{
			AccpObjectID:           obj.Attributes["accp_object_id"],
			OverallConfidenceScore: trySetFloat(obj.Attributes["overall_confidence_score"], errors),
			TravellerID:            obj.Attributes["traveller_id"],
			Number:                 obj.Attributes["number"],
			CountryCode:            obj.Attributes["country_code"],
			Type:                   obj.Attributes["type"],
			LastUpdated:            trySetTimestamp("PhoneHistory.last_updated", obj.Attributes["last_updated"], errors),
			LastUpdatedBy:          obj.Attributes["last_updated_by"],
			IsVerified:             trySetBool(obj.Attributes["is_verified"]),
		}
		nRecs = obj.TotalCount
		phoneHistoryRecs = append(phoneHistoryRecs, rec)
	}
	return phoneHistoryRecs, nRecs
}

func generateAlternateProfileIdRecords(profile *profilemodel.Profile, errors *[]string) ([]model.AlternateProfileID, int64) {
	alternateProfileIdRecs := []model.AlternateProfileID{}
	nRecs := int64(0)
	for _, obj := range profile.ProfileObjects {
		if obj.Type != constant.ACCP_RECORD_ALTERNATE_PROFILE_ID {
			continue
		}
		rec := model.AlternateProfileID{
			AccpObjectID: obj.Attributes[constant.ACCP_OBJECT_ID_KEY],
			TravellerID:  obj.Attributes[constant.ACCP_OBJECT_TRAVELLER_ID_KEY],
			Name:         obj.Attributes["name"],
			Description:  obj.Attributes["description"],
			Value:        obj.Attributes["value"],
		}
		nRecs = obj.TotalCount
		alternateProfileIdRecs = append(alternateProfileIdRecs, rec)
	}
	return alternateProfileIdRecs, nRecs
}

func trySetDate(fieldName string, val string, errors *[]string) time.Time {
	if val == "" {
		return time.Time{}
	}
	supportedFormats := []string{"2006-01-02", "20060102", "2006/01/02"}
	for _, format := range supportedFormats {
		parsed, err := utils.TryParseTime(val, format)
		if err == nil {
			return parsed
		}
	}
	*errors = append(*errors, "Unsupported date format for field "+fieldName+" ("+val+"). Supported formats are: "+strings.Join(supportedFormats, ","))
	return time.Time{}
}

func trySetTimestamp(fieldName string, val string, errors *[]string) time.Time {
	if val == "" {
		return time.Time{}
	}
	parsed, err := utils.TryParseTime(val, "2006-01-02T15:04:05.999999Z")
	if err != nil {
		*errors = append(*errors, "Unsupported Timestamp format for field "+fieldName+" ("+val+"). Supported format is 2006-01-02T15:04:05.999999Z")
	}
	return parsed
}

func trySetInt(val string, errors *[]string) int {
	if val == "" {
		return 0
	}
	parsed, err := utils.TryParseInt(val)
	if err != nil {
		*errors = append(*errors, err.Error())
	}
	return parsed
}

func trySetFloat(val string, errors *[]string) float64 {
	if val == "" {
		return 0.0
	}
	parsed, err := utils.TryParseFloat(val)
	if err != nil {
		*errors = append(*errors, err.Error())
	}
	return parsed
}

func trySetBool(val string) bool {
	if val == "true" || val == "True" || val == "TRUE" {
		return true
	}
	return false
}
