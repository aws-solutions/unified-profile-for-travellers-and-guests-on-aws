// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package traveller

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strconv"
	"strings"
	customerprofiles "tah/upt/source/storage"
	"tah/upt/source/tah-core/core"
	constants "tah/upt/source/ucp-backend/src/business-logic/constants"
	model "tah/upt/source/ucp-backend/src/business-logic/model/common"
	common "tah/upt/source/ucp-backend/src/business-logic/usecase/admin"
	"tah/upt/source/ucp-backend/src/business-logic/usecase/registry"
	"tah/upt/source/ucp-common/src/constant/admin"
	travModel "tah/upt/source/ucp-common/src/model/traveler"
	travSvc "tah/upt/source/ucp-common/src/services/traveller"
	utils "tah/upt/source/ucp-common/src/utils/utils"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/events"
)

type GetProfileSummary struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetProfileSummary() *GetProfileSummary {
	return &GetProfileSummary{name: "GetProfileSummary"}
}

func (u *GetProfileSummary) Name() string {
	return u.name
}
func (u *GetProfileSummary) Tx() core.Transaction {
	return *u.tx
}
func (u *GetProfileSummary) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetProfileSummary) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetProfileSummary) Registry() *registry.Registry {
	return u.reg
}

func (u *GetProfileSummary) AccessPermission() admin.AppPermission {
	return admin.PublicAccessPermission
}

func (u *GetProfileSummary) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw := model.RequestWrapper{
		ID: req.PathParameters["id"],
	}
	return rw, nil
}

func (u *GetProfileSummary) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Debug("[GetProfileSummary] Validating request")
	if u.reg.Env["ACCP_DOMAIN_NAME"] == "" {
		return errors.New("no domain found in context, cannot update promptConfig")
	}
	if rq.ID == "" {
		return errors.New("[GetProfileSummary] no id provided, cannot generate profile summary")
	}
	if !utils.IsUUID(rq.ID) {
		return errors.New("invalid Connect ID format")
	}
	return nil
}

func (u *GetProfileSummary) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func (u *GetProfileSummary) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	u.tx.Info("[GetProfileSummary] getting summary for profile: ", req.ID)

	promptConfig := []model.DynamoDomainConfig{}

	err := u.reg.PortalConfigDB.FindStartingWith(common.PORTAL_CONFIG_PROMPT_PK, u.reg.Env["ACCP_DOMAIN_NAME"], &promptConfig)
	if err != nil || len(promptConfig) != 1 {
		u.tx.Error("[GetProfileSummary] error getting prompt configuration, error: ", err)
		return model.ResponseWrapper{}, err
	}

	if !promptConfig[0].IsActive {
		u.tx.Error("[GetProfileSummary] summary generation has been turned off")
		return model.ResponseWrapper{}, errors.New("summary generation has been turned off")
	}

	promptFragment, err := buildPromptForSummary(u, promptConfig[0].Value, req.ID)
	if err != nil {
		u.tx.Error("[GetProfileSummary] error building prompt, error: ", err)
		return model.ResponseWrapper{}, err
	}

	answer, _, err := u.reg.Bedrock.InvokeClaudeModel(promptFragment, constants.BEDROCK_TEMPERATURE, constants.BEDROCK_TOP_P, constants.BEDROCK_TOP_K, constants.BEDROCK_MAX_TOKENS_TO_SAMPLE)
	if err != nil {
		u.tx.Error("[GetProfileSummary] error generating summary: ", err)
		return model.ResponseWrapper{}, err
	}

	return model.ResponseWrapper{ProfileSummary: &answer}, nil
}

func buildPromptForSummary(u *GetProfileSummary, prompt, id string) (string, error) {
	traveller, err := travSvc.RetreiveTraveller(*u.tx, u.reg.Accp, id, []customerprofiles.PaginationOptions{})
	if err != nil {
		u.tx.Error("[GetProfileSummary] error getting information for traveller ", id, ", error: ", err)
		return "", err
	}

	simpleTraveler := TravelerToSimpleTraveler(traveller)
	json, err := json.Marshal(simpleTraveler)
	if err != nil {
		u.tx.Error("[GetProfileSummary] error marshalling object, error: ", err)
		return "", err
	}

	return prompt + constants.PROMPT_REQUIREMENTS + string(json), nil
}

