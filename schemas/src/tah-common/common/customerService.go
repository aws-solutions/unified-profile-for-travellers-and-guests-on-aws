package common

import (
	"encoding/json"
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

type CustomerServiceInteraction struct {
	ObjectVersion            int                  `json:"objectVersion"`
	ModelVersion             string               `json:"modelVersion"`               //Version of the AWS travel and hospitality model
	SessionID                string               `json:"sessionId"`                  //session ID
	Channel                  string               `json:"channel"`                    //voice, Chat,  email,  sms
	LoyaltyID                string               `json:"loyaltyId,omitempty"`        //loyalty id
	FirstName                string               `json:"firstName,omitempty"`        //first name
	LastName                 string               `json:"lastName,omitempty"`         //last name
	Email                    string               `json:"email,omitempty"`            //email
	PhoneNumber              string               `json:"phoneNumber,omitempty"`      //phone number
	RelatedBookingID         string               `json:"relatedBookingId,omitempty"` //related booking id
	StartTime                time.Time            `json:"startTime"`                  //time stamp
	InteractionType          string               `json:"interactionType,omitempty"`  //interaction type
	Conversation             []ConversationItem   `json:"conversation,omitempty"`     //conversation item. can also be used for email, SMS...
	EndTime                  time.Time            `json:"endTime,omitempty"`          //time stamp
	Duration                 int                  `json:"duration,omitempty"`         //duration
	Status                   string               `json:"status,omitempty"`           //status
	Language                 Language             `json:"language,omitempty"`         //language
	OverallSentimentAgent    core.Float           `json:"overallSentimentAgent,omitempty"`
	OverallSentimentCustomer core.Float           `json:"overallSentimentCustomer,omitempty"`
	TalkTime                 int64                `json:"talkTime,omitempty"`
	NonTalkTime              int64                `json:"nonTalkTime,omitempty"`
	Interruptions            int64                `json:"interruptions,omitempty"`
	Summary                  string               `json:"summary,omitempty"` //conversation summary
	AlternateProfileIDs      []AlternateProfileID `json:"alternateProfileIds,omitempty"`
	CampaignJobID            string               `json:"campaignJobId,omitempty"`
	CampaignStrategy         string               `json:"campaignStrategy,omitempty"`
	CampaignProgram          string               `json:"campaignProgram,omitempty"`
	CampaignProduct          string               `json:"campaignProduct,omitempty"`
	CampaignName             string               `json:"campaignName,omitempty"`
	Category                 string               `json:"category,omitempty"`
	Subject                  string               `json:"subject,omitempty"`
	LoyaltyProgramName       string               `json:"loyaltyProgramName,omitempty"`
	IsVoiceOTP               bool                 `json:"isVoiceOtp,omitempty"`
	ExtendedData             interface{}          `json:"extendedData,omitempty"`
}

type ConversationItem struct {
	From         string      `json:"from,omitempty"`       //customer/agent/bot
	To           string      `json:"to,omitempty"`         //customer/agent/bot
	Content      string      `json:"content"`              //content
	StartTime    time.Time   `json:"startTime"`            //start tim of the
	EndTime      time.Time   `json:"endTime,omitempty"`    //to be used for audio conversation transcript
	Type         string      `json:"type,omitempty"`       //text
	ID           string      `json:"id,omitempty"`         //text
	Attachment   Attachment  `json:"attachment,omitempty"` //text
	Sentiment    string      `json:"sentiment,omitempty"`  //text
	ExtendedData interface{} `json:"extendedData,omitempty"`
}

type Attachment struct {
	Type        string `json:"type,omitempty"`
	Url         string `json:"url,omitempty"`
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Size        string `json:"size,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Content     string `json:"content"`
}

func (gp CustomerServiceInteraction) MarshalJSON() ([]byte, error) {
	type Alias CustomerServiceInteraction
	return json.Marshal(&struct {
		Alias
		EndTime string `json:"endTime"`
	}{
		Alias:   (Alias)(gp),
		EndTime: gp.EndTime.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}
