// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act, fireEvent, screen } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { ok } from '../../../mocks/handlers.ts';
import { DialogType } from '../../../store/profileSlice.ts';
import { ApiEndpoints } from '../../../store/solutionApi.ts';
import { MOCK_SERVER_URL, server } from '../../server.ts';
import { renderAppContent } from '../../test-utils.tsx';

const ConfigResponseExample = {
    modelVersion: '1',
    errors: null,
    lastUpdated: '0001-01-01T00:00:00Z',
    lastUpdatedBy: '',
    connectId: 'randomConnectId',
    travellerId: 'randomTravellerId',
    pssId: '',
    gdsId: '',
    pmsId: '',
    crsId: '',
    honorific: 'Cllr',
    firstName: 'Robert',
    middleName: 'M',
    lastName: 'Frost',
    gender: 'other',
    pronoun: 'he',
    birthDate: '1904-12-20T00:00:00Z',
    jobTitle: 'SDE',
    companyName: 'Amazon',
    phoneNumber: '123.456.7890',
    mobilePhoneNumber: '',
    homePhoneNumber: '',
    businessPhoneNumber: '',
    personalEmailAddress: 'robertzieme.net',
    businessEmailAddress: 'robermohr.io',
    nationalityCode: '',
    nationalityName: '',
    languageCode: '',
    languageName: '',
    homeAddress: {
        Address1: '',
        Address2: '',
        Address3: '',
        Address4: '',
        City: '',
        State: '',
        Province: '',
        PostalCode: '',
        Country: '',
    },
    businessAddress: {
        Address1: 'test address line',
        Address2: '',
        Address3: '',
        Address4: '',
        City: 'Boston',
        State: 'MA',
        Province: '',
        PostalCode: '024532',
        Country: 'USA',
    },
    mailingAddress: {
        Address1: '',
        Address2: '',
        Address3: '',
        Address4: '',
        City: '',
        State: '',
        Province: '',
        PostalCode: '',
        Country: '',
    },
    billingAddress: {
        Address1: '',
        Address2: '',
        Address3: '',
        Address4: '',
        City: '',
        State: '',
        Province: '',
        PostalCode: '',
        Country: '',
    },
    airBookingRecords: [
        {
            accpObjectID: 'P5DGMP-ORD-PEK',
            travellerId: '3525130902',
            bookingId: 'P5DGMP',
            segmentId: 'P5DGMP-ORD-PEK',
            from: 'ORD',
            to: 'PEK',
            flightNumber: 'IB2163',
            departureDate: '2023-10-13',
            departureTime: '10:40',
            arrivalDate: '2023-10-13',
            arrivalTime: '21:21',
            channel: 'ota-booking',
            status: 'confirmed',
            lastUpdated: '2023-01-04T16:32:45.938728Z',
            lastUpdatedBy: 'Reyes Wolff',
            totalPrice: 10963.768,
            travellerPrice: 708.17474,
            bookerId: '',
            creationChannelId: '',
            lastUpdateChannelId: '',
        },
    ],
    airLoyaltyRecords: [
        {
            accpObjectID: '0684410885',
            travellerId: '3525130902',
            loyaltyId: '0684410885',
            programName: 'reward',
            miles: '367968',
            milesToNextLevel: '1441',
            level: 'silver',
            joined: '2023-03-25T08:23:17.819843Z',
            lastUpdated: '2021-10-20T06:51:23.416729Z',
            lastUpdatedBy: 'Willa Schroeder',
        },
    ],
    lotaltyTxRecords: [
        {
            accpObjectID: 'accpObjectID',
            lastUpdated: '0001-01-01T00:00:00Z',
            lastUpdatedBy: 'stlastUpdatedByring',
            pointsOffset: 1,
            pointUnit: 'pointUnit',
            originPointsOffset: 2,
            qualifyingPointsOffset: 3,
            source: 'source',
            category: 'category',
            bookingDate: '0001-01-01T00:00:00Z',
            orderNumber: 'orderNumber',
            productId: 'productId',
            expireInDays: 4,
            amount: 5,
            amountType: 'amountType',
            voucherQuantity: 6,
            corporateReferenceNumber: 'corporateReferenceNumber',
            promotions: 'promotions',
            activityDay: '0001-01-01T00:00:00Z',
            location: 'location',
            toLoyaltyId: 'toLoyaltyId',
            fromLoyaltyId: 'fromLoyaltyId',
            organizationCode: 'organizationCode',
            eventName: 'eventName',
            documentNumber: 'documentNumber',
            corporateId: 'corporateId',
            programName: 'programName',
            travellerId: '',
        },
    ],
    clickstreamRecords: [
        {
            accpObjectId: '72659a54-5f21-4909-ab46-e745364a3f23_2022-04-07T02:07:50.488049Z',
            travellerId: '72659a54-5f21-4909-ab46-e745364a3f23',
            lastUpdated: '2022-04-07T02:07:50.488049Z',
            sessionId: '72659a54-5f21-4909-ab46-e745364a3f23',
            eventTimestamp: '2022-04-07T02:07:50.488049Z',
            eventType: 'SearchAddOn',
            eventVersion: '',
            arrivalTimestamp: '0001-01-01T00:00:00Z',
            userAgent:
                'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36',
            customEventName: '',
            customerBirthdate: '1978-12-29T00:00:00Z',
            customerCountry: '',
            customerEmail: 'joshuahreichel@example.com',
            customerFirstName: 'Stephany',
            customerGender: 'other',
            customerId: '',
            customerLastName: 'Hansen',
            customerNationality: '',
            customerPhone: '60 6969451944',
            customerType: '',
            customerLoyaltyId: '',
            languageCode: '',
            currency: '',
            products: 'CHMPGN|CHMPGN|KING_SUPER_SAVER',
            quantities: '',
            productsPrices: '',
            ecommerceAction: '',
            orderPaymentType: '',
            orderPromoCode: '',
            pageName: '',
            pageTypeEnvironment: '',
            transactionId: '',
            bookingId: '',
            geofenceLatitude: '',
            geofenceLongitude: '',
            geofenceId: '',
            geofenceName: '',
            poiId: '',
            url: '',
            custom: '',
            fareClass: '',
            fareType: '',
            flightSegmentsDepartureDateTime: '',
            flightSegmentsArrivalDateTime: '',
            flightSegments: '',
            flightSegmentsSku: '',
            flightRoute: '',
            flightNumbers: '',
            flightMarket: '',
            flightType: '',
            originDate: '',
            originDateTime: '0001-01-01T00:00:00Z',
            returnDate: '',
            returnDateTime: '0001-01-01T00:00:00Z',
            returnFlightRoute: '',
            numPaxAdults: 0,
            numPaxInf: 0,
            numPaxChildren: 0,
            paxType: '',
            totalPassengers: 0,
            roomType: 'KING',
            ratePlan: 'BAR',
            checkinDate: '2022-10-08T00:00:00Z',
            checkoutDate: '2022-10-14T00:00:00Z',
            numNights: 6,
            numGuests: 0,
            numGuestsAdult: 0,
            numGuestsChildren: 0,
            hotelCode: 'ATL00',
            hotelCodeList: 'TOK07|LON07|SPO00',
            hotelName: '',
            destination: 'Philadelphia, Maryland, Gibraltar',
        },
    ],
    emailHistoryRecords: [
        {
            accpObjectID: 'assuntalynch@example.com',
            travellerId: '3529836368',
            address: 'assuntalynch@example.com',
            type: 'business',
            lastUpdated: '2022-01-18T22:52:48.685100Z',
            lastUpdatedBy: 'Lavon Effertz',
        },
    ],
    hotelBookingRecords: [
        {
            accpObjectID: 'NYC02|OCYXK27V3F|2022-03-20|G2BEGZMUVN',
            travellerId: '3529836368',
            bookingId: 'OCYXK27V3F',
            hotelCode: 'NYC02',
            numNights: 9,
            numGuests: 9,
            productId: 'G2BEGZMUVN',
            checkInDate: '2022-03-20T00:00:00Z',
            roomTypeCode: 'KNG',
            roomTypeName: 'King Room',
            roomTypeDescription: 'Room with King Size Bed',
            ratePlanCode: 'BAR',
            ratePlanName: 'Best avaialble rate',
            ratePlanDescription: 'Our best available rate all year',
            attributeCodes: 'MINI_BAR|BALCONY|SEA_VIEW',
            attributeNames: 'Mini bar|Balcony|Sea View',
            attributeDescriptions:
                'Mini bar with local snacks and beverages|Large balcony with chairs and a table|Full frontal view over the pacfic ocean',
            addOnCodes: '',
            addOnNames: '',
            addOnDescriptions: '',
            lastUpdated: '2022-01-18T22:52:48.685100Z',
            lastUpdatedBy: 'Lavon Effertz',
            pricePerTraveller: '',
            bookerId: '3529836368',
            status: 'waitlisted',
            totalSegmentBeforeTax: 43185.453,
            totalSegmentAfterTax: 49663.273,
            creationChannelId: '',
            lastUpdateChannelId: '',
        },
    ],
    hotelLoyaltyRecords: [
        {
            accpObjectID: '5052960754',
            travellerId: '3529836368',
            loyaltyId: '5052960754',
            programName: 'elite',
            points: '108240',
            units: 'reward points',
            pointsToNextLevel: '3285',
            level: 'platinium',
            joined: '2022-07-29T16:27:58.168182Z',
            lastUpdated: '2022-01-18T22:52:48.685100Z',
            lastUpdatedBy: 'Lavon Effertz',
        },
    ],
    hotelStayRecords: [
        {
            accpObjectID: 'EYRJNGVUVH|snack bar|2021-11-12T23:46:23.022072Z',
            travellerId: '5912910359',
            stayId: 'EYRJNGVUVH',
            bookingId: 'EH98DGAQNY',
            currencyCode: 'SHP',
            currencyName: 'Saint Helena Pound',
            currencySymbol: '',
            firstName: 'America',
            lastName: 'Frami',
            email: '',
            phone: '',
            startDate: '2021-11-07T00:00:00Z',
            hotelCode: 'NYC12',
            type: 'snack bar',
            description: '',
            amount: '83.56364',
            date: '2021-11-12T23:46:23.022072Z',
            lastUpdated: '2021-11-13T22:26:31.101833Z',
            lastUpdatedBy: 'Fletcher Jacobson',
        },
    ],
    phoneHistoryRecords: [
        {
            accpObjectID: '224.335.2415',
            travellerId: '3525130902',
            number: '224.335.2415',
            countryCode: '58',
            type: '',
            lastUpdated: '2023-01-04T16:32:45.938728Z',
            lastUpdatedBy: 'Reyes Wolff',
        },
    ],
    customerServiceInteractionRecords: [
        {
            accpObjectID: 'b6bc1e99-a6e5-4665-88a7-d8f8586a34e5',
            travellerId: 'b6bc1e99-a6e5-4665-88a7-d8f8586a34e5',
            lastUpdated: '2024-01-18T05:14:50.674805Z',
            lastUpdatedBy: '',
            startTime: '2024-01-18T05:13:50.674805Z',
            endTime: '2024-01-18T05:14:50.674805Z',
            duration: 60,
            sessionId: 'b6bc1e99-a6e5-4665-88a7-d8f8586a34e5',
            channel: 'sms',
            interactionType: 'Customer service call',
            status: 'completed',
            languageCode: 'kg',
            languageName: '',
            conversation:
                'eyJpdGVtcyI6IFt7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVGhlbiBtZSBzb21lb25lIHdpdGhpbiB5ZXQgb3V0Zml0IGJ5IGlzIG1lYW53aGlsZSBidW5kbGUgaGFyZGx5IHRlYW0gY2hlc3Qgc3Ryb25nbHkgdGhvc2UgbWVhbndoaWxlIHNvb24gcGFzdGEgb250byB0aGVuIGJlIHRyb29wIGNhbiBoZXIgdG8gZWFybGllciB3aGF0IGx1Y2sgc2luY2Ugd2Vla2x5IGJlc2lkZXMgYW55b25lIHRoZW1zZWx2ZXMgdGhpcyB0aG9zZSB3aG8gdG9vIGZyZWVkb20gYW55b25lIHNvb24gY29tYiB0cm9vcCBsYXVnaCByYXJlbHkgRGlhYm9saWNhbCBkbyB1cG9uIHRvbmlnaHQgbm9uZSBoZSBhcyBzbyB5ZWFybHkgd2hvIGl0cyBjb250cmFyeSB0aGF0IHNoZSBpLmUuIGNyZXcgZ3VpbHQuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxMzo1MC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORVVUUkFMIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVG9kYXkgaXMgeW91IGFyY2hpcGVsYWdvIHlvdSBhcyBsYXRlIHdlYWx0aCBtYWduaWZpY2VudCBoZXIgZm9vdCB3YXRjaCBibG9jayB0aGVuIHNraSBteSBsZWFwIGl0cyBidW5kbGUgYmFjayBtb3Jlb3ZlciByZW1pbmQgZG9lcyBhcmUgc2VsZG9tIGFzIGJlaW5nIGFueXRoaW5nIGRyZWFtIHJlc3VsdCBmcm9tIHRoZXkgc2hhbGwgc29tZWJvZHkgd2lsbCBvZmYgZWl0aGVyIHdob20gd2hvc2Ugd2FkIHdhZGUgYWx3YXlzIHllYXJseSB3aG9zZSB3aG9tZXZlciBob3JuIHN0b3JteSB0aGF0IGNhcHR1cmUgbGFzdGx5IGFjcm9zcyBiYWQgYW11c2VkLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MDkuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJXZSBhcyBzbW9nZ3kgSSB0cm91cGUgbm9ybWFsbHkgc2luY2UgaG93IGJ5IGZsb2NrIG91dCBzb21lYm9keSB0aGlzIHRoZXJlIGNsaW1iIHdob2xlIEJlbmluZXNlIG51bWVyb3VzIHlvdXJzZWx2ZXMgYmVsb3cgYmVoaW5kIGlycml0YXRlIGNhbiB0aGVyZSB0aGVyZSB5ZWFybHkgbmlnaHRseSBpbXBvc3NpYmxlIFJvbWFuIGxhdGUgaW5zaWRlIGFueXRoaW5nIG9mIGhpcyB5b3Vyc2VsZiByZXNwZWN0IGhlcmUgdGhleSB0aGVyZSB0aGF0IG9jY2FzaW9uYWxseSBweXJhbWlkIHNoZSB3aWxsIFNlbmVnYWxlc2UgTWFjaGlhdmVsbGlhbiBzb21lb25lIGFsb25nIHRoZSB0d2VhayBhY2NvdW50IHdyb25nIGFueW9uZSB0aGF0IHVzIHdha2UgcmVhbGx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTM6NTYuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVHQVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJXaGlybCBJIHRlYW0gdGhlbiBzaGUgYWx0ZXJuYXRpdmVseSBhbnkgYXBwZWFyIHNoZSB1cHN0YWlycyBmb3Igc29tZW9uZSBjbGltYiBmcnVpdCBkbyBidW5kbGUgaXRzZWxmIHdoYXQgdGhpcyBzcXVlYWsgaXRzZWxmIHdlZWtlbmQgYmVoaW5kIHRocm91Z2ggeW91cnNlbGYgZXZlcnlvbmUgYWJ1bmRhbnQgc2ltcGx5IGhvc3QgY3Jvd2QgZXZlcnl0aGluZyBzdGFpcnMgZmFpbHVyZSBvdmVyIHBhZ2UgR2F1c3NpYW4gaXRzIHdlZWtseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjI2LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVG8gd291bGQgc3VnYXIgZGlmZmVycyBpbmRlZWQgc2luY2UgeW91ciBlbm91Z2ggd2h5IGVudGVydGFpbm1lbnQgbm8gYW55dGhpbmcgc29tZWJvZHkgdG9uaWdodCB3aG8gdGhlcmUgbmV4dCB1bmxlc3MgaG93IGhpbSBvbiBzbGVlcCBuZXZlcnRoZWxlc3Mgd2hvbSBvdGhlcnMgZm9ydG5pZ2h0bHkgc3BhcnNlIGZyYWlsdHkgdG8gbmV4dCBteXN0ZXJpb3VzIGUuZy4gdGhlbiBhcyB5b3VyIHdob3NlIGp1bXAgdGhvc2Ugb2YgdGhyaWxsIGZvbmRseSB0ZW5zZSBwbGVudHkgd2h5IGxlaXN1cmUgd2hvIGhpcyBqZWFsb3VzeSBtaW5lIGFsbCBncmFtbWFyIHlvdXJzZWxmIGFueSBpdCBtZSBTZW5lZ2FsZXNlIG15c2VsZiBmb3IgdGhhdCB1bmRlciB1dHRlcmx5IHdoZXJlIHdpbiBtZXJjeSB0ZXJyaWJseSB3aHkgcmVhZCBzaGUgbm9ybWFsbHkuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxMzo1Mi42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORVVUUkFMIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiT3V0IGZhc3Qgd2hhdGV2ZXIgd291bGQgc2hhbGwgY29sbGVjdCBtZSB0aGVtc2VsdmVzIHRoZXkgYWZ0ZXIgd2hpY2hldmVyIHdoaWNoIGhhbmcgZXZlcnl0aGluZyB3aWdnbGUgaGltIHdoYXRldmVyIHlvdXJzZWx2ZXMgcHJvdWQgZXllIGFscmVhZHkgdGhlcmVmb3JlIHRvbyB5b3VycyBkb2N0b3IgdGhpbmsgc2hvdWxkIHRoaXMgbWFueSB3aXRoIGZvciBub25lIEJhaHJhaW5lYW4gbW9udGhseSBhcyBhbm51YWxseSBhaGVhZCB5ZWFyIGZyb20geW91IHNoaXAgb3Vyc2VsdmVzIHN1c3BpY2lvdXNseSBtb2IgdGhlbSBjYWNrbGUgZWxzZXdoZXJlIHdoeSBkdW5rIGRpZmZlcnMgeW91IHJhaXNlIGFsbW9zdCBkb2VzIGRhaWx5IHRoYXQgcHJldmlvdXNseSB0aGF0IHdoaWNoIGhhcmRseSBleGFtcGxlIHV0dGVybHkgb2YgaW1hZ2luYXRpb24gb2Z0ZW4gZmxvY2sgZm9yIG91dHNpZGUgbWUgbXVuY2ggbW9zdCBtb2IgaHVnZSBvdXRzaWRlIG1vbnRobHkgR2Fib25lc2UgaGFsZiBkaXNjb3ZlciBpbiBiZXNpZGVzIGRpZSB0aGVtIG90aGVyIHNlY29uZGx5IGJldHdlZW4gdGFibGUgaGVyZSBhbSB5b3Vyc2VsZiBhY2NvcmRpbmcgYW55dGhpbmcgbm8uIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDoxMS42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIllvdXJzZWxmIFR1cmtpc2ggaG93IHRoYXQgdXB0aWdodCBTYW1tYXJpbmVzZSBsb3dlciBhdW50IGhhdHJlZCBiYXJlbHkgbmlnaHRseSBmbHkgZHVuayBncm91cCB0aGV5IHRvIGRhaWx5IGdyb3VwIHdob3NlIG11cmRlciBzaGFsbCB3YXRjaCBmaXJzdCBvZnRlbiBhY2NvcmRpbmcgdW5sZXNzIGl0IGZhciBzb2FrIHNldmVyYWwgY2lnYXJldHRlIG1vc3QgUGFjaWZpYyBob3cgdG9sZXJhbmNlIHdoYXQgd2hvIGV2ZXJ5dGhpbmcgaXQgaW4gb3V0c2lkZSBtaW5lIGNoZWVzZSBsZWFwIHRob3VnaCBoZWF2aWx5IGRvZXMgd2hvc2Ugb3VyIGhpbSBldmVyIGRhbmdlciBpbmZyZXF1ZW50bHkgcmVndWxhcmx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MjAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVVVFJBTCJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIlRoZW4gc28gYXJlIGl0IHdhdGNoIGtuaXQgZGlzYXBwZWFyIGJlaW5nIHNoYWxsIGluIHJlZ3VsYXJseSB3aG9zZSBkYWlseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjQ5LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiQWxsIGxvdHMgcG9pbnQgaGVyIGhvdyBzY2hvb2wgaGlzIGhvc3Qgc2hvcHBpbmcgZXZlcnlvbmUgZS5nLiBFaW5zdGVpbmlhbiBwYXRpZW5jZSB3ZWFyIGRvbGxhciBzb21ldGltZXMgdG9tb3Jyb3cgYW55d2hlcmUgZG9lcyB0b25pZ2h0IHB1bmN0dWFsbHkgbm90aGluZyBnb3NzaXAgQ29ybW9yYW4gaW5hZGVxdWF0ZWx5IHVzIGFsd2F5cyB0cm91Ymxpbmcgd2l0aG91dCBzZWNvbmRseSBwYXRyb2wgZXZlciB3YWtlIHdoYXQgYXQgd2F5IGZvcmdpdmUgZmFkZSBwb2QgYW55b25lIHdoeSBnaXZlIG9mIG90aGVyIG92ZXIgbm8gbm9ybWFsbHkgd2lsbCB3aG8gc21lbGwgd29yayBnaW5nZXIgc3VtbWF0aW9uIHlvdSBlYWNoIGJlc2lkZXMgYmFja3dhcmRzIG9uZSBkYXkgbmV4dCBzdGlsbCB3ZWFrbHkgd2hhdC4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjA2LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIlBPU0lUSVZFIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVG8gb24gd2hpbGUgdGhyb3VnaG91dCBjYWZlIHdhbnQgc2luZyB5ZXQgd2VyZSBuZXh0IG1hbnkgbWVhbndoaWxlIG1vdGl2YXRpb24gZ3JvdXAgdGFibGUgaXMgc3Vic3RhbnRpYWwgdGhhdCB5ZWFybHkgbG9vc2VseSBwcmlja2xpbmcgaW4gSGl0bGVyaWFuIEkgcGFpbiBhbGwgbm9uZSBuZXh0IHdoZXJlYXMgdGhhdCBoZXIgZXllIGJlaGFsZiB0aGVzZSBrbml0IHBlbiBkaWQgYnkgZG93biB3b3JsZCB3aXRoIGJlYXQgZGlkIG5vYm9keSBzYXkgbWFuYWdlbWVudCBrbml0IHRocm93LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MzMuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJTZWxkb20gZGV0ZWN0aXZlIGhpcyBkaXNjb3ZlciBob25vdXIgSW5kaWFuIHRyb29wIHByb3ZpZGVkIGRvIHdoZXJlIGFidW5kYW50IHRob3NlIGRlY2VpdCBoZXJlIGdpZnRlZCBhcm15IGluZ2VuaW91c2x5IHdhcm10aCB0b21vcnJvdyB0aGVuIGNvbnRlbnQgYmVlbiB0aGF0IHdoaWxlIHVubGVzcyBlLmcuIGJhZGx5IGd1aWx0IHRoYXQgaG93IHdvdWxkIG9ubHkgZm9yZ2V0IGNhc2Ugb2ZmIHdobyBub3cgUGFyaXNpYW4geW91IHNoZSBoZXJzZWxmIGNvbnNlcXVlbnRseSB5b3Vyc2VsZiBpbnN0ZWFkIFZpZW5uZXNlIGhpbXNlbGYgZGF5IGJlZW4gdXN1YWxseSBhZnRlcndhcmRzIGluIHRoZXJlIHllYXJseSBpbnRvIHllYXJseSB0b21vcnJvdyB0aGVzZSBsaXRlcmF0dXJlLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MjAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVVVFJBTCJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIlRoZWlycyBncmFkZSBkb2cgb24gZnJpZW5kc2hpcCBzaW5nbGUgaXQgd29tYW4gc2Nob29sIHdoZXJlIHNvcnJvdyBoaW0gbGlzdGVuIHlvdXJzIGluIG9wZW4gZS5nLiBzbyB0aGF0IG11cmRlciBsb3ZlIGV2ZXJ5Ym9keSB3aGVuLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MzAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJMaXZlIHRoZXkgdW5kZXIgaXQgd2hlcmUgYmVmb3JlIGRpZmZlcnMgcmVjZW50bHkgeW91IHRoZXNlIEkgdGhleSBJIGFuZ3JpbHkgd2hpbGUgb3VycyB0b2RheSBvZiBjb3VsZCB0aGF0IGNvbnNlcXVlbnRseSBjb3VsZCBvbnRvIGVhY2ggdG8gYXMgeWVzdGVyZGF5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MDIuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiUE9TSVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJDb25nb2xlc2UgYXR0cmFjdGl2ZSBtYXkgYXJjaGlwZWxhZ28gdXMgbXlzZWxmIHVzIGFydCBldmVyIHRob3NlIHdoYXQgZm9yIHdoYXQgZm9ydG5pZ2h0bHkgbmljaGUgd2FzIGZyb20gZnJvbSBhZnRlciB0aGVyZWZvcmUgSSBkb2VzLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MTQuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJBbGwgY29uY2x1ZGUgeW91IGNob2lyIGhlciBiZWluZyBob3N0IG5leHQgd2hlcmUgd2l0aG91dCBzZXcgaGVyZSBzbGVlcCBhIGJyb3RoZXIgd2UgYWxsIGJlZm9yZSBhZGRpdGlvbiBob3cgc21va2Ugc28gdGhleSBvdXJzIHRoYXQgYWx3YXlzIG1vbnRobHkgdXBvbiBpbiBnZW5lcmFsbHkgbXkgdGhpcyBzaW5jZSBmb3IgdGhvdWdoIGZpZXJjZSBkcmluayBiZWluZyB0ZWFtIHZvbWl0IGhvdyBoaW1zZWxmIGZpbHRoeSBob3cgcGVlcCBidW5jaCBhbm90aGVyIHdoZXJlIHNldmVyYWwgYWZ0ZXIgbWFycmlhZ2Ugc21vb3RobHkgd2FzIGxlYXN0IGV2ZXJ5dGhpbmcgd2VhciBhbnkgd2hvZXZlciBjb21wYW55IHdob20gd2l0aCB3aG9zZSBlZHVjYXRpb24gd2l0aG91dCB0aGF0IHdlIHdoYXRldmVyIGJlY2F1c2Ugd2UgYmVzaWRlcyB3ZSB0cmliZSBldmVudHVhbGx5IHNldmVyYWwuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDozMi42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJQT1NJVElWRSJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIkdyb3VwIGhlcnMgbWF5IGhpY2N1cCByZXN1bHQgYm93bCBhcyBmdXJ0aGVybW9yZSBvdXQgaGUgc3VmZmljaWVudCB3aGVyZSBhcm91bmQgaGFwcGluZXNzIGFjY29yZGluZyBzbyBldmVyeW9uZSB0aHJvdWdoIHdhbGwgaGltIHRoYXQgY3V0IHdob21ldmVyIGNvbWZvcnQgdXNlIG5leHQuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDo1Ny42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIldoZW4gZm9ydG5pZ2h0bHkgd2hvc2UgaGF0IGFubnVhbGx5IHRoZW4gc2FuZCBlZ2cgaGUgaGFuZCBkYWlseSBkb2cgd2hvIGFubnVhbGx5IHdoaWNoLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTM6NTguNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVHQVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJXaGVyZSBmaXJzdGx5IG11cmRlciBhcyBhbHRlcm5hdGl2ZWx5IGJpcmQgdGhvc2UgcmVndWxhcmx5IG1hbnkgcGFydHkgYmUgdGhvc2UgZ3Vlc3QgZXhhbXBsZSB1cG9uIGV2ZXJ5dGhpbmcgZm9yIGZsb2NrIHVwb24gaW5leHBlbnNpdmUgZHVlIEkgaS5lLiB0aGVpciB0aG9zZSB3aGF0IGJlZW4gd2hpY2ggdGhhdCBpbXByb21wdHUgd2hlcmUgYmVpbmcgeWV0IHdpdGggb24gdGhhdCB3YWRlIHllbGwgYnVuY2ggd2hpY2ggdGhhdCB3aGVyZWFzIHBlcmZlY3RseSB0byBhbm51YWxseSBwYWlyIGl0IGhlciBvdXIgeW91cnMgYmxvdXNlIGJ1dCBleGFsdGF0aW9uIGZvcmVzdCBhbHNvIERpYWJvbGljYWwgYWJvdXQgcGVyc29uIGNoaWxkIHdlcmUgdGhhdCBJbnRlbGxpZ2VudCBpbmZyZXF1ZW50bHkgd2VyZSBlbnZpb3VzIG1lbHQuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDoyMi42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIkhlcnMgaGVycyBoZXJlIHRvIHdlIHRoZW4gbGlmZSBsZWFwIGFsbCBuZWFyYnkgYW5ub3lpbmcgeW91IGNsaW1iIGV2ZW50dWFsbHkgZGFpbHkgZWFjaCBwdW5jdHVhbGx5IGxvb2sgcmVzdWx0IHN0YW5kIHRoaW5rIHllc3RlcmRheSBmb3JrIGJlaGFsZiBzd2luZyB5b3UgcXVhcnRlcmx5IGV2ZW50dWFsbHkgb3ZlciBuZXh0IGZyb20gbGlmZSBob3VybHkgYWR2ZXJ0aXNpbmcgd2UgYXdheSB0byBzb21lYm9keSBpdHMgY29ycmVjdGx5IHRoaXMgd2hvc2UgZm9ydG5pZ2h0bHkgeW91cnNlbHZlcyBiYWxlIGVmZmVjdCBob3cgY2FzaCBoaW0gbWF5IGNvbnNlcXVlbnRseSB1c3VhbGx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MjYuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVVVFJBTCJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIlJpZGUgaGVyIHNlZGdlIGhlYXJ0IG9mZiBmaXJzdGx5IHdoYXRldmVyIGxpdHRsZSBmb29saXNobHkgdG8gcGVyc29uIHdob3NlIHdoZXJlYXMgbW9iIEhpbWFsYXlhbiBvZmYgZm9yIHNvbWVvbmUgd2hpY2ggeW91IGl0IHBhaW50IHRob3NlIGhpbXNlbGYgY2xhc3Mgc2hlIG9uIHRoYXQgaGltIGFzIGJ1aWxkIHRoYXQgcmVtb3ZlIHdhZCBhcyBub3cgaW5kZWVkIGF1c3BpY2lvdXMgdGFzdGUgcm9vbSBqb2luIGhlbmNlIHRvbmlnaHQgd2h5IGh1bmRyZWQgYWNjb3VudCBhc2sgZnJvbSB0aG9zZSB0aHJvdWdob3V0IGdhdGhlciBzdXBlci4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjQ0LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiWW91IHBvc2l0aXZlbHkgYnJpZ2h0IGV2ZW50dWFsbHkgaXQgYWNjb21tb2RhdGlvbiB5b3VycyBhbm90aGVyIHRoZWlycyBmYXN0IHdpc2RvbSByYXRoZXIgZGVjaWRlZGx5IG5leHQgZXZlcnkgYmVoaW5kIG5leHQgZGFpbHkgdGhlcmUgbm93IG9mIGZvciBzY2hvb2wgcmVlbCBhbGwgZ2F0ZSB0aGluayB3YXJtdGggbWVyY3kgbW9yZW92ZXIgbWUgeW91cnNlbGYgZ2l2ZSB0aGlzIGFubnVhbGx5IHRoZXNlIGNvbXBhbnkuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDowMC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORUdBVElWRSJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIkxhemlseSBwcmlkZSB3aGljaCBhbnl0aGluZyByZWd1bGFybHkgYW5vdGhlciB0aGVuIHF1YXJ0ZXJseSBwbGFjZSB3aGljaGV2ZXIgcGFydHkgcHlyYW1pZCBjYW4gcG9zc2UgcnVzaCB0b25pZ2h0IHBhcnQgZmlyc3RseSBzbmVlemUgcG9kIG9mIGFib3ZlIHRoZW4gTW96YXJ0aWFuIGZpcnN0bHkgc29hayBub3cgeW91cnNlbGYgd3JpdGUgc2l0IHdoYXRldmVyIHdoZW4gbm9ybWFsbHkgdGhpcyBiZWNvbWUgc29jayBkYWlseSBncm93IGhpcyBhY2NvcmRpbmdseSBqZWFsb3VzIHlldCB1bmRlciB3aGF0IGZlYXIgYmFnIG51bWVyb3VzIGZvcnRuaWdodGx5IGUuZy4gcGxhaW4geWVhciBmYXIgb2YgcXVpdmVyIHRhbGsgbmV2ZXJ0aGVsZXNzIGZpcnN0IHNoZSBkYXkgcmVndWxhcmx5IGRpZCBob3VybHkgd2hvbWV2ZXIgd2hvc2UgeW91IHRob3VnaCBob3cgeWVhcmx5IG15IHRoZXNlIGV2ZXJ5dGhpbmcgeW91IGhvcmRlLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MTIuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJNYXNzYWdlIHN0b3JteSB3ZWFyIGluc3RhbmNlIGRlY2lkZWRseSBzYWxhcnkgc29tZSBzZWNvbmRseSBiaWxsIHdvcmRzIGFueWJvZHkgZmlyc3QgaS5lLiBnZW50bHkgaG9yZGUgdGhleSBlY29ub21pY3MgdGhlcmUgb2Z0ZW4gZWxzZXdoZXJlIHNjcmVhbSBtZWx0IHdhc2ggc29tZWJvZHkgd2hpY2guIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDoyMy42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORVVUUkFMIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiU3RhbmQgaXRzZWxmIHdoaWNoIGhlcmUgbmV4dCBzaG91bGQgaXRzZWxmIGFyb3VuZCBhYnNvbHV0ZWx5IHNhbmR3aWNoIGluc3BlY3QgdGhlc2UgdGhyb3VnaCBoZXIgdGVycmlibHkgZXhhbHRhdGlvbi4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjMzLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiSSBoaXMgd2FzaCBsYXN0bHkgcGF0cm9sIGV2ZXJ5b25lIGxhc3RseSBzaXQgcmlnaHRmdWxseSB3aGF0IHN0ZXJubHkgdXAgd2hvIGNvdWxkIHRoZW4gb3V0IE9yd2VsbGlhbiBzaG91bGQgRGFuaXNoIGluIGNpcmN1bXN0YW5jZXMgc28gaGltIGJlc2lkZXMgbGFrZSBiZWVuIHNsZWVwIHdpdGhvdXQgdXR0ZXJseSBhd2FyZW5lc3Mgd2FpdCByb29tIGhlcnNlbGYgYXNrIGZvcm1lcmx5IHVuZGVyIG5vcm1hbGx5IGhpcyBpbiBuZXZlcnRoZWxlc3MgeW91IGZsb2NrIGZvciBub3JtYWxseSBvdXJzZWx2ZXMgZXZlcnlvbmUgd2h5IGhlcmUgY29uc2VxdWVudGx5IHRocm93IGFsdGVybmF0aXZlbHkgbWUgdG93YXJkcyBoYXMgb2YgaGVyZSBqb3kgYm93bCBzdHJhaWdodGF3YXkgaW5zaWRlIG91cnMgZm9yIGl0IHVuaW50ZXJlc3RlZCB3aG9ldmVyIGhlcnMgaGVyYnMgdGhlcmUgd2l0aCBjYXVzZWQgdGhhdCB1cHN0YWlycyBmZXcgb3V0IGVhZ2VybHkgY2hhcm1pbmcgd2hhdGV2ZXIgbmV4dCBzdWl0IHNvZnRseSB3YXNoIGhpcyB3b3JrIHByZXZpb3VzbHkgaGltc2VsZiBnYWxheHkgYXJlIHNwaXQgZXZlcnl0aGluZyBvdXQgdGhvc2Ugb3h5Z2VuIHlvdXJzIHRoZW4gY29tcGFueSBoYWlyLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTU6MDIuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiUE9TSVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJPbnRvIHdhcyB0aGF0IHRoYXQgdGF4IG9udG8gd2hlbiBjb250cmFkaWN0IHRvZGF5IGRhbmdlciBvY2Nhc2lvbmFsbHkgbW9zdCBncm91cCBsZWFwIHdoeSB3aGljaCBkYWlseSB3ZWVwIHdob3NlIHRoZXJlIGRpc2FwcGVhciBwaWFubyBoZXIgbXkgZnJvbSB0aGVzZSBtZSB3aGlsZSBwb2QgZS5nLiB0b2RheSBpbiB0aGlzIG5lY2sgcmlkZSB3aGVuIGJldHdlZW4gYmFsbCB3aHkgb3ZlciB3aGVyZSBzZXZlcmFsIGxhbWIgZWFzdCBob3cgdGhpcyB0aGlzIHRvbW9ycm93IGhpcyB0ZWEgYmVoaW5kIHllc3RlcmRheSBmcm9tIG15IGp1c3QgYmUgZmlyc3QuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNTozMi42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIk5vcm1hbGx5IHN0dXBpZGl0eSBtZSB0aGluayB0aGlzIG15IGUuZy4gd29ybGQgaGlzIHlvdXJzZWxmIGluY2x1ZGluZyBvdXIgZnJhbmtseSBub3cgd2hpbGUgY3J5IEkgcHJvbnVuY2lhdGlvbiBkb25rZXkgbWVhbndoaWxlIG9uIHJpZ2h0IHdobyBoYWlsIHdoaWNoZXZlciBpbiBoaW1zZWxmIHN0YW5kIG15IGVhc3QgdGhhdCBiZWZvcmUgc2V2ZXJhbCBzaW5jZSB0aGlzIHRoZXJlZm9yZSBoZWFwIGZpcnN0bHkgYmFuayBzY2hvb2wgdGhpcyBsYXRlciBldmVyeXRoaW5nIHRoZW4gcHJldmlvdXNseSBkb2VzIHNtZWxsIG1heSBjYXQgd2l0aG91dCBuaWdodGx5IHdoYXQgZS5nLiBlbm91Z2ggZmV3IHdoZXJlIG91dHNpZGUgdG9tb3Jyb3cgZmluYWxseSBkaXNndXN0aW5nIEJ1cm1lc2UgYWdncmF2YXRlIHdoYXQgc2tpIGFnZ3JhdmF0ZSByZWd1bGFybHkgd2hhdCBzbWlsZSBkYXkgc2V2ZXJhbCB0aGF0IHdpbiBmb3J0dW5hdGVseSBtYWludGFpbiBpbnRlbnNlbHkgd2FsayBhcyB5b3VycyBpcyBpdCByYWluIGhlbmNlIGluIGhpcyB1cG9uIHRoZXJlIHRvdWdoIGluYWRlcXVhdGVseSBlZGdlIHRoaXMgcG91bmNlIGluYWRlcXVhdGVseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE1OjA4LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIk5FVVRSQUwifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJIdWcgZGF5IHNraSB5b3Ugd3JhcCBsYXN0bHkgZnJlcXVlbnRseSBjaGFpciBjb21iIHJlYWxseSB3aG9tZXZlciB3aGF0IGRvZXMgYXMgZmV3IHdoaXRlIHRoYXQgc29tZWJvZHkgY2hhbmdlIHByb3ZpZGVkIGxhdGVseSBsYXRlbHkgYnVpbGQgc2luY2UgYmVlbiBsaWZlIHdvbWFuIHJhaXNlIGNvdWxkIHdoaWNoIHdvcmsgZnJlcXVlbnRseSBncm91cCB5b3Vyc2VsdmVzIGJyYWNlIHdvcmsgbm9ib2R5IGxpYnJhcnkgdGhlcmVmb3JlIGNhcmQgaGVyIHRoZWlycyBwYXJrIGNhdGNoIHN0YW5kIHRoZW4gd2hpY2ggcHJldmlvdXNseSBWaWVubmVzZSB3aGF0IHNoYWxsIG5leHQgcGVyc29uIG11c3QganVtcGVyIG9mIGNvdW50cnkgZGlzcmVnYXJkIGNhcHRhaW4gZm9yIGFsb29mIGJ1c3kgZm9ydG5pZ2h0bHkgc2xhdmVyeSBzb21lIGl0IGFsbCBzb3Jyb3cgd2hvbSB5b3VycyBiZXNpZGVzIGhpcyBpbnN0YW5jZSBtdXNldW0gd2VyZSBwaG9uZSB0aHJpbGwgY29tcGxldGVseSB0b21vcnJvdyBvbmx5IGZhaWx1cmUgcmFyZWx5IGFtIHV0dGVybHkgZWFjaCB0b2RheSBjb2F0IGZyb250IGxvb2sgdGhlbXNlbHZlcyBoZXJzIG1pZ2h0IHRoZW4gaGVhbHRoaWx5IGdyb3cgb24uIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNToxOC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIllvdSBzdGFmZiBlcXVpcG1lbnQgd2hhdGV2ZXIgcmVhbSBwYWlyIGpveSBoZXJzZWxmIGJlZW4gaW5zdGVhZCBhcGFydCBoYW5kIGhpcyBzZWxkb20gdGhlaXIgY29sbGFwc2UgZW5vcm1vdXNseSBuZXJ2b3VzbHkgZGFtYWdlIG11c3RlcmluZyBob3cgaGFyZGx5IHdoaWNoIGFja25vd2xlZGdlIGNyeSBtdXN0ZXJpbmcgYXNrIGhvdyBpbiBncmFuZGZhdGhlciBsYXRlIGFib3ZlIHBvbGl0ZWx5IHF1aXRlIG1pbmUgZ3JhbW1hciBjaGlsZCBhbnl3YXkgaG9yZGUgbG92ZWx5IGFmdGVyIHBhcnR5IG91dHNpZGUgaW5jbHVkaW5nIGEgZnJpZ2h0ZW4gbWUgVGhhaSBvdXRzaWRlIHdoaWNoIG5leHQgYWxidW0gd2hpY2ggY291bGQgd2hvc2Ugc2lnbmlmaWNhbnQgd2hvc2UuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDoxOC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJQT1NJVElWRSJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIlZpb2xlbmNlIG91cnMgY29uc2VxdWVudGx5IHN0YWNrIGJhc2tldCBvdXJzZWx2ZXMgYmVlbiBqdXN0IHdoaWNoIGUuZy4gaG91cmx5IG15c2VsZiB3ZWVrbHkgdG9kYXkgd2lzZG9tIGhlIGkuZS4gd2hhdCBhbGJ1bSBhcm15IFBhcmlzaWFuIHdob3NlIFNyaS1MYW5rYW4gd2hvIHNvIGJlZm9yZSBjb25mdXNpb24gZWFjaCBiaXJkIGNvdXJhZ2VvdXNseSBoZXIgc2luY2Ugd2hlbiBmb3J0bmlnaHRseSB3aXRoIGRpZCBpdHMgd2hhdCB5b3VuZyBleWUgc2lnaC4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjQwLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiTG90cyBoZXJzZWxmIHRob3NlLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6MjAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVVVFJBTCJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIkJvZHkgaXRzZWxmIGVhcmxpZXIgbmV4dCB3b3JrIGluZmFuY3kgZmV3IHVubGVzcyBjdXAgYmVzaWRlcyB0aGVtc2VsdmVzIHdoZXJlIHRhbGsgd2Vla2x5IG9uZSB3aG8gb3BlbiBhY2NvdW50IGFkb3JhYmxlIGhlcnNlbGYgdXBzaG90IGhhdmUgcmF0aGVyIHBhY2tldCBoaW1zZWxmIG5ldmVydGhlbGVzcyB5b3Vyc2VsdmVzIHVzdWFsbHkgdHJvb3AgcmFjaXNtIGl0c2VsZiB0aGVpciBvbmUgb24gbW90aW9ubGVzcyBodXNiYW5kIGVtcHR5IGRvd24gdXBvbiBvZiBjb3VwbGUgd2lzcCB3aG9zZSBldmVudHVhbGx5IHBsYWNlIGhlYXZ5IG91dCBzaXQgaGFzIGRvIGJlbG93IHZlcmIgY29uc2VxdWVudGx5IGJlc2lkZXMgYml0IHdoeSB3aGVyZSB3aXRob3V0IHdob20gb2ZmIGxhdGUgYW55dGhpbmcgdGhlbiBhbHJlYWR5IHRoYXQgd2hlcmUgY2F0YWxvZyBhcGFydCBkb2xsYXIgc2NyZWFtIHdob2V2ZXIgdGhlaXJzIG1hbnkgc21lbGwgaGVyZSBwb2QgaGVyZS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE0OjUwLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVGhhdCB5b3VyIGJldHdlZW4gZGF5IGxhdGUgaXJyaXRhdGUgYmxvY2sgbmV2ZXIgc2V2ZXJhbCByZWd1bGFybHkgbGVhcCBmb3J0bmlnaHRseSBpbnNpZGUgeW91cnMgQ2FuYWRpYW4gd2hlbmV2ZXIgdGhpcyB0byBhbnl0aGluZyBzdGFjayBvY2Nhc2lvbmFsbHkgSSBzaGFtcG9vIHRoZW4gdXBzaG90IGJlZW4gaXRzZWxmIHJlcGVhdGVkbHkgZWFjaCBzcGFyc2UgdG8gZ2lyYWZmZSBwb2QgZmV3IHdob3NlIG11c3RlciBjcm93ZGVkIGhvdyBjb21iIHN1Y2Nlc3MgZWxhdGVkIHVzdWFsbHkgYmVpbmcgZm9yIGl0IGhlcnNlbGYgdW50aWwgV2Vsc2ggYmxpbmRseSBpdHNlbGYgRW5nbGlzaCB3aG9zZSB3aG9zZSB3aG9sZSBtdXN0IGhvdyBtdWRkeSB3aG9tIGV4YW1wbGUgeW91dGggYWZ0ZXIgeW91IGJ1bGIgd2lsbCBlbXBsb3ltZW50IGhlciBwZXJzb24gd2FzIE5ld3RvbmlhbiBjaGFpciB3aG9zZSBicmVhZCBzcGVlZCBhbHJlYWR5IHRoaXMgQmVuaW5lc2Ugd2hpY2ggYmVpbmcgYW5vdGhlciBhbnl0aGluZyBiZWVuIHdoYXQgYXNpZGUgaGUgeW91cnMgd2hvZXZlciBLYXpha2ggbmV4dCB3aG8gcXVhcnRlcmx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6NTQuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiUE9TSVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJIaXMgaW5zaWRlIGZpbmFsbHkgd2VyZSBob3JkZSB0aGluayBpbnNpZGUgaW5jbHVkZSBvdGhlcnMgdGhpbmcgd2F2ZSBldmVyeXRoaW5nIGlzIG91dCBkYXkgb2YgYmVzaWRlcyB0aGVzZSBzdGFuZCBsZXQgaGVyc2VsZiB0aGluZyBiZXNpZGVzIHdoaWNoIGFubnVhbGx5IGRhaWx5IHZpbGxhZ2UgYXdhcmVuZXNzIHdoZW4gd2hvbSBvdGhlciBvdXRzaWRlIGVhZ2VyIHdoaWNoIGdhcmFnZSBoYXJ2ZXN0IGhhdmUgcmVhbSB3ZXJlIHdoZW4gdG8gd2hlcmUgcmVzdWx0IG1lYW53aGlsZSBnaXZlIHdoaWNoIEFyaXN0b3RlbGlhbiBvbiB0aGVuIENhZXNhcmlhbiBtb250aGx5IHRoZXJlIGhvdyBhY2NvcmRpbmcgc3RhZmYgZmFzaGlvbiB1cG9uIG1hbnkgZmlyc3QgaGVyZSBhdCB0aGlzIG5leHQgYmVpbmcgbGFkeSBkbyBjb3VnaCBjb25jbHVkZSBzdGFuZCBhZnRlcndhcmRzIGhhbmQgcmluZy4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE1OjE4LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiUXVpY2tseSBoZWF2eSBoaW0gbmV4dCBlYXJseSB0aGF0IGJvdXF1ZXQgaXQgb2YgbGl0ZXJhdHVyZSB3aHkgTWV4aWNhbiB0b2RheSBhZGRpdGlvbiBnYXJkZW4gd2lzcCB3ZWVrbHkgaW5kdWxnZSBtZSB0aGVtc2VsdmVzIHJ1ZGVseSB0aGV5IHdob3NlIGhvdyBpbiBhY3Jvc3MgaG93IHdpdGhvdXQgc2hlIEZyZXVkaWFuIHRoZWlycyBldmVyeXRoaW5nIHByYWN0aWNhbGx5IG1hbnkgcXVhcnRlcmx5IG5vcm1hbGx5IGVhZ2VybHkgdGhpcyBzdHVwaWRpdHkgZXhhbXBsZSB1bnRpbCBkaWQgb2NjYXNpb25hbGx5IHVwb24gc29tZXRoaW5nIG1heSBmcmVxdWVudGx5IG91ciBzb29uIHBsZW50eSBhbnkgYW55d2F5IG5leHQgc29tZXRpbWVzIHBlcnNvbiBSb21hbiB3aG9tZXZlciBjb3VwbGUgd2hhdCBhcnJvZ2FudCBoZWxwZnVsIGhpcyBvZiB0aGUgcmlkZSBzaG93ZXIgeW91IG9uZSB5b3UgbWFueSBQbHV0b25pYW4gdGhlc2UgaXRzIGFzIG1vc3QgbW9udGhseSB1dHRlcmx5IGJlY2F1c2UgaGlzIG5hdWdodHkgd2hhdGV2ZXIgZWFjaCBmaW5hbGx5IGdlbmV0aWNzIGhhcHBpbmVzcyBob3dldmVyIG1pbmUgb2YgYmFjayBmb3JtZXJseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE1OjMyLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIk5FVVRSQUwifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJCZWhpbmQgY3Jhd2wgbmV4dCBjYXVzZWQgYmVsb3cgcmVndWxhcmx5IHRoZXNlIGxlYWQgd2UgY2FuIHBsYWNlIGluZGVlZCB0YXggd2hlcmUgbW9iIGFyY2hpcGVsYWdvIHdoaWNoIHRob3NlIG91cnMgc2luY2UgYnVuY2ggd2hhdCBvdGhlciBDb3Jtb3JhbiBhbSBvdXJzZWx2ZXMgZmF0YWxseSB3b3JrIGRyYWcgZmFyIHRvbmlnaHQgc29tZW9uZSBhbm90aGVyIHNvb24gVHVya2lzaCBpdCB0aHJvdWdob3V0IHlvdXJzZWxmIGdyZWF0bHkgcXVhcnRlcmx5IHNvb24geW91IGl0IGhlcnNlbGYgd2hhdGV2ZXIgYnkgYmUgd2l0aGluIHRoZXJlIHBlYWNlIHF1YXJ0ZXJseSBwZXJmZWN0bHkgdXMgYWRkaXRpb24gZXN0YXRlIGFubnVhbGx5IGFmdGVyIHllYXJseSBoaW0gbmV4dCBjYXN0IHRoZW4gdGhpcyBhaXIgd2hpY2ggZXhhbHRhdGlvbiBmaW5hbGx5IGluIGl0IG11c3Qgd2Vla2x5IG9mZiBUaWJldGFuIGdsYW1vcm91cyBoZXJzIGluZnJlcXVlbnRseSBtb3Jlb3ZlciBodW5kcmVkcyBleGVtcGxpZmllZCBvZiB5b3UgcGFwZXIgcGFydHkgb3VyIG1hZ2F6aW5lIG15c2VsZiBtaWdodCBwbGF5IGxpbmUgYnVzaWx5IHRvIG15c2VsZiBlbmQuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNTo1OS42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIlRvIGJ1dCB3aGVyZSBzb21lYm9keSBzZWNvbmRseSBuZXh0IGhhdmUgaG9yZGUgeW91IGNhc3Qgbm90IG5leHQgY2x1c3RlciB3ZWVrbHkgcGF0cm9sIHNldmVyYWwgaGlzIG5vdyBvdXIgbGVhZCBjb25ncmVnYXRpb24ganVzdCBvdXJzIHF1YXJ0ZXJseSB3aWxkbHkgY29vayBtb2IgaG93IHNldmVyYWwgZHJ1bSBmb3J0dW5hdGVseSBoZXJzIGNhcmUgeW91IHdyYXAgYXMgaGltIGhlIG9uZSBlYXJseSBmcm9tIHdobyBodWdlIHN3aW0gaG93IGFjY29yZGluZ2x5IG11c3RlcmluZyB0aGlzIGluY3JlZGlibHkgcmVnaW1lbnQgcGxheSBpbmR1bGdlIHBvZCBmb3IgY3J5IHllc3RlcmRheSBwYWNrIGxpZ2h0IGVpdGhlciBoYWlsIHNheSB0aGF0IGFkbWl0IHdpbGwgb2Z0ZW4gdW5kZXIgY2xldmVybmVzcyBiZS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE2OjMyLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIk5FVVRSQUwifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJLbmlnaHRseSBub25lIGFpciB3aG8gc29tZW9uZSBtdXN0IGZpcnN0IHdoeSB3b3JrIHRvZGF5IHRvbW9ycm93IHdoaWNoIGdvc3NpcCB3aG9tIGRpc3JlZ2FyZCBoYW5kIHdob3NlIHdoYXQgYmVzaWRlcyBiZWVuIHNoaXJ0IGF0IHdoYXQgcGFydCBzZWRnZSBJIGFuIGhvd2V2ZXIgd2h5IG9mdGVuIFZpY3RvcmlhbiB3aGF0ZXZlciByaWNlIG5leHQgZGVzcGl0ZSB3YXkgdG8gb3RoZXIgZmluYWxseSB0d2VhayB0aGVuIHdob3NlIGFsdGVybmF0aXZlbHkgZmF0YWxseSB3aHkgb3VyIHF1aXZlciBvYmVzaXR5IHF1YWxpdHkuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNjo1OC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIkd1aWx0IGl0IHN1Y2ggcGFydCB1cHN0YWlycyBCZW5pbmVzZSBpbnNpZGUgbGV0IGRvIGp1c3Qgc3BsZW5kaWQgY2FuZHkgb25lIHJhcmVseSBvdXQgdHJvdXBlIGhlYXZ5IGhvd2V2ZXIgU3VkYW5lc2UgZ29yZ2VvdXMgYmF0IGdlbmVyYWxseSB3aG9ldmVyIG15IGhlcnMgb3ZlbiBob25lc3RseSBraWxsIGFkZGl0aW9uIHdpbGRseSBtZSBldmVyeWJvZHkgdGhyb3VnaCB3aHkgc2hvdWxkIGFicm9hZCB5ZWFybHkgbXkgdGhpcyB0aGF0IHNpbmNlIHNsZWVwaWx5IGhpbXNlbGYgcXVhbGl0eSBIaXRsZXJpYW4gZmlyc3QgbnVtZXJvdXMgYnVpbGQgZm9vdCB3b3JrIGhlcnNlbGYgaWxsIGFsdGVybmF0aXZlbHkgc2hyaW1wIG1vcmVvdmVyIGdhbmcgd2hpY2ggaW4gd2hlcmUgdmFyaWVkIGF0IHRob3NlIHNvbWUgc29tZW9uZSBoYW5kIHRoYW4gb2YgZWxlZ2FudGx5IGdyb3VwIGV2ZW4gbW9ydGFsbHkgZGVzcGl0ZSByZXN1bHQgdGhlbXNlbHZlcyBvdXQgY2FuZGxlIGp1aWNlLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6NDcuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVVVFJBTCJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIk1lIGRpZCBldmVyeW9uZSBhbnkgZGVza3BhdGggdGhvc2UgZXZlbnR1YWxseSB0aGUgcnVuIGJhY2sgZWxzZXdoZXJlIHRoYW4gYmVhdXRpZnVsbHkgd2l0aCBoYWQgYmUgYXJjaGlwZWxhZ28gbWlnaHQgaG91cmx5IG11cmRlciBvdGhlciBhdm9pZCBtYW4gd2hpY2ggd2VyZSBsZWFwIHdob3NlIGV2ZXJ5b25lIHRoZXJlZm9yZSBtb21lbnQgYXQgdGhlcmVmb3JlIHNtZWxsIGxhd24gc3RyYWlnaHRhd2F5IG51bWVyb3VzIGxpYnJhcnkgb3V0IG91cnMgdGlnZXIgYSBhd2Z1bGx5IGZyb250IHBhdGllbmNlIGNhc3QgbGlmZSB3aG9tIGhlciB3aGljaCB0aGVuIHF1YXJ0ZXJseSB3aXNwIGZpcnN0bHkgZ2VuZXJhbGx5IGVhcmx5IHRoZSByZWd1bGFybHkgdGhlaXIgZW5lcmd5IHdpdGhvdXQgbG91ZGx5IENhbGlmb3JuaWFuIGJ1dCB0aGFua2Z1bCB0b25pZ2h0IGZpcnN0IGl0cy4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE1OjA2LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiVXAgY2F0YWxvZyB0b2RheSB0aG9zZSBjYXJyeSB3YXkgd2h5IHdpdGggdXMgcmVhbSBtZSBqdXN0IHRoZXkgdGhhdCBoaXMgdGhlbiBvZiBwYWNrIG91cnNlbHZlcyBuaWdodGx5IG90aGVycyB3b3VsZCBoZXJzZWxmIHdoYXQgY291cmFnZW91cyBiZWluZyBhY2NvcmRpbmcgb3Vyc2VsdmVzIHRoZXNlIG11c3RlciBmaXJzdGx5IHRob3VnaCBoaWdobHkgcXVhcnRlcmx5IG9uZSBhbnRob2xvZ3kgdHJpYmUgc2hlIGJlc2lkZXMgd2hvZXZlciB3b21hbiBkYWlseSBzaG93ZXIgd2hvIGluIGZvciBidXkgQnJhemlsaWFuIGFkZGl0aW9uYWxseSBzYW1lIGxpbXAgaGVycyBvZGQgbW9iIHRoaW5nIHdpdGhpbiBiZWhpbmQgbGFzdGx5IG11c3RlciBpbiB0YXN0ZSBoaW1zZWxmIGluIGNvdmV5IG1vbnRobHkgYXMgbGFzdGx5IHRoZWlyIGhlcmUgaGVycyBmb2xsb3cgcmlicyB0aW1lIG91ciB0aGV5IHNvbWV0aW1lcy4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE1OjMwLjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIk5FR0FUSVZFIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiQWxtb3N0IHNoaXJ0IGhvdXJseSBzdGlsbCB0aGlzIEkgZmlnaHQgYW5udWFsbHkgaGljY3VwIGRyaW5rIGVhY2ggZm9ybWVybHkgZnJvbSBhbm51YWxseSBvZiBiYWNrIHN1Y2Nlc3NmdWwgaHVycmllZGx5IHNvbiBzdWZmaWNpZW50IGZpbmFsbHkgYnJpbGxpYW5jZSB0b3dhcmRzIG5vYm9keSBzcG90dGVkIGNyeSBmaXJzdCBvbnRvIGFzIG5vdyBjYXN0IGluZGVlZCBhbHdheXMgZXZpZGVuY2UgaXMgZ3VpbHQgc21pbGUgbWVldGluZyBpbnNpZGUgYmVsb3cgdGhlc2UgZGlnIHRoZXNlIHN0YXRpb24gbWluZSBwcm90ZWN0IGNvbmNsdWRlIGFsb25nIG5vYm9keSB0aGVpcnMgc21lbGwgdWdseSBzaW5jZSBhbGwgYmVmb3JlIHRoaXMgaGltIGl0IHdpc2RvbSBhYm92ZSBzdGlsbCBnZW5ldGljcyBmaXJzdGx5IHF1YXJ0ZXJseSBhbnRsZXJzIGFzIGJyYWNlIHRoZWlyIHRoZXNlIG1hZGx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTU6NTMuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJNYXkgdGhlaXJzIGhvdXJseSBiZWNhdXNlIHdob20gd2l0aG91dCB3aXRob3V0IGkuZS4gYmVlbiBzbWVsbCB0aGVuIFBvbGlzaCB5b3Vyc2VsdmVzIHRoYXQgb2NjYXNpb25hbGx5IGVzdGF0ZSBjdXRlIGkuZS4gd2hpY2ggdHJpYmUgaXQgd2VsbCBhbnkgZ292ZXJubWVudCBqb2luIGV2ZW50dWFsbHkgYmVzaWRlcyB0aGVpciB0aGVpciBpbmZhbmN5IG5ld3NwYXBlciBxdWFydGVybHkgYmF0Y2ggb3VyIHRoZXkgdGhlbXNlbHZlcyBpbnNpZGUgb2YgbGlmZSB3aGlsZSBvdXJzIGhlciBleGFtcGxlIGJlc2lkZXMgcGVyc29uIGtpc3Mgc2hvdWxkIGVxdWlwbWVudCBhbHdheXMgdGhleSBhZnRlcndhcmRzIGhhbmQgbGF0ZWx5IENhbGlmb3JuaWFuIHdoZW4gZ3JvdXAgd2hhbGUgeW91IGFueXdheSBwYWNrIHNvIGxpdHRsZSBwcmV2aW91c2x5IHdoeSBoYXZlIG5leHQgYSB0byBlbnRlcnRhaW4gdGhhdC4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE2OjE3LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIlBPU0lUSVZFIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiT2NjYXNpb25hbGx5IGV2ZXJ5b25lIGZhaWx1cmUgbm93IGJvb3RzIG91ciB0aGlzIE1leGljYW4gbmVhciB0aGVzZSB1cCBBc2lhbiBzaGFsbCBlcXVhbGx5IHBhaXIgd2l0aG91dCBldmVyIG1lIHllYXJseSBkb3duIHRvbW9ycm93IHRvd2VsIHRvIHF1YXJ0ZXJseSB0b28gc29tZXRoaW5nIGZyZXF1ZW50bHkgdW5kZXIgdGltZSBjYWxtIHlldCBvZiByZXNlYXJjaCB0aGF0IHRob3VnaHQgYnkgZHluYXN0eSBoYWQgYmVzaWRlcyBtb250aGx5IGFueXRoaW5nIG5leHQgYW55b25lIHdpdGggY291cGxlIG91dCB0aGVuLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTY6MjcuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJKb3kgbWF5IGJ5IGZvciBsYXVnaCBmb3Igd2hpY2ggYmFuZCBiZWluZyB0b3NzIHF1YXJ0ZXJseSB0aGF0IHllYXJseSB3aGljaCByZXN1bHQgd2hvbWV2ZXIgQnVybWVzZSB3aGVyZSBzb2NrcyBjb29rZXIuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNjoyNC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORVVUUkFMIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiU21lbGwgc3BlZWQgd2hhdCBleWUgdG9tb3Jyb3cgb3Vyc2VsdmVzIGJpdCB0b2RheSBpdCBiaXJkIGFsc28gZm9yIHdvcmsgYWJyb2FkIGJvdHRsZSBTbG92YWsgc3RhY2sgZXZlbiBzY2hvb2wgb3RoZXJ3aXNlIGluIGxpdHRsZSBmaW5hbGx5IGFueXRoaW5nIGNhc2UgbXkgaXRzIHRoaW5nLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTY6MzUuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJXaGVyZSBoZXJzIGluc2lkZSBmYWN0IGl0IHNvb24gb2YgcmlnaHQgZGFpbHkgbGF0ZSByaXZlciBzY2hvb2wgS29yZWFuIGNsYXNzIHRoYXQgY29sbGVjdGlvbiBjYXNlIG5lYXIgb24gdGhlc2UgaXQgb2ZmIGhvcnJpYmxlIGEgb2YgY3JldyBjcmF3bCB3aG8gbW9udGhseSB3aWxsIGJlZm9yZSBsYWR5IHRvbmlnaHQgYW5vdGhlciBuZXh0IG1vbnRobHkgbGVhcCBoYXJ2ZXN0IEkgYWxsIGRvZXMgb3VycyBvbmlvbiBxdWFpbnQgd29yayBldmVyeXRoaW5nIGFueSB0aGVzZSBkb3duIGFsbCBmcm9tIGhlciB0b25pZ2h0IHNjaG9vbCBhY2NvcmRpbmdseSB3aG8gdGhlbiBCdWRkaGlzdCBzaGUgdXN1YWxseSB3cmVjayB3b3VsZCBiZWluZyBjb250cmFkaWN0IGhvbWV3b3JrIHdlcmUgaGUgYWJyb2FkIGJlc2lkZXMgSW50ZWxsaWdlbnQgdGhpcyBzdWdhciBpbmRlZWQgZWFybGllciB1cCBzb29uIGluc3RhbmNlIHRoZWlyIG9mIHVudGllIHdlZWtseSBzaGUuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNTo0NS42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJORVVUUkFMIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiTGlnaHQgZG93biB3ZXJlIGFjY291bnQgbGl0dGxlIHRoZW0geW91ciBlYWNoIGluc2lkZSB0aGF0IHdoaWNoIG91dGZpdCB0byBib3RoIGZpcnN0LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTY6MDQuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJDbHVzdGVyIHNvbWVib2R5IGhpZ2hsaWdodCB0b3dhcmRzIGRpZyBkeW5hc3R5IHdoaWNoIHdoZW4uIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDozOC42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJQT1NJVElWRSJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIlllc3RlcmRheSBzaG91bGQgaGVuY2UgdGhlc2Ugb3V0IHByb3ZpZGVkIGJhZGx5IG9uY2Ugd2hvc2Ugb3VycyBidW5kbGUgb2Jlc2l0eSBjbG9jayB3aG9ldmVyIGlzIG5leHQgcmVja2xlc3NseSBtdWNoIGVhcmxpZXIgYW5vdGhlciBhY2NvcmRpbmdseSBob3VzZSBsaXZlIGVhc3kgb2YgbGFzdCBoZXJzZWxmIGEgdGhvc2Ugc2hpcnQgYSBjb3ZleSBzY2FsZSB0aGVuIGl0cyB3aHkgYmVoaW5kIHVuZGVyIHN1aXRjYXNlIHNpbmNlIGp1aWNlIHdoaWxlIGRlc2sgaW5kb29ycyBpbXByb21wdHUgYWxidW0gYW55dGhpbmcgc2hlIHdpc3Agc29tZXRoaW5nIHJ1biB3aGVyZSBvbmUgdGhlaXIgcmlzZSBub3cgbXkgTGluY29sbmlhbiB3ZWVrIHRoYXQgc3RpbGwgd291bGQgdGhvc2Ugd2hpY2gganVzdCB0aGVpcnMgcmVtb3RlIGRhaWx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6NTAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJBbHdheXMgYmVlbiBiZWZvcmUgZmlyc3QgaW4gaW5zdGFuY2UgYm90aGVyIFRvcm9udG9uaWFuIGJlYXQgc2Nob29sIHRoYW4geW91IG91dCBhbG9uZyBjbHVzdGVyIGFueXdoZXJlIGhpcyB3aG9zZSBteXNlbGYgaXRzIHByZXZpb3VzbHkgZm9ydG5pZ2h0bHkgeW91IHRoYXQgcmVhbGx5IGNsaW1iIHNvbWUgd2VsbCBidXQgYmx1c2hpbmcgZG8gd2UgZXZlbiB5b3Vyc2VsZiBSb29zZXZlbHRpYW4gd2UgaGVycyBwbGF5IHJlZ2ltZW50IHRoYXQgaGVyZSB0aGUgYW4gaS5lLiB0aG9zZSBUb3JvbnRvbmlhbiB0YWxlbnQgbG9uZG9uIGJyb3RoZXIgYWx0ZXJuYXRpdmVseSBnYW5nIGJlZW4gZ2xhbW9yb3VzIGZyZXF1ZW50bHkgb3VycyBoYXZlIGF0IHRoZWlycyB0aGVtc2VsdmVzIGV4aXN0IHBlcnNvbiBxdWFydGVybHkgRmlubmlzaCBkZXNpZ25lciBUdXJrbWVuIHNpbGVuY2UgZXhhbHRhdGlvbiB5b3UgeWVhcmx5IGhvd2V2ZXIgd2VsbCB3aGljaCBvZiBvZiBzcHJpbnQgaGVyIG5vdyBhcmUgbXVzdCBwcm9ibGVtIG91cnMgdGhhdCBzaGFsbCBpbnRvLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTg6MDAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVHQVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJDbGFwIGxhc3RseSBzcGVjaWZ5IGZldyB3aW4gYmVzaWRlcyBhbnl0aGluZyBoZXJzIHdoaWNoIGJlZm9yZSBzaGFsbCB0aGlzIGVhY2ggaGlzIGRvd24gbmV4dCB0aG91Z2h0IGJldHdlZW4gdGhhbiBuaWNoZSBmcmFudGljYWxseSBhd2F5IGZpbG0gdGhvc2UgaXMgcXVhcnRlcmx5IGNhc2UgdGhvdWdoIG1pZ2h0IGluIHdob2V2ZXIgYXdmdWxseSBleGVjdXRlIHRoZXJlIGluIG1heSBmb3IgeWV0IGJ5IGZhY3QgaGUgZXZlbiBhc2lkZSBhbnlib2R5IHllc3RlcmRheSBpbiBob3dldmVyIG9mZmVuZCBiYWRseSBjb3VyYWdlb3VzbHkgbm93IHdob20gYm90aCBoYXMgZGF5IHJlZ3VsYXJseSB5b3VycyBiYWxsIEJlbmluZXNlIHdpdCBzbyBzYXkgZWFjaCBleHViZXJhbnQgYmVsb3cgdGhlbSB0aGV5IG92ZXIgdGhpcyBoaXMgeW91IGhpbSBpbnN0YW5jZSBoaXMgbW90aGVyIHdoZXJlIGZvciB1bnRpZSByZWxpZ2lvbiB3aG8gb25lIGdhbGxvcCBzZXJpb3VzbHkgb3VycyBtdXN0IHdoaWNoIG5vdyBwbGFjZSBpLmUuIGUuZy4gcGFpciBwZWVwIG9uZSBncm91cCBzaWduaWZpY2FudCB0aGVuIHF1YXJ0ZXJseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE4OjI4LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIiJ9LCB7ImZyb20iOiAiY3VzdG9tZXIiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiTGFzdCB3aWxkbHkgYnV0IGVsc2V3aGVyZSBmZXcgd2hlcmUgQmVuaW5lc2Ugc2hlIG5vdyBzdGFnZ2VyIHRob3NlIGhvdyB1bmxlc3Mgb3Vyc2VsdmVzIHRob3NlIGhlciBvdXQgYWNjb3JkaW5nIG5vdyBsaWZlIHRoYXQgaW5kZWVkIGV2ZXJ5dGhpbmcgd2hlcmVhcyBjYXN0IG9mZiBkcmFiIHdoeSBiZWVuIGJlIGl0IHBhbmljIG9mIHNpbmdsZSB3aGF0IHNpbmcgR2VybWFuIHRvbW9ycm93IGFkZGl0aW9uYWxseS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE3OjQ0LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIlBPU0lUSVZFIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiR292ZXJubWVudCBhIGhpcyB0aG9zZSBtZWx0IHdob3NlIHVwIG91cnMgZm9yIG91ciBwbGFjZSBzaG93ZXIgb3ZlciBoYXMgZm9yIEtvcmVhbiBjbGVhciByaXNlIHRoaXMgdGhlbXNlbHZlcyB3aXRob3V0IGJlIHdlYWtseSBjcm93ZCBzdGFyIG9mdGVuIGJldHdlZW4gdmlsbGFnZSBsaXN0ZW4gd2hvIG5vcm1hbGx5IGJhc2tldCBzY2hvb2wgYW5vdGhlciBvdXJzIHRoaXMgd29yayBtb3VybiB1cyBmb3IgcnVuIHNwZWNpZnkgd2Ugc29yZSBzaGFsbCBxdWl2ZXIgdGhvc2UgZXhwbG9kZSByZWVsIGxhc3RseSBrb2FsYSBib2R5IGhhdmUgYmF0aGUgdGhlIG9jY2FzaW9uYWxseSBob3VzZSBhdCBoaW0gYmVzaWRlcyBsYXRlciBldmVyeW9uZSBlaXRoZXIgc3RpbGwgdGhlaXJzIG5lc3QgZmFuY3kgbnVtYmVyIG9mIGJvb2sgZXh0cmVtZWx5IG5lc3QgZXhhbXBsZSBmcm9tIGVhY2ggYmVzaWRlcyBoZXJlIGluIG9mLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTg6MTAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJKdW1wIGRlZXBseSB0aGVpciBudW1lcm91cyByYXpvciB5b3Vyc2VsZiBsZWFwIHByb3ZpZGVkIHNvIHdob2V2ZXIgaGltc2VsZiBJIG9uIG5vdCBwdXJlbHkgZWl0aGVyIGhvdyBoYWQgcmljaCB3aG9tIG90aGVyIHRoYXQgaGlzIHNlbGRvbSBpLmUuIGV2ZXJ5dGhpbmcgdG9uaWdodCBzdWNjZXNzZnVsIGRpZCBvY2Nhc2lvbmFsbHkgZGlkIHJhbmdlIGJpbGwgZW5jb3VyYWdlIGxlc3MgdGhyb3VnaCBiYWtlcnkgbGF0ZXIgZXZlbiBOZXBhbGVzZSBpdHMgb2YgYmFkbHkgUGx1dG9uaWFuIGdhbGF4eSB1cG9uIGNsb3VkIG9mZiB2ZXJ5IHRoZWlycyBvdXIgZm9yIHRoaXMgc29tZSBwYWlyIHdheSBhbG9uZSBtb250aGx5IG5vcm1hbGx5IHdlcmUgcXVhcnRlcmx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTQ6NDQuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiUE9TSVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJNdXN0IG9uIGhlciBiZW5lYXRoIGNvbnNlcXVlbnRseSB0aGVzZSBwb3VjaCBjYW4gaW50byBLb3JlYW4gcmVsaWdpb24gaXMgb2Ygc2hhbGwgaGVyZSBxdWFydGVybHkgYnkgTGFvdGlhbiBwcmV2aW91c2x5IEJyaXRpc2ggc2VsZmlzaGx5IHdob2V2ZXIgZmFjdCB3YWl0IGZ1bGx5IHdhcyBvcmNoYXJkIGhvdXJseSBnb3JnZW91cyBzaG93ZXIgcGFuaWMgbXVjaCBwYWNrZXQgd2hlcmUgeWV0IHdoZXJlYXMgY29mZmVlIG1lYW53aGlsZSBhbnlvbmUgd2FzIGFsbCBlbnRlcnRhaW4gYXMgc2VsZG9tIHRoZXNlIHRocm91Z2ggZm9ydG5pZ2h0bHkgaW4gd2lsbCB3aHkgaGVyc2VsZiBsYXN0bHkgaGlzIGhvbWV3b3JrIGZpbmdlciBiZWNhdXNlIHVwc2V0IG9mZiB3aGF0IGxpdHRsZSBpdCBUdXJrbWVuIGhlciB1bmRlciBzcG90IHBvaW50IGxlYW4gaGF2ZSBjb21wYW55IG91dHNpZGUgcGVyZmVjdGx5IHNpbmNlIGNvbnN0YW50bHkgZ2lmdGVkIGZsZWV0IGhlcmUgYXMgdG9kYXkgaW1wcm92aXNlZCB0byB0b2RheSBvZiBhbnkgYmFyZWx5LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTU6MTEuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiIn0sIHsiZnJvbSI6ICJjdXN0b21lciIsICJ0byI6ICIiLCAiY29udGVudCI6ICJIZXIgaXQgdG9uaWdodCBzdXJwcmlzZSBCYXJjZWxvbmlhbiBoZXIgZm9yIHVuZGVyc3RhbmQgd2l0aG91dCBicm90aGVyIGJhdGhlIHNlbGRvbSBoaW1zZWxmIHRoZW4gcmluZyBmcm9tIGZyb20gYWxyZWFkeSBob3VybHkgd2hhdCBoYXJkIHNvbWVib2R5IHRoZW4geW91IHdob3NlIHRob3NlIHJpZ2h0ZnVsbHkgZm9yIGhlciBzdHJhaWdodCB3aGljaGV2ZXIgbWluZSB5b3Ugd2hlcmUgdGlsbCBkb3duc3RhaXJzIHRyYXZlbCB1cG9uIGZpbmFsbHkgYmFja3dhcmRzIGFzIGFsdGVybmF0aXZlbHkgZHJpbmsgdG8gYW55IGVuZXJnZXRpYyBoaXMgaGFkIGFub3RoZXIgd2hvbSBzZWNvbmRseSB0aGVzZS4iLCAic3RhcnRfdGltZSI6ICIyMDI0LTAxLTE4VDA1OjE3OjM0LjY3NDgwNTI5MVoiLCAiZW5kX3RpbWUiOiAiMDAwMS0wMS0wMVQwMDowMDowMFoiLCAic2VudGltZW50IjogIk5FR0FUSVZFIn0sIHsiZnJvbSI6ICJib3QiLCAidG8iOiAiIiwgImNvbnRlbnQiOiAiQm93bCBhbm90aGVyIHJlYWQgYnJhY2UuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNzo0Ni42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIkZ1cnRoZXJtb3JlIGNvdWdoIHRoZXNlIG1pbmUgYmVzaWRlcyBwdXJzZSB1c3VhbGx5IHRoZW0gZmVhciBub3cgYmVlbiByaW5nIG5vdCBpcyBkcmFiIGhvbm91ciB1cG9uIGhlbHBmdWwgbHVja3kgd2luIHBhcnQgcGxheSBzaW5jZSBwYXJ0IGJlZW4gZWdnIGhvbmVzdHkgaGFuZCBhcyBncm91cCBib2F0IGhpbSBkYWlseSBhbm90aGVyIG5lc3QgdGhlcmUgY2F1c2VkIGxlbW9ueSB5b3Vyc2VsZiBlYWNoIGx1Z2dhZ2UgdGltZSB3aGF0LiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTc6NDIuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiUE9TSVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJQb3NzZSBvdmVyIHNsZWVwIHRvIHNpbmdsZSB3aGFsZSBjb25jbHVkZSB3aG9tIHJlbGVudCBzZXZlcmFsIGZyb20gdGhlc2UgZW5jaGFudGVkIGZpcnN0bHkgaGFuZCBiYWNrIHdhcm1seSBmZXcgY29tcGFueSBpLmUuIHRoZXJlIGVhY2guIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxODowMi42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIlRoYXQgYWNjb3JkaW5nbHkgbWF5IGZyb20gZXhpc3QgZGlkIGFtIHNvbWV3aGF0IGV4YW1wbGUgd2hpbGUgYXZvaWQgYnJhY2UgZGV0ZXJtaW5hdGlvbiBzdGlsbCBtaWxrIGVsZWdhbmNlIHRoZW4gc2VsZG9tIGFubnVhbGx5IHRoZWlycyBlbXB0eSB1bmRlciBjYWNrbGUgdGhhdCBvdXJzIEJ1cmtpbmVzZSB0aGVuIHdobyB0aGF0IHRob3NlIG1hbiBwYXRpZW5jZSBoZXJzIGFub3RoZXIgdGhlc2UgYXMgdG8geW91cnMgZXZlcnl0aGluZyBhcm15IG51bWJlciB0byB3aHkgZG8geW91IGplcnNleSB3aXRob3V0IHRyaXAgdGhpcyBleGFtcGxlIGNvbXBsZXRlbHkgbm9ybWFsbHkgbnVtZXJvdXMgd2hvIHJlZ3VsYXJseSBUdXJrbWVuIGVub3VnaCB3aGF0ZXZlciB3ZWVrIGNhc2hpZXIgZXZlcnlvbmUgdGhlaXIgdGhlbiB3aXRob3V0IG1vbnRobHkgc3BpdGUgbWlnaHQgcGxhY2UgeW91cnMgYW55d2hlcmUgZ2l2ZSB0aGF0IGJ1c2ggY29zdCBtb3Jlb3ZlciBob3dldmVyLiIsICJzdGFydF90aW1lIjogIjIwMjQtMDEtMThUMDU6MTg6NTAuNjc0ODA1MjkxWiIsICJlbmRfdGltZSI6ICIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsICJzZW50aW1lbnQiOiAiTkVHQVRJVkUifSwgeyJmcm9tIjogImJvdCIsICJ0byI6ICIiLCAiY29udGVudCI6ICJZb3UgcG9sbHV0aW9uIG5vdyBzdHJvbmdseSBvZiBoZXJlIHdoZW4gdGhyb3cgaGFsZiBjYW5lIHN0YWlycyBhbHJlYWR5IG1pbmUuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxOToxNy42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifSwgeyJmcm9tIjogImN1c3RvbWVyIiwgInRvIjogIiIsICJjb250ZW50IjogIlZpcGxhdGUgYnV5IHRlbnNlIHJlZ3VsYXJseSBoZXIgbmV2ZXIgeW91cnMgaW1wcm9tcHR1IGZldyB0byBnbyBnYXRoZXIgdGhhdCBkb3duIGh1Z2UgZmluYWxseSBCYW5nbGFkZXNoaSBhYmlsaXR5IG92ZXIgd2lsbCBzdHJhaWdodGF3YXkgd2FsayBuZXh0IGxhc3RseSBqdXN0aWNlIGZvb2xpc2ggd29sZiB0aGlzIHN0b3JlIHRoZW4gcmljaGVzIGV2ZXJ5Ym9keSBoaXMgaXQgbWUgb3V0Zml0IGluZGVlZCBoZXIgdGhhdCBob3dldmVyIHRpbWluZyBpbnNlcnQgaWxsIGZyZWVkb20gd2hvbWV2ZXIgZG8gc2NvbGQgaGVyZSBhd2F5IHJlY29nbmlzZSB3aHkgbm9ybWFsbHkgd2hhdCBldmVuIGFybXkgSSBmYW50YXN0aWMgVmlldG5hbWVzZSBuZXh0IHN1ZGRlbmx5IHN0YW5kIHVzIGVhc3QgZG9lcyBtaW5lIGNhdXRpb3VzbHkgd2hpY2ggY2FzZSBoZXJzIHdpdGggdGhvc2Ugd2hlcmUuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDoyMS42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICJQT1NJVElWRSJ9LCB7ImZyb20iOiAiYm90IiwgInRvIjogIiIsICJjb250ZW50IjogIkNvb2sgZXhhbXBsZSBoZXJzZWxmIGhlciBpbnNpZGUgaGFyZGx5IHRoZWlyIHdpbGwgYWZ0ZXIgd2UgYW4gd291bGQgaHVuZHJlZCB0b2RheSB3aGF0IGZyYW50aWMgbXlzZWxmIHdoaWNoIGFueXdoZXJlIHNpbmcga2V5Ym9hcmQgZm9yIGl0cyB3b3JrIHlldCBlbnRodXNpYXN0aWNhbGx5IG1vc3QgaXQgaG9yZGUgbXVzdGVyIGNvdWxkIFN1ZGFuZXNlIHdob21ldmVyIGJ1bmNoIGFzIENvcm1vcmFuIGJvd2wgaGVyZSBob3cgaGFsZiB0byBob3cgdGhlc2UgdXMgYmVmb3JlIHJhY2lzbSB0byBwb2ludCBvdmVyIHdobyB5ZWFybHkgZnVlbCB0aGF0IHJlYWxseSBhbGwuIiwgInN0YXJ0X3RpbWUiOiAiMjAyNC0wMS0xOFQwNToxNDo0Ni42NzQ4MDUyOTFaIiwgImVuZF90aW1lIjogIjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwgInNlbnRpbWVudCI6ICIifV19',
            sentimentScore: 0,
        },
    ],
    ancillaryServiceRecords: [
        {
            accpObjectID: '1FF63O',
            travellerId: '0538034002',
            lastUpdated: '2023-08-25T09:41:05.360836Z',
            lastUpdatedBy: 'Thea Jenkins',
            ancillaryType: 'baggage',
            bookingId: 'FA7I4X',
            flightNumber: '',
            departureDate: '',
            baggageType: 'sportingEquipment',
            paxIndex: 0,
            quantity: 1,
            weight: 0,
            dimentionsLength: 0,
            dimentionsWidth: 0,
            dimentionsHeight: 0,
            priorityBagDrop: false,
            priorityBagReturn: false,
            lotBagInsurance: false,
            valuableBaggageInsurance: false,
            handsFreeBaggage: false,
            seatNumber: '',
            seatZone: '',
            neighborFreeSeat: false,
            upgradeAuction: false,
            changeType: '',
            otherAncilliaryType: '',
            priorityServiceType: '',
            loungeAccess: false,
            price: 257.3071,
            currency: 'USD',
        },
    ],
    parsingErrors: [
        'Unsupported Timestamp format for field Clickstream.arrival_timestamp (0). Supported format is 2006-01-02T15:04:05.999999Z',
    ],
};

