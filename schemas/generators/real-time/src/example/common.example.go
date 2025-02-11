package example

import (
	"math/rand"
	"tah/upt/schemas/src/tah-common/air"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"tah/upt/schemas/src/tah-common/lodging"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func AirlineCode() string {
	return gofakeit.RandomString([]string{"AA", "DL", "AF", "LF", "IB", "UA", "RK"})
}
func AirportCode() string {
	return gofakeit.RandomString([]string{"ATL", "LAX", "ORD", "DFW", "DEN", "JFK", "SFO", "SEA", "MCO", "LAS", "CDG", "FRA", "LIM", "PEK", "HKG", "MAD", "DXB", "AMS"})
}
func HotelCode() string {
	return gofakeit.RandomString([]string{"ATL", "NYC", "PAR", "SGP", "TOK", "SPO", "LIM", "ROM", "LON"}) + gofakeit.Regex("[0-1]{1}[0-9]{1}")
}

func FlightNumbers(airlineCode string, nSegments int) []string {
	numbers := []string{}
	for i := 0; i < nSegments; i++ {
		numbers = append(numbers, FlightNumber(airlineCode))
	}
	return numbers
}

func buildFilePath() string {
	now := time.Now()
	startTime := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	return startTime.Format("2006/01/02/15") + "/" + gofakeit.UUID() + ".jsonl"
}

func FlightNumber(airlineCode string) string {
	return airlineCode + gofakeit.Regex("[0-9]{4}")
}

func BuildAirItinerary(airlineCode string, originDepartureDate time.Time, nStops int, withReturn bool, returnDepartureDate time.Time) (air.Itinerary, air.Itinerary) {
	departureDate := originDepartureDate
	origin := AirportCode()
	destination := AirportCode()
	for destination == origin {
		destination = AirportCode()
	}
	it := air.Itinerary{
		From:          origin,
		To:            destination,
		DepartureDate: departureDate.Format(core.DATE_FORMAT),
		DepartureTime: departureDate.Format(core.TIME_FORMAT),
		Status:        gofakeit.RandomString([]string{air.BOOKING_STATUS_CONFIRMED, air.BOOKING_STATUS_CANCELED, air.BOOKING_STATUS_TENTATIVE}),
	}
	itStops, retStops := buildStops(origin, destination, nStops)
	it.Segments = buildSegments(airlineCode, origin, departureDate, itStops)
	it.ArrivalDate = it.Segments[len(it.Segments)-1].ArrivalDate
	it.ArrivalTime = it.Segments[len(it.Segments)-1].ArrivalTime
	it.Duration = buildDuration(it.Segments).String()
	//return itinerary
	ret := air.Itinerary{}
	if withReturn {
		ret.To = origin
		ret.From = destination
		ret.Segments = buildSegments(airlineCode, destination, returnDepartureDate, retStops)
		ret.DepartureDate = returnDepartureDate.Format(core.DATE_FORMAT)
		ret.DepartureTime = returnDepartureDate.Format(core.TIME_FORMAT)
		ret.ArrivalDate = ret.Segments[len(it.Segments)-1].ArrivalDate
		ret.ArrivalTime = ret.Segments[len(it.Segments)-1].ArrivalTime
		ret.Duration = buildDuration(ret.Segments).String()
	}

	return it, ret
}

func buildDuration(segs []air.FlightSegment) time.Duration {
	departure, _ := time.Parse(core.DATE_FORMAT+" "+core.TIME_FORMAT, segs[0].DepartureDate+" "+segs[0].DepartureTime)
	arrival, _ := time.Parse(core.DATE_FORMAT+" "+core.TIME_FORMAT, segs[len(segs)-1].ArrivalDate+" "+segs[len(segs)-1].ArrivalTime)
	return arrival.Sub(departure)
}

func buildSegments(airlineCode string, origin string, departureDate time.Time, stops []string) []air.FlightSegment {
	lastStop := origin
	lastDeparture := departureDate
	lastArrival := departureDate
	segs := []air.FlightSegment{}
	for i, stop := range stops {
		//if not the first stop, adding layover time
		if i > 0 {
			lastDeparture = lastDeparture.Add(time.Hour*time.Duration(gofakeit.Number(1, 3)) + time.Minute*time.Duration(gofakeit.Number(0, 59)))
		}
		lastArrival = lastDeparture.Add(time.Hour*time.Duration(gofakeit.Number(1, 12)) + time.Minute*time.Duration(gofakeit.Number(0, 59)))
		segs = append(segs, air.FlightSegment{
			Rank:          i,
			From:          lastStop,
			To:            stop,
			DepartureDate: lastDeparture.Format(core.DATE_FORMAT),
			DepartureTime: lastDeparture.Format(core.TIME_FORMAT),
			ArrivalDate:   lastArrival.Format(core.DATE_FORMAT),
			ArrivalTime:   lastArrival.Format(core.TIME_FORMAT),
			FlightNumber:  FlightNumber(airlineCode),
		})
		lastStop = stop
	}
	return segs
}

func buildStops(origin string, destination string, nStops int) ([]string, []string) {
	retStops := make([]string, nStops+1, nStops+1)
	stops := []string{}
	retStops[nStops] = origin

	for i := 0; i < nStops; i++ {
		stop := AirportCode()
		for stop == destination || stop == origin {
			stop = AirportCode()
		}
		stops = append(stops, stop)
		retStops[nStops-i-1] = stop
	}
	stops = append(stops, destination)
	return stops, retStops
}

func BuildItinerary(originDepartureDate time.Time, nSegments int) ([]string, []string, [][]string) {
	departures := []string{}
	arrivals := []string{}
	segments := [][]string{}
	departureDate := originDepartureDate
	arrivalDate := departureDate.Add(time.Hour * time.Duration(gofakeit.Number(1, 12)))
	departures = append(departures, departureDate.Format(core.DATETIME_FORMAT))
	arrivals = append(arrivals, arrivalDate.Format(core.DATETIME_FORMAT))
	initialOrigin := AirportCode()
	origin := initialOrigin
	destination := AirportCode()
	for destination == origin {
		destination = AirportCode()
	}
	segments = append(segments, []string{origin, destination})
	for i := 1; i < nSegments; i++ {
		departureDate = arrivalDate.Add(time.Hour * time.Duration(gofakeit.Number(1, 2)))
		arrivalDate = departureDate.Add(time.Hour * time.Duration(gofakeit.Number(1, 12)))
		departures = append(departures, departureDate.Format(core.DATETIME_FORMAT))
		arrivals = append(arrivals, arrivalDate.Format(core.DATETIME_FORMAT))

		origin := destination
		//last segment should return to orgin
		if i == nSegments-1 {
			destination = initialOrigin
		} else {
			destination := AirportCode()
			for destination == origin {
				destination = AirportCode()
			}
		}
		segments = append(segments, []string{origin, destination})
	}
	return departures, arrivals, segments
}

func Passenger(nLoyaltyTx int) air.PassengerProfile {
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	pax := air.PassengerProfile{
		ModelVersion:  core.MODEL_VERSION,
		ID:            gofakeit.Regex("[0-9]{10}"),
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Gender:        gofakeit.RandomString([]string{gofakeit.Gender(), "other"}),
		LastUpdatedOn: createdOn.AddDate(0, gofakeit.Number(1, 4), 0),
		CreatedOn:     createdOn,
		LastUpdatedBy: gofakeit.Name(),
		CreatedBy:     gofakeit.Name(),

		Phones:          []common.Phone{},
		MiddleName:      gofakeit.FirstName(),
		DateOfBirth:     gofakeit.Date().Format(core.DATE_FORMAT),
		Language:        common.Language{Code: gofakeit.LanguageAbbreviation()},
		Nationality:     common.Country{Code: gofakeit.CountryAbr()},
		JobTitle:        gofakeit.JobTitle(),
		CompanyName:     gofakeit.Company(),
		Emails:          []common.Email{},
		LoyaltyPrograms: []air.LoyaltyProgram{},
		Addresses:       []common.Address{},
	}
	if pax.Gender == "female" {
		pax.Honorific = gofakeit.RandomString([]string{common.HONORIFIC_MRS,
			common.HONORIFIC_MISS,
			common.HONORIFIC_MRS,
			common.HONORIFIC_MS,
			common.HONORIFIC_DAME,
			common.HONORIFIC_LADY,
			""})
		pax.Pronoun = gofakeit.RandomString([]string{"she", ""})
	} else {
		pax.Honorific = gofakeit.RandomString([]string{common.HONORIFIC_MR,
			common.HONORIFIC_MX,
			common.HONORIFIC_SIR,
			common.HONORIFIC_DR,
			common.HONORIFIC_CLLR,
			common.HONORIFIC_LORD,
			""})
		pax.Pronoun = gofakeit.RandomString([]string{"he", "they", ""})
	}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		pax.Addresses = append(pax.Addresses, BuildAddress())
	}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		pax.Emails = append(pax.Emails, BuildEmail())
	}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		pax.Phones = append(pax.Phones, BuildPhone())
	}
	for i := 0; i < gofakeit.Number(0, 2); i++ {
		pax.LoyaltyPrograms = append(pax.LoyaltyPrograms, BuildPaxLoyalty(createdOn, nLoyaltyTx))
	}
	return pax
}

func Guest(nLoyaltyTx int, simpleGuest bool) lodging.GuestProfile {
	now := time.Now()
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	guest := lodging.GuestProfile{
		ModelVersion:    core.MODEL_VERSION,
		ID:              gofakeit.Regex("[0-9]{10}"),
		FirstName:       gofakeit.FirstName(),
		LastName:        gofakeit.LastName(),
		Gender:          gofakeit.RandomString([]string{gofakeit.Gender(), "other"}),
		LastUpdatedOn:   createdOn.AddDate(0, gofakeit.Number(1, 4), 0),
		CreatedOn:       createdOn,
		LastUpdatedBy:   gofakeit.Name(),
		CreatedBy:       gofakeit.Name(),
		Phones:          []common.Phone{},
		MiddleName:      gofakeit.FirstName(),
		DateOfBirth:     gofakeit.Date().Format(core.DATE_FORMAT),
		Language:        common.Language{Code: gofakeit.LanguageAbbreviation()},
		Nationality:     common.Country{Code: gofakeit.CountryAbr()},
		JobTitle:        gofakeit.JobTitle(),
		CompanyName:     gofakeit.Company(),
		Emails:          []common.Email{},
		LoyaltyPrograms: []lodging.LoyaltyProgram{},
		Addresses:       []common.Address{},
	}
	if guest.Gender == "female" {
		guest.Honorific = gofakeit.RandomString([]string{common.HONORIFIC_MRS,
			common.HONORIFIC_MISS,
			common.HONORIFIC_MRS,
			common.HONORIFIC_MS,
			common.HONORIFIC_DAME,
			common.HONORIFIC_LADY,
			""})
		guest.Pronoun = gofakeit.RandomString([]string{"she", ""})
	} else {
		guest.Honorific = gofakeit.RandomString([]string{common.HONORIFIC_MR,
			common.HONORIFIC_MX,
			common.HONORIFIC_SIR,
			common.HONORIFIC_DR,
			common.HONORIFIC_CLLR,
			common.HONORIFIC_LORD,
			""})
		guest.Pronoun = gofakeit.RandomString([]string{"he", "they", ""})
	}
	nAddr := gofakeit.Number(0, 3)
	if simpleGuest {
		nAddr = gofakeit.Number(0, 1)
	}
	for i := 0; i < nAddr; i++ {
		guest.Addresses = append(guest.Addresses, BuildAddress())
	}
	nEmails := gofakeit.Number(0, 3)
	if simpleGuest {
		nEmails = gofakeit.Number(0, 1)
	}
	for i := 0; i < nEmails; i++ {
		guest.Emails = append(guest.Emails, BuildEmail())
	}
	nPhone := gofakeit.Number(0, 3)
	if simpleGuest {
		nPhone = gofakeit.Number(0, 1)
	}
	for i := 0; i < nPhone; i++ {
		guest.Phones = append(guest.Phones, BuildPhone())
	}
	nLoyalty := gofakeit.Number(0, 2)
	if simpleGuest {
		nLoyalty = gofakeit.Number(0, 1)
	}
	for i := 0; i < nLoyalty; i++ {
		guest.LoyaltyPrograms = append(guest.LoyaltyPrograms, BuildGuestLoyalty(createdOn, nLoyaltyTx))
	}
	return guest
}

