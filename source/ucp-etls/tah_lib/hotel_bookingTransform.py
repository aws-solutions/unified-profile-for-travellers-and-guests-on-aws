
import uuid
import traceback
from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setBillingAddress, setTravellerId, getExternalId, setPaymentInfo, buildSerializedLists


def buildObjectRecord(rec):
    try:
        bookingRecs = []
        emailRecs = []
        phoneRecs = []
        loyaltyRecs = []

        for seg in rec["segments"]:
            cid = str(uuid.uuid1(node=None, clock_seq=None))
            for product in seg["products"]:
                addBookingRec(bookingRecs, rec, seg,
                              product, seg["holder"], cid)
                addEmailRecs(emailRecs, rec, seg, seg["holder"], cid)
                addPhoneRecs(phoneRecs, rec, seg, seg["holder"], cid)
                addLoyaltyRecs(loyaltyRecs, rec, seg, seg["holder"], cid)
            for addGuest in seg["additionalGuests"]:
                cid = str(uuid.uuid1(node=None, clock_seq=None))
                for product in seg["products"]:
                    addBookingRec(bookingRecs, rec, seg,
                                  product, addGuest, cid)
                    addEmailRecs(emailRecs, rec, seg, addGuest, cid)
                    addPhoneRecs(phoneRecs, rec, seg, addGuest, cid)
                    addLoyaltyRecs(loyaltyRecs, rec, seg, addGuest, cid)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        bookingRecs.append({"error": str(e), "trace": traceback_info})
    return {
        'hotel_booking_recs': bookingRecs,
        'common_email_recs': emailRecs,
        'common_phone_recs': phoneRecs,
        'guest_loyalty_recs': loyaltyRecs,
    }


def addBookingRec(bookingRecs, rec, seg, product, guest, cid):
    hotelBookingRec = {
        'model_version': rec["modelVersion"],
        'object_type': "hotel_booking",
        'last_updated': rec["lastUpdatedOn"],
        'last_updated_by': rec["lastUpdatedBy"],
        'booking_id': rec["id"],
        'hotel_code': seg["hotelCode"],
        'n_nights': rec["nNights"],
        'n_guests': rec["nGuests"],
        'product_id': product["id"],
        'check_in_date': rec["startDate"],
    }
    if "externalIds" in rec:
        rec['pms_id'] = getExternalId(rec["externalIds"], "pms")
        rec['crs_id'] = getExternalId(rec["externalIds"], "crs")
        rec['gds_id'] = getExternalId(rec["externalIds"], "gds")
    if "roomType" in rec:
        hotelBookingRec['room_type_code'] = rec["roomType"]["code"]
        hotelBookingRec['room_type_name'] = rec["roomType"]["name"]
        hotelBookingRec['room_type_description'] = rec["roomType"]["description"]
    if "ratePlan" in rec:
        hotelBookingRec['room_type_code'] = rec["ratePlan"]["code"]
        hotelBookingRec['room_type_name'] = rec["ratePlan"]["name"]
        hotelBookingRec['room_type_description'] = rec["ratePlan"]["description"]
    if "attributes" in rec:
        hotelBookingRec['attribute_codes'] = buildSerializedLists(
            rec["attributes"], "code", "|")
        hotelBookingRec['attribute_names'] = buildSerializedLists(
            rec["attributes"], "name", "|")
        hotelBookingRec['attribute_descriptions'] = buildSerializedLists(
            rec["attributes"], "description", "|")
    if "addOns" in rec:
        hotelBookingRec['attribute_codes'] = buildSerializedLists(
            rec["addOns"], "code", "|")
        hotelBookingRec['attribute_names'] = buildSerializedLists(
            rec["addOns"], "name", "|")
        hotelBookingRec['attribute_descriptions'] = buildSerializedLists(
            rec["addOns"], "description", "|")

    if "paymentInformation" in rec:
        setPaymentInfo(hotelBookingRec, rec["paymentInformation"])
        setBillingAddress(hotelBookingRec, rec["paymentInformation"])
    # Set primary option for phone/email/address
    setPrimaryEmail(hotelBookingRec, guest['emails'])
    setPrimaryPhone(hotelBookingRec, guest['phones'])
    setTravellerId(hotelBookingRec, guest, cid)
    setPrimaryAddress(hotelBookingRec, guest['addresses'])

    bookingRecs.append(hotelBookingRec)


def addEmailRecs(emailRecs, rec, seg, guest, cid):
    for email in guest["emails"]:
        historicalEmail = {
            'model_version': rec["modelVersion"],
            'object_type': "email_history",
            'last_updated': rec["lastUpdatedOn"],
            'last_updated_by': rec["lastUpdatedBy"],
            'address': email["address"],
            'type': email["type"],
        }
        setTravellerId(historicalEmail, guest, cid)
        emailRecs.append(historicalEmail)


def addPhoneRecs(phoneRecs, rec, seg, guest, cid):
    for phone in guest["phones"]:
        historicalPhone = {
            'model_version': rec["modelVersion"],
            'object_type': "phone_history",
            'last_updated': rec["lastUpdatedOn"],
            'last_updated_by': rec["lastUpdatedBy"],
            'number': phone["number"],
            'country_code': phone["countryCode"],
            'type': phone["type"],
        }
        setTravellerId(historicalPhone, guest, cid)
        phoneRecs.append(historicalPhone)


def addLoyaltyRecs(loyaltyRecs, rec, seg, guest, cid):
    if 'loyaltyPrograms' in guest:
        for loyalty in guest['loyaltyPrograms']:
            loyaltyRec = {
                'object_type': 'hotel_loyalty',
                'model_version': rec['modelVersion'],
                'last_updated': rec['lastUpdatedOn'],
                'last_updated_by': rec['lastUpdatedBy'],
                'id': loyalty['id'],
                'program_name': loyalty['programName'],
                'points': loyalty['points'],
                'units': loyalty['pointUnit'],
                'points_to_next_level': loyalty['pointsToNextLevel'],
                'level': loyalty['level'],
                'joined': loyalty['joined']
            }
            setTravellerId(loyaltyRec, guest, cid)
            loyaltyRecs.append(loyaltyRec)
