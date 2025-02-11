# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import traceback
import uuid
from typing import Any, TypedDict

from tah_lib.common import (
    AIR_LOYALTY_RECORDS_KEY,
    AIR_LOYALTY_TX_RECORDS_KEY,
    AIR_PROFILE_RECORDS_KEY,
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    COMMON_EMAIL_RECORDS_KEY,
    COMMON_PHONE_RECORDS_KEY,
    INVALID_CHARS_SINGLE_LINE,
    ValidationError,
    add_alternate_profile_ids,
    add_extended_data,
    deduplicate_list,
    log_error_to_sqs,
    parse_int,
    parse_number,
    set_primary_address,
    set_primary_email,
    set_primary_phone,
    set_timestamp,
    set_traveller_id,
    validate_email,
    validate_phone,
    validate_string,
)

BIZ_OBJECT_TYPE = "pax_profile"


class ObjectRecord(TypedDict):
    air_profile_recs: list[dict[str, Any]]
    common_email_recs: list[dict[str, Any]]
    common_phone_recs: list[dict[str, Any]]
    air_loyalty_recs: list[dict[str, Any]]
    air_loyalty_tx_recs: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    try:
        profile_recs = []
        email_recs = []
        phone_recs = []
        loyalty_recs = []
        loyalty_tx_recs = []
        alternate_profile_id_recs: list[dict[str, Any]] = []
        cid = str(uuid.uuid4())
        # Passenger Profile
        profile_rec = {
            "object_type": "pax_profile",
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "id": validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE),
            "last_updated": validate_string(
                rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "created_on": validate_string(
                rec.get("createdOn", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated_by": validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
            "created_by": validate_string(
                rec.get("createdBy", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "honorific": validate_string(
                rec.get("honorific", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "first_name": validate_string(
                rec.get("firstName", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "middle_name": validate_string(
                rec.get("middleName", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_name": validate_string(
                rec.get("lastName", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "gender": validate_string(rec.get("gender", ""), INVALID_CHARS_SINGLE_LINE),
            "pronoun": validate_string(
                rec.get("pronoun", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "date_of_birth": validate_string(
                rec.get("dateOfBirth", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "job_title": validate_string(
                rec.get("jobTitle", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "company": validate_string(
                rec.get("parentCompany", ""), INVALID_CHARS_SINGLE_LINE
            ),
        }
        profile_rec["nationality_code"] = validate_string(
            rec.get("nationality", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
        )
        profile_rec["nationality_name"] = validate_string(
            rec.get("nationality", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
        )
        profile_rec["language_code"] = validate_string(
            rec.get("language", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
        )
        profile_rec["language_name"] = validate_string(
            rec.get("language", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
        )

        # Set primary option for phone/email/address
        set_primary_email(profile_rec, rec.get("emails", []))
        set_primary_phone(profile_rec, rec.get("phones", []))
        set_traveller_id(profile_rec, rec, cid)
        profile_rec["accp_object_id"] = profile_rec["traveller_id"]
        set_timestamp(profile_rec)
        set_primary_address(profile_rec, rec.get("addresses", []))
        add_extended_data(profile_rec, rec)

        # Email Addresses
        for email in rec.get("emails", []):
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
                "type": validate_string(
                    email.get("type", ""), INVALID_CHARS_SINGLE_LINE
                ),
            }
            if (historical_email["address"]) in ["", None]:
                raise ValidationError("email address must be populated")
            historical_email["accp_object_id"] = (
                f"{historical_email['type']}{historical_email['address']}"
            )
            set_traveller_id(historical_email, rec, cid)
            set_timestamp(historical_email)
            add_extended_data(historical_email, email)
            email_recs.append(historical_email)

        # Phone Numbers
        for phone in rec.get("phones", []):
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
                "type": validate_string(
                    phone.get("type", ""), INVALID_CHARS_SINGLE_LINE
                ),
            }
            if (historical_phone["number"]) in ["", None]:
                raise ValidationError("phone number must be populated")
            historical_phone["accp_object_id"] = (
                f"{historical_phone['type']}{historical_phone['number']}"
            )
            set_traveller_id(historical_phone, rec, cid)
            set_timestamp(historical_phone)
            add_extended_data(historical_phone, phone)
            phone_recs.append(historical_phone)

        # Loyalty Programs
        for loyalty in rec.get("loyaltyPrograms", []):
            loyalty_rec = {
                "model_version": validate_string(
                    rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
                ),
                "object_type": "air_loyalty",
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
                "miles": parse_number(loyalty.get("miles", "")),
                "miles_to_next_level": parse_number(
                    loyalty.get("milesToNextLevel", "")
                ),
                "level": validate_string(
                    loyalty.get("level", ""), INVALID_CHARS_SINGLE_LINE
                ),
                "joined": validate_string(
                    loyalty.get("joined", ""), INVALID_CHARS_SINGLE_LINE
                ),
            }
            set_traveller_id(loyalty_rec, rec, cid)
            set_timestamp(loyalty_rec)
            add_extended_data(loyalty_rec, loyalty)
            loyalty_recs.append(loyalty_rec)

            for tx in loyalty.get("transactions", []):
                loyalty_tx_rec = {
                    "object_type": "loyalty_transaction",
                    "model_version": validate_string(
                        rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "last_updated": validate_string(
                        tx.get("timeStamp", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "accp_object_id": validate_string(
                        tx.get("id", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "last_updated_by": validate_string(
                        tx.get("agent", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "tx_id": tx_id,
                    "category": validate_string(
                        tx.get("category", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "points_offset": parse_number(tx.get("pointsOffset", "")),
                    "points_unit": validate_string(
                        tx.get("pointUnit", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "origin_points_offset": parse_number(
                        tx.get("originPointsOffset", "")
                    ),
                    "source": validate_string(
                        tx.get("source", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "booking_date": validate_string(
                        tx.get("bookingDate", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "order_number": validate_string(
                        tx.get("orderNumber", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "product_id": validate_string(
                        tx.get("productId", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "expire_in_days": parse_int(tx.get("expireInDays")),
                    "amount": parse_number(tx.get("amount")),
                    "amount_type": validate_string(
                        tx.get("amountType", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "voucher_quantity": parse_number(tx.get("voucherQuantity")),
                    "corporate_reference_number": validate_string(
                        tx.get("corporateReferenceNumber", ""),
                        INVALID_CHARS_SINGLE_LINE,
                    ),
                    "promotions": validate_string(
                        tx.get("promotions", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "location": validate_string(
                        tx.get("location", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "activity_day": validate_string(
                        tx.get("activityDay", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "to_loyalty_id": validate_string(
                        tx.get("toLoyaltyId", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "from_loyalty_id": validate_string(
                        tx.get("fromLoyaltyId", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "organization_code": validate_string(
                        tx.get("organizationCode", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "event_name": validate_string(
                        tx.get("eventName", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "document_number": validate_string(
                        tx.get("documentNumber", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "corporate_id": validate_string(
                        tx.get("corporateId", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                    "program_name": validate_string(
                        tx.get("programName", ""), INVALID_CHARS_SINGLE_LINE
                    ),
                }

                set_traveller_id(loyalty_tx_rec, rec, cid)
                set_timestamp(loyalty_tx_rec)
                add_extended_data(loyalty_tx_rec, tx)
                loyalty_tx_recs.append(loyalty_tx_rec)

        # add additional profile ids
        add_alternate_profile_ids(alternate_profile_id_recs, rec, cid, tx_id)

        profile_recs.append(profile_rec)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        AIR_PROFILE_RECORDS_KEY: profile_recs,
        COMMON_EMAIL_RECORDS_KEY: email_recs,
        COMMON_PHONE_RECORDS_KEY: phone_recs,
        AIR_LOYALTY_RECORDS_KEY: loyalty_recs,
        AIR_LOYALTY_TX_RECORDS_KEY: loyalty_tx_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }
