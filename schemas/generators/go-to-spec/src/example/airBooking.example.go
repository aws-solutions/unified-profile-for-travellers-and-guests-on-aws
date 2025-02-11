package example

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/core"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateAirBooking(passengers []air.PassengerProfile, nLines int) ([]byte, string, error) {
	file := ""
	nEventsPerFile := nLines
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

func buildAirBookingPath(booking air.Booking) string {
	return booking.LastUpdatedOn.Format("2006/01/02/15") + "/" + booking.ID + ".json"

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
	nPax := gofakeit.Number(1, 5)
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
		AncillaryServices:  BuildAncillaryServices(nPax),
		Itinerary:          it,
		Return:             ret,
		PassengerInfo:      air.PassengerInfo{Passengers: []air.PassengerProfile{}},
		Promocode:          gofakeit.Regex("[0-9A-Z]{10}"),
		Status:             gofakeit.RandomString([]string{air.BOOKING_STATUS_CANCELED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CANCELED, air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_TENTATIVE}),
		DayOfTravelEmail:   gofakeit.Email(),
		DayOfTravelPhone:   gofakeit.Phone(),
	}
	bookerIndex := gofakeit.Number(1, nPax)
	pax, remaning, hasPaxLeft := pickPax(passengers)
	for i := 0; i < nPax; i++ {
		pax, remaning, hasPaxLeft = pickPax(remaning)
		if !hasPaxLeft {
			break
		}
		if bookerIndex-1 == i {
			pax.IsBooker = true
		}
		booking.PassengerInfo.Passengers = append(booking.PassengerInfo.Passengers, pax)
	}
	booking.Price = BuildAirBookingPrice(booking.PassengerInfo.Passengers, it.Segments, ret.Segments)
	return booking
}

