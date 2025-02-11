package example

import (
	"encoding/json"
	"log"
	"tah/upt/schemas/src/tah-common/lodging"

	"github.com/brianvoe/gofakeit/v6"
)

func GenerateGuests(nLines int, simpleGuest bool) []lodging.GuestProfile {
	guests := []lodging.GuestProfile{}
	nEventsPerFile := nLines
	nTx := gofakeit.Number(1, 10)
	for i := 0; i < nEventsPerFile; i++ {
		guests = append(guests, CreateGuestProfileExample(nTx, simpleGuest))
	}
	return guests
}

func SerializeGuest(guests []lodging.GuestProfile) ([]byte, string, error) {
	file := ""
	for i, guest := range guests {
		log.Printf("serializing guest: %+v", i)
		bytes, err := json.Marshal(guest)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func buildGuestPath(guest lodging.GuestProfile) string {
	return guest.LastUpdatedOn.Format("2006/01/02/15") + "/" + guest.ID + ".json"

}

func CreateGuestProfileExample(nTx int, simpleGuest bool) lodging.GuestProfile {
	return Guest(nTx, simpleGuest)
}
