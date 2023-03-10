
import uuid
import traceback
from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setBillingAddress, setTravellerId, getExternalId, setPaymentInfo


def noneToList(val):
    if val is None:
        return []
    return val


def buildObjectRecord(rec):
    try:
        bookingRecs = []
        emailRecs = []
        phoneRecs = []
        loyaltyRecs = []
        if "passengerInfo" in rec and "passengers" in rec["passengerInfo"]:
            for pax in rec["passengerInfo"]["passengers"]:
                cid = str(uuid.uuid1(node=None, clock_seq=None))
                for seg in rec.get("itinerary", {}).get("segments", []):
                    addAirBookingRecord(bookingRecs, rec, pax, seg, cid)
                for seg in noneToList(rec.get("return", {}).get("segments", [])):
                    addAirBookingRecord(bookingRecs, rec, pax, seg, cid)
                for email in pax.get("emails", []):
                    historicalEmail = {
                        'model_version': rec.get("modelVersion", ""),
                        'object_type': "email_history",
                        'last_updated': rec.get("lastUpdatedOn", ""),
                        'last_updated_by': rec.get("lastUpdatedBy", ""),
                        'address': email.get("address", ""),
                        'type': email.get("type", ""),
                    }
                    setTravellerId(historicalEmail, pax, cid)
                    emailRecs.append(historicalEmail)
                for phone in pax.get("phones", []):
                    historicalPhone = {
                        'model_version': rec.get("modelVersion", ""),
                        'object_type': "phone_history",
                        'last_updated': rec.get("lastUpdatedOn", ""),
                        'last_updated_by': rec.get("lastUpdatedBy", ""),
                        'number': phone.get("number", ""),
                        'country_code': phone.get("countryCode", ""),
                        'type': phone.get("type", ""),
                    }
                    setTravellerId(historicalPhone, pax, cid)
                    phoneRecs.append(historicalPhone)
                for loyalty in pax.get("loyaltyPrograms", []):
                    loyaltyRec = {
                        'model_version': rec.get("modelVersion", ""),
                        'object_type': "air_loyalty",
                        'last_updated': rec.get("lastUpdatedOn", ""),
                        'last_updated_by': rec.get("lastUpdatedBy", ""),
                        "id": loyalty.get("id", ""),
                        "program_name": loyalty.get("programName", ""),
                        "miles": loyalty.get("id", ""),
                        "miles_to_next_level": loyalty.get("milesToNextLevel", ""),
                        "level": loyalty.get("level", ""),
                        "joined": loyalty.get("joined", "")
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
        'model_version': rec.get("modelVersion", ""),
        'object_type': "air_booking",
        'last_updated': rec.get("lastUpdatedOn", ""),
        'last_updated_by': rec.get("lastUpdatedBy", ""),
        'booking_id': rec.get("id", ""),
        'segment_id': buildSegmentId(rec, seg),
        'from': seg.get("from", ""),
        'to': seg.get("to", ""),
        'flight_number': seg.get("flightNumber", ""),
        'departure_date': seg.get("departureDate", ""),
        'departure_time': seg.get("departureTime", ""),
        'arrival_date': seg.get("arrivalDate", ""),
        'arrival_time': seg.get("arrivalTime", ""),
        'channel': rec.get("creationChannelId", ""),
        'status': rec.get("status", ""),
        "honorific": "Ms",
        "first_name": pax.get("firstName", ""),
        "middle_name": pax.get("middleName", ""),
        "last_name": pax.get("lastName", ""),
        "gender": pax.get("gender", ""),
        "pronoun": pax.get("pronoun", ""),
        "date_of_birth": pax.get("dateOfBirth", ""),
        "job_title": pax.get("jobTitle", ""),
        "company": pax.get("parentCompany", ""),
        "price": rec.get("price", {}).get("grandTotal", ""),
    }
    if "nationality" in rec:
        airBookingRec["nationality_code"] = rec.get(
            'nationality', {}).get("code")
        airBookingRec["nationality_name"] = rec.get(
            'nationality', {}).get("name")
    if "language" in rec:
        airBookingRec["language_code"] = rec.get('language', {}).get("code")
        airBookingRec["language_name"] = rec.get('language', {}).get("name")
    if "externalIds" in rec:
        airBookingRec['pss_id'] = getExternalId(
            rec.get("externalIds", ""), "pss")
        airBookingRec['gds_id'] = getExternalId(
            rec.get("externalIds", ""), "gds")

    setPaymentInfo(airBookingRec, rec.get("paymentInformation"))
    setBillingAddress(airBookingRec, rec.get("paymentInformation"))

    # Set primary option for phone/email/address
    setPrimaryEmail(airBookingRec, pax.get('emails', []))
    setPrimaryPhone(airBookingRec, pax.get('phones', []))
    setTravellerId(airBookingRec, pax, cid)
    setPrimaryAddress(airBookingRec, pax.get('addresses', []))
    recs.append(airBookingRec)


def buildSegmentId(rec, seg):
    return rec["id"]+"-"+seg.get("from", "")+"-"+seg.get("to", "")
