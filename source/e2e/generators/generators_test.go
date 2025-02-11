// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package generators

import (
	"testing"
)

func TestGenerateAirBookingRecords(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleAirBookingRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating air booking records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}

func TestGenerateClickstreamRecords(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleClickstreamRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating clickstream records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}
func TestGenerateCsiRecords(t *testing.T) {
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleCsiRecords(domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating csi records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}

func TestGenerateHotelBookingRecords(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleHotelBookingRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating hotel booking records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}

func TestGenerateHotelStayRecords(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleHotelStayRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating hotel stay records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}

func TestGenerateGuestProfile(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimpleGuestProfileRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating guest profile records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}

func TestGeneratePaxProfile(t *testing.T) {
	travellerId := "test_id"
	domainName := "test_domain"
	numRecs := 5
	res, err := GenerateSimplePaxProfileRecords(travellerId, domainName, numRecs)
	if err != nil {
		t.Errorf("[%v] error generating pax profile records: %v", t.Name(), err)
	}
	if len(res) != numRecs {
		t.Errorf("[%v] expected %v records, found %v", t.Name(), numRecs, len(res))
	}
}
