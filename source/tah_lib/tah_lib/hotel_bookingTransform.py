# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import traceback
import uuid
from typing import Any, TypedDict

from tah_lib.common import (
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    COMMON_EMAIL_RECORDS_KEY,
    COMMON_PHONE_RECORDS_KEY,
    FIELD_NAME_TRAVELLER_ID,
    GUEST_LOYALTY_RECORDS_KEY,
    HOTEL_BOOKING_RECORDS_KEY,
    INVALID_CHARS_SINGLE_LINE,
    ValidationError,
    add_alternate_profile_ids,
    add_extended_data,
    build_serialized_lists,
    deduplicate_list,
    get_external_id,
    log_error_to_sqs,
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

BIZ_OBJECT_TYPE = "hotel_booking"


class ObjectRecord(TypedDict):
    hotel_booking_recs: list[dict[str, Any]]
    common_email_recs: list[dict[str, Any]]
    common_phone_recs: list[dict[str, Any]]
    guest_loyalty_recs: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    try:
        booking_recs: list[dict[str, Any]] = []
        email_recs: list[dict[str, Any]] = []
        phone_recs: list[dict[str, Any]] = []
        loyalty_recs: list[dict[str, Any]] = []
        alternate_profile_id_recs: list[dict[str, Any]] = []

        if len(rec.get("segments", [])) == 0:
            print("Invalid format Exception")
            raise ValueError("Hotel booking must have at least 1 segment ")

        for seg in rec.get("segments", []):
            cid = str(uuid.uuid4())
            sid = str(uuid.uuid4())
            # we want to uniquely identify segment to be able to calculate prices
            if seg.get("id", "") == "":
                seg["id"] = sid
            add_email_recs(email_recs, rec, seg.get("holder", {}), cid, tx_id)
            add_phone_recs(phone_recs, rec, seg.get("holder", {}), cid, tx_id)
            add_loyalty_recs(loyalty_recs, rec, seg.get("holder", {}), cid, tx_id)
            add_alternate_profile_ids(
                alternate_profile_id_recs, seg.get("holder", {}), cid, tx_id
            )
            for product in seg.get("products", []):
                add_booking_rec(
                    booking_recs,
                    rec,
                    seg,
                    product,
                    seg.get("holder", {}),
                    cid,
                    True,
                    tx_id,
                )
            for add_guest in seg.get("additionalGuests", []):
                cid = str(uuid.uuid4())
                add_email_recs(email_recs, rec, add_guest, cid, tx_id)
                add_phone_recs(phone_recs, rec, add_guest, cid, tx_id)
                add_loyalty_recs(loyalty_recs, rec, add_guest, cid, tx_id)
                add_alternate_profile_ids(
                    alternate_profile_id_recs, add_guest, cid, tx_id
                )
                for product in seg.get("products", []):
                    add_booking_rec(
                        booking_recs, rec, seg, product, add_guest, cid, False, tx_id
                    )
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        HOTEL_BOOKING_RECORDS_KEY: booking_recs,
        COMMON_EMAIL_RECORDS_KEY: email_recs,
        COMMON_PHONE_RECORDS_KEY: phone_recs,
        GUEST_LOYALTY_RECORDS_KEY: loyalty_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }


def build_segment_product_id(
    rec: dict[str, Any], seg: dict[str, Any], product: dict[str, Any]
) -> str:
    hotel = validate_string(
        seg.get("hotelCode", rec.get("hotelCode", "")), INVALID_CHARS_SINGLE_LINE
    )
    conf_number = validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE)
    seg_id = validate_string(seg.get("id", ""), INVALID_CHARS_SINGLE_LINE)
    start_date = validate_string(seg.get("startDate", ""), INVALID_CHARS_SINGLE_LINE)
    product_str = validate_string(product.get("id", ""), INVALID_CHARS_SINGLE_LINE)
    return "|".join([hotel, conf_number, seg_id, start_date, product_str])


