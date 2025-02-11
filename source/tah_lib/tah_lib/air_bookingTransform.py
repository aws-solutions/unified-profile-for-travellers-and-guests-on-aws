# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import traceback
import uuid
from typing import Any, Literal, TypedDict

from tah_lib.common import (
    AIR_BOOKING_RECORDS_KEY,
    AIR_LOYALTY_RECORDS_KEY,
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    ANCILLARY_RECORDS_KEY,
    COMMON_EMAIL_RECORDS_KEY,
    COMMON_PHONE_RECORDS_KEY,
    FIELD_NAME_TRAVELLER_ID,
    INVALID_CHARS_SINGLE_LINE,
    ValidationError,
    add_alternate_profile_ids,
    add_extended_data,
    deduplicate_list,
    get_external_id,
    log_error_to_sqs,
    none_to_list,
    parse_bool,
    parse_number,
    set_billing_address,
    set_payment_info,
    set_primary_address,
    set_primary_email,
    set_primary_phone,
    set_timestamp,
    set_traveller_id,
    validate_email,
    validate_phone,
    validate_string,
)

BIZ_OBJECT_TYPE = "air_booking"


class ObjectRecord(TypedDict):
    air_booking_recs: list[dict[str, Any]]
    common_email_recs: list[dict[str, Any]]
    common_phone_recs: list[dict[str, Any]]
    air_loyalty_recs: list[dict[str, Any]]
    ancillary_recs: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    try:
        booking_recs: list[dict[str, Any]] = []
        email_recs: list[dict[str, Any]] = []
        phone_recs: list[dict[str, Any]] = []
        loyalty_recs: list[dict[str, Any]] = []
        ancillary_recs: list[dict[str, Any]] = []
        alternate_profile_id_recs: list[dict[str, Any]] = []
        if "passengerInfo" in rec and "passengers" in rec["passengerInfo"]:
            pax_index = -1.0
            for pax in rec["passengerInfo"]["passengers"]:
                pax_index += 1.0
                cid = str(uuid.uuid4())
                for seg in none_to_list(rec.get("itinerary", {}).get("segments", [])):
                    add_air_booking_record(
                        pax_index,
                        booking_recs,
                        rec,
                        pax,
                        seg,
                        cid,
                        tx_id,
                        alternate_profile_id_recs,
                    )
                for seg in none_to_list(rec.get("return", {}).get("segments", [])):
                    add_air_booking_record(
                        pax_index,
                        booking_recs,
                        rec,
                        pax,
                        seg,
                        cid,
                        tx_id,
                        alternate_profile_id_recs,
                    )
                email_recs = set_email_recs(email_recs, pax, cid, rec, tx_id)
                phone_recs = set_phone_recs(phone_recs, pax, cid, rec, tx_id)
                # LOYALTY PROFILES
                add_loyalty_profile(tx_id, rec, pax, loyalty_recs, cid)
                # ANCILLARY SERVICES
                ancillary = rec.get("ancillaryServices", {})
                add_ancillary_baggage(
                    tx_id, rec, ancillary, ancillary_recs, pax, pax_index, cid
                )
                add_ancillary_seat(
                    tx_id, rec, ancillary, ancillary_recs, pax, pax_index, cid
                )
                add_ancillary_flight_change(
                    tx_id, rec, ancillary, ancillary_recs, pax, pax_index, cid
                )
                add_ancillary_other(tx_id, rec, ancillary, ancillary_recs, pax, cid)
                add_ancillary_priority(tx_id, rec, ancillary, ancillary_recs, pax, cid)
        else:
            print("Invalid format Exception")
            print(rec)
            raise ValueError("Invalid Business object of type", BIZ_OBJECT_TYPE)

    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        AIR_BOOKING_RECORDS_KEY: booking_recs,
        COMMON_EMAIL_RECORDS_KEY: email_recs,
        COMMON_PHONE_RECORDS_KEY: phone_recs,
        AIR_LOYALTY_RECORDS_KEY: loyalty_recs,
        ANCILLARY_RECORDS_KEY: ancillary_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }


def add_loyalty_profile(
    tx_id: str,
    rec: dict[str, Any],
    pax: dict[str, Any],
    loyalty_recs: list[dict[str, Any]],
    cid: str,
) -> None:
    for loyalty in none_to_list(pax.get("loyaltyPrograms", [])):
        loyalty_rec = {
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "air_loyalty",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_string(
                loyalty.get("id", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "id": validate_string(loyalty.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "program_name": validate_string(
                loyalty.get("programName", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "miles": parse_number(loyalty.get("miles", "")),
            "miles_to_next_level": parse_number(loyalty.get("milesToNextLevel", "")),
            "level": validate_string(
                loyalty.get("level", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "joined": validate_string(
                loyalty.get("joined", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "enrollment_source": validate_string(
                loyalty.get("enrollmentSource", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "currency": validate_string(
                loyalty.get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "amount": parse_number(loyalty.get("amount", "")),
            "account_status": validate_string(
                loyalty.get("accountStatus", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "reason_for_close": validate_string(
                loyalty.get("reasonForClose", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "language_preference": validate_string(
                loyalty.get("languagePreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "display_preference": validate_string(
                loyalty.get("displayPreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "meal_preference": validate_string(
                loyalty.get("mealPreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "seat_preference": validate_string(
                loyalty.get("seatPreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "home_airport": validate_string(
                loyalty.get("homeAirport", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "date_time_format_preference": validate_string(
                loyalty.get("dateTimeFormatPreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "cabin_preference": validate_string(
                loyalty.get("cabinPreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "fare_type_preference": validate_string(
                loyalty.get("fareTypePreference", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "expert_mode": parse_bool(loyalty.get("expertMode", False)),
            "privacy_indicator": validate_string(
                loyalty.get("privacyIndicator", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "car_preference_vendor": validate_string(
                loyalty.get("carPreferenceVendor", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "car_preference_type": validate_string(
                loyalty.get("carPreferenceType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "special_accommodation_1": validate_string(
                loyalty.get("specialAccommodation1", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "special_accommodation_2": validate_string(
                loyalty.get("specialAccommodation2", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "special_accommodation_3": validate_string(
                loyalty.get("specialAccommodation3", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "special_accommodation_4": validate_string(
                loyalty.get("specialAccommodation4", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "special_accommodation_5": validate_string(
                loyalty.get("specialAccommodation5", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "marketing_opt_ins": validate_string(
                loyalty.get("marketingOptIns", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "renew_date": validate_string(
                loyalty.get("renewDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "next_bill_amount": parse_number(loyalty.get("nextBillAmount", "")),
            "clear_enroll_date": validate_string(
                loyalty.get("clearEnrollDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "clear_renew_date": validate_string(
                loyalty.get("clearRenewDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "clear_tier_level": validate_string(
                loyalty.get("clearTierLevel", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "clear_next_bill_amount": parse_number(
                loyalty.get("clearNextBillAmount", "")
            ),
            "clear_is_active": parse_bool(loyalty.get("clearIsActive", False)),
            "clear_auto_renew": parse_bool(loyalty.get("clearAutoRenew", False)),
            "clear_has_biometrics": parse_bool(
                loyalty.get("clearHasBiometrics", False)
            ),
            "clear_has_partner_pricing": parse_bool(
                loyalty.get("clearHasPartnerPricing", False)
            ),
            "tsa_type": validate_string(
                loyalty.get("tsaType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tsa_seq_num": validate_string(
                loyalty.get("tsaSeqNum", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tsa_number": validate_string(
                loyalty.get("tsaNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(loyalty_rec, pax, cid)
        set_timestamp(loyalty_rec)
        add_extended_data(loyalty_rec, loyalty)
        loyalty_recs.append(loyalty_rec)


def add_ancillary_baggage(
    tx_id: str,
    rec: dict[str, Any],
    ancillary: dict[str, Any],
    ancillary_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    pax_index: float,
    cid: str,
) -> None:
    for bag in none_to_list(ancillary.get("ancillaryServiceBaggage", [])):
        ancillary_rec = {
            "tx_id": tx_id,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "ancillary_service",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_string(
                bag.get("id", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "ancillary_type": "baggage",
            "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "flight_number": validate_string(
                bag.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "departure_date": validate_string(
                bag.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "baggage_type": validate_string(
                bag.get("baggageType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "pax_index": parse_number(bag.get("paxIndex", 0)),
            "quantity": parse_number(bag.get("quantity", 0)),
            "weight": parse_number(bag.get("weight", "")),
            "dimentions_length": parse_number(
                bag.get("dimentions", {}).get("length", 0)
            ),
            "dimentions_width": parse_number(bag.get("dimentions", {}).get("width", 0)),
            "dimentions_height": parse_number(
                bag.get("dimentions", {}).get("height", 0)
            ),
            "priority_bag_drop": parse_bool(bag.get("priorityBagDrop", False)),
            "priority_bag_return": parse_bool(bag.get("priorityBagReturn", False)),
            "lot_bag_insurance": parse_bool(bag.get("lotBagInsurance", False)),
            "valuable_baggage_insurance": parse_bool(
                bag.get("valuableBaggageInsurance", False)
            ),
            "hands_free_baggage": parse_bool(bag.get("handsFreeBaggage", False)),
            "price": parse_number(bag.get("price", {}).get("total", 0)),
            "currency": validate_string(
                bag.get("price", {}).get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(ancillary_rec, pax, cid)
        set_timestamp(ancillary_rec)
        add_extended_data(ancillary_rec, bag)
        if pax.get("isBooker", False) or pax_index == ancillary_rec["pax_index"]:
            ancillary_recs.append(ancillary_rec)


def add_ancillary_seat(
    tx_id: str,
    rec: dict[str, Any],
    ancillary: dict[str, Any],
    ancillary_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    pax_index: float,
    cid: str,
) -> None:
    for seat in none_to_list(ancillary.get("ancillaryServiceSeat", [])):
        ancillary_rec = {
            "tx_id": tx_id,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "ancillary_service",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_id(seat.get("id", "")),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "ancillary_type": "seat",
            "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "flight_number": validate_string(
                seat.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "departure_date": validate_string(
                seat.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "pax_index": parse_number(seat.get("paxIndex", 0)),
            "seat_number": validate_string(
                seat.get("seatNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "seat_zone": validate_string(
                seat.get("seatZone", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "neighbor_free_seat": parse_bool(seat.get("neighborFreeSeat", False)),
            "upgrade_auction": parse_bool(seat.get("upgradeAuction", False)),
            "price": parse_number(seat.get("price", {}).get("total", 0)),
            "currency": validate_string(
                seat.get("price", {}).get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(ancillary_rec, pax, cid)
        set_timestamp(ancillary_rec)
        add_extended_data(ancillary_rec, seat)
        if pax.get("isBooker", False) or pax_index == ancillary_rec["pax_index"]:
            ancillary_recs.append(ancillary_rec)


def add_ancillary_flight_change(
    tx_id: str,
    rec: dict[str, Any],
    ancillary: dict[str, Any],
    ancillary_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    pax_index: float,
    cid: str,
) -> None:
    for change in none_to_list(ancillary.get("ancillaryServiceFlightChange", [])):
        ancillary_rec = {
            "tx_id": tx_id,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "ancillary_service",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_id(change.get("id", "")),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "ancillary_type": "change",
            "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "flight_number": validate_string(
                change.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "departure_date": validate_string(
                change.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "pax_index": parse_number(change.get("paxIndex", 0)),
            "change_type": validate_string(
                change.get("changeType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "price": parse_number(change.get("price", {}).get("total", 0)),
            "currency": validate_string(
                change.get("price", {}).get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(ancillary_rec, pax, cid)
        add_extended_data(ancillary_rec, change)
        set_timestamp(ancillary_rec)
        if pax.get("isBooker", False) or pax_index == ancillary_rec["pax_index"]:
            ancillary_recs.append(ancillary_rec)


def add_ancillary_other(
    tx_id: str,
    rec: dict[str, Any],
    ancillary: dict[str, Any],
    ancillary_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    cid: str,
) -> None:
    for other in none_to_list(ancillary.get("ancillaryServiceOther", [])):
        ancillary_rec = {
            "tx_id": tx_id,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "ancillary_service",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_id(other.get("id", "")),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "ancillary_type": "other",
            "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "flight_number": validate_string(
                other.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "departure_date": validate_string(
                other.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "other_ancilliary_type": validate_string(
                other.get("otherAncilliaryType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "price": parse_number(other.get("price", {}).get("total", 0)),
            "currency": validate_string(
                other.get("price", {}).get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(ancillary_rec, pax, cid)
        set_timestamp(ancillary_rec)
        add_extended_data(ancillary_rec, other)
        if pax.get("isBooker", False):
            ancillary_recs.append(ancillary_rec)


def add_ancillary_priority(
    tx_id: str,
    rec: dict[str, Any],
    ancillary: dict[str, Any],
    ancillary_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    cid: str,
) -> None:
    for priority in none_to_list(ancillary.get("ancillaryServicePriority", [])):
        ancillary_rec = {
            "tx_id": tx_id,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "ancillary_service",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "accp_object_id": validate_id(priority.get("id", "")),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "ancillary_type": "priority",
            "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "flight_number": validate_string(
                priority.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "departure_date": validate_string(
                priority.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "priority_service_type": validate_string(
                priority.get("priorityServiceType", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "lounge_access": parse_bool(priority.get("loungeAccess", "")),
            "price": parse_number(priority.get("price", {}).get("total", 0)),
            "currency": validate_string(
                priority.get("price", {}).get("currency", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        set_traveller_id(ancillary_rec, pax, cid)
        set_timestamp(ancillary_rec)
        add_extended_data(ancillary_rec, priority)
        if pax.get("isBooker", False):
            ancillary_recs.append(ancillary_rec)


def set_email_recs(
    email_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    cid: str,
    rec: dict[str, Any],
    tx_id: str,
) -> list[dict[str, Any]]:
    for email in pax.get("emails", []):
        historical_email = {
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "email_history",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "address": validate_email(email.get("address", "")),
            "type": validate_string(email.get("type", ""), INVALID_CHARS_SINGLE_LINE),
            "is_verified": parse_bool(email.get("isVerified", False)),
        }
        if (historical_email["address"]) in ["", None]:
            raise ValidationError("email address must be populated")
        historical_email["accp_object_id"] = (
            f"{historical_email['type']}{historical_email['address']}"
        )
        set_traveller_id(historical_email, pax, cid)
        set_timestamp(historical_email)
        add_extended_data(historical_email, email)
        email_recs.append(historical_email)
    return email_recs


def set_phone_recs(
    phone_recs: list[dict[str, Any]],
    pax: dict[str, Any],
    cid: str,
    rec: dict[str, Any],
    tx_id: str,
) -> list[dict[str, Any]]:
    for phone in pax.get("phones", []):
        historical_phone = {
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "phone_history",
            "last_updated": rec.get("lastUpdatedOn", ""),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "number": validate_phone(phone.get("number", "")),
            "country_code": validate_string(
                str(phone.get("countryCode", "")), INVALID_CHARS_SINGLE_LINE
            ),
            "type": validate_string(phone.get("type", ""), INVALID_CHARS_SINGLE_LINE),
            "is_verified": parse_bool(phone.get("isVerified", False)),
        }
        if (historical_phone["number"]) in ["", None]:
            raise ValidationError("phone number must be populated")
        historical_phone["accp_object_id"] = (
            f"{historical_phone['type']}{historical_phone['number']}"
        )
        set_traveller_id(historical_phone, pax, cid)
        set_timestamp(historical_phone)
        add_extended_data(historical_phone, phone)
        phone_recs.append(historical_phone)
    return phone_recs


def get_price_for_traveller(index: float, rec: dict[str, Any]) -> float | Literal[0]:
    for pi in none_to_list(rec.get("price", {}).get("priceItems", [])):
        if pi.get("passengerIndex", 0) == index:
            return parse_number(pi.get("price", {}).get("total", ""))
    return 0


def add_air_booking_record(
    pax_index: float,
    recs: list[dict[str, Any]],
    rec: dict[str, Any],
    pax: dict[str, Any],
    seg: dict[str, Any],
    cid: str,
    tx_id: str,
    alt_ids: list[dict[str, Any]],
) -> None:
    air_booking_rec = {
        "model_version": validate_string(
            rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "object_type": "air_booking",
        "last_updated": rec.get("lastUpdatedOn", ""),
        "last_updated_by": validate_string(
            rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "tx_id": tx_id,
        "accp_object_id": validate_id(build_segment_id(rec, seg)),
        "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
        "creation_channel_id": validate_string(
            rec.get("creationChannelId", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "last_update_channel_id": validate_string(
            rec.get("lastUpdateChannelId", ""), INVALID_CHARS_SINGLE_LINE
        ),
        # this field will be mapped to the profile conveying the "current booking" id
        "last_booking_id": validate_string(
            rec.get("id", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "segment_id": validate_string(
            build_segment_id(rec, seg), INVALID_CHARS_SINGLE_LINE
        ),
        "from": validate_string(seg.get("from", ""), INVALID_CHARS_SINGLE_LINE),
        "to": validate_string(seg.get("to", ""), INVALID_CHARS_SINGLE_LINE),
        "flight_number": validate_string(
            seg.get("flightNumber", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "departure_date": validate_string(
            seg.get("departureDate", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "departure_time": validate_string(
            seg.get("departureTime", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "arrival_date": validate_string(
            seg.get("arrivalDate", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "arrival_time": validate_string(
            seg.get("arrivalTime", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "channel": validate_string(
            rec.get("creationChannelId", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "status": validate_string(
            seg.get("status", rec.get("status", "")), INVALID_CHARS_SINGLE_LINE
        ),
        "honorific": validate_string(
            pax.get("honorific", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "first_name": validate_string(
            pax.get("firstName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "middle_name": validate_string(
            pax.get("middleName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "last_name": validate_string(
            pax.get("lastName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "gender": validate_string(pax.get("gender", ""), INVALID_CHARS_SINGLE_LINE),
        "pronoun": validate_string(pax.get("pronoun", ""), INVALID_CHARS_SINGLE_LINE),
        "date_of_birth": validate_string(
            pax.get("dateOfBirth", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "job_title": validate_string(
            pax.get("jobTitle", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "company": validate_string(
            pax.get("parentCompany", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "price": parse_number(rec.get("price", {}).get("grandTotal", "")),
        "traveller_price": get_price_for_traveller(pax_index, rec),
        "day_of_travel_email": validate_string(
            rec.get("dayOfTravelEmail", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "day_of_travel_phone": validate_string(
            rec.get("dayOfTravelPhone", ""), INVALID_CHARS_SINGLE_LINE
        ),
    }
    air_booking_rec["nationality_code"] = validate_string(
        pax.get("nationality", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    air_booking_rec["nationality_name"] = validate_string(
        pax.get("nationality", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    air_booking_rec["language_code"] = validate_string(
        pax.get("language", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    air_booking_rec["language_name"] = validate_string(
        pax.get("language", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    air_booking_rec["pss_id"] = validate_string(
        get_external_id(rec.get("externalIds", []), "pss"), INVALID_CHARS_SINGLE_LINE
    )
    air_booking_rec["gds_id"] = validate_string(
        get_external_id(rec.get("externalIds", []), "gds"), INVALID_CHARS_SINGLE_LINE
    )

    set_payment_info(air_booking_rec, rec.get("paymentInformation", {}))
    set_billing_address(air_booking_rec, rec.get("paymentInformation", {}))

    # Set primary option for phone/email/address
    set_primary_email(air_booking_rec, pax.get("emails", []))
    set_primary_phone(air_booking_rec, pax.get("phones", []))
    set_traveller_id(air_booking_rec, pax, cid)
    set_timestamp(air_booking_rec)
    set_primary_address(air_booking_rec, none_to_list(pax.get("addresses", [])))
    # must be set after to set the traveller ID
    if pax.get("isBooker", False):
        air_booking_rec["booker_id"] = air_booking_rec[FIELD_NAME_TRAVELLER_ID]

    recs.append(air_booking_rec)
    add_alternate_profile_ids(alt_ids, pax, cid, tx_id)
    add_extended_data(air_booking_rec, pax)


def validate_id(id_value: Any) -> Any:
    return validate_string(id_value, INVALID_CHARS_SINGLE_LINE)


def build_segment_id(rec: dict[str, Any], seg: dict[str, Any]) -> Any:
    return (
        validate_string(rec["id"], INVALID_CHARS_SINGLE_LINE)
        + "-"
        + validate_string(seg.get("from", ""), INVALID_CHARS_SINGLE_LINE)
        + "-"
        + validate_string(seg.get("to", ""), INVALID_CHARS_SINGLE_LINE)
    )
