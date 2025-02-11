package example

import (
	"encoding/json"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/lodging"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateCustomerServiceInteractionGuests(guests []lodging.GuestProfile, nLines int) []common.CustomerServiceInteraction {
	interactions := []common.CustomerServiceInteraction{}
	nEventsPerFile := nLines
	for i := 0; i < nEventsPerFile; i++ {
		interactions = append(interactions, CreateGuestsCustomerServiceInteraction(guests))
	}
	return interactions
}

func GenerateCustomerServiceInteractionPax(paxs []air.PassengerProfile, nLines int) []common.CustomerServiceInteraction {
	interactions := []common.CustomerServiceInteraction{}
	nEventsPerFile := nLines
	for i := 0; i < nEventsPerFile; i++ {
		interactions = append(interactions, CreatePaxCustomerServiceInteraction(paxs))
	}
	return interactions
}

func SerializeCustomerServiceInteractions(interactions []common.CustomerServiceInteraction) ([]byte, string, error) {
	file := ""
	for _, interaction := range interactions {
		bytes, err := json.Marshal(interaction)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func CreateGuestsCustomerServiceInteraction(guests []lodging.GuestProfile) common.CustomerServiceInteraction {
	guest := guests[gofakeit.Number(0, len(guests)-1)]
	loyaltyID := ""
	if len(guest.LoyaltyPrograms) > 0 {
		loyaltyID = guest.LoyaltyPrograms[0].ID
	}
	email := ""
	if len(guest.Emails) > 0 {
		email = guest.Emails[0].Address
	}
	phone := ""
	if len(guest.Phones) > 0 {
		phone = guest.Phones[0].Number
	}
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))

	conversation := BuildChatConversation(createdOn)
	interaction := common.CustomerServiceInteraction{
		ObjectVersion:      1,
		ModelVersion:       "1.0",
		SessionID:          gofakeit.UUID(),
		Channel:            gofakeit.RandomString([]string{"voice", "chat", "email", "sms"}),
		LoyaltyID:          loyaltyID,
		FirstName:          guest.FirstName,
		LastName:           guest.LastName,
		Email:              email,
		PhoneNumber:        phone,
		RelatedBookingID:   gofakeit.Regex("[0-9A-Z]{10}"),
		StartTime:          createdOn,
		InteractionType:    "Customer service call",
		EndTime:            computeEndTime(conversation),
		Conversation:       conversation,
		Status:             "completed",
		Language:           common.Language{Code: gofakeit.LanguageAbbreviation()},
		CampaignJobID:      gofakeit.UUID(),
		CampaignStrategy:   gofakeit.RandomString([]string{"outbound", "inbound", "retention", "acquisition"}),
		CampaignProgram:    gofakeit.RandomString([]string{"loyalty", "promotion", "service", "sales"}),
		CampaignProduct:    gofakeit.RandomString([]string{"flight", "hotel", "package", "insurance"}),
		CampaignName:       gofakeit.Company() + " " + gofakeit.RandomString([]string{"Summer", "Winter", "Spring", "Fall"}) + " Campaign",
		Category:           gofakeit.RandomString([]string{"booking", "loyalty", "complaint", "inquiry", "feedback"}),
		Subject:            gofakeit.RandomString([]string{"Flight Change", "Refund Request", "Loyalty Points", "General Information", "Technical Support"}),
		LoyaltyProgramName: gofakeit.RandomString([]string{"Miles Plus", "Elite Rewards", "Premium Club", "Traveler Benefits"}),
		IsVoiceOTP:         gofakeit.Bool(),
	}
	interaction.Duration = int(interaction.EndTime.Sub(interaction.StartTime).Seconds())

	if len(guest.AlternateProfileIDs) > 0 {
		interaction.AlternateProfileIDs = append(interaction.AlternateProfileIDs, guest.AlternateProfileIDs...)
	}

	return interaction
}

func computeEndTime(conversation []common.ConversationItem) time.Time {
	return conversation[len(conversation)-1].StartTime.Add(time.Duration(gofakeit.Number(1, 15)) * time.Second)
}

func BuildChatConversation(createdOn time.Time) []common.ConversationItem {
	nSegments := gofakeit.Number(1, 50)
	items := []common.ConversationItem{}
	agentOrBot := gofakeit.RandomString([]string{"agent", "bot"})
	startTime := createdOn
	for i := 0; i < nSegments; i++ {
		rqTime := startTime.Add(time.Duration(gofakeit.Number(1, 10)*i) * time.Second)
		request := common.ConversationItem{
			From:      "customer",
			Content:   gofakeit.Sentence(gofakeit.Number(2, 100)),
			StartTime: rqTime,
			Sentiment: gofakeit.RandomString([]string{"POSITIVE", "NEGATIVE", "NEUTRAL"}),
		}
		response := common.ConversationItem{
			From:      agentOrBot,
			Content:   gofakeit.Sentence(gofakeit.Number(2, 100)),
			StartTime: rqTime.Add(time.Duration(gofakeit.Number(10, 30)) * time.Second),
		}
		if agentOrBot == "agent" {
			response.Sentiment = gofakeit.RandomString([]string{"POSITIVE", "NEGATIVE", "NEUTRAL"})
		}
		items = append(items, request)
		items = append(items, response)
	}
	return items
}

func CreatePaxCustomerServiceInteraction(paxes []air.PassengerProfile) common.CustomerServiceInteraction {
	pax := paxes[gofakeit.Number(0, len(paxes)-1)]
	loyaltyID := ""
	if len(pax.LoyaltyPrograms) > 0 {
		loyaltyID = pax.LoyaltyPrograms[0].ID
	}
	email := ""
	if len(pax.Emails) > 0 {
		email = pax.Emails[0].Address
	}
	phone := ""
	if len(pax.Phones) > 0 {
		phone = pax.Phones[0].Number
	}
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))

	conversation := BuildChatConversation(createdOn)
	interaction := common.CustomerServiceInteraction{
		ObjectVersion:    1,
		ModelVersion:     "1.0",
		SessionID:        gofakeit.UUID(),
		Channel:          gofakeit.RandomString([]string{"voice", "chat", "email", "sms"}),
		LoyaltyID:        loyaltyID,
		FirstName:        pax.FirstName,
		LastName:         pax.LastName,
		Email:            email,
		PhoneNumber:      phone,
		RelatedBookingID: gofakeit.Regex("[0-9A-Z]{10}"),
		StartTime:        createdOn,
		InteractionType:  "Customer service call",
		EndTime:          computeEndTime(conversation),
		Conversation:     conversation,
		Status:           "completed",
		Language:         common.Language{Code: gofakeit.LanguageAbbreviation()},
	}
	interaction.Duration = int(interaction.EndTime.Sub(interaction.StartTime).Seconds())
	return interaction
}