func TravelerToSimpleTraveler(traveler travModel.Traveller) travModel.SimpleTraveler {
	log.Printf("TravelerToSimpleTraveler")
	interactions := []travModel.Interaction{}
	recentEventCount := 100
	for _, rec := range traveler.AirBookingRecords[:min(recentEventCount, len(traveler.AirBookingRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type: "Air Booking",
			Date: rec.LastUpdated,
			Description: ` Flight ` + rec.FlightNumber + ` from ` + rec.From + ` to ` + rec.To + ` on
			` + rec.DepartureDate + ` at ` + rec.DepartureTime + ` purchased for ` + fmt.Sprintf("%v", rec.TotalPrice) + ` USD`,
		})
	}
	for _, rec := range traveler.AirLoyaltyRecords[:min(recentEventCount, len(traveler.AirLoyaltyRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type:        "Loyalty Profile (Air)",
			Date:        rec.LastUpdated,
			Description: `Loyalty profile for program ` + rec.ProgramName + ` with status ` + rec.Level + ` and ` + fmt.Sprintf("%v", rec.Miles) + ` (` + fmt.Sprintf("%v", rec.MilesToNextLevel) + ` to next level)`,
		})
	}
	for _, rec := range traveler.HotelLoyaltyRecords[:min(recentEventCount, len(traveler.HotelLoyaltyRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type:        "LoyaltyProfile (Hotel)",
			Date:        rec.LastUpdated,
			Description: `Loyalty profile for program ` + rec.ProgramName + ` with status ` + rec.Level + ` and ` + fmt.Sprintf("%v", rec.Points) + ` (` + fmt.Sprintf("%v", rec.PointsToNextLevel) + ` to next level)`,
		})
	}
	for _, rec := range traveler.HotelStayRecords[:min(recentEventCount, len(traveler.HotelStayRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type:        "Hotel Charge",
			Date:        rec.LastUpdated,
			Description: `Puchase for ` + rec.Type + `(` + rec.Description + `) on ` + rec.Date.Format("2006-01-02") + ` for ` + fmt.Sprintf("%v", rec.Amount) + ` ` + rec.CurrencyCode,
		})
	}
	for _, rec := range traveler.HotelBookingRecords[:min(recentEventCount, len(traveler.HotelBookingRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type:        "Hotel Booking",
			Date:        rec.LastUpdated,
			Description: ` Booked ` + fmt.Sprintf("%v", rec.NumNights) + ` nights at ` + rec.HotelCode + ` from ` + rec.CheckInDate.Format("2006-01-02") + ` for ` + fmt.Sprintf("%v", rec.NumGuests) + ` guests for ` + fmt.Sprintf("%v", rec.TotalSegmentAfterTax) + ` in local currency`,
		})
	}

	clickstreamRecords := getRecentInteractions(traveler.ClickstreamRecords, 500)
	for _, rec := range clickstreamRecords {
		csAsString := filterNullClickstreamAttributes(rec)
		interactions = append(interactions, travModel.Interaction{
			Type:        "Web and Mobile Event",
			Date:        rec.LastUpdated,
			Description: rec.EventType + ` event from Web channel. Here is the clickstream data: ` + csAsString,
		})
	}

	for _, rec := range traveler.AncillaryServiceRecords[:min(recentEventCount, len(traveler.AncillaryServiceRecords))] {
		interactions = append(interactions, travModel.Interaction{
			Type:        "Ancillary Service Purchase",
			Date:        rec.LastUpdated,
			Description: `Ancillary Purchase of type ` + rec.AncillaryType + ` for ` + fmt.Sprintf("%v", rec.Price) + ` ` + rec.Currency,
		})
	}

	for _, rec := range traveler.CustomerServiceInteractionRecords[:min(recentEventCount, len(traveler.CustomerServiceInteractionRecords))] {
		duration := rec.EndTime.Sub(rec.StartTime)
		description := `Conversation of type` + rec.InteractionType + ` lasted for ` + fmt.Sprintf("%.2f", duration.Seconds()) + `s. It had a sentiment score of ` + strconv.FormatFloat(rec.SentimentScore, 'f', 3, 64)
		parsedConversation, err := parseConversation(rec.Conversation)
		if err == nil {
			description += `. Here is the actual conversation: ` + parsedConversation
		}
		interactions = append(interactions, travModel.Interaction{
			Type:        "Customer Service Conversation",
			Date:        rec.LastUpdated,
			Description: description,
		})
	}

	simpleTraveler := travModel.SimpleTraveler{
		// Metadata
		ModelVersion: "1",
		LastUpdated:  traveler.LastUpdated,

		// Profile IDs
		ConnectID:   traveler.ConnectID,
		Domain:      traveler.Domain,
		TravellerID: traveler.TravellerID,
		PSSID:       traveler.PSSID,
		GDSID:       traveler.GDSID,
		PMSID:       traveler.PMSID,
		CRSID:       traveler.CRSID,

		// Profile Data
		Honorific:            traveler.Honorific,
		FirstName:            traveler.FirstName,
		MiddleName:           traveler.MiddleName,
		LastName:             traveler.LastName,
		Gender:               traveler.Gender,
		Pronoun:              traveler.Pronoun,
		BirthDate:            traveler.BirthDate,
		JobTitle:             traveler.JobTitle,
		CompanyName:          traveler.CompanyName,
		PhoneNumber:          traveler.PhoneNumber,
		MobilePhoneNumber:    traveler.MobilePhoneNumber,
		HomePhoneNumber:      traveler.HomePhoneNumber,
		BusinessPhoneNumber:  traveler.BusinessPhoneNumber,
		PersonalEmailAddress: traveler.PersonalEmailAddress,
		BusinessEmailAddress: traveler.BusinessEmailAddress,
		NationalityCode:      traveler.NationalityCode,
		NationalityName:      traveler.NationalityName,
		LanguageCode:         traveler.LanguageCode,
		LanguageName:         traveler.LanguageName,

		HomeAddress:     traveler.HomeAddress,
		BusinessAddress: traveler.BusinessAddress,
		MailingAddress:  traveler.MailingAddress,
		BillingAddress:  traveler.BillingAddress,

		Interactions: interactions,
	}
	return simpleTraveler
}

