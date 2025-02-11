package example

import (
	"encoding/json"
	"fmt"
	"strconv"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"tah/upt/schemas/src/tah-common/lodging"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateAirClickstream(passengers []air.PassengerProfile, nLines int) ([]byte, string, error) {
	sessionID := gofakeit.UUID()
	file := ""
	nEventsPerFile := nLines
	now := time.Now()
	startTime := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	pax := passengers[gofakeit.Number(0, len(passengers)-1)]

	//Air params
	nSegments := gofakeit.Number(1, 4)
	originDepartureDate := gofakeit.DateRange(startTime.AddDate(0, 0, 1), now.AddDate(1, 0, 0))
	departureDates, arrivalDates, segments := BuildItinerary(originDepartureDate, nSegments)
	airlineCode := gofakeit.RandomString([]string{"AA", "DL", "AF", "AC", "LH", "LA", "BA", "JL", "EK"})
	for i := 0; i < nEventsPerFile; i++ {
		click := CreateAirClickeventExample(pax, airlineCode, sessionID, startTime, i, departureDates, arrivalDates, segments)
		if gofakeit.Bool() {
			click = CreateGeofenceEvent(sessionID, startTime, i)
		}
		bytes, err := json.Marshal(click)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func GenerateHotelClickstream(guests []lodging.GuestProfile, bookings []lodging.Booking, nLines int) ([]byte, string, error) {
	sessionID := gofakeit.UUID()
	file := ""
	nEventsPerFile := nLines
	now := time.Now()
	startTime := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	//is tru is the click stream converst to a booking (we will see a booking id)
	withBooking := gofakeit.Bool()
	guest := guests[gofakeit.Number(0, len(guests)-1)]
	booking := bookings[gofakeit.Number(0, len(bookings)-1)]
	if withBooking {
		guest = booking.Segments[0].Holder
	}

	//Lodging params
	nNights := core.Float(gofakeit.Number(1, 10))
	checkInDate := gofakeit.DateRange(startTime.AddDate(0, 0, 1), now.AddDate(1, 0, 0))
	addr := gofakeit.Address()
	destination := addr.City + ", " + addr.State + ", " + addr.Country
	//we randomly have user log in or out
	userLogsIn := gofakeit.Bool()
	//position of the login event in the clickstream
	logInEventPosistion := gofakeit.Number(0, nEventsPerFile-1)

	for i := 0; i < nEventsPerFile; i++ {
		click := CreateLodgingClickeventExample(guest, sessionID, startTime, i, checkInDate, nNights, HotelCode(), destination)
		if gofakeit.Bool() {
			click = CreateGeofenceEvent(sessionID, startTime, i)
		}
		if userLogsIn && logInEventPosistion == i {
			click = CreateLoginEvent(guest, sessionID, startTime, i)
		}
		//last events is the booking
		if i == nEventsPerFile-1 && withBooking {
			click = CreateBookingEvent(booking, sessionID, startTime, i)
		}
		bytes, err := json.Marshal(click)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func CreateBookingEvent(booking lodging.Booking, sessionID string, startTime time.Time, seqNum int) common.ClickEvent {
	event := common.ClickEvent{
		ModelVersion:   "1.0",
		EventType:      common.EVENT_START_BOOKING,
		EventTimestamp: BuildTimestamp(startTime, seqNum),
		Device:         BuildDevice(),
		Session:        common.EventSession{ID: sessionID},
		Attributes:     []common.EventAttribute{},
	}
	//Guest data
	bookingID := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_BOOKING_ID,
		StringValue: booking.ID,
	}
	event.Attributes = append(event.Attributes, bookingID)
	return event
}

func CreateLoginEvent(guest lodging.GuestProfile, sessionID string, startTime time.Time, seqNum int) common.ClickEvent {
	event := common.ClickEvent{
		ModelVersion:   "1.0",
		EventType:      common.EVENT_LOGIN,
		EventTimestamp: BuildTimestamp(startTime, seqNum),
		Device:         BuildDevice(),
		Session:        common.EventSession{ID: sessionID},
		Attributes:     []common.EventAttribute{},
	}
	//Guest data
	guestID := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_ID,
		StringValue: guest.ID,
	}
	event.Attributes = append(event.Attributes, guestID)
	return event
}

func CreateGeofenceEvent(sessionID string, startTime time.Time, seqNum int) common.ClickEvent {
	event := common.ClickEvent{
		ModelVersion:   "1.0",
		EventType:      common.EVENT_CROSS_GEOFENCE,
		EventTimestamp: BuildTimestamp(startTime, seqNum),
		Device:         BuildDevice(),
		Session:        common.EventSession{ID: sessionID},
		Attributes:     []common.EventAttribute{},
	}
	//Guest data
	lat := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_GEOFENCE_LATITUDE,
		StringValue: fmt.Sprintf("%f", gofakeit.Latitude()),
	}
	lng := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_GEOFENCE_LONGITUDE,
		StringValue: fmt.Sprintf("%f", gofakeit.Longitude()),
	}
	idVal := gofakeit.RandomString([]string{"hotel_parking_1", "parc_entrance", "restaurant_1"})
	geofenceNames := map[string]string{"hotel_parking_1": "Hotel Parking", "parc_entrance": "Parc Entrance", "restaurant_1": "Restaurant"}
	id := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_GEOFENCE_ID,
		StringValue: idVal,
	}

	name := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_GEOFENCE_NAME,
		StringValue: geofenceNames[idVal],
	}

	event.Attributes = append(event.Attributes, lat)
	event.Attributes = append(event.Attributes, lng)
	event.Attributes = append(event.Attributes, name)
	event.Attributes = append(event.Attributes, id)
	return event
}

