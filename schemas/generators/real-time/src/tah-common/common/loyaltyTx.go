package common

import (
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

type LoyaltyTx struct {
	ID                       string             `json:"id"`                    //transaction ID
	TimeStamp                time.Time          `json:"timeStamp"`             //transaction timestamp
	Category                 string             `json:"category"`              //transaction category
	Description              string             `json:"description"`           //transaction descriprtion
	PointsOffset             core.Float         `json:"pointsOffset"`          //positive or negative offset of poiunts associated to the transaction
	PointUnit                string             `json:"pointUnit"`             //unnit of points
	OriginPointsOffset       core.Float         `json:"originPointsOffset"`    //The original value of the transaction in points without any subsequent bonus added by each promotion.
	QualifyingPointOffset    core.Float         `json:"qualifyingPointOffset"` //The original value of the transaction in points without any subsequent bonus added by each promotion.
	Source                   string             `json:"source"`                //Source of the transaction
	Agent                    string             `json:"agent"`                 //If transaction is performed by an agent, name of the agent
	BookingDate              time.Time          `json:"bookingDate"`           //The transaction booking date. If the transaction is an airline transaction, then this attribute indicates the date when the booking is made for travel. There is no default value, dependency, or limitation associated with this attribute
	OrderNumber              string             `json:"orderNumber"`
	ProductId                string             `json:"productId"`
	ExpireInDays             int64              `json:"expireInDays"` //The date when the accrued points will expire, in days.",
	Amount                   core.Float         `json:"amount"`
	AmountType               string             `json:"amountType"` //ammount type of specified currency assocuated withthe transaction: payment/refund
	Currency                 Currency           `json:"currency"`
	VoucherQuantity          int64              `json:"voucherQuantity"`
	CancReason               CancellationReason `json:"cancReason"`
	ParentTransactionId      string             `json:"parentTransactionId"`
	CorporateReferenceNumber string             `json:"corporateReferenceNumber"` //The reference to a third-party corporate member from the program to get the points split for the individual and the corporation or company, if an individual member completes a transaction on behalf of the corporation or company.
	Promotions               string             `json:"promotions"`               //The loyalty promotions resource is used to retrieve the details of a loyalty promotion that are applied to the transaction.
	Location                 string             `json:"location"`                 //property restaurant venue whene the transaction was processed
	ActivityDay              time.Time          `json:"activityDay"`              //The transaction activity day. This attribute represents the day when an activity, either accrual or redemption, occurs.
	ToLoyaltyID              string             `json:"toLoyaltyId"`              //The loyaltyId to whom the points were transferred.
	FromLoyaltyID            string             `json:"fromLoyaltyId"`            //The unique identification number of the member to whom points are transferred.
	OrganizationCode         string             `json:"organizationCode"`         //The organization code associated with the member transaction.
	VoucherIDs               []string           `json:"voucherIds"`               //The transaction vouchers resource is used to retrieve the details of loyalty vouchers that are applied to the transaction.
	EventName                string             `json:"eventName"`                //The attribute indicates the event name for which the transaction is created.
	DocumentNumber           string             `json:"documentNumber"`
	Comment                  Comment            `json:"Comment"`
	ExternalTransaction      string             `json:"externalTransaction"` //The details of the transaction coming from a partner's external application
	Status                   string             `json:"status"`
	PaymentInformation       PaymentInformation `json:"paymentInformation"`
	CorporateID              string             `json:"corporateId"` //The attribute represents the corporate reference member identifier in the case of the individual member performing a transaction on behalf of corporate
	ProgramName              string             `json:"programName"`
	ExternalOrderIDs         []ExternalID       `json:"externalOrderIds"`
}
