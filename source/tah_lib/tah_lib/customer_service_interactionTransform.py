# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import base64
import json
import traceback
import uuid
from collections.abc import Sequence
from typing import Any, TypedDict

from tah_lib.common import (
    ALTERNATE_PROFILE_ID_RECORDS_KEY,
    CONVERSATION_RECORDS_KEY,
    INVALID_CHARS_MULTI_LINE,
    INVALID_CHARS_SINGLE_LINE,
    ValidationError,
    add_alternate_profile_ids,
    add_extended_data,
    deduplicate_list,
    log_error_to_sqs,
    parse_bool,
    parse_number,
    safe_list_get,
    set_timestamp,
    validate_email,
    validate_phone,
    validate_string,
    validate_text,
)

BIZ_OBJECT_TYPE = "customer_service_interaction"


class ObjectRecord(TypedDict):
    conversation_recs: list[dict[str, Any]]
    alternate_profile_id_recs: list[dict[str, Any]]


def build_object_record(
    rec: dict[str, Any], err_queue_url: str, tx_id: str
) -> ObjectRecord:
    new_recs = []
    alternate_profile_id_recs: list[dict[str, Any]] = []
    try:
        if len(rec.get("conversation", [])) == 0:
            print("Invalid format Exception")
            raise ValueError(
                "Customer Service Interaction must have at least 1 conversation item "
            )
        new_rec = {}
        # overall interaction data
        new_rec["model_version"] = validate_string(
            rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["object_type"] = "customer_service_interaction"
        new_rec["last_updated"] = validate_string(
            rec.get("endTime", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["tx_id"] = tx_id
        new_rec["traveller_id"] = validate_string(
            rec.get("sessionId", ""), INVALID_CHARS_SINGLE_LINE
        )
        if new_rec["traveller_id"] == "":
            new_rec["traveller_id"] = str(uuid.uuid4())
        new_rec["accp_object_id"] = validate_string(
            rec.get("sessionId", ""), INVALID_CHARS_SINGLE_LINE
        )
        if new_rec["accp_object_id"] == "":
            new_rec["accp_object_id"] = str(uuid.uuid4())
        new_rec["channel"] = validate_string(
            rec.get("channel", ""), INVALID_CHARS_SINGLE_LINE
        )

        # traveler loyalty ID
        new_rec["loyalty_id"] = validate_string(
            rec.get("loyaltyId", ""), INVALID_CHARS_SINGLE_LINE
        )
        # related loyalty ID mentioned in the conversation
        new_rec["related_loyalty_id1"] = validate_string(
            safe_list_get(rec.get("relatedLoyaltyIds", []), 0, ""),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["related_loyalty_id2"] = validate_string(
            safe_list_get(rec.get("relatedLoyaltyIds", []), 1, ""),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["related_loyalty_id3"] = validate_string(
            safe_list_get(rec.get("relatedLoyaltyIds", []), 2, ""),
            INVALID_CHARS_SINGLE_LINE,
        )

        # traveler Booking ID
        new_rec["booking_id"] = validate_string(
            rec.get("relatedBookingId", ""), INVALID_CHARS_SINGLE_LINE
        )
        # related booking mentioned in the conversation
        new_rec["related_booking_id1"] = validate_string(
            safe_list_get(rec.get("relatedBookings", []), 0, ""),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["related_booking_id2"] = validate_string(
            safe_list_get(rec.get("relatedBookings", []), 1, ""),
            INVALID_CHARS_SINGLE_LINE,
        )
        new_rec["related_booking_id3"] = validate_string(
            safe_list_get(rec.get("relatedBookings", []), 2, ""),
            INVALID_CHARS_SINGLE_LINE,
        )

        # traveler First Name, last names, email, phone
        new_rec["first_name"] = validate_string(
            rec.get("firstName", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["last_name"] = validate_string(
            rec.get("lastName", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["email"] = validate_email(rec.get("email", ""))
        new_rec["email_business"] = validate_email(rec.get("businessEmail", ""))
        new_rec["phone_number"] = validate_phone(rec.get("phoneNumber", ""))

        # Campaign fields
        new_rec["campaign_job_id"] = validate_string(
            rec.get("campaignJobId", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["campaign_strategy"] = validate_string(
            rec.get("campaignStrategy", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["campaign_program"] = validate_string(
            rec.get("campaignProgram", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["campaign_product"] = validate_string(
            rec.get("campaignProduct", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["campaign_name"] = validate_string(
            rec.get("campaignName", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["category"] = validate_string(
            rec.get("category", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["subject"] = validate_string(
            rec.get("subject", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["loyalty_program_name"] = validate_string(
            rec.get("loyaltyProgramName", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["is_voice_otp"] = parse_bool(rec.get("isVoiceOtp", False))

        # TSA ID
        new_rec["tsa_precheck_number"] = validate_string(
            rec.get("tsaPreCheckNumber", ""), INVALID_CHARS_SINGLE_LINE
        )
        # CART ID
        new_rec["cart_id"] = validate_string(
            rec.get("cartId", ""), INVALID_CHARS_SINGLE_LINE
        )

        new_rec["interaction_type"] = validate_string(
            rec.get("interactionType", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["start_time"] = validate_string(
            rec.get("startTime", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["end_time"] = validate_string(
            rec.get("endTime", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["duration"] = parse_number(rec.get("duration", ""))
        new_rec["status"] = validate_string(
            rec.get("status", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["summary"] = validate_text(
            rec.get("summary", ""), INVALID_CHARS_MULTI_LINE
        )
        new_rec["language_code"] = validate_string(
            rec.get("language", {}).get("code", ""), INVALID_CHARS_SINGLE_LINE
        )
        new_rec["language_name"] = validate_string(
            rec.get("language", {}).get("name", ""), INVALID_CHARS_SINGLE_LINE
        )
        conversation: dict[str, Any] = {}
        conversation["items"] = []

        validate_conversation_length(rec.get("conversation", []))
        for conversation_item in rec.get("conversation", []):
            # conversation data
            item = {}
            item["from"] = validate_string(
                conversation_item.get("from", ""), INVALID_CHARS_SINGLE_LINE
            )
            item["to"] = validate_string(
                conversation_item.get("to", ""), INVALID_CHARS_SINGLE_LINE
            )
            item["content"] = conversation_item.get("content", "")
            item["start_time"] = validate_string(
                conversation_item.get("startTime", ""), INVALID_CHARS_SINGLE_LINE
            )
            item["end_time"] = validate_string(
                conversation_item.get("endTime", ""), INVALID_CHARS_SINGLE_LINE
            )
            item["sentiment"] = validate_string(
                conversation_item.get("sentiment", ""), INVALID_CHARS_SINGLE_LINE
            )
            conversation["items"].append(item)
            add_extended_data(item, conversation_item)

        add_extended_data(new_rec, rec)
        conversation_json = json.dumps(conversation)
        conversation_json_gzip_b64 = base64.b64encode(conversation_json.encode("utf-8"))
        new_rec["conversation"] = conversation_json_gzip_b64.decode("utf-8")
        set_timestamp(new_rec)
        new_recs.append(new_rec)

        add_alternate_profile_ids(
            alternate_profile_id_recs, rec, new_rec["traveller_id"], tx_id
        )
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)
        log_error_to_sqs(e, rec, err_queue_url, BIZ_OBJECT_TYPE, tx_id)
    return {
        CONVERSATION_RECORDS_KEY: new_recs,
        ALTERNATE_PROFILE_ID_RECORDS_KEY: deduplicate_list(alternate_profile_id_recs),
    }


# the TEXT column max size is 1GB and base 64 encoding adds 33% size so theoretical limit is 41M
# but we want to restrict to a max of 500 messages of 1000 char each.
def validate_conversation_length(conversation: Sequence[dict[str, Any]]) -> None:
    s = 0
    if len(conversation) > 500:
        raise ValidationError("Maximum number of conversation items supported is 500.")
    for conversation_item in conversation:
        s += len(conversation_item.get("content", ""))
    if s > 500000:
        raise ValidationError(
            "Maximum number of conversation items supported is 500000 after base64 encoding."
        )