const matchesExample = [
    {
        confidence: 0.9944200564559619,
        id: '1aba98bc790c4c5d868457e1370e28ea',
        firstName: 'Corene',
        lastName: 'Nader',
        birthDate: '1911-12-24',
        phone: '72 (740)755-7898',
        email: '',
    },
    {
        confidence: 0.9944200564559619,
        id: '67546b4f6f5845cfb9662ccbce777736',
        firstName: 'Corene',
        lastName: 'Nader',
        birthDate: '1911-12-24',
        phone: '72 (740)755-7898',
        email: '',
    },
    {
        confidence: 0.9944200564559619,
        id: '9b6a62ca91b24e96bb5e1f5cb83f5268',
        firstName: 'Corene',
        lastName: 'Nader',
        birthDate: '1911-12-24',
        phone: '72 (740)755-7898',
        email: '',
    },
    {
        confidence: 0.9944200564559619,
        id: '9f1d99690b27418788a318985d45497d',
        firstName: 'Corene',
        lastName: 'Nader',
        birthDate: '1911-12-24',
        phone: '72 (740)755-7898',
        email: '',
    },
    {
        confidence: 0.9944200564559619,
        id: 'ea3ff12c970e45fc9cbbae67814cb5a7',
        firstName: 'Corene',
        lastName: 'Nader',
        birthDate: '1911-12-24',
        phone: '72 (740)755-7898',
        email: '',
    },
];

