package main

import (
	"encoding/json"
	"log"
	air "tah/upt/schemas/src/tah-common/air"
	core "tah/upt/schemas/src/tah-common/core"
	lodging "tah/upt/schemas/src/tah-common/lodging"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	log.Printf("Testing AWS T&H Common package version " + core.MODEL_VERSION)
	airBooking := air.Booking{}
	now := time.Now()
	hotelBooking := lodging.Booking{ID: "hotel_booking_id", LastUpdatedOn: now}
	hotelStay := lodging.Stay{ID: "hotel_stay_id", LastUpdatedOn: now}
	guestProfile := lodging.GuestProfile{ID: "hotel_profile_id", LastUpdatedOn: now}
	log.Printf("Business object: %+v", airBooking)
	log.Printf("Business object: %+v", hotelBooking)
	log.Printf("Business object: %+v", hotelStay)
	log.Printf("Business object: %+v", guestProfile)

	//Test importable interface
	if hotelBooking.DataID() != "hotel_booking_id" {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
	if hotelStay.DataID() != "hotel_stay_id" {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
	if guestProfile.DataID() != "hotel_profile_id" {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
	if hotelBooking.LastUpdaded() != now {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
	if hotelStay.LastUpdaded() != now {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
	if guestProfile.LastUpdaded() != now {
		t.Errorf("[models] invalid implementation of importable inteface")
	}
}

type TestSerializable struct {
	Value core.Float `json:"value"`
}

func TestSerializationFloats(t *testing.T) {
	val, _ := json.Marshal(TestSerializable{Value: 1.2345678})
	str := string(val)
	if str != `{"value":1.2345678}` {
		t.Errorf("[TestSerializationFloats] error serializing floats %v", str)
	}
	val, _ = json.Marshal(TestSerializable{Value: 1})
	str = string(val)
	if str != `{"value":1.0}` {
		t.Errorf("[TestSerializationFloats] error serializing floats %v", str)
	}
}
