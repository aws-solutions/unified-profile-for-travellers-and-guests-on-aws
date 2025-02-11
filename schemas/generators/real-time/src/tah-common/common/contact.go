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
	Type    string `json:"type"`
	Address string `json:"address"`
	Primary bool   `json:"primary"`
}
type Phone struct {
	Type        string `json:"type"`
	Number      string `json:"number"`
	Primary     bool   `json:"primary"`
	CountryCode int64  `json:"countryCode"`
}
type Address struct {
	Type       string        `json:"type"`
	Line1      string        `json:"line1"`
	Line2      string        `json:"line2"`
	Line3      string        `json:"line3"`
	Line4      string        `json:"line4"`
	City       string        `json:"city"`
	State      StateProvince `json:"state"`
	Province   StateProvince `json:"province"`
	PostalCode string        `json:"postalCode"`
	Country    Country       `json:"country"`
	Primary    bool          `json:"primary"`
}
type StateProvince struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
