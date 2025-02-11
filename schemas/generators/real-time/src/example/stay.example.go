package example

import (
	"encoding/json"
	"strconv"
	"tah/upt/schemas/src/tah-common/common"
	"tah/upt/schemas/src/tah-common/core"
	"tah/upt/schemas/src/tah-common/lodging"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateHotelStay(hotelBooking lodging.Booking) ([]byte, string, error) {
	stay := CreateStayExample(hotelBooking)
	bytes, err := json.Marshal(stay)
	return bytes, buildStayPath(stay), err
}

func buildStayPath(stay lodging.Stay) string {
	return stay.LastUpdatedOn.Format("2006/01/02/15") + "/" + stay.ID + ".json"
}

func CreateStayExample(hotelBooking lodging.Booking) lodging.Stay {
	checkIn := hotelBooking.CheckIn()
	checkOut := hotelBooking.CheckOut()
	createdOn := gofakeit.DateRange(checkIn, checkOut)
	updatedOn := gofakeit.DateRange(createdOn, checkOut)
	stay := lodging.Stay{
		ObjectVersion: gofakeit.Number(1, 20),
		ID:            gofakeit.Regex("[0-9A-Z]{10}"),
		ModelVersion:  core.MODEL_VERSION,
		LastUpdatedOn: updatedOn,
		CreatedOn:     createdOn,
		LastUpdatedBy: gofakeit.Name(),
		CreatedBy:     gofakeit.Name(),
		ReservationID: hotelBooking.ID,
		GuestID:       hotelBooking.Holder.ID,
		Currency:      hotelBooking.Currency,
		FirstName:     hotelBooking.Holder.FirstName,
		LastName:      hotelBooking.Holder.LastName,
		Email:         chooseEmail(hotelBooking.Holder.Emails),
		Phone:         choosePhone(hotelBooking.Holder.Phones),
		StartDate:     hotelBooking.StartDate,
		HotelCode:     hotelBooking.HotelCode,
		Revenue:       []lodging.StayRevenueItem{},
	}
	for _, seg := range hotelBooking.Segments {
		for _, p := range seg.Price.PricePerNight {
			date, _ := time.Parse(core.DATE_FORMAT, p.Date)
			for _, app := range p.AmountPerProduct {
				stay.Revenue = append(stay.Revenue, lodging.StayRevenueItem{
					Type:        "Room Charge " + p.Date,
					Description: app.ProductDescription,
					Currency:    hotelBooking.Currency,
					Amount:      app.Amount,
					Date:        date,
				})
			}
		}
		for i := 0; i < gofakeit.Number(0, 10); i++ {
			stay.Revenue = append(stay.Revenue, lodging.StayRevenueItem{
				Type:     gofakeit.RandomString([]string{"snack bar", "just-walk-out-purchase", "restaurant", "sodas"}),
				Currency: hotelBooking.Currency,
				Amount:   core.Float(gofakeit.Float64Range(5.0, 100.0)),
				Date:     gofakeit.DateRange(checkIn, checkOut),
			})
		}
	}

	return stay
}

func chooseEmail(emails []common.Email) string {
	if len(emails) > 0 {
		return emails[0].Address
	}
	return ""
}
func choosePhone(phones []common.Phone) string {
	if len(phones) > 0 {
		return strconv.Itoa(int(phones[0].CountryCode)) + phones[0].Number
	}
	return ""
}