func parseConversation(conversation string) (string, error) {
	var decoded []byte
	var err error

	// Attempt to decode as base64
	decoded, err = base64.StdEncoding.DecodeString(conversation)
	if err == nil {
		return string(decoded), nil
	}

	log.Printf("error decoding conversation as base64, attempting to decode as base32")
	decoded, err = base32.StdEncoding.DecodeString(conversation)
	if err == nil {
		return string(decoded), nil
	}

	// Check if the input string is valid UTF-8
	if utf8.ValidString(conversation) {
		return conversation, nil
	}

	return "", fmt.Errorf("input string is not valid UTF-8 and could not be decoded as base64 or base32")
}

func filterNullClickstreamAttributes(cs travModel.Clickstream) string {
	v := reflect.ValueOf(cs)
	t := v.Type()

	var result strings.Builder

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i).Type
		fieldName := t.Field(i).Name

		switch fieldType.Kind() {
		case reflect.String:
			if field.String() != "" && field.String() != "<nil>" {
				result.WriteString(fmt.Sprintf("%s: %s, ", fieldName, field.String()))
			}
		case reflect.Float64:
			if field.Float() != 0.0 {
				result.WriteString(fmt.Sprintf("%s: %.2f, ", fieldName, field.Float()))
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				result.WriteString(fmt.Sprintf("%s: %d, ", fieldName, field.Int()))
			}
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Time{}) {
				if !field.Interface().(time.Time).IsZero() {
					result.WriteString(fmt.Sprintf("%s: %s, ", fieldName, field.Interface().(time.Time).Format(time.RFC3339)))
				}
			} else {
				result.WriteString(fmt.Sprintf("%s: %v, ", fieldName, field))
			}
		default:
			result.WriteString(fmt.Sprintf("%s: %v, ", fieldName, field))
		}
	}

	return strings.TrimSuffix(result.String(), ", ")
}

func getRecentInteractions(cs []travModel.Clickstream, n int) []travModel.Clickstream {
	slices.SortFunc(cs, func(a, b travModel.Clickstream) int {
		if a.EventTimestamp.Before(b.EventTimestamp) {
			return 1
		} else if a.EventTimestamp.After(b.EventTimestamp) {
			return -1
		}
		return 0
	})

	if len(cs) < n {
		n = len(cs)
	}
	return cs[:n]
}
