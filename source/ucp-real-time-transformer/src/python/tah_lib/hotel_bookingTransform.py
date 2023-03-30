
import uuid
import traceback
from datetime import datetime

from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setBillingAddress, setTravellerId, getExternalId, setPaymentInfo, buildSerializedLists, setTimestamp, parseNumber, logErrorToSQS

BIZ_OBJECT_TYPE = "hotel_booking"


def buildObjectRecord(rec, errQueueUrl):
    try:
        bookingRecs = []
        emailRecs = []
        phoneRecs = []
        loyaltyRecs = []

        if len(rec.get("segments", [])) == 0:
            print("Invalid format Exception")
            raise Exception("Hotel booking must have at least 1 segment ")

        for seg in rec.get("segments", []):
            cid = str(uuid.uuid1(node=None, clock_seq=None))
            for product in seg.get("products", []):
                addBookingRec(bookingRecs, rec, seg,
                              product, seg.get("holder", {}), cid)
                addEmailRecs(emailRecs, rec, seg, seg.get("holder", {}), cid)
                addPhoneRecs(phoneRecs, rec, seg, seg.get("holder", {}), cid)
                addLoyaltyRecs(loyaltyRecs, rec, seg,
                               seg.get("holder", {}), cid)
            for addGuest in seg.get("additionalGuests", []):
                cid = str(uuid.uuid1(node=None, clock_seq=None))
                for product in seg.get("products", []):
                    addBookingRec(bookingRecs, rec, seg,
                                  product, addGuest, cid)
                    addEmailRecs(emailRecs, rec, seg, addGuest, cid)
                    addPhoneRecs(phoneRecs, rec, seg, addGuest, cid)
                    addLoyaltyRecs(loyaltyRecs, rec, seg, addGuest, cid)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        logErrorToSQS(e, rec, errQueueUrl, BIZ_OBJECT_TYPE)
    return {
        'hotel_booking_recs': bookingRecs,
        'common_email_recs': emailRecs,
        'common_phone_recs': phoneRecs,
        'guest_loyalty_recs': loyaltyRecs,
    }


def addBookingRec(bookingRecs, rec, seg, product, guest, cid):
    hotelBookingRec = {
        'model_version': rec.get("modelVersion", ""),
        'object_type': "hotel_booking",
        'last_updated': rec.get("lastUpdatedOn", ""),
        'last_updated_by': rec.get("lastUpdatedBy", ""),
        'booking_id': rec.get("id", ""),
        'hotel_code': seg.get("hotelCode", ""),
        'n_nights': parseNumber(rec.get("nNights", "")),
        'n_guests': parseNumber(rec.get("nGuests", "")),
        'product_id': product.get("id", ""),
        'check_in_date': rec.get("startDate", ""),
        'honorific': guest.get('honorific', ''),
        'first_name': guest.get('firstName', ''),
        'middle_name': guest.get('middleName', ''),
        'last_name': guest.get('lastName', ''),
        'gender': guest.get('gender', ''),
        'pronoun': guest.get('pronoun', ''),
        'date_of_birth': guest.get('dateOfBirth', ''),
        'job_title': guest.get('jobTitle', ''),
        'company': guest.get('parentCompany', ''),
        'totalAfterTax': parseNumber(seg.get('price', {}).get("totalAfterTax", "")),
        'totalBeforeTax': parseNumber(seg.get('price', {}).get("totalBeforeTax", "")),
    }
    hotelBookingRec["nationality_code"] = rec.get(
        'nationality', {}).get("code", "")
    hotelBookingRec["nationality_name"] = rec.get(
        'nationality', {}).get("name", "")
    hotelBookingRec["language_code"] = rec.get('language', {}).get("code", "")
    hotelBookingRec["language_name"] = rec.get(
        'language', {}).get("name", "")
    hotelBookingRec['pms_id'] = getExternalId(
        rec.get("externalIds", []), "pms")
    hotelBookingRec['crs_id'] = getExternalId(
        rec.get("externalIds", []), "crs")
    hotelBookingRec['gds_id'] = getExternalId(
        rec.get("externalIds", []), "gds")
    hotelBookingRec['room_type_code'] = rec.get(
        "roomType", {}).get("code", "")
    hotelBookingRec['room_type_name'] = rec.get(
        "roomType", {}).get("name", "")
    hotelBookingRec['room_type_description'] = rec.get(
        "roomType", {}).get("description", "")
    hotelBookingRec['room_type_code'] = rec.get(
        "ratePlan", {}).get("code", "")
    hotelBookingRec['room_type_name'] = rec.get(
        "ratePlan", {}).get("name", "")
    hotelBookingRec['room_type_description'] = rec.get(
        "ratePlan", {}).get("description", "")
    hotelBookingRec['attribute_codes'] = buildSerializedLists(
        rec.get("attributes", []), "code", "|")
    hotelBookingRec['attribute_names'] = buildSerializedLists(
        rec.get("attributes", []), "name", "|")
    hotelBookingRec['attribute_descriptions'] = buildSerializedLists(
        rec.get("attributes", []), "description", "|")
    hotelBookingRec['attribute_codes'] = buildSerializedLists(
        rec.get("addOns", []), "code", "|")
    hotelBookingRec['attribute_names'] = buildSerializedLists(
        rec.get("addOns", []), "name", "|")
    hotelBookingRec['attribute_descriptions'] = buildSerializedLists(
        rec.get("addOns", []), "description", "|")
    setPaymentInfo(hotelBookingRec, rec.get("paymentInformation", {}))
    setBillingAddress(hotelBookingRec, rec.get("paymentInformation", {}))
    # Set primary option for phone/email/address
    setPrimaryEmail(hotelBookingRec, guest.get('emails', []))
    setPrimaryPhone(hotelBookingRec, guest.get('phones', []))
    setPrimaryAddress(hotelBookingRec, guest.get('addresses', []))
    # set traveller ID
    setTravellerId(hotelBookingRec, guest, cid)
    setTimestamp(hotelBookingRec)
    bookingRecs.append(hotelBookingRec)


