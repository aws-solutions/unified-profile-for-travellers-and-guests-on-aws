package common

var PHONE_TYPE_PERSONAL = "personal"
var PHONE_TYPE_BUSINESS = "business"
var PHONE_TYPE_HOME = "home"
var PHONE_TYPE_MOBILE = "mobile"

var EMAIL_TYPE_PERSONAL = "personal"
var EMAIL_TYPE_BUSINESS = "business"

var ADDRESS_TYPE_BUSINESS = "business"
var ADDRESS_TYPE_MAILING = "mailing"
var ADDRESS_TYPE_SHIPPING = "shipping"

var HONORIFIC_MR = "Mr"
var HONORIFIC_MX = "Mx"
var HONORIFIC_SIR = "Sir"
var HONORIFIC_DR = "Dr"
var HONORIFIC_CLLR = "Cllr"
var HONORIFIC_LORD = "Lord"

var HONORIFIC_MRS = "Mrs"
var HONORIFIC_MISS = "Miss"
var HONORIFIC_MS = "Ms"
var HONORIFIC_DAME = "Dame"
var HONORIFIC_LADY = "Lady"

type Email struct {
	Type         string      `json:"type,omitempty"`
	Address      string      `json:"address"`
	Primary      bool        `json:"primary,omitempty"`
	IsVerified   bool        `json:"isVerified,omitempty"`
	ExtendedData interface{} `json:"extendedData,omitempty"`
}
type Phone struct {
	Type         string      `json:"type,omitempty"`
	Number       string      `json:"number"`
	Primary      bool        `json:"primary,omitempty"`
	CountryCode  int64       `json:"countryCode,omitempty"`
	IsVerified   bool        `json:"isVerified,omitempty"`
	ExtendedData interface{} `json:"extendedData,omitempty"`
}
type Address struct {
	Type       string        `json:"type,omitempty"`
	Line1      string        `json:"line1,omitempty"`
	Line2      string        `json:"line2,omitempty"`
	Line3      string        `json:"line3,omitempty"`
	Line4      string        `json:"line4,omitempty"`
	City       string        `json:"city,omitempty"`
	State      StateProvince `json:"state,omitempty"`
	Province   StateProvince `json:"province,omitempty"`
	PostalCode string        `json:"postalCode,omitempty"`
	Country    Country       `json:"country,omitempty"`
	Primary    bool          `json:"primary"`
}
type StateProvince struct {
	Code string `json:"code"`
	Name string `json:"name,omitempty"`
}
type Country struct {
	Code string `json:"code"`
	Name string `json:"name,omitempty"`
}
