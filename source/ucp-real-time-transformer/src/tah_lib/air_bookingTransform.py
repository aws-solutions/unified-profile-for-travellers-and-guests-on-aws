
import uuid
import traceback
from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setBillingAddress, setTravellerId, getExternalId, setPaymentInfo


def buildObjectRecord(rec):
    try:
        bookingRecs = []
        emailRecs = []
        phoneRecs = []
        loyaltyRecs = []
        if "passengerInfo" in rec and "passengers" in rec["passengerInfo"]:
            for pax in rec["passengerInfo"]["passengers"]:
                cid = str(uuid.uuid1(node=None, clock_seq=None))
                for seg in rec["itinerary"]["segments"]:
                    addAirBookingRecord(bookingRecs, rec, pax, seg, cid)
                if "return" in rec and rec["return"]["segments"] is not None:
                    for seg in rec["return"]["segments"]:
                        addAirBookingRecord(bookingRecs, rec, pax, seg, cid)
                for email in pax["emails"]:
                    historicalEmail = {
                        'model_version': rec["modelVersion"],
                        'object_type': "email_history",
                        'last_updated': rec["lastUpdatedOn"],
                        'last_updated_by': rec["lastUpdatedBy"],
                        'address': email["address"],
                        'type': email["type"],
                    }
                    setTravellerId(historicalEmail, pax, cid)
                    emailRecs.append(historicalEmail)
                for phone in pax["phones"]:
                    historicalPhone = {
                        'model_version': rec["modelVersion"],
                        'object_type': "phone_history",
                        'last_updated': rec["lastUpdatedOn"],
                        'last_updated_by': rec["lastUpdatedBy"],
                        'number': phone["number"],
                        'country_code': phone["countryCode"],
                        'type': phone["type"],
                    }
                    setTravellerId(historicalPhone, pax, cid)
                    phoneRecs.append(historicalPhone)
                for loyalty in pax["loyaltyPrograms"]:
                    loyaltyRec = {
                        'model_version': rec["modelVersion"],
                        'object_type': "air_loyalty",
                        'last_updated': rec["lastUpdatedOn"],
                        'last_updated_by': rec["lastUpdatedBy"],
                        "id": loyalty["id"],
                        "program_name": loyalty["programName"],
                        "miles": loyalty["id"],
                        "miles_to_next_level": loyalty["milesToNextLevel"],
                        "level": loyalty["level"],
                        "joined": loyalty["joined"]
                    }
                    setTravellerId(loyaltyRec, pax, cid)
                    loyaltyRecs.append(loyaltyRec)

    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        bookingRecs.append({"error": str(e), "trace": traceback_info})
    return {
        'air_booking_recs': bookingRecs,
        'common_email_recs': emailRecs,
        'common_phone_recs': phoneRecs,
        'air_loyalty_recs': loyaltyRecs,
    }


def addAirBookingRecord(recs, rec, pax, seg, cid):
    airBookingRec = {
        'model_version': rec["modelVersion"],
        'object_type': "air_booking",
        'last_updated': rec["lastUpdatedOn"],
        'last_updated_by': rec["lastUpdatedBy"],
        'booking_id': rec["id"],
        'segment_id': buildSegmentId(rec, seg),
        'from': seg["from"],
        'to': seg["to"],
        'flight_number': seg["flightNumber"],
        'departure_date': seg["departureDate"],
        'departure_time': seg["departureTime"],
        'arrival_date': seg["arrivalDate"],
        'arrival_time': seg["arrivalTime"],
        'channel': rec["creationChannelId"],
        'status': rec["status"],
        "honorific": "Ms",
        "first_name": pax["firstName"],
        "middle_name": pax["middleName"],
        "last_name": pax["lastName"],
        "gender": pax["gender"],
        "pronoun": pax["pronoun"],
        "date_of_birth": pax["dateOfBirth"],
        "job_title": pax["jobTitle"],
        "company": pax["parentCompany"],
        "price": rec["price"]["grandTotal"],
    }
    if "nationality" in rec:
        airBookingRec["nationality_code"] = rec['nationality']["code"]
        airBookingRec["nationality_name"] = rec['nationality']["name"]
    if "language" in rec:
        airBookingRec["language_code"] = rec['language']["code"]
        airBookingRec["language_name"] = rec['language']["name"]
    if "externalIds" in rec:
        airBookingRec['pss_id'] = getExternalId(rec["externalIds"], "pss")
        airBookingRec['gds_id'] = getExternalId(rec["externalIds"], "gds")

    if "paymentInformation" in rec:
        setPaymentInfo(airBookingRec, rec["paymentInformation"])
        setBillingAddress(airBookingRec, rec["paymentInformation"])

    # Set primary option for phone/email/address
    setPrimaryEmail(airBookingRec, pax['emails'])
    setPrimaryPhone(airBookingRec, pax['phones'])
    setTravellerId(airBookingRec, pax, cid)
    setPrimaryAddress(airBookingRec, pax['addresses'])
    recs.append(airBookingRec)


def buildSegmentId(rec, seg):
    return rec["id"]+"-"+seg["from"]+"-"+seg["to"]
