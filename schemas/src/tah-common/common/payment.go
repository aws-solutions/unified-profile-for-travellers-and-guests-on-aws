package common

var PAYMENT_TYPE_CREDIT_CARD = "credit_card"
var PAYMENT_TYPE_BANK_ACCOUNT = "bank_account"

// Details of a direct billing arrangement.
var PAYMENT_TYPE_DIRECT_BILL = "direct_bill"
var PAYMENT_TYPE_VOUCHER = "voucher"
var PAYMENT_TYPE_LOYALTY_REDEMPTION = "loyalty_redemption"

// A miscellaneous charges order (MCO), also known as miscellaneous purpose document (MPD), is an accountable IATA document, similar to an old-style airline ticket, used to process the payment of travel arrangements
var PAYMENT_TYPE_MCO = "miscellaneous_charges_order"
var PAYMENT_TYPE_CASH = "cash"

type Currency struct {
	Code   string `json:"code"`
	Name   string `json:"name,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}

type PaymentInformation struct {
	PaymentType string                `json:"paymentType"`
	CcInfo      CreditCardInformation `json:"ccInfo,omitempty"`
	AchRouting  string                `json:"routingNumber,omitempty"`
	AchAccount  string                `json:"accountNumber,omitempty"`
	VoucherID   string                `json:"voucherID,omitempty"`
	Address     Address               `json:"address,omitempty"`
}

type CreditCardInformation struct {
	Token    string  `json:"token"`
	CardType string  `json:"cardType,omitempty"`
	CardExp  string  `json:"cardExp"`
	CardCvv  string  `json:"cardCvv"`
	Name     string  `json:"name"`
	Address  Address `json:"address,omitempty"`
}