def add_booking_rec(
    booking_recs: list[dict[str, Any]],
    rec: dict[str, Any],
    seg: dict[str, Any],
    product: dict[str, Any],
    guest: dict[str, Any],
    cid: str,
    is_booker: bool,
    tx_id: str,
) -> None:
    hotel_booking_rec = {
        "model_version": validate_string(
            rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "object_type": "hotel_booking",
        "accp_object_id": build_segment_product_id(rec, seg, product),
        # segment ID is generated upstream if not provided
        "segment_id": validate_string(seg.get("id", ""), INVALID_CHARS_SINGLE_LINE),
        "last_updated": validate_string(
            rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "last_updated_by": validate_string(
            rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "tx_id": tx_id,
        "booking_id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
        # this field will be mapped to the profile conveying the "current booking" id
        "last_booking_id": validate_string(
            rec.get("id", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "creation_channel_id": validate_string(
            rec.get("creationChannelId", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "last_update_channel_id": validate_string(
            rec.get("lastUpdateChannelId", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "hotel_code": validate_string(
            seg.get("hotelCode", rec.get("hotelCode", "")), INVALID_CHARS_SINGLE_LINE
        ),
        "n_nights": parse_number(seg.get("nNights", rec.get("nNights", ""))),
        "n_guests": parse_number(seg.get("nGuests", rec.get("nGuests", ""))),
        "product_id": validate_string(product.get("id", ""), INVALID_CHARS_SINGLE_LINE),
        "check_in_date": validate_string(
            seg.get("startDate", rec.get("startDate", "")), INVALID_CHARS_SINGLE_LINE
        ),
        "status": validate_string(
            seg.get("status", rec.get("status", "")), INVALID_CHARS_SINGLE_LINE
        ),
        "honorific": validate_string(
            guest.get("honorific", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "first_name": validate_string(
            guest.get("firstName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "middle_name": validate_string(
            guest.get("middleName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "last_name": validate_string(
            guest.get("lastName", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "gender": validate_string(guest.get("gender", ""), INVALID_CHARS_SINGLE_LINE),
        "pronoun": validate_string(guest.get("pronoun", ""), INVALID_CHARS_SINGLE_LINE),
        "date_of_birth": validate_string(
            guest.get("dateOfBirth", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "job_title": validate_string(
            guest.get("jobTitle", ""), INVALID_CHARS_SINGLE_LINE
        ),
        "company": validate_string(
            guest.get("parentCompany", ""), INVALID_CHARS_SINGLE_LINE
        ),
        # these prices relate to the segement not the whole booking and are not traveller specific
        "total_segment_after_tax": parse_number(
            seg.get("price", {}).get("totalAfterTax", "")
        ),
        "total_segment_before_tax": parse_number(
            seg.get("price", {}).get("totalBeforeTax", "")
        ),
    }
    hotel_booking_rec["nationality_code"] = validate_string(
        rec.get("nationality", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["nationality_name"] = validate_string(
        rec.get("nationality", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["language_code"] = validate_string(
        rec.get("language", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["language_name"] = validate_string(
        rec.get("language", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["pms_id"] = validate_string(
        get_external_id(rec.get("externalIds", []), "pms"), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["crs_id"] = validate_string(
        get_external_id(rec.get("externalIds", []), "crs"), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["gds_id"] = validate_string(
        get_external_id(rec.get("externalIds", []), "gds"), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["room_type_code"] = validate_string(
        product.get("roomType", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["room_type_name"] = validate_string(
        product.get("roomType", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["room_type_description"] = validate_string(
        product.get("roomType", {}).get("description", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["rate_plan_code"] = validate_string(
        product.get("ratePlan", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["rate_plan_name"] = validate_string(
        product.get("ratePlan", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["rate_plan_description"] = validate_string(
        product.get("ratePlan", {}).get("description", ""), INVALID_CHARS_SINGLE_LINE
    )
    hotel_booking_rec["attribute_codes"] = build_serialized_lists(
        product.get("attributes", []), "code", "|"
    )
    hotel_booking_rec["attribute_names"] = build_serialized_lists(
        product.get("attributes", []), "name", "|"
    )
    hotel_booking_rec["attribute_descriptions"] = build_serialized_lists(
        product.get("attributes", []), "description", "|"
    )
    hotel_booking_rec["add_on_codes"] = build_serialized_lists(
        product.get("addOns", []), "code", "|"
    )
    hotel_booking_rec["add_on_names"] = build_serialized_lists(
        product.get("addOns", []), "name", "|"
    )
    hotel_booking_rec["add_on_descriptions"] = build_serialized_lists(
        product.get("addOns", []), "description", "|"
    )
    set_payment_info(hotel_booking_rec, rec.get("paymentInformation", {}))
    set_billing_address(hotel_booking_rec, rec.get("paymentInformation", {}))
    # Set primary option for phone/email/address
    set_primary_email(hotel_booking_rec, guest.get("emails", []))
    set_primary_phone(hotel_booking_rec, guest.get("phones", []))
    set_primary_address(hotel_booking_rec, guest.get("addresses", []))
    # set traveller ID
    set_traveller_id(hotel_booking_rec, guest, cid)
    set_timestamp(hotel_booking_rec)
    add_extended_data(hotel_booking_rec, rec)
    if is_booker:
        hotel_booking_rec["booker_id"] = hotel_booking_rec[FIELD_NAME_TRAVELLER_ID]
    booking_recs.append(hotel_booking_rec)


def add_email_recs(
    email_recs: list[dict[str, Any]],
    rec: dict[str, Any],
    guest: dict[str, Any],
    cid: str,
    tx_id: str,
) -> None:
    for email in guest.get("emails", []):
        historical_email = {
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "email_history",
            "last_updated": validate_string(
                rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "address": validate_email(email.get("address", "")),
            "type": validate_string(email.get("type", ""), INVALID_CHARS_SINGLE_LINE),
        }
        if (historical_email["address"]) in ["", None]:
            raise ValidationError("email address must be populated")
        historical_email["accp_object_id"] = (
            f"{historical_email['type']}{historical_email['address']}"
        )
        set_traveller_id(historical_email, guest, cid)
        set_timestamp(historical_email)
        add_extended_data(historical_email, email)
        email_recs.append(historical_email)


def add_phone_recs(
    phone_recs: list[dict[str, Any]],
    rec: dict[str, Any],
    guest: dict[str, Any],
    cid: str,
    tx_id: str,
) -> None:
    for phone in guest.get("phones", []):
        historical_phone = {
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "object_type": "phone_history",
            "last_updated": validate_string(
                rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "number": validate_phone(phone.get("number", "")),
            "country_code": validate_string(
                str(phone.get("countryCode", "")), INVALID_CHARS_SINGLE_LINE
            ),
            "type": validate_string(phone.get("type", ""), INVALID_CHARS_SINGLE_LINE),
        }
        if (historical_phone["number"]) in ["", None]:
            raise ValidationError("phone number must be populated")
        historical_phone["accp_object_id"] = (
            f"{historical_phone['type']}{historical_phone['number']}"
        )
        set_traveller_id(historical_phone, guest, cid)
        set_timestamp(historical_phone)
        add_extended_data(historical_phone, phone)
        phone_recs.append(historical_phone)


def add_loyalty_recs(
    loyalty_recs: list[dict[str, Any]],
    rec: dict[str, Any],
    guest: dict[str, Any],
    cid: str,
    tx_id: str,
) -> None:
    if "loyaltyPrograms" in guest:
        for loyalty in guest.get("loyaltyPrograms", []):
            loyalty_rec = {
                "object_type": "hotel_loyalty",
                "model_version": validate_string(
                    rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
                ),
                "last_updated": validate_string(
                    rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
                ),
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
                "points": parse_number(loyalty.get("points", "")),
                "units": validate_string(
                    loyalty.get("pointUnit", ""), INVALID_CHARS_SINGLE_LINE
                ),
                "points_to_next_level": parse_number(
                    loyalty.get("pointsToNextLevel", "")
                ),
                "level": validate_string(
                    loyalty.get("level", ""), INVALID_CHARS_SINGLE_LINE
                ),
                "joined": validate_string(
                    loyalty.get("joined", ""), INVALID_CHARS_SINGLE_LINE
                ),
            }
            set_traveller_id(loyalty_rec, guest, cid)
            set_timestamp(loyalty_rec)
            add_extended_data(loyalty_rec, loyalty)
            loyalty_recs.append(loyalty_rec)