func CreateLodgingClickeventExample(guest lodging.GuestProfile, sessionID string, startTime time.Time, seqNum int, checkinDate time.Time, nNights core.Float, hotelCode string, destination string) common.ClickEvent {
	nProducts := gofakeit.Number(1, 10)
	event := common.ClickEvent{
		ModelVersion:   "1.0",
		EventType:      BuildEventName(),
		EventTimestamp: BuildTimestamp(startTime, seqNum),
		Device:         BuildDevice(),
		Session:        common.EventSession{ID: sessionID},
		Attributes:     []common.EventAttribute{},
	}
	endDate := checkinDate.AddDate(0, 0, int(nNights))

	//Guest data
	firstName := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_FIRST_NAME,
		StringValue: guest.FirstName,
	}
	lastName := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_LAST_NAME,
		StringValue: guest.LastName,
	}
	if len(guest.Emails) > 0 {
		email := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_CUSTOMER_EMAIL,
			StringValue: guest.Emails[0].Address,
		}
		event.Attributes = append(event.Attributes, email)
	}
	if len(guest.Phones) > 0 {
		number := guest.Phones[0].Number
		if guest.Phones[0].CountryCode > 0 {
			number = strconv.Itoa(int(guest.Phones[0].CountryCode)) + " " + number
		}
		phone := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_CUSTOMER_PHONE,
			StringValue: number,
		}
		event.Attributes = append(event.Attributes, phone)
	}
	if len(guest.LoyaltyPrograms) > 0 {
		loyalty := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_LOYALTY_ID,
			StringValue: guest.LoyaltyPrograms[0].ID,
		}
		event.Attributes = append(event.Attributes, loyalty)
	}

	birthDate := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_BIRTHDATE,
		StringValue: guest.DateOfBirth,
	}
	language := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_LANGUAGE_CODE,
		StringValue: guest.Language.Code,
	}
	gender := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_GENDER,
		StringValue: guest.Gender,
	}
	event.Attributes = append(event.Attributes, gender)
	event.Attributes = append(event.Attributes, language)
	event.Attributes = append(event.Attributes, birthDate)
	event.Attributes = append(event.Attributes, firstName)
	event.Attributes = append(event.Attributes, lastName)

	//Lodging data
	checkin := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_CHECKIN_DATE,
		StringValue: checkinDate.Format(core.DATE_FORMAT),
	}
	checkout := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_CHECKOUT_DATE,
		StringValue: endDate.Format(core.DATE_FORMAT),
	}
	destinationAttr := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_DESTINATION,
		StringValue: destination,
	}
	hotelCodeAttr := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_HOTEL_CODE,
		StringValue: hotelCode,
	}
	hotelCodes := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         lodging.EVENT_ATTRIBUTE_HOTEL_CODES,
		StringValues: []string{},
	}
	numNights := common.EventAttribute{
		Type:     common.EVENT_ATTRIBUTE_TYPE_NUMBER,
		Name:     lodging.EVENT_ATTRIBUTE_N_NIGHTS,
		NumValue: nNights,
	}
	roomType := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_ROOM_TYPE,
		StringValue: gofakeit.RandomString([]string{"DBL", "KING", "SUITE"}),
	}
	ratePlan := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        lodging.EVENT_ATTRIBUTE_RATE_PLAN,
		StringValue: gofakeit.RandomString([]string{"BAR", "SUPER_SAVER", "NON_REFUNDABLE"}),
	}
	productsQty := common.EventAttribute{
		Type:      common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:      lodging.EVENT_ATTRIBUTE_QUANTITIES,
		NumValues: []core.Float{},
	}
	productsCodes := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         lodging.EVENT_ATTRIBUTE_PRODUCTS,
		StringValues: []string{},
	}
	productsPrices := common.EventAttribute{
		Type:      common.EVENT_ATTRIBUTE_TYPE_NUMBERS,
		Name:      lodging.EVENT_ATTRIBUTE_PRODUCTS_PRICES,
		NumValues: []core.Float{},
	}
	for i := 0; i < nProducts; i++ {
		productsQty.NumValues = append(productsQty.NumValues, core.Float(gofakeit.Number(1, 5)))
		productsCodes.StringValues = append(productsCodes.StringValues, gofakeit.RandomString([]string{"DBL-BAR", "KING_SUPER_SAVER", "SUITE-BAR", "BKFST", "CHMPGN"}))
		productsPrices.NumValues = append(productsPrices.NumValues, core.Float(gofakeit.Number(100, 1000)))
		hotelCodes.StringValues = append(hotelCodes.StringValues, HotelCode())
	}
	event.Attributes = append(event.Attributes, productsQty)
	event.Attributes = append(event.Attributes, productsCodes)
	event.Attributes = append(event.Attributes, destinationAttr)
	event.Attributes = append(event.Attributes, numNights)
	event.Attributes = append(event.Attributes, roomType)
	event.Attributes = append(event.Attributes, ratePlan)
	event.Attributes = append(event.Attributes, checkin)
	event.Attributes = append(event.Attributes, checkout)
	event.Attributes = append(event.Attributes, hotelCodeAttr)
	event.Attributes = append(event.Attributes, hotelCodes)
	return event
}

