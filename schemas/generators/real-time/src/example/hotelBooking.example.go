package example

import (
	"encoding/json"
	"log"
	"strings"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"tah/upt/schemas/src/tah-common/lodging"
	"time"

	"strconv"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateHotelBookings(guests []lodging.GuestProfile) []lodging.Booking {
	bookings := []lodging.Booking{}
	nEventsPerFile := gofakeit.Number(1, 500)
	for i := 0; i < nEventsPerFile; i++ {
		bookings = append(bookings, CreateHotelBookingExample(guests, false))
	}
	return bookings
}

func SerializeHotelBookings(bookings []lodging.Booking) ([]byte, string, error) {
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

func buildBookingPath(booking lodging.Booking) string {
	return booking.LastUpdatedOn.Format("2006/01/02/15") + "/" + booking.HotelCode + "/" + booking.ID + ".json"
}

func CreateHotelBookingExample(guests []lodging.GuestProfile, isSimpleBooking bool) lodging.Booking {
	cur := gofakeit.Currency()
	currency := common.Currency{
		Code: cur.Short,
		Name: cur.Long,
	}
	hotelCode := HotelCode()
	groupID := gofakeit.Regex("[0-9A-Z]{10}")
	now := time.Now()
	holderIndex := gofakeit.Number(0, len(guests)-1)
	holder := guests[holderIndex]
	createdOn := gofakeit.DateRange(now.AddDate(-2, 0, 0), now.AddDate(0, 0, -1))
	updatedOn := createdOn.AddDate(0, gofakeit.Number(1, 4), 0)
	//stay date should be between creation date +1 and a year from now
	checkInDate := gofakeit.DateRange(createdOn.AddDate(0, 0, 1), createdOn.AddDate(1, 0, 0)).Format(core.DATE_FORMAT)
	booking := lodging.Booking{
		ObjectVersion:      gofakeit.Number(1, 20),
		ModelVersion:       core.MODEL_VERSION,
		LastUpdatedOn:      updatedOn,
		CreatedOn:          createdOn,
		LastUpdatedBy:      gofakeit.Name(),
		CreatedBy:          gofakeit.Name(),
		CreationChannelId:  gofakeit.RandomString([]string{"web", "mobile", "ota-anycompany", "ota-booking", "gds"}),
		PaymentInformation: BuildPaymentInformation(),
		GroupId:            groupID,
		Currency:           currency,
		Comments:           []common.Comment{},
		HotelCode:          hotelCode,
		ID:                 gofakeit.Regex("[0-9A-Z]{10}"),
		Holder:             holder,
		Segments:           []lodging.BookingSegment{},
	}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		booking.Comments = append(booking.Comments, BuildComment(createdOn, "comment", "", ""))
	}
	nSegments := gofakeit.Number(1, 3)
	if isSimpleBooking {
		nSegments = 1
	}
	for i := 0; i < nSegments; i++ {
		booking.Segments = append(booking.Segments, BuildBookingSegment(checkInDate, i, hotelCode, currency, groupID, holderIndex, guests, isSimpleBooking))
	}
	for i := 0; i < gofakeit.Number(0, 2); i++ {
		booking.ExternalID = append(booking.ExternalID, BuildExternalID())
	}

	booking.Status = ComputeStatus(booking.Segments)
	booking.StartDate = ComputeStartDate(booking.Segments)
	booking.EndDate = ComputeEndDate(booking.Segments)
	booking.NGuests = ComputeNGuests(booking.Segments)
	booking.NNights = ComputeNNights(booking.Segments)
	if booking.Status == lodging.STATUS_CANCELLED {
		booking.CancelReason = BuildCancelReason(updatedOn)
	}
	return booking
}

func BuildBookingSegment(startDate string, seqNum int, hotelCode string, currency common.Currency, groupID string, holderIndex int, allGuests []lodging.GuestProfile, isSimpleBooking bool) lodging.BookingSegment {
	nGuests := gofakeit.Number(1, 10)
	nNights := int64(gofakeit.Number(1, 10))

	if isSimpleBooking {
		nGuests = gofakeit.Number(1, 2)
		nNights = int64(gofakeit.Number(1, 2))
	}

	products := []lodging.Product{}
	segHolderIndex := holderIndex
	segHolder := allGuests[segHolderIndex]
	if seqNum > 0 {
		segHolderIndex = gofakeit.Number(0, len(allGuests)-1)
		segHolder = allGuests[segHolderIndex]
	}
	for i := 0; i < gofakeit.Number(1, nGuests); i++ {
		products = append(products, BuildHotelProduct())
	}
	segment := lodging.BookingSegment{
		ID:                 gofakeit.Regex("[0-9A-Z]{10}"),
		HotelCode:          hotelCode,
		NNights:            nNights,
		NGuests:            int64(nGuests),
		StartDate:          startDate,
		Products:           products,
		Holder:             segHolder,
		PaymentInformation: BuildPaymentInformation(),
		GroupId:            groupID,
		Status:             BuildBookingStatus(),
		Price:              BuildBookingPrice(nNights, startDate, products, currency),
		AdditionalGuests:   []lodging.GuestProfile{},
	}

	for i := 0; i < gofakeit.Number(0, 3); i++ {
		segment.Products = append(segment.Products, BuildHotelProduct())
	}
	guestsWithoutHolder := append(allGuests[:segHolderIndex], allGuests[segHolderIndex+1:]...)
	guest, remainingGuests, hasGuestLeft := pickGuest(guestsWithoutHolder)
	for i := 0; i < nGuests-1; i++ {
		guest, remainingGuests, hasGuestLeft = pickGuest(remainingGuests)
		if !hasGuestLeft {
			break
		}
		segment.AdditionalGuests = append(segment.AdditionalGuests, guest)
	}
	return segment
}

// ipck geusts and returns the remaning unpicked guest
func pickGuest(remaningGuests []lodging.GuestProfile) (lodging.GuestProfile, []lodging.GuestProfile, bool) {
	if len(remaningGuests) == 0 {
		return lodging.GuestProfile{}, []lodging.GuestProfile{}, false
	}
	i := gofakeit.Number(0, len(remaningGuests)-1)
	guest := remaningGuests[i]
	return guest, append(remaningGuests[:i], remaningGuests[i+1:]...), true
}

func ComputeStatus(segs []lodging.BookingSegment) string {
	allCancelled := true
	for _, seg := range segs {
		if seg.Status != lodging.STATUS_CANCELLED {
			allCancelled = false
		}
	}
	if allCancelled {
		return lodging.STATUS_CANCELLED
	}
	return lodging.STATUS_CONFIRMED
}
func ComputeStartDate(segs []lodging.BookingSegment) string {
	date := time.Now()
	for _, seg := range segs {
		startDate, _ := time.Parse(core.DATE_FORMAT, seg.StartDate)
		if startDate.Before(date) {
			date = startDate
		}
	}
	return date.Format(core.DATE_FORMAT)
}
func ComputeEndDate(segs []lodging.BookingSegment) string {
	date := time.Time{}
	for _, seg := range segs {
		start, _ := time.Parse(core.DATE_FORMAT, seg.StartDate)
		endDate := start.AddDate(0, 0, int(seg.NNights))
		if endDate.After(date) {
			date = endDate
		}
	}
	return date.Format(core.DATE_FORMAT)
}
func ComputeNGuests(segs []lodging.BookingSegment) int64 {
	nGuests := int64(0)
	for _, seg := range segs {
		if seg.NGuests > nGuests {
			nGuests = seg.NGuests
		}
	}
	return nGuests
}

func ComputeNNights(segs []lodging.BookingSegment) int64 {
	start, _ := time.Parse(core.DATE_FORMAT, ComputeStartDate(segs))
	end, _ := time.Parse(core.DATE_FORMAT, ComputeEndDate(segs))
	return int64(end.Sub(start).Hours() / 24)
}

func BuildCancelReason(lastUpdatedTime time.Time) common.CancellationReason {
	return common.CancellationReason{
		Reason:  gofakeit.RandomString([]string{"plan_changed", "mobile", "ota-anycompany", "ota-booking", "gds"}),
		Comment: BuildComment(lastUpdatedTime, "cancellation_reason", "canccellation reason", ""),
	}
}

func BuildExternalID() common.ExternalID {
	return common.ExternalID{
		ID:     gofakeit.Regex("[0-9A-Z]{10}"),
		IdName: "booking_id",
		OriginatingSystem: gofakeit.RandomString([]string{
			lodging.ORIGINATING_SYSTEN_PMS,
			lodging.ORIGINATING_SYSTEN_CRS,
			lodging.ORIGINATING_SYSTEN_OTA,
			lodging.ORIGINATING_SYSTEN_GDS,
		}),
	}
}

func BuildComment(createdOn time.Time, typ string, title string, text string) common.Comment {
	if title == "" {
		title = gofakeit.Sentence(20)
	}
	if text == "" {
		text = gofakeit.Sentence(100)
	}
	return common.Comment{
		Type:                 typ,
		Language:             common.Language{Code: gofakeit.LanguageAbbreviation()},
		Title:                title,
		Text:                 text,
		CreatedDateTime:      createdOn,
		CreatedBy:            gofakeit.Name(),
		LastModifiedDateTime: createdOn,
		LastModifiedBy:       gofakeit.Name(),
	}
}

func BuildBookingStatus() string {
	return gofakeit.RandomString([]string{
		lodging.STATUS_CONFIRMED,
		lodging.STATUS_CANCELLED,
		lodging.STATUS_WAITLISTED,
		lodging.STATUS_IN_HOUSE,
		lodging.STATUS_CHECKED_OUT})
}

func BuildBookingPrice(nNights int64, startDate string, products []lodging.Product, currency common.Currency) lodging.BookingPrice {
	nights := createNightsArray(startDate, nNights)
	ppn := BuildPricePerNights(nights, products, currency)
	tpn := BuildTaxPerNights(ppn, currency)
	pps := BuildPricePerStay(ppn)
	tps := BuildTaxPerStay(tpn)
	br := BuildBusinessRules()
	tppnbt := BuildTotalPricePerNightBeforeTaxes(ppn)
	tppnat := BuildTotalPricePerNightAfterTaxes(ppn, tpn)
	tpppbt := BuildTotalPricePerProductBeforeTaxes(ppn)
	tpppat := BuildTotalPricePerProductAfterTaxes(ppn, tpn)
	tbt := BuildTotalTotalBeforeTax(ppn)
	tat := BuildTotalTotalAfterTax(ppn, tpn)
	return lodging.BookingPrice{
		PricePerNight:                   ppn,
		TaxePerNight:                    tpn,
		PricePerStay:                    pps,
		TaxPerStay:                      tps,
		BusinessRules:                   br,
		TotalPricePerNightBeforeTaxes:   tppnbt,
		TotalPricePerNightAfterTaxes:    tppnat,
		TotalPricePerProductBeforeTaxes: tpppbt,
		TotalPricePerProductAfterTaxes:  tpppat,
		TotalBeforeTax:                  tbt,
		TotalAfterTax:                   tat,
	}
}

func createNightsArray(startDate string, nNights int64) []string {
	dates := []string{}
	parseTime, err := time.Parse(core.DATE_FORMAT, startDate)
	if err != nil {
		log.Printf("Warning: invalid time format %v", startDate)
	}
	dates = append(dates, startDate)
	for i := 1; i < int(nNights); i++ {
		dates = append(dates, parseTime.AddDate(0, 0, i).Format(core.DATE_FORMAT))
	}
	return dates
}

func BuildPricePerNights(nights []string, products []lodging.Product, currency common.Currency) []lodging.AmountPerNightItem {
	ppn := []lodging.AmountPerNightItem{}
	for i, night := range nights {
		app := []lodging.AmountPerProductItem{}
		for j, product := range products {
			app = append(app, lodging.AmountPerProductItem{
				ProductID:          product.ID,
				ProductDescription: product.Description,
				Amount:             BuildPrice(product),
				Label:              "Nightly price for product " + strconv.Itoa(j) + " (" + product.ID + ")",
				Currency:           currency,
			})
		}
		ppn = append(ppn, lodging.AmountPerNightItem{
			Date:             night,
			Label:            "Prices for night " + strconv.Itoa(i) + " (" + night + ")",
			AmountPerProduct: app,
			Currency:         currency,
		})
	}
	return ppn
}

func BuildTaxPerNights(ppn []lodging.AmountPerNightItem, currency common.Currency) []lodging.AmountPerNightItem {
	tpn := []lodging.AmountPerNightItem{}
	for i, p := range ppn {
		tpp := []lodging.AmountPerProductItem{}
		for j, app := range p.AmountPerProduct {
			tpp = append(tpp, lodging.AmountPerProductItem{
				ProductID: app.ProductID,
				Amount:    app.Amount * 0.15,
				Label:     "15 percent city and sales Taxes for product " + strconv.Itoa(j) + " (" + app.ProductID + ")",
				Currency:  currency,
			})
		}
		tpn = append(tpn, lodging.AmountPerNightItem{
			Date:             p.Date,
			Label:            "Prices for night " + strconv.Itoa(i) + " (" + p.Date + ")",
			AmountPerProduct: tpp,
			Currency:         currency,
		})
	}
	return tpn
}

func BuildPricePerStay(ppn []lodging.AmountPerNightItem) []lodging.AmountPerProductItem {
	return []lodging.AmountPerProductItem{}
}
func BuildTaxPerStay(ppn []lodging.AmountPerNightItem) []lodging.AmountPerProductItem {
	return []lodging.AmountPerProductItem{}
}
func BuildBusinessRules() []lodging.BusinessRule {
	return []lodging.BusinessRule{}
}
func BuildTotalPricePerNightBeforeTaxes(ppn []lodging.AmountPerNightItem) []lodging.TotalAmountPerNightItem {
	return []lodging.TotalAmountPerNightItem{}
}
func BuildTotalPricePerNightAfterTaxes(ppn []lodging.AmountPerNightItem, tpn []lodging.AmountPerNightItem) []lodging.TotalAmountPerNightItem {
	return []lodging.TotalAmountPerNightItem{}
}
func BuildTotalPricePerProductBeforeTaxes(ppn []lodging.AmountPerNightItem) []lodging.TotalAmountPerProductItem {
	return []lodging.TotalAmountPerProductItem{}
}
func BuildTotalPricePerProductAfterTaxes(ppn []lodging.AmountPerNightItem, tpn []lodging.AmountPerNightItem) []lodging.TotalAmountPerProductItem {
	return []lodging.TotalAmountPerProductItem{}
}
func BuildTotalTotalBeforeTax(ppn []lodging.AmountPerNightItem) core.Float {
	total := core.Float(0.0)
	for _, p := range ppn {
		for _, app := range p.AmountPerProduct {
			total += core.Float(app.Amount)
		}
	}
	return total
}
func BuildTotalTotalAfterTax(ppn []lodging.AmountPerNightItem, tpn []lodging.AmountPerNightItem) core.Float {
	total := core.Float(0.0)
	for _, p := range ppn {
		for _, app := range p.AmountPerProduct {
			total += core.Float(app.Amount)
		}
	}
	for _, p := range tpn {
		for _, app := range p.AmountPerProduct {
			total += core.Float(app.Amount)
		}
	}
	return total
}

func BuildPrice(product lodging.Product) core.Float {
	rooms := map[string][]float64{
		"SUITE": []float64{400.0, 2000.0},
		"DBL":   []float64{100.0, 300.0},
		"KNG":   []float64{200.0, 600.0},
	}
	ratePlans := map[string]float64{
		"BAR":         1.0,
		"SUPER_SAVER": 0.6,
		"XMAS_WEEK":   1.8,
	}
	attributes := map[string]float64{
		"BALCONY":  0.1,
		"MINI_BAR": 0.05,
		"SEA_VIEW": 2.0,
	}
	addOns := map[string][]float64{
		"BKFT":    []float64{10.0, 40.0},
		"TICKETS": []float64{100.0, 1000.0},
		"SPA":     []float64{200.0, 400.0},
	}
	price := 0.0
	price += gofakeit.Float64Range(rooms[product.RoomType.Code][0], rooms[product.RoomType.Code][1]) * ratePlans[product.RatePlan.Code]
	for _, att := range product.Attributes {
		price += price * attributes[att.Code]
	}
	for _, add := range product.AddOns {
		price += gofakeit.Float64Range(addOns[add.Code][0], addOns[add.Code][1])
	}
	return core.Float(price)
}

func BuildHotelProduct() lodging.Product {
	rooms := map[string]lodging.RoomType{
		"SUITE": lodging.RoomType{Code: "SUITE", Name: "Suite", Description: "Luxury suite for a comfortable stays"},
		"DBL":   lodging.RoomType{Code: "DBL", Name: "Double room", Description: "Room with Double bed"},
		"KNG":   lodging.RoomType{Code: "KNG", Name: "King Room", Description: "Room with King Size Bed"},
	}
	ratePlans := map[string]lodging.RatePlan{
		"BAR":         lodging.RatePlan{Code: "BAR", Name: "Best available rate", Description: "Our best available rate all year"},
		"SUPER_SAVER": lodging.RatePlan{Code: "SUPER_SAVER", Name: "Super Saver rate", Description: "Lower price for advanced purchase"},
		"XMAS_WEEK":   lodging.RatePlan{Code: "XMAS_WEEK", Name: "Xmas special rate", Description: "Special rate for the chrismas season"},
	}
	attributes := map[string]lodging.Attribute{
		"BALCONY":  lodging.Attribute{Code: "BALCONY", Name: "Balcony", Description: "Large balcony with chairs and a table"},
		"MINI_BAR": lodging.Attribute{Code: "MINI_BAR", Name: "Mini bar", Description: "Mini bar with local snacks and beverages"},
		"SEA_VIEW": lodging.Attribute{Code: "SEA_VIEW", Name: "Sea View", Description: "Full frontal view over the pacfic ocean"},
	}
	addOns := map[string]lodging.AddOn{
		"BKFT":    lodging.AddOn{Code: "BKFT", Name: "Suite", Description: "Our famous continental breakfast"},
		"TICKETS": lodging.AddOn{Code: "TICKETS", Name: "Suite", Description: "Tickets for our world reknown local theater"},
		"SPA":     lodging.AddOn{Code: "SPA", Name: "Suite", Description: "Luxury Spa session "},
	}

	product := lodging.Product{
		ID:         gofakeit.Regex("[0-9A-Z]{10}"),
		RoomType:   rooms[gofakeit.RandomString([]string{"SUITE", "DBL", "KNG"})],
		RatePlan:   ratePlans[gofakeit.RandomString([]string{"BAR", "SUPER_SAVER", "XMAS_WEEK"})],
		Attributes: []lodging.Attribute{},
		AddOns:     []lodging.AddOn{},
	}
	//map to ensure we don't add attributes twice. TODO: move this to a helper function
	selected := map[string]bool{}
	attNames := []string{}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		attrCode := gofakeit.RandomString([]string{"BALCONY", "MINI_BAR", "SEA_VIEW"})
		if !selected[attrCode] {
			attNames = append(attNames, attributes[attrCode].Name)
			product.Attributes = append(product.Attributes, attributes[attrCode])
			selected[attrCode] = true
		}
	}
	for i := 0; i < gofakeit.Number(0, 3); i++ {
		product.AddOns = append(product.AddOns, addOns[gofakeit.RandomString([]string{"BKFT", "TICKETS", "SPA"})])
	}

	product.Description = product.RoomType.Name + " with " + strings.Join(attNames, ",") + " ( " + product.RatePlan.Name + ")"
	return product
}