func BuildPaxLoyalty(createdOn time.Time, nTx int) air.LoyaltyProgram {
	txs := []common.LoyaltyTx{}
	programName := gofakeit.RandomString([]string{"ascend", "elite", "reward", "club"})
	for i := 0; i < nTx; i++ {
		txs = append(txs, BuildLoyaltyTx(createdOn, programName, "air"))
	}
	return air.LoyaltyProgram{
		ID:               gofakeit.Regex("[0-9]{10}"),
		ProgramName:      programName,
		Miles:            computePoints(txs),
		MilesToNextLevel: core.Float(gofakeit.Number(0, 10000)),
		Level:            gofakeit.RandomString([]string{"bronze", "silver", "gold", "platinium"}),
		Joined:           createdOn,
		Transactions:     txs,
	}
}

func BuildGuestLoyalty(createdOn time.Time, nTx int) lodging.LoyaltyProgram {
	txs := []common.LoyaltyTx{}
	programName := gofakeit.RandomString([]string{"ascend", "elite", "reward", "club"})
	for i := 0; i < nTx; i++ {
		txs = append(txs, BuildLoyaltyTx(createdOn, programName, "lodging"))
	}
	return lodging.LoyaltyProgram{
		ID:                gofakeit.Regex("[0-9]{10}"),
		ProgramName:       programName,
		Points:            computePoints(txs),
		PointUnit:         gofakeit.RandomString([]string{"points", "miles", "reward points"}),
		PointsToNextLevel: core.Float(gofakeit.Number(0, 10000)),
		Level:             gofakeit.RandomString([]string{"bronze", "silver", "gold", "platinium"}),
		Joined:            createdOn,
		Transactions:      txs,
	}
}