func CreateAirClickeventExample(pax air.PassengerProfile, airlineCode string, sessionID string, startTime time.Time, seqNum int, departureDates []string, arrivalDates []string, segments [][]string) common.ClickEvent {
	nProducts := gofakeit.Number(1, 10)
	evName := BuildEventName()
	event := common.ClickEvent{
		ModelVersion:   "1.0",
		EventType:      evName,
		EventTimestamp: BuildTimestamp(startTime, seqNum),
		Device:         BuildDevice(),
		Session:        common.EventSession{ID: sessionID},
		Attributes:     []common.EventAttribute{},
	}

	//Guest data
	firstName := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_FIRST_NAME,
		StringValue: pax.FirstName,
	}
	lastName := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_LAST_NAME,
		StringValue: pax.LastName,
	}
	if len(pax.Emails) > 0 {
		email := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_CUSTOMER_EMAIL,
			StringValue: pax.Emails[0].Address,
		}
		event.Attributes = append(event.Attributes, email)
	}
	if len(pax.Phones) > 0 {
		number := pax.Phones[0].Number
		if pax.Phones[0].CountryCode > 0 {
			number = strconv.Itoa(int(pax.Phones[0].CountryCode)) + " " + number
		}
		phone := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_CUSTOMER_PHONE,
			StringValue: number,
		}
		event.Attributes = append(event.Attributes, phone)
	}
	if len(pax.LoyaltyPrograms) > 0 {
		loyalty := common.EventAttribute{
			Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
			Name:        common.EVENT_ATTRIBUTE_LOYALTY_ID,
			StringValue: pax.LoyaltyPrograms[0].ID,
		}
		event.Attributes = append(event.Attributes, loyalty)
	}

	birthDate := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_BIRTHDATE,
		StringValue: pax.DateOfBirth,
	}
	language := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_LANGUAGE_CODE,
		StringValue: pax.Language.Code,
	}
	customerID := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_ID,
		StringValue: pax.ID,
	}
	gender := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        common.EVENT_ATTRIBUTE_CUSTOMER_GENDER,
		StringValue: pax.Gender,
	}
	event.Attributes = append(event.Attributes, gender)
	event.Attributes = append(event.Attributes, customerID)
	event.Attributes = append(event.Attributes, language)
	event.Attributes = append(event.Attributes, birthDate)
	event.Attributes = append(event.Attributes, firstName)
	event.Attributes = append(event.Attributes, lastName)

	fareClass := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_FARE_CLASS,
		StringValue: gofakeit.RandomString([]string{"Y", "J", "F"}),
	}
	fareType := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_FARE_TYPE,
		StringValue: gofakeit.RandomString([]string{"P", "N", "O"}),
	}
	flightMarket := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_FLIGHT_MARKET,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	flightType := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_FLIGHT_TYPE,
		StringValue: gofakeit.RandomString([]string{"one-way", "round-trip"}),
	}
	origineDate := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_ORIGIN_DATE,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	origineDateTime := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_ORIGIN_DATE_TIME,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	returnDateTime := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_RETURNING_DATE_TIME,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	returnDate := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_RETURNING_DATE,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	flightRoute := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_RETURNING_FLIGHT_ROUTE,
		StringValue: gofakeit.RandomString([]string{"DOM,DOM"}),
	}
	nPaxAdults := common.EventAttribute{
		Type:     common.EVENT_ATTRIBUTE_TYPE_NUMBER,
		Name:     air.EVENT_ATTRIBUTE_NUM_PAX_ADULTS,
		NumValue: core.Float(gofakeit.Number(1, 6)),
	}
	nPaxInfant := common.EventAttribute{
		Type:     common.EVENT_ATTRIBUTE_TYPE_NUMBER,
		Name:     air.EVENT_ATTRIBUTE_NUM_PAX_INF,
		NumValue: core.Float(gofakeit.Number(0, 1)),
	}
	nPaxChildren := common.EventAttribute{
		Type:     common.EVENT_ATTRIBUTE_TYPE_NUMBER,
		Name:     air.EVENT_ATTRIBUTE_NUM_PAX_CHILDREN,
		NumValue: core.Float(gofakeit.Number(0, 2)),
	}
	paxType := common.EventAttribute{
		Type:        common.EVENT_ATTRIBUTE_TYPE_STRING,
		Name:        air.EVENT_ATTRIBUTE_PAX_TYPE,
		StringValue: gofakeit.RandomString([]string{"adult", "child", "infant"}),
	}
	totalPax := common.EventAttribute{
		Type:     common.EVENT_ATTRIBUTE_TYPE_NUMBER,
		Name:     air.EVENT_ATTRIBUTE_TOTAL_PASSENGERS,
		NumValue: core.Float(gofakeit.Number(1, 6)),
	}

	productsQty := common.EventAttribute{
		Type:      common.EVENT_ATTRIBUTE_TYPE_NUMBERS,
		Name:      air.EVENT_ATTRIBUTE_QUANTITIES,
		NumValues: []core.Float{},
	}
	productsCodes := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         air.EVENT_ATTRIBUTE_PRODUCTS,
		StringValues: []string{},
	}
	productsPrices := common.EventAttribute{
		Type:      common.EVENT_ATTRIBUTE_TYPE_NUMBERS,
		Name:      air.EVENT_ATTRIBUTE_PRODUCTS_PRICES,
		NumValues: []core.Float{},
	}
	flightSegmentsDepartureDateTime := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         air.EVENT_ATTRIBUTE_FLIGHT_SEGMENTS_DEPARTURE_DATE_TIME,
		StringValues: []string{},
	}
	flightSegmentsArrivalDateTime := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         air.EVENT_ATTRIBUTE_FLIGHT_SEGMENTS_ARRIVAL_DATE_TIME,
		StringValues: []string{},
	}
	flightSegmentsSku := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         air.EVENT_ATTRIBUTE_FLIGHT_SEGMENT_SKU,
		StringValues: FlightNumbers(airlineCode, len(segments)),
	}
	flightSegments := common.EventAttribute{
		Type:         common.EVENT_ATTRIBUTE_TYPE_STRINGS,
		Name:         air.EVENT_ATTRIBUTE_FLIGHT_SEGMENTS,
		StringValues: []string{},
	}
	for i := 0; i < nProducts; i++ {
		productsQty.NumValues = append(productsQty.NumValues, core.Float(gofakeit.Number(1, 5)))
		productsCodes.StringValues = append(productsCodes.StringValues, gofakeit.RandomString([]string{"FARE", "SEAT", "LUGG"}))
		productsPrices.NumValues = append(productsPrices.NumValues, core.Float(gofakeit.Number(100, 1000)))
	}
	for i := 0; i < len(segments); i++ {
		flightSegmentsDepartureDateTime.StringValues = append(flightSegmentsDepartureDateTime.StringValues, departureDates[i])
		flightSegmentsArrivalDateTime.StringValues = append(flightSegmentsArrivalDateTime.StringValues, arrivalDates[i])
		flightSegments.StringValues = append(flightSegments.StringValues, segments[i][0]+"-"+segments[i][1])
	}

	if len(pax.AlternateProfileIDs) > 0 {
		event.AlternateProfileIDs = append(event.AlternateProfileIDs, pax.AlternateProfileIDs...)
	}

	event.Attributes = append(event.Attributes, productsQty)
	event.Attributes = append(event.Attributes, productsCodes)
	event.Attributes = append(event.Attributes, productsPrices)
	event.Attributes = append(event.Attributes, fareClass)
	event.Attributes = append(event.Attributes, fareType)
	event.Attributes = append(event.Attributes, flightSegmentsDepartureDateTime)
	event.Attributes = append(event.Attributes, flightSegmentsArrivalDateTime)
	event.Attributes = append(event.Attributes, flightSegmentsSku)
	event.Attributes = append(event.Attributes, flightMarket)
	event.Attributes = append(event.Attributes, flightType)
	event.Attributes = append(event.Attributes, origineDate)
	event.Attributes = append(event.Attributes, origineDateTime)
	event.Attributes = append(event.Attributes, returnDateTime)
	event.Attributes = append(event.Attributes, returnDate)
	event.Attributes = append(event.Attributes, flightRoute)
	event.Attributes = append(event.Attributes, nPaxAdults)
	event.Attributes = append(event.Attributes, nPaxInfant)
	event.Attributes = append(event.Attributes, nPaxChildren)
	event.Attributes = append(event.Attributes, paxType)
	event.Attributes = append(event.Attributes, totalPax)

	return event
}

