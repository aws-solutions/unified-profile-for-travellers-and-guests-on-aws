package common

import (
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

type CustomerServiceInteraction struct {
	ObjectVersion            int                `json:"objectVersion"`
	ModelVersion             string             `json:"modelVersion"`     //Version of the AWS travel and hospitality model
	SessionID                string             `json:"sessionId"`        //session ID
	Channel                  string             `json:"channel"`          //voice, Chat,  email,  sms
	LoyaltyID                string             `json:"loyaltyId"`        //loyalty id
	FirstName                string             `json:"firstName"`        //first name
	LastName                 string             `json:"lastName"`         //last name
	Email                    string             `json:"email"`            //email
	PhoneNumber              string             `json:"phoneNumber"`      //phone number
	RelatedBookingID         string             `json:"relatedBookingId"` //related booking id
	StartTime                time.Time          `json:"startTime"`        //time stamp
	InteractionType          string             `json:"interactionType"`  //interaction type
	Conversation             []ConversationItem `json:"conversation"`     //conversation
	EndTime                  time.Time          `json:"endTime"`          //time stamp
	Duration                 int                `json:"duration"`         //duration
	Status                   string             `json:"status"`           //status
	Language                 Language           `json:"language"`         //language
	OverallSentimentAgent    core.Float         `json:"overallSentimentAgent"`
	OverallSentimentCustomer core.Float         `json:"overallSentimentCustomer"`
	TalkTime                 int64              `json:"talkTime"`
	NonTalkTime              int64              `json:"nonTalkTime"`
	Interruptions            int64              `json:"interruptions"`
	Summary                  string             `json:"summary"` //conversation summary
}

type ConversationItem struct {
	From       string     `json:"from"`       //customer/agent/bot
	To         string     `json:"to"`         //customer/agent/bot
	Content    string     `json:"content"`    //content
	StartTime  time.Time  `json:"startTime"`  //start tim of the
	EndTime    time.Time  `json:"endTime"`    //to be used for audio conversation stranscript
	Type       string     `json:"type"`       //text
	ID         string     `json:"id"`         //text
	Attachment Attachment `json:"attachment"` //text
	Sentiment  string     `json:"sentiment"`  //text
}

type Attachment struct {
	Type        string `json:"type"`
	Url         string `json:"url"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}
