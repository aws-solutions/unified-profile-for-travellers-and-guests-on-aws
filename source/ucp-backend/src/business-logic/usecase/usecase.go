package usecase

import (
	"log"
	"strconv"
	customerprofiles "tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model"
	accpmappings "tah/ucp/src/business-logic/model/accp-mappings"
	"time"
)

var AVAIL_DB_TIMEFORMAT string = "20060102"

// Key Names for Business Objects
const HOTEL_BOOKING string = "hotel_booking"
const HOTEL_STAY_REVENUE string = "hotel_stay_revenue"
const CLICKSTREAM string = "clickstream"
const AIR_BOOKING string = "air_booking"
const GUEST_PROFILE string = "guest_profile"
const PASSENGER_PROFILE string = "passenger_profile"

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

var SEARCH_KEY_LAST_NAME = "LastName"
var SEARCH_KEY_FIRST_NAME = "FirstName"
var SEARCH_KEY_EMAIL = "PersonalEmailAddress"
var SEARCH_KEY_PHONE = "PhoneNumber"
var SEARCH_KEY_ACCOUNT_NUMBER = "AccountNumber"
var SEARCH_KEY_CONF_NUMBER = "confirmationNumber"

var DOMAIN_TAG_ENV_NAME = "envName"

func RetreiveUCPProfile(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	profile, err := profilesSvc.GetProfile(rq.ID)
	return model.ResWrapper{Profiles: []model.Traveller{profileToTraveller(profile)}, Matches: profileToMatches(profile)}, err
}

func DeleteUCPProfile(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) error {
	return profilesSvc.DeleteProfile(rq.ID)
}

func SearchUCPProfile(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	profiles := []customerprofiles.Profile{}
	var err error
	if rq.SearchRq.LastName != "" {
		profiles, err = profilesSvc.SearchProfiles(SEARCH_KEY_LAST_NAME, []string{rq.SearchRq.LastName})
	}
	if rq.SearchRq.LoyaltyID != "" {
		profiles, err = profilesSvc.SearchProfiles(SEARCH_KEY_ACCOUNT_NUMBER, []string{rq.SearchRq.LoyaltyID})
	}
	if rq.SearchRq.Phone != "" {
		profiles, err = profilesSvc.SearchProfiles(SEARCH_KEY_PHONE, []string{rq.SearchRq.Phone})
	}
	if rq.SearchRq.Email != "" {
		profiles, err = profilesSvc.SearchProfiles(SEARCH_KEY_EMAIL, []string{rq.SearchRq.Email})
	}
	if err != nil {
		return model.ResWrapper{}, err
	}
	return model.ResWrapper{Profiles: profilesToTravellers(profiles)}, nil
}

func profileToMatches(profile customerprofiles.Profile) []model.Match {
	matches := []model.Match{}
	for _, m := range profile.Matches {
		matches = append(matches, model.Match{
			ConfidenceScore: m.ConfidenceScore,
			ID:              m.ProfileID,
			FirstName:       m.FirstName,
			LastName:        m.LastName,
			BirthDate:       m.BirthDate,
			PhoneNumber:     m.PhoneNumber,
			EmailAddress:    m.EmailAddress,
		})
	}
	//Demo stub. to remove
	matches = append(matches, model.Match{
		ConfidenceScore: 0.99,
		ID:              "4caf3d602b84460db31650159dcff896",
		FirstName:       "Geoffroy",
		LastName:        "Rollat",
	})

	return matches
}

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
		BirthDate:  parseTime(profile.BirthDate, "2006-01-02"),
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
				Points:  parseFloat64(order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_POINTS]),
				Program: order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_PROGRAM],
				Joined:  parseTime(order.Attributes[ATTRIBUTE_KEY_LOYALTY_PROFILE_JOINED], "2006-01-02"),
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
				TotalPrice: parseFloat64(order.TotalPrice),
				//To be added or configured in enrichments phase
				Currency:  "USD",
				Products:  order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_PRODUCTS],
				NNight:    parseInt64(order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_N_NIGHTS]),
				HotelCode: order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_HOTEL_CODE],
				//To be fixed. should be custome attribute
				Channel: order.Name,
				NGuests: parseInt64(order.Attributes[ATTRIBUTE_KEY_HOTEL_BOOKING_N_GUESTS]),
				Source:  "to be added",
			}
			bookings = append(bookings, booking)
		}

	}
	return bookings
}

func parseFloat64(val string) float64 {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing float: %s", err)
		return 0
	}
	return parsed
}

func parseInt64(val string) int64 {
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing int: %s", err)
		return 0
	}
	return parsed
}

func parseTime(t string, layout string) time.Time {
	parsed, err := time.Parse(layout, t)
	if err != nil {
		log.Printf("[WARNING] error while parsing date: %s", err)
		return time.Time{}
	}
	return parsed
}

func RetreiveUCPConfig(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	dom, err := profilesSvc.GetDomain()
	if err != nil {
		return model.ResWrapper{}, err
	}
	domain := model.Domain{
		Name:            dom.Name,
		NObjects:        dom.NObjects,
		NProfiles:       dom.NProfiles,
		MatchingEnabled: dom.MatchingEnabled}
	mappings, err2 := profilesSvc.GetMappings()
	if err2 != nil {
		return model.ResWrapper{}, err2
	}
	domain.Mappings = parseMappings(mappings)

	integration, err3 := profilesSvc.GetIntegrations()
	if err3 != nil {
		return model.ResWrapper{}, err3
	}
	domain.Integrations = parseIntegrations(integration)

	return model.ResWrapper{UCPConfig: model.UCPConfig{Domains: []model.Domain{domain}}}, err
}