func BuildEventName() string {
	return gofakeit.RandomString([]string{common.EVENT_START_SESSION,
		common.EVENT_LOGIN,
		common.EVENT_LOGOUT,
		common.EVENT_SEARCH_DESTINATION,
		common.EVENT_SEARCH_EXPERIENCE,
		common.EVENT_SELECT_DESTINATION,
		common.EVENT_SELECT_EXPERIENCE,
		common.EVENT_SELECT_ORIGIN,
		common.EVENT_SELECT_N_TRAVELLER,
		common.EVENT_SELECT_START_DATE,
		common.EVENT_SELECT_END_DATE,
		common.EVENT_SELECT_START_TIME,
		common.EVENT_SELECT_END_TIME,
		common.EVENT_SELECT_N_ROOM,
		common.EVENT_SELECT_N_NIGHT,
		common.EVENT_SEARCH_FLIGHT,
		common.EVENT_SEARCH_MULTI_AVAIL,
		common.EVENT_SEARCH_SINGLE_AVAIL,
		common.EVENT_VIEW_PRODUUCT,
		common.EVENT_SELECT_PRODUCT,
		common.EVENT_START_BOOKING,
		common.EVENT_UPDATE_BOOKING,
		common.EVENT_CONFIRM_BOOKING,
		common.EVENT_CANCEL_BOOKING,
		common.EVENT_RETRIEVE_BOOKING,
		common.EVENT_SEARCH_BOOKING,
		common.EVENT_IGNORE_BOOKING,
		common.EVENT_REGRET,
		common.EVENT_SEARCH_ADD_ON,
		common.EVENT_VIEW_ADD_ON,
		common.EVENT_SELECT_ADD_ON,
		common.EVENT_REMOVE_ADD_ON,
		common.EVENT_CUSTOM,
		common.EVENT_CROSS_GEOFENCE,
		common.EVENT_SIGN_UP,
		common.EVENT_CHECK_IN,
		common.EVENT_ATTRIBUTE_ORIGIN,
		common.EVENT_ATTRIBUTE_DESTINATION,
		common.EVENT_ATTRIBUTE_MARKETING_CARRIER,
		common.EVENT_ATTRIBUTE_LENGTH_OF_STAY,
		common.EVENT_ATTRIBUTE_ERROR,
		common.EVENT_ATTRIBUTE_SELECTED_SEATS,
	})
}

func BuildDevice() common.EventDevice {
	return common.EventDevice{
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		IpAddress: gofakeit.IPv4Address(),
	}
}

func BuildTimestamp(startTime time.Time, seqNum int) time.Time {
	return startTime.Add(time.Second * time.Duration(seqNum*10)).Add(time.Second * time.Duration(gofakeit.Number(0, 9)))
}
