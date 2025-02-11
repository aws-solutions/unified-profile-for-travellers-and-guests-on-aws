package example

import (
	"encoding/json"
	"math/rand"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/core"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateAirBooking(passengers []air.PassengerProfile) ([]byte, string, error) {
	file := ""
	nEventsPerFile := gofakeit.Number(1, 500)
	for i := 0; i < nEventsPerFile; i++ {
		booking := CreateAirBookingExample(passengers)
		bytes, err := json.Marshal(booking)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func SerializeAirBookings(bookings []air.Booking) ([]byte, string, error) {
	file := ""
	for _, booking := range bookings {
		bytes, err := json.Marshal(booking)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func CreateAirBookingExample(passengers []air.PassengerProfile) air.Booking {
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	airlineCode := AirlineCode()
	originDepartureDate := gofakeit.DateRange(createdOn.AddDate(0, 0, 1), createdOn.AddDate(1, 0, 0))
	returnDepartureDate := gofakeit.DateRange(originDepartureDate.AddDate(0, 0, 3), createdOn.AddDate(0, 0, 15))
	withReturn := gofakeit.Bool()
	nStops := gofakeit.Number(0, 2)
	it, ret := BuildAirItinerary(airlineCode, originDepartureDate, nStops, withReturn, returnDepartureDate)
	booking := air.Booking{
		ID:                 gofakeit.Regex("[0-9A-Z]{6}"),
		ObjectVersion:      gofakeit.Number(1, 20),
		ModelVersion:       core.MODEL_VERSION,
		LastUpdatedOn:      createdOn.AddDate(0, gofakeit.Number(1, 4), 0),
		CreatedOn:          createdOn,
		LastUpdatedBy:      gofakeit.Name(),
		CreatedBy:          gofakeit.Name(),
		CreationChannelId:  gofakeit.RandomString([]string{"web", "mobile", "ota-anycompany", "ota-booking", "gds"}),
		PaymentInformation: BuildPaymentInformation(),
		Itinerary:          it,
		Return:             ret,
		PassengerInfo:      air.PassengerInfo{Passengers: []air.PassengerProfile{}},
		Promocode:          gofakeit.Regex("[0-9A-Z]{10}"),
		Status:             gofakeit.RandomString([]string{air.BOOKING_STATUS_CANCELED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CANCELED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_TENTATIVE}),
	}
	bookerIndex := gofakeit.Number(0, len(passengers)-1)
	passengers[bookerIndex].IsBooker = true
	booking.PassengerInfo.Passengers = passengers
	booking.Price = BuildAirBookingPrice(booking.PassengerInfo.Passengers, it.Segments, ret.Segments)
	return booking
}

func BuildAirBookingPrice(passengers []air.PassengerProfile, it []air.FlightSegment, ret []air.FlightSegment) air.BookingPrice {
	ppps := []air.PricePerPassengerPerSegments{}
	for i := range passengers {
		for _, seg := range it {
			fareClass := gofakeit.RandomString([]string{"A", "J", "F"})
			ppps = append(ppps, air.PricePerPassengerPerSegments{
				SegmentRank:    seg.Rank,
				PassengerIndex: i,
				FareClass:      fareClass,
				Price:          BuildFarePrice(fareClass),
			})
		}
	}

	return air.BookingPrice{PriceItems: ppps, GrandTotal: computeGrandTotal(ppps)}
}

func computeGrandTotal(ppps []air.PricePerPassengerPerSegments) core.Float {
	gt := core.Float(0.0)
	for _, ppp := range ppps {
		gt += ppp.Price.Total
	}
	return gt
}

func BuildFarePrice(fareClass string) air.Price {
	pricesByClass := map[string][]float64{
		"A": {200.0, 400.0},
		"J": {500.0, 1000.0},
		"F": {2000.0, 5000.0},
	}
	min := pricesByClass[fareClass][0]
	max := pricesByClass[fareClass][1]
	price := air.Price{
		Base: core.Float(min + rand.Float64()*(max-min)),
		Fees: []air.Fee{
			{
				Amount: 16.50,
				Type:   "PASSENGER FACILITY CHARGE (UNITED STATES)",
			},
			{
				Amount: 17.20,
				Type:   "SEGMENT TAX (UNITED STATES)",
			},
			{
				Amount: 48.07,
				Type:   "TRANSPORTATION TAX (UNITED STATES)",
			},
			{
				Amount: 11.20,
				Type:   "US SECURITY FEE (UNITED STATES)",
			},
		},
	}
	price.Total = computeTotal(price)
	return price
}

func computeTotal(p air.Price) core.Float {
	total := p.Base
	for _, f := range p.Fees {
		total += f.Amount
	}
	return core.Float(total)
}