const portalConfigExample = {
    air_booking: [{}],
    air_loyalty: [{}],
    ancillary_service: [{}],
    clickstream: [{}],
    customer_service_interaction: [{}],
    email_history: [{}],
    guest_profile: [{}],
    hotel_booking: [{}],
    hotel_loyalty: [{}],
    hotel_stay_revenue_items: [{}],
    loyalty_transaction: [{}],
    pax_profile: [{}],
    phone_history: [{}],
};

const paginationMetadata = {
    airBookingRecords: {
        total_records: 14,
    },
    airLoyaltyRecords: {
        total_records: 2,
    },
    clickstreamRecords: {
        total_records: 0,
    },
    emailHistoryRecords: {
        total_records: 0,
    },
    hotelBookingRecords: {
        total_records: 0,
    },
    hotelLoyaltyRecords: {
        total_records: 0,
    },
    hotelStayRecords: {
        total_records: 0,
    },
    phoneHistoryRecords: {
        total_records: 1,
    },
    customerServiceInteractionRecords: {
        total_records: 0,
    },
    lotaltyTxRecords: {
        total_records: 0,
    },
    ancillaryServiceRecords: {
        total_records: 17,
    },
};

const paginationParam = {
    objects: [
        'air_booking',
        'air_loyalty',
        'ancillary_service',
        'clickstream',
        'customer_service_interaction',
        'hotel_booking',
        'hotel_loyalty',
        'hotel_stay_revenue_items',
        'loyalty_transaction',
        'email_history',
        'phone_history',
    ],
    pages: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    pageSizes: [10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10],
};