func parseIntegrations(profileIntegrations []customerprofiles.Integration) []model.Integration {
	integrations := []model.Integration{}
	for _, pi := range profileIntegrations {
		integrations = append(integrations, model.Integration{
			Source:         pi.Source,
			Target:         pi.Target,
			Status:         pi.Status,
			StatusMessage:  pi.StatusMessage,
			LastRun:        pi.LastRun,
			LastRunStatus:  pi.LastRunStatus,
			LastRunMessage: pi.LastRunMessage,
			Trigger:        pi.Trigger,
		})
	}
	return integrations
}

func ListUcpDomains(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	profileDomains, err := profilesSvc.ListDomains()
	if err != nil {
		return model.ResWrapper{}, err
	}
	domains := []model.Domain{}
	for _, dom := range profileDomains {
		if rq.EnvName == dom.Tags[DOMAIN_TAG_ENV_NAME] {
			domains = append(domains, model.Domain{
				Name:        dom.Name,
				Created:     dom.Created,
				LastUpdated: dom.LastUpdated,
			})
		}
	}
	return model.ResWrapper{UCPConfig: model.UCPConfig{Domains: domains}}, err
}

func CreateUcpDomain(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig, KMS_KEY_PROFILE_DOMAIN string, CONNECT_PROFILE_SOURCE_BUCKET string) (model.ResWrapper, error) {
	err := profilesSvc.CreateDomain(rq.Domain.Name, true, KMS_KEY_PROFILE_DOMAIN, map[string]string{DOMAIN_TAG_ENV_NAME: rq.EnvName})
	if err != nil {
		return model.ResWrapper{}, err
	}

	businessMap := map[string]func() []customerprofiles.FieldMapping{
		HOTEL_BOOKING:      accpmappings.BuildHotelBookingMapping,
		HOTEL_STAY_REVENUE: accpmappings.BuildHotelStayMapping,
		CLICKSTREAM:        accpmappings.BuildClickstreamMapping,
		AIR_BOOKING:        accpmappings.BuildAirBookingMapping,
		GUEST_PROFILE:      accpmappings.BuildGuestProfileMapping,
		PASSENGER_PROFILE:  accpmappings.BuildPassengerProfileMapping,
	}

	for keyBusiness := range businessMap {
		err = profilesSvc.CreateMapping(keyBusiness,
			"Primary Mapping for the "+keyBusiness+" object", businessMap[keyBusiness]())
		if err != nil {
			log.Printf("[CreateUcpDomain] Error creating Mapping: %s. deleting domain", err)
			err2 := profilesSvc.DeleteDomain()
			if err2 != nil {
				log.Printf("[CreateUcpDomain][warning] Error cleaning up domain after failed mapping creation %v", err2)
			}
			return model.ResWrapper{}, err
		}
		_, err4 := profilesSvc.PutIntegration(keyBusiness, CONNECT_PROFILE_SOURCE_BUCKET)
		if err4 != nil {
			log.Printf("Error creating integration %s", err4)
			return model.ResWrapper{}, err4
		}
	}
	return model.ResWrapper{}, err
}

func DeleteUcpDomain(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	err := profilesSvc.DeleteDomain()
	return model.ResWrapper{}, err
}

func parseMappings(profileMappings []customerprofiles.ObjectMapping) []model.ObjectMapping {
	modelMappings := []model.ObjectMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, parseMapping(profileMapping))
	}
	return modelMappings
}
func parseMapping(profileMapping customerprofiles.ObjectMapping) model.ObjectMapping {
	return model.ObjectMapping{
		Name:   profileMapping.Name,
		Fields: parseFieldMappings(profileMapping.Fields),
	}
}
func parseFieldMappings(profileMappings []customerprofiles.FieldMapping) []model.FieldMapping {
	modelMappings := []model.FieldMapping{}
	for _, profileMapping := range profileMappings {
		modelMappings = append(modelMappings, parseFieldMapping(profileMapping))
	}
	return modelMappings
}
func parseFieldMapping(profileMappings customerprofiles.FieldMapping) model.FieldMapping {
	return model.FieldMapping{
		Type:   profileMappings.Type,
		Source: profileMappings.Source,
		Target: profileMappings.Target,
	}
}

func MergeUCPConfig(rq model.UCPRequest) (model.ResWrapper, error) {
	return model.ResWrapper{}, nil
}
func ListUCPIngestionError(rq model.UCPRequest, profilesSvc customerprofiles.CustomerProfileConfig) (model.ResWrapper, error) {
	errs, totalErrors, err := profilesSvc.GetErrors()
	if err != nil {
		return model.ResWrapper{}, err
	}
	ingErrors := []model.IngestionErrors{}
	for _, ingErr := range errs {
		ingErrors = append(ingErrors, model.IngestionErrors{
			Reason:  ingErr.Reason,
			Message: ingErr.Message,
		})
	}
	return model.ResWrapper{IngestionErrors: ingErrors, TotalErrors: totalErrors}, nil
}
