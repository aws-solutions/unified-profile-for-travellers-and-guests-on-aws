package example

import (
	"encoding/json"
	"tah/upt/schemas/src/tah-common/air"

	"github.com/brianvoe/gofakeit/v6"
)

func GeneratePassengerProfiles(nLines int) []air.PassengerProfile {
	paxs := []air.PassengerProfile{}
	nEventsPerFile := nLines
	nTx := gofakeit.Number(1, 10)
	for i := 0; i < nEventsPerFile; i++ {
		paxs = append(paxs, CreatePassengerProfileExample(nTx))
	}
	return paxs
}

func SerializePassengers(paxs []air.PassengerProfile) ([]byte, string, error) {
	file := ""
	for _, pax := range paxs {
		bytes, err := json.Marshal(pax)
		if err != nil {
			return []byte(file), buildFilePath(), err
		}
		file += string(bytes) + "\n"
	}
	return []byte(file), buildFilePath(), nil
}

func CreatePassengerProfileExample(nTx int) air.PassengerProfile {
	return Passenger(nTx)
}
