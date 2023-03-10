package traveller

import (
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/traveller"
	"tah/ucp/src/business-logic/utils"
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

func profilesToTravellers(profiles []customerprofiles.Profile) []model.Traveller {
	travellers := []model.Traveller{}
	for _, profile := range profiles {
		travellers = append(travellers, profileToTraveller(profile))
	}
	return travellers
}
func profileToTraveller(profile customerprofiles.Profile) model.Traveller {
	tv := model.Traveller{
		ID:         profile.ProfileId,
		FirstName:  profile.FirstName,
		MiddleName: profile.MiddleName,
		LastName:   profile.LastName,
		Phones:     []model.TraceableString{},
		Emails:     []model.TraceableString{},
		JobTitles:  []model.TraceableString{},
		BirthDate:  utils.ParseTime(profile.BirthDate, "2006-01-02"),
		Gender:     profile.Gender,
		Title:      profile.Attributes["honorific"],
	}
	if profile.PhoneNumber != "" {
		tv.Phones = append(tv.Phones, model.TraceableString{
			Value: profile.PhoneNumber,
		})
	}
	if profile.MobilePhoneNumber != "" {
		tv.Phones = append(tv.Phones, model.TraceableString{
			Value: profile.PhoneNumber,
		})
	}
	if profile.BusinessPhoneNumber != "" {
		tv.Phones = append(tv.Phones, model.TraceableString{
			Value: profile.PhoneNumber,
		})
	}
	if profile.EmailAddress != "" {
		tv.Emails = append(tv.Emails, model.TraceableString{
			Value: profile.EmailAddress,
		})
	}
	if profile.PersonalEmailAddress != "" {
		tv.Emails = append(tv.Emails, model.TraceableString{
			Value: profile.PersonalEmailAddress,
		})
	}
	if profile.BusinessEmailAddress != "" {
		tv.Emails = append(tv.Emails, model.TraceableString{
			Value: profile.BusinessEmailAddress,
		})
	}
	if profile.Attributes["jobTitle"] != "" {
		tv.JobTitles = append(tv.JobTitles, model.TraceableString{
			Value: profile.Attributes["jobTitle"],
		})
	}
	tv.HotelBookings = parseHotelBookingFromOrders(profile.Orders)
	tv.LoyaltyProfiles = parseLoyaltyProfilesFromOrders(profile.Orders)
	tv.Searches = parseSeachesFromOrders(profile.Orders)
	return tv
}

func parseSeachesFromOrders(orders []customerprofiles.Order) []model.Search {
	searches := []model.Search{}
	for _, order := range orders {
		if order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_SESSION_ID] != "" {
			search := model.Search{
				Date:        order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_START_DATE],
				Origin:      order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_ORIGIN],
				Destination: order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_START_DATE],
				Hotel:       order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_HOTEL],
				Sailing:     order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_SAILING],
			}
			searches = append(searches, search)

		}
	}
	return searches
}

func parseLoyaltyProfilesFromOrders(orders []customerprofiles.Order) []model.LoyaltyProfileSummary {
	profiles := []model.LoyaltyProfileSummary{}
	for _, order := range orders {
		if order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_CONFIRMATION_NUMBER] == "" &&
			order.Attributes[ATTRIBUTE_KEY_CLICKSTREAM_SESSION_ID] == "" &&
			order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_LOYALTY_ID] != "" {
			profile := model.LoyaltyProfileSummary{
				ID:      order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_LOYALTY_ID],
				Status:  order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_STATUS],
				Points:  utils.ParseFloat64(order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_POINTS]),
				Program: order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_PROGRAM],
				Joined:  utils.ParseTime(order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_JOINED], "2006-01-02"),
			}
			profiles = append(profiles, profile)

		}
	}
	return profiles
}

func parseHotelBookingFromOrders(orders []customerprofiles.Order) []model.HotelBookingSummary {
	bookings := []model.HotelBookingSummary{}
	for _, order := range orders {
		if order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_CONFIRMATION_NUMBER] != "" {
			booking := model.HotelBookingSummary{
				ID:         order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_CONFIRMATION_NUMBER],
				StartDate:  order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_START_DATE],
				TotalPrice: utils.ParseFloat64(order.TotalPrice),
				//To be added or configured in enrichments phase
				Currency:  "USD",
				Products:  order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_PRODUCTS],
				NNight:    utils.ParseInt64(order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_N_NIGHTS]),
				HotelCode: order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_HOTEL_CODE],
				//To be fixed. should be custome attribute
				Channel: order.Name,
				NGuests: utils.ParseInt64(order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_N_GUESTS]),
				Source:  "to be added",
			}
			bookings = append(bookings, booking)
		}

	}
	return bookings
}