def addEmailRecs(emailRecs, rec, seg, guest, cid):
    for email in guest.get("emails", []):
        historicalEmail = {
            'model_version': rec.get("modelVersion", ""),
            'object_type': "email_history",
            'last_updated': rec.get("lastUpdatedOn", ""),
            'last_updated_by': rec.get("lastUpdatedBy", ""),
            'address': email.get("address", ""),
            'type': email.get("type", ""),
        }
        setTravellerId(historicalEmail, guest, cid)
        setTimestamp(historicalEmail)
        emailRecs.append(historicalEmail)


def addPhoneRecs(phoneRecs, rec, seg, guest, cid):
    for phone in guest.get("phones", []):
        historicalPhone = {
            'model_version': rec.get("modelVersion", ""),
            'object_type': "phone_history",
            'last_updated': rec.get("lastUpdatedOn", ""),
            'last_updated_by': rec.get("lastUpdatedBy", ""),
            'number': phone.get("number", ""),
            'country_code': phone.get("countryCode", ""),
            'type': phone.get("type", ""),
        }
        setTravellerId(historicalPhone, guest, cid)
        setTimestamp(historicalPhone)
        phoneRecs.append(historicalPhone)


def addLoyaltyRecs(loyaltyRecs, rec, seg, guest, cid):
    if 'loyaltyPrograms' in guest:
        for loyalty in guest.get('loyaltyPrograms', []):
            loyaltyRec = {
                'object_type': 'hotel_loyalty',
                'model_version': rec.get('modelVersion', ''),
                'last_updated': rec.get('lastUpdatedOn', ''),
                'last_updated_by': rec.get('lastUpdatedBy', ''),
                'id': loyalty.get('id', ''),
                'program_name': loyalty.get('programName', ''),
                'points': loyalty.get('points', ''),
                'units': loyalty.get('pointUnit', ''),
                'points_to_next_level': loyalty.get('pointsToNextLevel', ''),
                'level': loyalty.get('level', ''),
                'joined': loyalty.get('joined', '')
            }
            setTravellerId(loyaltyRec, guest, cid)
            setTimestamp(loyaltyRec)
            loyaltyRecs.append(loyaltyRec)