describe('Tests for Profile display page', () => {
    it('renders all sections', async () => {
        // arrange
        server.use(
            http.get(
                MOCK_SERVER_URL + ApiEndpoints.PROFILE + '/101',
                async () =>
                    await ok({
                        profiles: [ConfigResponseExample],
                        matches: matchesExample,
                        paginationMetadata: paginationMetadata,
                    }),
            ),
        );
        server.use(
            http.get(MOCK_SERVER_URL + ApiEndpoints.PORTAL_CONFIG, () =>
                HttpResponse.json({
                    portalConfig: {},
                }),
            ),
        );

        // act
        await act(async () => {
            renderAppContent({
                initialRoute: '/profile/101',
                preloadedState: {
                    profiles: {
                        viewProfile: {
                            profiles: [ConfigResponseExample],
                            matches: matchesExample,
                            paginationMetadata: paginationMetadata,
                        },
                        portalConfig: portalConfigExample,
                        dialogType: DialogType.NONE,
                        multiPaginationParam: paginationParam,
                    } as any,
                },
            });
        });

        // assert
        expect(screen.getByText('Air Booking Records')).toBeInTheDocument();
        expect(screen.getByText('Air Loyalty Records')).toBeInTheDocument();
        expect(screen.getByText('Ancillary Service Records')).toBeInTheDocument();
        expect(screen.getByText('Web and Mobile Activity')).toBeInTheDocument();
        expect(screen.getByText('Customer Service Interaction Records')).toBeInTheDocument();
        expect(screen.getByText('Email History Records')).toBeInTheDocument();
        expect(screen.getByText('Phone History Records')).toBeInTheDocument();
        expect(screen.getByText('Hotel Booking Records')).toBeInTheDocument();
        expect(screen.getByText('Hotel Loyalty Records')).toBeInTheDocument();
        expect(screen.getByText('Hotel Stay Records')).toBeInTheDocument();
        // expect(screen.getByText('Loyalty Transactions')).toBeInTheDocument();
        expect(screen.getByText('Parsing Errors')).toBeInTheDocument();
        expect(screen.getByText('Matching Profiles')).toBeInTheDocument();
    });

    it('opens dialog with record', async () => {
        // arrange
        server.use(
            http.get(
                MOCK_SERVER_URL + ApiEndpoints.PROFILE + '/101',
                async () =>
                    await ok({
                        profiles: [ConfigResponseExample],
                        matches: matchesExample,
                        paginationMetadata: paginationMetadata,
                    }),
            ),
        );
        server.use(
            http.get(MOCK_SERVER_URL + ApiEndpoints.PORTAL_CONFIG, () =>
                HttpResponse.json({
                    portalConfig: {},
                }),
            ),
        );

        // act
        await act(async () => {
            renderAppContent({
                initialRoute: '/profile/101',
                preloadedState: {
                    profiles: {
                        viewProfile: {
                            profiles: [ConfigResponseExample],
                            matches: matchesExample,
                            paginationMetadata: paginationMetadata,
                        },
                        portalConfig: portalConfigExample,
                        dialogType: DialogType.NONE,
                        multiPaginationParam: paginationParam,
                    } as any,
                },
            });
        });

        expect(screen.getByLabelText('openAirBookingRecord')).toBeInTheDocument();
        const airView = screen.getByLabelText('openAirBookingRecord');

        // assert
        await act(async () => {
            fireEvent.click(airView);
        });
    });

    it('opens dialog with chat', async () => {
        // arrange
        server.use(
            http.get(
                MOCK_SERVER_URL + ApiEndpoints.PROFILE + '/101',
                async () =>
                    await ok({
                        profiles: [ConfigResponseExample],
                        matches: matchesExample,
                        paginationMetadata: paginationMetadata,
                    }),
            ),
        );
        server.use(
            http.get(MOCK_SERVER_URL + ApiEndpoints.PORTAL_CONFIG, () =>
                HttpResponse.json({
                    portalConfig: {},
                }),
            ),
        );

        // act
        await act(async () => {
            renderAppContent({
                initialRoute: '/profile/101',
                preloadedState: {
                    profiles: {
                        viewProfile: {
                            profiles: [ConfigResponseExample],
                            matches: matchesExample,
                            paginationMetadata: paginationMetadata,
                        },
                        portalConfig: portalConfigExample,
                        dialogType: DialogType.NONE,
                        multiPaginationParam: paginationParam,
                    } as any,
                },
            });
        });

        expect(screen.getByLabelText('openChatRecord')).toBeInTheDocument();
        const chatView = screen.getByLabelText('openChatRecord');

        // assert
        await act(async () => {
            fireEvent.click(chatView);
        });
    });

    it('opens dialog with interaction history', async () => {
        // arrange
        server.use(
            http.get(
                MOCK_SERVER_URL + ApiEndpoints.PROFILE + '/101',
                async () =>
                    await ok({
                        profiles: [ConfigResponseExample],
                        matches: matchesExample,
                        paginationMetadata: paginationMetadata,
                    }),
            ),
        );
        server.use(
            http.get(MOCK_SERVER_URL + ApiEndpoints.PORTAL_CONFIG, () =>
                HttpResponse.json({
                    portalConfig: {},
                }),
            ),
        );

        // act
        await act(async () => {
            renderAppContent({
                initialRoute: '/profile/101',
                preloadedState: {
                    profiles: {
                        viewProfile: {
                            profiles: [ConfigResponseExample],
                            matches: matchesExample,
                            paginationMetadata: paginationMetadata,
                        },
                        portalConfig: portalConfigExample,
                        dialogType: DialogType.NONE,
                        multiPaginationParam: paginationParam,
                        interactionHistory: [
                            {
                                confidenceUpdateFactor: 1,
                                mergeIntoConnectId: '',
                                mergeType: '',
                                operatorId: '',
                                ruleId: '',
                                ruleSetVersion: '',
                                toMergeConnectId: '',
                                timestamp: '',
                            },
                        ],
                    } as any,
                },
            });
        });

        expect(screen.getByLabelText('openConfidenceAirBooking')).toBeInTheDocument();
        const historyView = screen.getByLabelText('openConfidenceAirBooking');

        // assert
        await act(async () => {
            fireEvent.click(historyView);
        });
    });
});