func pickPax(remainingPax []air.PassengerProfile) (air.PassengerProfile, []air.PassengerProfile, bool) {
	if len(remainingPax) == 0 {
		return air.PassengerProfile{}, []air.PassengerProfile{}, false
	}
	i := gofakeit.Number(0, len(remainingPax)-1)
	guest := remainingPax[i]
	return guest, append(remainingPax[:i], remainingPax[i+1:]...), true
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

func BuildAncillaryServices(nPax int) air.AncillaryServices {
	baggage := BuildAncillaryServiceBaggage(nPax)
	seats := BuildAncillaryServiceSeat(nPax)
	flightChanges := BuildAncillaryServiceFlightChange()
	otherAncillaries := BuildAncillaryServiceOther()
	priority := BuildAncillaryServicePriority()
	price := BuildAncillaryServicePrice(baggage, seats, flightChanges, otherAncillaries, priority)
	return air.AncillaryServices{
		AncillaryServiceBaggage:      baggage,
		AncillaryServiceSeat:         seats,
		AncillaryServiceFlightChange: flightChanges,
		AncillaryServiceOther:        otherAncillaries,
		AncillaryServicePriority:     priority,
		Price:                        price,
	}
}

func BuildAncillaryServicePrice(baggage []air.AncillaryServiceBaggage, seats []air.AncillaryServiceSeat, flightChanges []air.AncillaryServiceFlightChange, otherAncillaries []air.AncillaryServiceOther, priority []air.AncillaryServicePriority) air.AncillaryServicePrice {
	price := air.AncillaryServicePrice{}
	for _, b := range baggage {
		price.Total += b.Price.Total
	}
	for _, s := range seats {
		price.Total += s.Price.Total
	}
	for _, f := range flightChanges {
		price.Total += f.Price.Total
	}
	for _, o := range otherAncillaries {
		price.Total += o.Price.Total
	}
	for _, p := range priority {
		price.Total += p.Price.Total
	}
	return price
}

func BuildAncillaryServiceBaggage(nPax int) []air.AncillaryServiceBaggage {
	services := []air.AncillaryServiceBaggage{}
	if gofakeit.Bool() {
		price := gofakeit.Float64Range(50.0, 120.0)
		services = append(services, air.AncillaryServiceBaggage{
			ID:          gofakeit.Regex("[0-9A-Z]{6}"),
			BaggageType: air.BAGGAGE_TYPE_CARRY_ON,
			Quantity:    gofakeit.Number(1, nPax),
			Price: air.AncillaryServicePrice{
				ServiceName:       air.BAGGAGE_TYPE_CARRY_ON,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	if gofakeit.Bool() {
		price := gofakeit.Float64Range(100.0, 300.0)
		services = append(services, air.AncillaryServiceBaggage{
			ID:          gofakeit.Regex("[0-9A-Z]{6}"),
			BaggageType: air.BAGGAGE_TYPE_CHECKED_BAG,
			Quantity:    gofakeit.Number(1, nPax),
			Price: air.AncillaryServicePrice{
				ServiceName:       air.BAGGAGE_TYPE_CHECKED_BAG,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	if gofakeit.Bool() {
		price := gofakeit.Float64Range(100.0, 300.0)
		services = append(services, air.AncillaryServiceBaggage{
			ID:          gofakeit.Regex("[0-9A-Z]{6}"),
			BaggageType: air.BAGGAGE_TYPE_OVERWEIGHT_BAG,
			Weight:      core.Float(gofakeit.Float64Range(25.0, 100.0)),
			Quantity:    gofakeit.Number(1, nPax),
			Price: air.AncillaryServicePrice{
				ServiceName:       air.BAGGAGE_TYPE_OVERWEIGHT_BAG,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	if gofakeit.Bool() {
		price := gofakeit.Float64Range(100.0, 300.0)
		services = append(services, air.AncillaryServiceBaggage{
			ID:          gofakeit.Regex("[0-9A-Z]{6}"),
			BaggageType: air.BAGGAGE_TYPE_OVERSIZED_BAG,
			Quantity:    gofakeit.Number(1, nPax),
			Dimensions: air.BagDimensions{
				Length: gofakeit.Number(20, 100),
				Width:  gofakeit.Number(20, 100),
				Height: gofakeit.Number(20, 100),
			},
			Price: air.AncillaryServicePrice{
				ServiceName:       air.BAGGAGE_TYPE_OVERSIZED_BAG,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	if gofakeit.Bool() {
		price := gofakeit.Float64Range(100.0, 300.0)
		services = append(services, air.AncillaryServiceBaggage{
			ID:          gofakeit.Regex("[0-9A-Z]{6}"),
			BaggageType: air.BAGGAGE_TYPE_SPORTING_EQUIPMENT,
			Quantity:    gofakeit.Number(1, nPax),
			Dimensions: air.BagDimensions{
				Length: gofakeit.Number(20, 100),
				Width:  gofakeit.Number(20, 100),
				Height: gofakeit.Number(20, 100),
			},
			Price: air.AncillaryServicePrice{
				ServiceName:       air.BAGGAGE_TYPE_SPORTING_EQUIPMENT,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	return services
}
func BuildAncillaryServiceSeat(nPax int) []air.AncillaryServiceSeat {
	pricesByZone := map[string]float64{
		air.SEAT_ZONE_BUSINESS:       0.0,
		air.SEAT_ZONE_FIRST:          0.0,
		air.SEAT_ZONE_EXTRA_LEG_ROOM: gofakeit.Float64Range(10.0, 50.0),
		air.SEAT_ZONE_FRONT_AIRCRAFT: gofakeit.Float64Range(20.0, 60.0),
		air.SEAT_ZONE_ECONOMY:        0.0,
		air.SEAT_ZONE_ECONOMY_PLUS:   gofakeit.Float64Range(70.0, 200.0),
	}
	seatZone := gofakeit.RandomString([]string{air.SEAT_ZONE_LIE_FLAT,
		air.SEAT_ZONE_BUSINESS,
		air.SEAT_ZONE_FIRST,
		air.SEAT_ZONE_EXTRA_LEG_ROOM,
		air.SEAT_ZONE_FRONT_AIRCRAFT,
		air.SEAT_ZONE_ECONOMY,
		air.SEAT_ZONE_ECONOMY_PLUS})
	services := []air.AncillaryServiceSeat{}
	for i := 0; i < nPax; i++ {
		services = append(services, air.AncillaryServiceSeat{
			ID:         gofakeit.Regex("[0-9A-Z]{6}"),
			PaxIndex:   i,
			SeatNumber: gofakeit.RandomString([]string{"A", "B", "C", "D", "E", "F"}) + strconv.Itoa(gofakeit.Number(1, 50)),
			SeatZone:   seatZone,
			Price: air.AncillaryServicePrice{
				ServiceName:       "Seat selection: " + seatZone,
				AmountBeforeTaxes: core.Float(pricesByZone[seatZone]),
				Currency:          "USD",
				Total:             core.Float(pricesByZone[seatZone]),
			},
			NeighborFreeSeat: gofakeit.Bool(),
			UpgradeAuction:   gofakeit.Bool(),
			Available:        gofakeit.Bool(),
		})
	}
	return services
}

func BuildAncillaryServiceFlightChange() []air.AncillaryServiceFlightChange {
	services := []air.AncillaryServiceFlightChange{}
	price := gofakeit.Float64Range(100.0, 300.0)
	if gofakeit.Bool() {
		services = append(services, air.AncillaryServiceFlightChange{
			ID:         gofakeit.Regex("[0-9A-Z]{6}"),
			ChangeType: "Flight Date change",
			Price: air.AncillaryServicePrice{
				ServiceName:       "Flight Date change",
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	return services
}
func BuildAncillaryServiceOther() []air.AncillaryServiceOther {
	services := []air.AncillaryServiceOther{}
	price := gofakeit.Float64Range(100.0, 300.0)
	otherType := gofakeit.RandomString([]string{air.OTHER_ANCILLARY_TYPE_INFANT_CHARGE, air.OTHER_ANCILLARY_TYPE_UNACCOMPANIED_MINOR})
	if gofakeit.Bool() {
		services = append(services, air.AncillaryServiceOther{
			ID:                  gofakeit.Regex("[0-9A-Z]{6}"),
			OtherAncilliaryType: otherType,
			Price: air.AncillaryServicePrice{
				ServiceName:       otherType,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	return services
}

func BuildAncillaryServicePriority() []air.AncillaryServicePriority {
	services := []air.AncillaryServicePriority{}
	price := gofakeit.Float64Range(100.0, 300.0)
	svcType := gofakeit.RandomString([]string{air.PRIORITY_BOARDING, air.PRIORITY_SECURITY, air.PRIORITY_BAG_DROP, air.PRIORITY_CHECKIN, air.PRIORITY_BAG_RETURN})
	if gofakeit.Bool() {
		services = append(services, air.AncillaryServicePriority{
			ID:                  gofakeit.Regex("[0-9A-Z]{6}"),
			PriorityServiceType: svcType,
			LoungeAccess:        gofakeit.Bool(),
			Price: air.AncillaryServicePrice{
				ServiceName:       svcType,
				AmountBeforeTaxes: core.Float(price),
				Currency:          "USD",
				Total:             core.Float(price),
			},
		})
	}
	return services
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
