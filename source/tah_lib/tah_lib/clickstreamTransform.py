# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import traceback
import uuid
from collections.abc import Iterable
from typing import Any, TypedDict

from tah_lib.common import (
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    CLICKSTREAM_RECORDS_KEY,
    FIELD_NAME_LAST_UPDATED,
    FIELD_NAME_TRAVELLER_ID,
    INVALID_CHARS_SINGLE_LINE,
    INVALID_CHARS_URL,
    add_alternate_profile_ids,
    add_extended_data,
    deduplicate_list,
    log_error_to_sqs,
    parse_bool,
    parse_number,
    replace_and_join,
    set_timestamp,
    validate_email,
    validate_phone,
    validate_string,
)

BIZ_OBJECT_TYPE = "clickstream"


class ObjectRecord(TypedDict):
    clickstream_recs: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    alternate_profile_id_recs: list[dict[str, Any]] = []
    new_recs = []
    try:

        new_rec = {}
        # accp Object fields
        new_rec["model_version"] = validate_string(
            rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["object_type"] = BIZ_OBJECT_TYPE
        new_rec["tx_id"] = tx_id
        new_rec[FIELD_NAME_TRAVELLER_ID] = validate_string(
            rec.get("userId", ""), INVALID_CHARS_SINGLE_LINE
        )

        # events fields
        new_rec["event_type"] = validate_string(
            rec.get("event_type", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["event_timestamp"] = validate_string(
            rec.get("event_timestamp", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["arrival_timestamp"] = validate_string(
            str(rec.get("arrival_timestamp", "")), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["event_version"] = validate_string(
            rec.get("event_version", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["user_agent"] = validate_string(
            rec.get("device", {}).get("useragent", ""), []
        )
        new_rec["ip_address"] = validate_string(
            rec.get("device", {}).get("ipAddress", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["session_id"] = validate_string(
            rec.get("session", {}).get("id", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["custom_event_name"] = validate_string(
            get_attr(rec.get("attributes", []), "custom_event_name"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["error"] = validate_string(
            get_attr(rec.get("attributes", []), "error"), INVALID_CHARS_SINGLE_LINE
        )

        # hotel Fields
        new_rec["destination"] = validate_string(
            get_attr(rec.get("attributes", []), "destination"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["checkin_date"] = validate_string(
            get_attr(rec.get("attributes", []), "checkin_date"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["checkout_date"] = validate_string(
            get_attr(rec.get("attributes", []), "checkout_date"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["room_type"] = validate_string(
            get_attr(rec.get("attributes", []), "room_type"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["rate_plan"] = validate_string(
            get_attr(rec.get("attributes", []), "rate_plan"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["hotel_code"] = validate_string(
            get_attr(rec.get("attributes", []), "hotel_code"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["hotel_name"] = validate_string(
            get_attr(rec.get("attributes", []), "hotel_name"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["hotel_code_list"] = validate_string(
            get_attr(rec.get("attributes", []), "hotel_code_list"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["num_nights"] = parse_number(
            get_attr(rec.get("attributes", []), "num_nights")
        )
        new_rec["num_guests"] = parse_number(
            get_attr(rec.get("attributes", []), "num_guests")
        )
        new_rec["num_guest_adults"] = parse_number(
            get_attr(rec.get("attributes", []), "num_guest_adults")
        )
        new_rec["num_guest_children"] = parse_number(
            get_attr(rec.get("attributes", []), "num_guest_children")
        )

        # Cross industry
        new_rec["quantities"] = validate_string(
            get_attr(rec.get("attributes", []), "quantities"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["products"] = validate_string(
            get_attr(rec.get("attributes", []), "products"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["products_prices"] = validate_string(
            get_attr(rec.get("attributes", []), "products_prices"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["currency"] = validate_string(
            get_attr(rec.get("attributes", []), "currency"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["ecommerce_action"] = validate_string(
            get_attr(rec.get("attributes", []), "ecommerce_action"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["order_payment_type"] = validate_string(
            get_attr(rec.get("attributes", []), "order_payment_type"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["order_promo_code"] = validate_string(
            get_attr(rec.get("attributes", []), "order_promo_code"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["page_name"] = validate_string(
            get_attr(rec.get("attributes", []), "page_name"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["page_type_environment"] = validate_string(
            get_attr(rec.get("attributes", []), "page_type_environment"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["transaction_id"] = validate_string(
            get_attr(rec.get("attributes", []), "transaction_id"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["url"] = validate_string(
            get_attr(rec.get("attributes", []), "url"), INVALID_CHARS_URL
        )
        new_rec["booking_id"] = validate_string(
            get_attr(rec.get("attributes", []), "booking_id"), INVALID_CHARS_SINGLE_LINE
        )

        # Geofence
        new_rec["geofence_latitude"] = validate_string(
            get_attr(rec.get("attributes", []), "geofence_latitude"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["geofence_longitude"] = validate_string(
            get_attr(rec.get("attributes", []), "geofence_longitude"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["geofence_id"] = validate_string(
            get_attr(rec.get("attributes", []), "geofence_id"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["geofence_name"] = validate_string(
            get_attr(rec.get("attributes", []), "geofence_name"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["poi_id"] = validate_string(
            get_attr(rec.get("attributes", []), "poi_id"), INVALID_CHARS_SINGLE_LINE
        )

        # Customer fields
        new_rec["customer_birthdate"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_birthdate"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_country"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_country"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_email"] = validate_email(
            get_attr(rec.get("attributes", []), "customer_email")
        )
        new_rec["customer_first_name"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_first_name"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_gender"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_gender"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_id"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_id"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_last_name"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_last_name"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_nationality"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_nationality"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_phone"] = validate_phone(
            get_attr(rec.get("attributes", []), "customer_phone")
        )
        new_rec["customer_loyalty_id"] = validate_string(
            get_attr(rec.get("attributes", []), "loyalty_id"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["language_code"] = validate_string(
            get_attr(rec.get("attributes", []), "language_code"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["customer_type"] = validate_string(
            get_attr(rec.get("attributes", []), "customer_type"),
            INVALID_CHARS_SINGLE_LINE,
        )

        # Air attributes
        new_rec["fare_class"] = validate_string(
            get_attr(rec.get("attributes", []), "fare_class"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["fare_type"] = validate_string(
            get_attr(rec.get("attributes", []), "fare_type"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["flight_segments_departure_date_time"] = validate_string(
            get_attr(rec.get("attributes", []), "flight_segments_departure_date_time"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["flight_numbers"] = validate_string(
            get_attr(rec.get("attributes", []), "flight_segment_sku"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["flight_market"] = validate_string(
            get_attr(rec.get("attributes", []), "flight_market"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["flight_type"] = validate_string(
            get_attr(rec.get("attributes", []), "flight_type"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["origin_date"] = validate_string(
            get_attr(rec.get("attributes", []), "origin_date"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["origin_date_time"] = validate_string(
            get_attr(rec.get("attributes", []), "origin_date_time"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["return_date_time"] = validate_string(
            get_attr(rec.get("attributes", []), "returning_date_time"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["return_date"] = validate_string(
            get_attr(rec.get("attributes", []), "returning_date"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["origin_flight_route"] = validate_string(
            get_attr(rec.get("attributes", []), "origin_flight_route"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["return_flight_route"] = validate_string(
            get_attr(rec.get("attributes", []), "returning_flight_route"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["num_pax_adults"] = parse_number(
            get_attr(rec.get("attributes", []), "num_Pax_Adults")
        )
        new_rec["num_pax_inf"] = parse_number(
            get_attr(rec.get("attributes", []), "num_pax_inf")
        )
        new_rec["num_pax_children"] = parse_number(
            get_attr(rec.get("attributes", []), "num_Pax_Children")
        )
        new_rec["pax_type"] = validate_string(
            get_attr(rec.get("attributes", []), "pax_type"), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["total_passengers"] = parse_number(
            get_attr(rec.get("attributes", []), "total_passengers")
        )
        new_rec["selected_seats"] = validate_string(
            get_attr(rec.get("attributes", []), "selected_seats"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["length_of_stay"] = validate_string(
            get_attr(rec.get("attributes", []), "length_of_stay"),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["origin"] = validate_string(
            get_attr(rec.get("attributes", []), "origin"), INVALID_CHARS_SINGLE_LINE
        )

        # Custom Events fields
        new_rec["custom"] = validate_string(
            get_attr(rec.get("attributes", []), "custom"), INVALID_CHARS_SINGLE_LINE
        )

        new_rec["aws_account_id"] = validate_string(
            rec.get("awsAccountId", ""), INVALID_CHARS_SINGLE_LINE
        )

        # session ID is a mandatory field
        if new_rec["session_id"] == "":
            new_rec["session_id"] = str(uuid.uuid4())

        new_rec[FIELD_NAME_TRAVELLER_ID] = new_rec["session_id"]
        new_rec[FIELD_NAME_LAST_UPDATED] = new_rec["event_timestamp"]

        # We use a combination of session_id and event_timestamp to create a unique object ID for the record
        new_rec["accp_object_id"] = (
            new_rec["session_id"] + "_" + new_rec["event_timestamp"]
        )
        # If no timestamp is provided we set a timestamp for the object
        set_timestamp(new_rec)
        # cleaning up None data
        to_delete = []
        for key in new_rec:
            if new_rec[key] is None:
                to_delete.append(key)
        for key in to_delete:
            del new_rec[key]

        new_recs.append(new_rec)
        add_alternate_profile_ids(
            alternate_profile_id_recs, rec, new_rec["session_id"], tx_id
        )
        add_extended_data(new_rec, rec)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        CLICKSTREAM_RECORDS_KEY: new_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }


def get_attr(
    attributes: Iterable[dict[str, Any]], name: str
) -> str | float | bool | None:
    selected_att = {}
    for att in attributes:
        if att["name"] == name:
            selected_att = att
            break
    if "type" in selected_att:
        if selected_att["type"] == "string":
            return str(selected_att["stringValue"])
        elif selected_att["type"] == "strings":
            arr = []
            for val in selected_att["stringValues"]:
                arr.append(str(val))
            return replace_and_join(arr, "|")
        elif selected_att["type"] == "number":
            return parse_number(selected_att.get("numValue", 0))
        elif selected_att["type"] == "numbers":
            arr = []
            for val in selected_att["numValues"]:
                arr.append(str(val))
            return replace_and_join(arr, "|")
        elif selected_att["type"] == "boolean":
            return parse_bool(selected_att["boolValue"])
    return None