func BuildLoyaltyTx(createdOn time.Time, programName string, industry string) common.LoyaltyTx {
	date := createdOn.Add(time.Hour * time.Duration(gofakeit.Number(1, 200*24)))
	cur := gofakeit.Currency()
	currency := common.Currency{
		Code: cur.Short,
		Name: cur.Long,
	}
	offset := core.Float(-500.0 + rand.Float64()*(5000.0+500.0))
	pointUnit := gofakeit.RandomString([]string{"points", "reward points"})
	source := gofakeit.RandomString([]string{"web", "mobile", "pms"})
	productID := gofakeit.RandomString([]string{"SUITE", "DBL", "KNG"})
	location := gofakeit.RandomString([]string{"digital", "on property " + HotelCode()})
	if industry == "air" {
		pointUnit = "miles"
		source = gofakeit.RandomString([]string{"web", "mobile", "airport"})
		productID = gofakeit.RandomString([]string{"FARE", "SEAT", "LUGG"})
		location = gofakeit.RandomString([]string{"digital", "at airport " + AirportCode()})

	}
	return common.LoyaltyTx{
		ID:                       gofakeit.Regex("[0-9]{10}"),
		TimeStamp:                date,
		Category:                 gofakeit.RandomString([]string{"redeem", "purchase", "refund"}),
		Description:              gofakeit.RandomString([]string{"Redempton of points for stay", "Purchase of points", "Refund of points"}),
		PointsOffset:             offset,
		PointUnit:                pointUnit,
		OriginPointsOffset:       offset,
		QualifyingPointOffset:    offset,
		Source:                   source,
		Agent:                    gofakeit.Name(),
		BookingDate:              date,
		OrderNumber:              gofakeit.Regex("[0-9]{10}"),
		ProductId:                productID,
		ExpireInDays:             int64(gofakeit.Number(1, 200)),
		Amount:                   core.Float(100.0 + rand.Float64()*(1000.0-100.0)),
		AmountType:               gofakeit.RandomString([]string{"redeem", "purchase", "refund"}),
		Currency:                 currency,
		VoucherQuantity:          int64(gofakeit.Number(0, 3)),
		CancReason:               BuildCancelReason(date),
		ParentTransactionId:      "",
		CorporateReferenceNumber: gofakeit.Regex("[A-Z0-9]{10}"),
		Promotions:               gofakeit.Regex("[A-Z0-9]{10}"),
		Location:                 location,
		ActivityDay:              date,
		ToLoyaltyID:              "",
		FromLoyaltyID:            "",
		OrganizationCode:         gofakeit.Company(),
		VoucherIDs:               []string{gofakeit.Regex("[A-Z0-9]{10}"), gofakeit.Regex("[A-Z0-9]{10}")},
		DocumentNumber:           gofakeit.Regex("[A-Z0-9]{10}"),
		Comment:                  BuildComment(date, "comment from agent", "", ""),
		ExternalTransaction:      "",
		Status:                   gofakeit.RandomString([]string{"processed", "processed", "processed", "processed", "processed", "in-progress", "cancelled", "failed"}),
		PaymentInformation:       BuildPaymentInformation(),
		CorporateID:              gofakeit.Company(),
		ProgramName:              programName,
		ExternalOrderIDs:         []common.ExternalID{BuildExternalID(), BuildExternalID()},
	}
}

