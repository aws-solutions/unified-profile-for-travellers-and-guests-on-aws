// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package profilemodel

import (
	"strings"
	"time"
)

/********************************************************
* Setup
*********************************************************/

type ProfileObject struct {
	ID                  string
	Type                string
	JSONContent         string
	Attributes          map[string]string
	AttributesInterface map[string]interface{}
	TotalCount          int64 //total count of object for the given type(to be used for pagination)
}

type Address struct {
	Address1   string `json:"Address1"`
	Address2   string `json:"Address2"`
	Address3   string `json:"Address3"`
	Address4   string `json:"Address4"`
	City       string `json:"City"`
	State      string `json:"State"`
	Province   string `json:"Province"`
	PostalCode string `json:"PostalCode"`
	Country    string `json:"Country"`
}

func (a Address) Get(fieldName string) string {
	switch fieldName {
	case "Address1":
		return a.Address1
	case "Address2":
		return a.Address2
	case "Address3":
		return a.Address3
	case "Address4":
		return a.Address4
	case "City":
		return a.City
	case "State":
		return a.State
	case "Province":
		return a.Province
	case "PostalCode":
		return a.PostalCode
	case "Country":
		return a.Country
	}
	return ""
}

func (a *Address) Set(fieldName string, value string) {
	switch fieldName {
	case "Address1":
		a.Address1 = value
	case "Address2":
		a.Address2 = value
	case "Address3":
		a.Address3 = value
	case "Address4":
		a.Address4 = value
	case "City":
		a.City = value
	case "State":
		a.State = value
	case "Province":
		a.Province = value
	case "PostalCode":
		a.PostalCode = value
	case "Country":
		a.Country = value
	}
}

/********************************************************
* Profile
*********************************************************/

type Profile struct {
	LastUpdated          time.Time
	Domain               string            `json:"domain"`
	ProfileId            string            `json:"ProfileId"`
	AccountNumber        string            `json:"AccountNumber"`
	FirstName            string            `json:"FirstName"`
	MiddleName           string            `json:"MiddleName"`
	LastName             string            `json:"LastName"`
	BirthDate            string            `json:"BirthDate"`
	Gender               string            `json:"Gender"`
	PhoneNumber          string            `json:"PhoneNumber"`
	MobilePhoneNumber    string            `json:"MobilePhoneNumber"`
	HomePhoneNumber      string            `json:"HomePhoneNumber"`
	BusinessPhoneNumber  string            `json:"BusinessPhoneNumber"`
	PersonalEmailAddress string            `json:"PersonalEmailAddress"`
	BusinessEmailAddress string            `json:"BusinessEmailAddress"`
	EmailAddress         string            `json:"EmailAddress"`
	Attributes           map[string]string `json:"Attributes"`
	Address              Address           `json:"Address"`
	MailingAddress       Address           `json:"MailingAddress"`
	BillingAddress       Address           `json:"BillingAddress"`
	ShippingAddress      Address           `json:"ShippingAddress"`
	BusinessName         string            `json:"BusinessName"`
	ProfileObjects       []ProfileObject
	Matches              []Match
}

func (p Profile) Get(fieldName string) string {
	switch fieldName {
	case "FirstName":
		return p.FirstName
	case "LastName":
		return p.LastName
	case "MiddleName":
		return p.MiddleName
	case "BirthDate":
		return p.BirthDate
	case "Gender":
		return p.Gender
	case "PhoneNumber":
		return p.PhoneNumber
	case "MobilePhoneNumber":
		return p.MobilePhoneNumber
	case "HomePhoneNumber":
		return p.HomePhoneNumber
	case "BusinessPhoneNumber":
		return p.BusinessPhoneNumber
	case "EmailAddress":
		return p.EmailAddress
	case "BusinessEmailAddress":
		return p.BusinessEmailAddress
	case "PersonalEmailAddress":
		return p.PersonalEmailAddress
	}
	if strings.HasPrefix(fieldName, "Attributes.") {
		fieldName := strings.ReplaceAll(fieldName, "Attributes.", "")
		return p.Attributes[fieldName]
	}
	if strings.HasPrefix(fieldName, "Address") {
		addrField := strings.ReplaceAll(fieldName, "Address.", "")
		return p.Address.Get(addrField)
	}
	if strings.HasPrefix(fieldName, "ShippingAddress") {
		addrField := strings.ReplaceAll(fieldName, "ShippingAddress.", "")
		return p.ShippingAddress.Get(addrField)
	}
	if strings.HasPrefix(fieldName, "MailingAddress") {
		addrField := strings.ReplaceAll(fieldName, "MailingAddress.", "")
		return p.MailingAddress.Get(addrField)
	}
	if strings.HasPrefix(fieldName, "BillingAddress") {
		addrField := strings.ReplaceAll(fieldName, "BillingAddress.", "")
		return p.BillingAddress.Get(addrField)
	}
	return ""
}

func (p *Profile) Set(fieldName string, value string) {
	switch fieldName {
	case "FirstName":
		p.FirstName = value
	case "LastName":
		p.LastName = value
	case "MiddleName":
		p.MiddleName = value
	case "BirthDate":
		p.BirthDate = value
	case "Gender":
		p.Gender = value
	case "PhoneNumber":
		p.PhoneNumber = value
	case "MobilePhoneNumber":
		p.MobilePhoneNumber = value
	case "HomePhoneNumber":
		p.HomePhoneNumber = value
	case "BusinessPhoneNumber":
		p.BusinessPhoneNumber = value
	case "EmailAddress":
		p.EmailAddress = value
	case "BusinessEmailAddress":
		p.BusinessEmailAddress = value
	case "PersonalEmailAddress":
		p.PersonalEmailAddress = value
	}
	if strings.HasPrefix(fieldName, "Attributes.") {
		fieldName := strings.ReplaceAll(fieldName, "Attributes.", "")
		if p.Attributes == nil {
			p.Attributes = make(map[string]string)
		}
		p.Attributes[fieldName] = value
	}
	if strings.HasPrefix(fieldName, "Address") {
		addrField := strings.ReplaceAll(fieldName, "Address.", "")
		p.Address.Set(addrField, value)
	}
	if strings.HasPrefix(fieldName, "ShippingAddress") {
		addrField := strings.ReplaceAll(fieldName, "ShippingAddress.", "")
		p.ShippingAddress.Set(addrField, value)
	}
	if strings.HasPrefix(fieldName, "MailingAddress") {
		addrField := strings.ReplaceAll(fieldName, "MailingAddress.", "")
		p.MailingAddress.Set(addrField, value)
	}
	if strings.HasPrefix(fieldName, "BillingAddress") {
		addrField := strings.ReplaceAll(fieldName, "BillingAddress.", "")
		p.BillingAddress.Set(addrField, value)
	}
}

type Match struct {
	ConfidenceScore float64
	ProfileID       string
	FirstName       string
	LastName        string
	BirthDate       string
	PhoneNumber     string
	EmailAddress    string
}
