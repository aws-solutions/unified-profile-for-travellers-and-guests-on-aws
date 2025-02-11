# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import traceback
import uuid
from typing import Any, TypedDict

from tah_lib.common import (
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    HOTEL_STAY_REVENUE_ITEMS_RECORDS_KEY,
    INVALID_CHARS_SINGLE_LINE,
    add_alternate_profile_ids,
    add_extended_data,
    deduplicate_list,
    log_error_to_sqs,
    parse_number,
    set_timestamp,
    validate_email,
    validate_phone,
    validate_string,
)

BIZ_OBJECT_TYPE = "hotel_stay"


def build_segment_id(rec: dict[str, Any], item: dict[str, Any]) -> str:
    stay_id = validate_string(rec.get("id", ""), INVALID_CHARS_SINGLE_LINE)
    purchase_type = validate_string(item.get("type", ""), INVALID_CHARS_SINGLE_LINE)
    date = validate_string(item.get("date", ""), INVALID_CHARS_SINGLE_LINE)
    return "|".join([stay_id, purchase_type, date])


class ObjectRecord(TypedDict):
    hotel_stay_revenue_items: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    new_recs = []
    alternate_profile_id_recs: list[dict[str, Any]] = []
    try:
        if len(rec.get("revenue", [])) == 0:
            print("Invalid format Exception")
            raise ValueError("Hotel stay must have at least 1 revenue item ")

        for item in rec.get("revenue", []):

            new_rec = {}
            new_rec["model_version"] = validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["object_type"] = "hotel_stay_revenue_items"
            new_rec["last_updated"] = validate_string(
                rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["accp_object_id"] = build_segment_id(rec, item)
            new_rec["created_on"] = validate_string(
                rec.get("createdOn", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["last_updated_by"] = validate_string(
                rec.get("lastUpdatedBy", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["tx_id"] = tx_id
            new_rec["created_by"] = validate_string(
                rec.get("createdBy", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["id"] = validate_string(
                rec.get("id", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["booking_id"] = validate_string(
                rec.get("bookingId", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["traveller_id"] = validate_string(
                rec.get("guestId", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["currency_code"] = validate_string(
                rec.get("currency", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["currency_name"] = validate_string(
                rec.get("currency", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["currency_symbol"] = validate_string(
                rec.get("currency", {}).get("symbol", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["first_name"] = validate_string(
                rec.get("firstName", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["last_name"] = validate_string(
                rec.get("lastName", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["email"] = validate_email(rec.get("email", ""))
            new_rec["phone"] = validate_phone(str(rec.get("phone", "")))
            new_rec["start_date"] = validate_string(
                rec.get("startDate", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["hotel_code"] = validate_string(
                rec.get("hotelCode", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["type"] = validate_string(
                item.get("type", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["description"] = validate_string(
                item.get("description", ""), INVALID_CHARS_SINGLE_LINE
            )
            new_rec["amount"] = parse_number(item.get("amount", ""))
            new_rec["date"] = validate_string(
                item.get("date", ""), INVALID_CHARS_SINGLE_LINE
            )
            if new_rec["traveller_id"] == "":
                new_rec["traveller_id"] = str(uuid.uuid4())
            set_timestamp(new_rec)
            add_extended_data(new_rec, item)
            new_recs.append(new_rec)

            #   Add Alternate Profile IDs using modified "rec":
            #       Original rec["id"] is not used as the traveller_id; rec["guestId"] is
            #       Send a dictionary containing the original keys except for "id"
            #       The `set_traveller_id` invocation will use `new_rec["traveller_id"]` as the fallback and correctly link the record to the profile
            exclude_keys = {"id"}
            abridged_dict = {
                key: rec[key] for key in set(list(rec.keys())) - exclude_keys
            }
            add_alternate_profile_ids(
                alternate_profile_id_recs, abridged_dict, new_rec["traveller_id"], tx_id
            )
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        HOTEL_STAY_REVENUE_ITEMS_RECORDS_KEY: new_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }
