package common

import (
	"encoding/json"
	"tah/upt/schemas/src/tah-common/core"
	"time"
)

type LoyaltyTx struct {
	ID                       string             `json:"id"`                              //transaction ID
	TimeStamp                time.Time          `json:"timeStamp"`                       //transaction timestamp
	Category                 string             `json:"category,omitempty"`              //transaction category
	Description              string             `json:"description,omitempty"`           //transaction description
	PointsOffset             core.Float         `json:"pointsOffset,omitempty"`          //positive or negative offset of points associated to the transaction
	PointUnit                string             `json:"pointUnit,omitempty"`             //unit of points
	OriginPointsOffset       core.Float         `json:"originPointsOffset,omitempty"`    //The original value of the transaction in points without any subsequent bonus added by each promotion.
	QualifyingPointOffset    core.Float         `json:"qualifyingPointOffset,omitempty"` //The original value of the transaction in points without any subsequent bonus added by each promotion.
	Source                   string             `json:"source"`                          //Source of the transaction
	Agent                    string             `json:"agent,omitempty"`                 //If transaction is performed by an agent, name of the agent
	BookingDate              time.Time          `json:"bookingDate,omitempty"`           //The transaction booking date. If the transaction is an airline transaction, then this attribute indicates the date when the booking is made for travel. There is no default value, dependency, or limitation associated with this attribute
	OrderNumber              string             `json:"orderNumber,omitempty"`
	ProductId                string             `json:"productId,omitempty"`
	ExpireInDays             int64              `json:"expireInDays,omitempty"` //The date when the accrued points will expire, in days.",
	Amount                   core.Float         `json:"amount,omitempty"`
	AmountType               string             `json:"amountType,omitempty"` //amount type of specified currency associated with the transaction: payment/refund
	Currency                 Currency           `json:"currency,omitempty"`
	VoucherQuantity          int64              `json:"voucherQuantity,omitempty"`
	CancReason               CancellationReason `json:"cancReason,omitempty"`
	ParentTransactionId      string             `json:"parentTransactionId,omitempty"`
	CorporateReferenceNumber string             `json:"corporateReferenceNumber,omitempty"` //The reference to a third-party corporate member from the program to get the points split for the individual and the corporation or company, if an individual member completes a transaction on behalf of the corporation or company.
	Promotions               string             `json:"promotions,omitempty"`               //The loyalty promotions resource is used to retrieve the details of a loyalty promotion that are applied to the transaction.
	Location                 string             `json:"location,omitempty"`                 //property restaurant venue when the transaction was processed
	ActivityDay              time.Time          `json:"activityDay,omitempty"`              //The transaction activity day. This attribute represents the day when an activity, either accrual or redemption, occurs.
	ToLoyaltyID              string             `json:"toLoyaltyId,omitempty"`              //The loyaltyId to whom the points were transferred.
	FromLoyaltyID            string             `json:"fromLoyaltyId,omitempty"`            //The unique identification number of the member to whom points are transferred.
	OrganizationCode         string             `json:"organizationCode,omitempty"`         //The organization code associated with the member transaction.
	VoucherIDs               []string           `json:"voucherIds,omitempty"`               //The transaction vouchers resource is used to retrieve the details of loyalty vouchers that are applied to the transaction.
	EventName                string             `json:"eventName,omitempty"`                //The attribute indicates the event name for which the transaction is created.
	DocumentNumber           string             `json:"documentNumber,omitempty"`
	Comment                  Comment            `json:"Comment,omitempty"`
	ExternalTransaction      string             `json:"externalTransaction,omitempty"` //The details of the transaction coming from a partner's external application
	Status                   string             `json:"status,omitempty"`
	PaymentInformation       PaymentInformation `json:"paymentInformation,omitempty"`
	CorporateID              string             `json:"corporateId,omitempty"` //The attribute represents the corporate reference member identifier in the case of the individual member performing a transaction on behalf of corporate
	ProgramName              string             `json:"programName,omitempty"`
	ExternalOrderIDs         []ExternalID       `json:"externalOrderIds,omitempty"`
	ExtendedData             interface{}        `json:"extendedData,omitempty"`
}

func (loyaltyTx LoyaltyTx) MarshalJSON() ([]byte, error) {
	type Alias LoyaltyTx
	return json.Marshal(&struct {
		Alias
		TimeStamp string `json:"timeStamp"`
	}{
		Alias:     (Alias)(loyaltyTx),
		TimeStamp: loyaltyTx.TimeStamp.Format(core.INGEST_TIMESTAMP_FORMAT),
	})
}