func computePoints(txs []common.LoyaltyTx) core.Float {
	points := core.Float(0.0)
	for _, tx := range txs {
		points += tx.PointsOffset
		if points < 0 {
			points = core.Float(0.0)
		}
	}
	return points
}

func BuildEmail() common.Email {
	return common.Email{
		Address: gofakeit.Email(),
		Primary: gofakeit.Bool(),
		Type:    gofakeit.RandomString([]string{"", common.EMAIL_TYPE_PERSONAL, common.EMAIL_TYPE_BUSINESS})}
}
func BuildPhone() common.Phone {
	return common.Phone{
		Type:        gofakeit.RandomString([]string{"", common.PHONE_TYPE_PERSONAL, common.PHONE_TYPE_BUSINESS, common.PHONE_TYPE_HOME, common.PHONE_TYPE_MOBILE}),
		Number:      gofakeit.RandomString([]string{gofakeit.Phone(), gofakeit.PhoneFormatted()}),
		CountryCode: int64(gofakeit.Number(1, 99)),
		Primary:     gofakeit.Bool(),
	}
}

func BuildPaymentInformation() common.PaymentInformation {
	pi := common.PaymentInformation{
		PaymentType: gofakeit.RandomString([]string{
			common.PAYMENT_TYPE_BANK_ACCOUNT,
			common.PAYMENT_TYPE_CREDIT_CARD,
			common.PAYMENT_TYPE_VOUCHER,
			common.PAYMENT_TYPE_LOYALTY_REDEMPTION,
			common.PAYMENT_TYPE_CASH,
		}),
	}
	if pi.PaymentType == common.PAYMENT_TYPE_CREDIT_CARD {
		pi.CcInfo = common.CreditCardInformation{
			Token:    gofakeit.Regex("[0-9]{16}"),
			CardType: gofakeit.CreditCardType(),
			CardExp:  gofakeit.CreditCardExp(),
		}
		if gofakeit.Bool() {
			pi.CcInfo.Address = BuildAddress()
		}
	}
	if pi.PaymentType == common.PAYMENT_TYPE_BANK_ACCOUNT {
		pi.AchRouting = gofakeit.AchRouting()
		pi.AchAccount = gofakeit.AchAccount()
	}
	if pi.PaymentType == common.PAYMENT_TYPE_VOUCHER {
		pi.VoucherID = gofakeit.Regex("[0-9A-Z]{16}")
	}
	if gofakeit.Bool() {
		pi.Address = BuildAddress()
	}
	return pi
}

func BuildAddress() common.Address {
	address := common.Address{
		Type:       gofakeit.RandomString([]string{"", common.ADDRESS_TYPE_BUSINESS, common.ADDRESS_TYPE_MAILING, common.ADDRESS_TYPE_SHIPPING}),
		Line1:      gofakeit.Street(),
		Line2:      "more content line 2",
		Line3:      "more content line 3",
		Line4:      "more content line 4",
		City:       gofakeit.City(),
		PostalCode: gofakeit.Zip(),
		Country:    common.Country{Code: gofakeit.CountryAbr()},
		Primary:    gofakeit.Bool(),
	}
	addState := gofakeit.Bool()
	if addState {
		address.State = common.StateProvince{
			Name: gofakeit.State(),
			Code: gofakeit.StateAbr(),
		}
	} else {
		address.Province = Province()
	}
	return address
}

func Province() common.StateProvince {
	provinces := map[string]string{
		"AB": "Alberta",
		"BC": "British Columbia",
		"ON": "Ontario",
		"QC": "Quebec",
	}
	code := gofakeit.RandomString([]string{"AB", "BC", "ON", "QC"})
	return common.StateProvince{
		Name: provinces[code],
		Code: code,
	}
}
