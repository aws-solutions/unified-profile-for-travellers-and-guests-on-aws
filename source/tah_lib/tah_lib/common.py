# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import base64
import json
import os
import re
import traceback
from collections.abc import Iterable, Sequence
from datetime import datetime
from typing import TYPE_CHECKING, Any, Final, TypedDict, TypeVar

import boto3
import requests
from botocore.config import Config

if TYPE_CHECKING:
    from mypy_boto3_dynamodb.client import DynamoDBClient
else:
    DynamoDBClient = object

SUPPORTED_ADRESS_TYPES: Final = ["", "billing", "business", "mailing"]
SUPPORTED_ADRESS_FIELDS: Final = [
    "line1",
    "line2",
    "line3",
    "line4",
    "city",
    "state",
    "province",
    "postal_code",
    "country",
    "is_primary",
]
FIELD_NAME_TRAVELLER_ID: Final = "traveller_id"
FIELD_NAME_LAST_UPDATED: Final = "last_updated"
# field added to be used in S3 partitining which requires removing some special char
FIELD_NAME_LAST_UPDATED_PARTITION: Final = "last_updated_partition"
# 2021-12-09T07:59:14.873255Z
DATETIME_FORMAT: Final = "%Y-%m-%dT%H:%M:%S.%fZ"

SQS_MES_ATTR_UCP_ERROR_TYPE: Final = "UcpErrorType"
SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: Final = "BusinessObjectTypeName"
SQS_MES_ATTR_MESSAGE: Final = "Message"
SQS_MES_ATTR_TXID: Final = "TransactionId"

# max field size supported by DB
MAX_FIELD_SIZE: Final = 255
MAX_TEXT_SIZE: Final = 5000
# max size of a phone number taking into account extension and many chars
MAX_PHONE_SIZE: Final = 30

# maximum length of extended data
MAX_EXTENDED_DATA = int(os.getenv("MAX_EXTENDED_DATA", "5000"))

# characters that should not be supported in string fields
INVALID_CHARS_SINGLE_LINE: Final = [
    "\\",
    "\b",
    "\f",
    "\n",
    "\r",
    "\t",
    "{",
    "}",
    "<",
    ">",
    "^",
    "=",
]
INVALID_CHARS_MULTI_LINE: Final[list[str]] = []
INVALID_CHARS_URL: Final = ["\n", "\r"]

EMAIL_REGEXP: Final = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,63}\b"
URL_REGEXP: Final = (
    r"((http|https)\:\/\/)?[a-zA-Z0-9\.\/\?\:@\-_=#]+\.([a-zA-Z]){2,6}([a-zA-Z0-9\.\&\/\?\:@\-_=#])*"
)

AIR_BOOKING_RECORDS_KEY: Final = "air_booking_recs"
AIR_LOYALTY_RECORDS_KEY: Final = "air_loyalty_recs"
AIR_LOYALTY_TX_RECORDS_KEY: Final = "air_loyalty_tx_recs"
AIR_PROFILE_RECORDS_KEY: Final = "air_profile_recs"
ALTERNATE_PROFILE_ID_RECORDS_KEY: Final = "alternate_profile_id_recs"
ANCILLARY_RECORDS_KEY: Final = "ancillary_recs"
CLICKSTREAM_RECORDS_KEY: Final = "clickstream_recs"
COMMON_EMAIL_RECORDS_KEY: Final = "common_email_recs"
COMMON_PHONE_RECORDS_KEY: Final = "common_phone_recs"
CONVERSATION_RECORDS_KEY: Final = "conversation_recs"
GUEST_LOYALTY_RECORDS_KEY: Final = "guest_loyalty_recs"
GUEST_PROFILE_RECORDS_KEY: Final = "guest_profile_recs"
HOTEL_BOOKING_RECORDS_KEY: Final = "hotel_booking_recs"
HOTEL_LOYALTY_RECORDS_KEY: Final = "hotel_loyalty_recs"
HOTEL_LOYALTY_TX_RECORDS_KEY: Final = "hotel_loyalty_tx_recs"
HOTEL_STAY_REVENUE_ITEMS_RECORDS_KEY: Final = "hotel_stay_revenue_items"

ALTERNATE_PROFILE_ID_ACCP_OBJECT_TYPE: Final = "alternate_profile_id"


class ValidationError(Exception):
    def __init__(self, message: str) -> None:
        super().__init__(message)


class NoBookmarkError(Exception):
    def __init__(self, message: str) -> None:
        super().__init__(message)


# resilient float parsing to be used to avoid cases where both int and float would come in for the same value which would lead to
# CSV data source does not support struct<int:int,double:double> data type.
# pyspark.sql.utils.AnalysisException: CSV data source does not support struct<int:int,double:double> data type.
# we return 0 if we can't parse the number
def parse_number(num: Any) -> float:
    try:
        return float(num)
    except Exception:
        return 0.0


def parse_int(num: Any) -> int:
    try:
        return int(num)
    except Exception:
        return 0


def parse_bool(value: Any) -> bool:
    try:
        if isinstance(value, bool):
            return value

        if isinstance(value, str):
            lower_value = value.lower()
            return lower_value == "true" or lower_value == "false"

        return False
    except Exception:
        return False


# validate input string by checking for invalid characters and max length
def validate_string(strVal: Any, INVALID_CHAR: Iterable[str]) -> Any:
    if strVal == "":
        return ""
    if strVal is None:
        return None
    strVal_str = str(strVal)
    if len(strVal_str) > MAX_FIELD_SIZE:
        raise ValidationError(
            "field value of size "
            + str(len(strVal_str))
            + "  exceeds the maximum field length of "
            + str(MAX_FIELD_SIZE)
        )
    for ch in INVALID_CHAR:
        if ch in strVal_str:
            raise ValidationError("field value contains invalid characters " + ch)
    return strVal_str


def validate_text(strVal: Any, INVALID_CHAR: Iterable[str]) -> Any:
    if strVal == "" or strVal is None:
        return ""
    strVal = str(strVal)
    if len(strVal) > MAX_TEXT_SIZE:
        raise ValidationError(
            "field value of size "
            + str(len(strVal))
            + "  exceeds the maximum field length of "
            + str(MAX_TEXT_SIZE)
        )
    for ch in INVALID_CHAR:
        if ch in strVal:
            raise ValidationError("field value contains invalid characters " + ch)
    return strVal


def validate_email(strVal: Any) -> Any:
    if strVal == "":
        return ""
    if strVal is None:
        return None
    validate_string(strVal, INVALID_CHARS_SINGLE_LINE)
    if not re.fullmatch(EMAIL_REGEXP, strVal):
        print("invalid email format: ", strVal)
        raise ValidationError("Email field contains invalid email format")
    return strVal


def validate_url(strVal: str | None) -> str | None:
    if strVal == "":
        return ""
    if strVal is None:
        return None
    if len(strVal) > MAX_FIELD_SIZE:
        raise ValidationError(
            "field value exceeds the maximum field length of " + str(MAX_FIELD_SIZE)
        )
    if not re.fullmatch(URL_REGEXP, strVal):
        print("invalid url format: ", strVal)
        raise ValidationError("URL field contains invalid url format")
    return strVal


def validate_phone(strVal: Any) -> Any:
    if strVal == "":
        return ""
    if strVal is None:
        return None
    validate_string(strVal, INVALID_CHARS_SINGLE_LINE)
    if len(strVal) > MAX_PHONE_SIZE:
        raise ValidationError(
            "phone number field value exceeds the maximum field length of "
            + str(MAX_PHONE_SIZE)
        )
    return strVal


def set_traveller_id(
    rec: dict[str, Any], guest_biz_object: dict[str, Any], uid: str
) -> None:
    rec[FIELD_NAME_TRAVELLER_ID] = guest_biz_object.get("id", "")
    if rec[FIELD_NAME_TRAVELLER_ID] == "":
        rec[FIELD_NAME_TRAVELLER_ID] = uid


# setting default timestamp if does not exist
def set_timestamp(rec: dict[str, Any]) -> None:
    if rec.get(FIELD_NAME_LAST_UPDATED, "") == "":
        rec[FIELD_NAME_LAST_UPDATED] = datetime.now().strftime(DATETIME_FORMAT)
    else:
        # Validate existing timestamp format
        try:
            datetime.strptime(rec[FIELD_NAME_LAST_UPDATED], DATETIME_FORMAT)
        except ValueError as ex:
            raise ValidationError(f"invalid last updated time: {ex}")
    partition_str = datetime.strptime(
        rec[FIELD_NAME_LAST_UPDATED].split(".")[0], "%Y-%m-%dT%H:%M:%S"
    ).strftime("%Y-%m-%d-%H")
    rec[FIELD_NAME_LAST_UPDATED_PARTITION] = partition_str


def add_alternate_profile_ids(
    alt_prof_ids_list: list[dict[str, Any]], rec: dict[str, Any], cid: str, tx_id: str
) -> None:
    """
    Add Alternate Profile IDs to extracted record.

    Parameters
    ----------
    alt_prof_ids : List
        Existing List of AltProfIds to append to; passed by reference
    rec : Obj
        Object containing key "alternateProfileIds"
    cid : Str
        Fallback Traveller ID if rec.get("id") == None
    tx_id : Str
        Transaction ID
    """
    for alt_prof_id_rec in rec.get("alternateProfileIds", []):
        # Create alternate profile id record that links to the empty profile
        alt_proj_id_obj = {
            "object_type": ALTERNATE_PROFILE_ID_ACCP_OBJECT_TYPE,
            "model_version": validate_string(
                rec.get("modelVersion", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "last_updated": validate_string(
                rec.get("lastUpdatedOn", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "value": validate_string(
                alt_prof_id_rec.get("value", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "name": validate_string(
                alt_prof_id_rec.get("name", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "description": validate_string(
                alt_prof_id_rec.get("description", ""), INVALID_CHARS_SINGLE_LINE
            ),
            "tx_id": tx_id,
        }
        set_timestamp(alt_prof_id_rec)
        alt_proj_id_obj[FIELD_NAME_TRAVELLER_ID] = alt_proj_id_obj["accp_object_id"] = (
            f'{alt_proj_id_obj["name"]}_{alt_proj_id_obj["value"]}'
        )
        alt_prof_ids_list.append(alt_proj_id_obj)

        # Create alternate profile id record that links to the main (hydrated) profile
        hydrated_rec = alt_proj_id_obj.copy()
        set_traveller_id(hydrated_rec, rec, cid)
        alt_prof_ids_list.append(hydrated_rec)


T = TypeVar("T", bound=dict[str, Any])


def deduplicate_list(list_: Iterable[T]) -> list[T]:
    """
    Deduplicate a list of objects, preserving their order within the list

    Parameters
    ----------
    list: List[any]
        List of objects to be deduplicated

    Returns
    -------
    Deduplicated list, with original order preserved
    """
    visited = set()
    deduped_list: list[T] = []
    for obj in list_:
        t = tuple(sorted(obj.items()))
        if t not in visited:
            visited.add(t)
            deduped_list.append(obj)

    return deduped_list


# this function set the primary email according to the following logic
# 1- if one or multiple emails are flagged as Primary, we choose the first one
# 2- if not we choose the first email in the list.
# 3- then we use the email type to choose whether we set the biz_email key or 'email' key
# 3- if the list is empty we set both keys to empty strings to allow creation of the column in the CSV
def set_primary_email(rec: dict[str, Any], emails: Sequence[dict[str, Any]]) -> None:
    rec["email_business"] = ""
    rec["email"] = ""
    rec["email_primary"] = ""
    if len(emails) == 0:
        return
    for email in emails:
        etype = email.get("type", "").lower()
        if etype == "business":
            rec["email_business"] = email.get("address", "")
        else:
            rec["email"] = email.get("address", "")
        if email.get("primary", False):
            rec["email_primary"] = etype


# this function sets the primary phone using the same logic than the email function
# we have 4 types of phones: default, business, mobile and home
def set_primary_phone(rec: dict[str, Any], phones: Sequence[dict[str, Any]]) -> None:
    rec["phone"] = ""
    rec["phone_home"] = ""
    rec["phone_mobile"] = ""
    rec["phone_business"] = ""
    rec["phone_primary"] = ""
    if len(phones) == 0:
        return
    for phone in phones:
        ptype = phone.get("type", "").lower()
        if ptype == "business":
            rec["phone_business"] = phone.get("number", "")
        elif ptype == "home":
            rec["phone_home"] = phone.get("number", "")
        elif ptype == "mobile":
            rec["phone_mobile"] = phone.get("number", "")
        else:
            rec["phone"] = phone.get("number", "")
        if phone.get("primary", False):
            rec["phone_primary"] = ptype


# this function sets the primary address  using the same logic than the email function
def set_primary_address(
    rec: dict[str, Any], addresses: Sequence[dict[str, Any]]
) -> None:
    # initializing all address types if they do not exist on the record
    for addr_type in SUPPORTED_ADRESS_TYPES:
        if addr_type != "":
            addr_type += "_"
        for field in SUPPORTED_ADRESS_FIELDS:
            if "address_" + addr_type + field not in rec:
                if field == "is_primary":
                    rec["address_" + addr_type + field] = False
                else:
                    rec["address_" + addr_type + field] = ""
    if len(addresses) == 0:
        return
    for address in addresses:
        set_address_from_type(rec, address)


def set_address_from_type(rec: dict[str, Any], primary_address: dict[str, Any]) -> None:
    atype = primary_address.get("type", "").lower()
    if atype == "business":
        set_address(rec, "business", primary_address)
    elif atype == "shipping":
        set_address(rec, "shipping", primary_address)
    elif atype == "mailing":
        set_address(rec, "mailing", primary_address)
    else:
        set_address(rec, "", primary_address)
    rec["address_type"] = atype


def set_billing_address(rec: dict[str, Any], payment_info: dict[str, Any]) -> None:
    if "ccInfo" in payment_info and "address" in payment_info["ccInfo"]:
        set_address(rec, "billing", payment_info["ccInfo"]["address"])
    elif "address" in payment_info:
        set_address(rec, "billing", payment_info["address"])
    else:
        rec["address_billing_line1"] = ""
        rec["address_billing_line2"] = ""
        rec["address_billing_line3"] = ""
        rec["address_billing_line4"] = ""
        rec["address_billing_city"] = ""
        rec["address_billing_state"] = ""
        rec["address_billing_province"] = ""
        rec["address_billing_postal_code"] = ""
        rec["address_billing_country"] = ""


def set_address(rec: dict[str, Any], add_type: str, address: dict[str, Any]) -> None:
    if add_type not in SUPPORTED_ADRESS_TYPES:
        return
    if address is None:
        address = {}
    if add_type != "":
        add_type += "_"
    rec["address_" + add_type + "line1"] = address.get("line1", "")
    rec["address_" + add_type + "line2"] = address.get("line2", "")
    rec["address_" + add_type + "line3"] = address.get("line3", "")
    rec["address_" + add_type + "line4"] = address.get("line4", "")
    rec["address_" + add_type + "postal_code"] = address.get("postalCode", "")
    rec["address_" + add_type + "city"] = address.get("city", "")
    rec["address_" + add_type + "state"] = address.get("state", {}).get("code", "")
    rec["address_" + add_type + "province"] = address.get("province", {}).get(
        "code", ""
    )
    rec["address_" + add_type + "country"] = address.get("country", {}).get("code", "")
    rec["address_" + add_type + "is_primary"] = address.get("primary", False)


def set_payment_info(rec: dict[str, Any], payment_info: dict[str, Any]) -> None:
    rec["payment_type"] = payment_info.get("paymentType", "")
    rec["cc_token"] = payment_info.get("ccInfo", {}).get("token", "")
    rec["cc_type"] = payment_info.get("ccInfo", {}).get("cardType", "")
    rec["cc_exp"] = payment_info.get("ccInfo", {}).get("cardExp", "")
    rec["cc_cvv"] = payment_info.get("ccInfo", {}).get("cardCvv", "")
    rec["cc_name"] = payment_info.get("ccInfo", {}).get("name", "")


def get_external_id(external_ids: Iterable[dict[str, str]] | None, system: str) -> str:
    if external_ids is None:
        return ""
    for ex_id in external_ids:
        if ex_id.get("originatingSystem", "") == system:
            return ex_id.get("id", "")
    return ""


def build_serialized_lists(list_: Iterable[dict[str, str]], key: str, sep: str) -> str:
    parts = []
    for item in list_:
        if key in item:
            item[key].replace(sep, "")
            parts.append(item[key])
    return sep.join(parts)


def replace_and_join(array: Iterable[str], sep: str) -> str:
    parts = []
    for item in array:
        item.replace(sep, "")
        parts.append(item)
    return sep.join(parts)


U = TypeVar("U")


def none_to_list(val: list[U]) -> U | list[U]:
    if val is None:
        return []
    return val


def log_error_to_sqs(
    e: Exception, rec: dict[str, Any], err_queue_url: str, biz_obj_type: str, tx_id: str
) -> None:
    if err_queue_url == "":
        print(tx_id, "-No queue URL provided")
        return
    try:
        print(tx_id, "-writing error to queue: ", err_queue_url)
        append_solution_identifier = build_solution_header("SO0244", "1.0.0")
        config = Config(**append_solution_identifier)
        client = boto3.client("sqs", config=config)
        res = client.send_message(
            QueueUrl=err_queue_url,
            MessageBody=json.dumps(rec),
            MessageAttributes={
                SQS_MES_ATTR_UCP_ERROR_TYPE: {
                    "StringValue": "transformer_exception",
                    "DataType": "String",
                },
                SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: {
                    "StringValue": biz_obj_type,
                    "DataType": "String",
                },
                SQS_MES_ATTR_MESSAGE: {"StringValue": str(e), "DataType": "String"},
                SQS_MES_ATTR_TXID: {"StringValue": tx_id, "DataType": "String"},
            },
        )
        print(tx_id, "-[INFO] Message sent to queue", res)
    except Exception as e2:
        print(
            tx_id,
            "-[ERROR] An exception occured when trying to write error to queue",
            e2,
        )
        traceback_info = traceback.format_exc()
        print(tx_id, "-", traceback_info)


class JobPredicates(TypedDict):
    pdp: str | None
    cpp: str | None


# Control the date range for data being processed based on the last job run.
# The last job run's timestamp is used to build a push down predicate and
# catalog partition predicate. These allow us to effeciently process new
# data without reprocessing old data or having to scan extra partitions.
# Read more: https://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-etl-partitions.html
def get_job_predicates(
    dynamodb_client: DynamoDBClient,
    dynamo_table_name: str,
    biz_object_name: str,
    domain_name: str,
) -> JobPredicates:
    pk = "glue_job_bookmark"
    sk = biz_object_name + "_" + domain_name
    print("[get_job_predicates] getting job bookmark for " + sk)
    response = dynamodb_client.get_item(
        TableName=dynamo_table_name,
        Key={
            "item_id": {"S": pk},
            "item_type": {"S": sk},
        },
    )
    try:
        bookmark = response.get("Item", {}).get("bookmark", {}).get("S")
        if bookmark is None:
            raise NoBookmarkError("no bookmark found")
        print("[get_job_predicates] bookmark returned: " + bookmark)
        y, m, d, _ = bookmark.split("/")
        # check if it's 1) a later year, 2) later month of same year, or 3) same/later day of same month and year
        is_later_year = "(year > '" + y + "')"
        is_later_month_same_year = "(year = '" + y + "' AND month > '" + m + "')"
        is_same_or_later_day = (
            "(year = '" + y + "' AND month = '" + m + "' AND day >= '" + d + "')"
        )
        pdp = (
            is_later_year
            + " OR "
            + is_later_month_same_year
            + " OR "
            + is_same_or_later_day
        )
        cpp = (
            is_later_year
            + " OR "
            + is_later_month_same_year
            + " OR "
            + is_same_or_later_day
        )
        return {"pdp": pdp, "cpp": cpp}
    except Exception:
        print("[get_job_predicates] no bookmark found")
        return {"pdp": None, "cpp": None}


# Send operational metrics to AWS Solutions Builder service
def send_solution_metrics(
    solution_id: str,
    solution_version: str,
    metrics_uuid: str,
    send_anonymized_data: str,
    payload: dict[str, Any],
) -> tuple[bool, int | None]:
    # Only send metrics if customer allows anonymous logging
    if send_anonymized_data.lower() != "yes":
        return False, None
    url = "https://metrics.awssolutionsbuilder.com/generic"
    data = {
        "Solution": solution_id,
        "Version": solution_version,
        "Timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f"),
        "UUID": metrics_uuid,
        "Data": payload,
    }
    json_data = json.dumps(data)
    headers = {"Content-Type": "application/json"}
    response = requests.post(url, data=json_data, headers=headers)
    if response.status_code != 200:
        print(
            "[send_solution_metrics] Error sending operational metrics",
            response.status_code,
        )
        return True, response.status_code
    else:
        return True, None


class SolutionHeader(TypedDict):
    user_agent_extra: str


def build_solution_header(solution_id: str, solution_version: str) -> SolutionHeader:
    solution_identifier = (
        "AWSSOLUTION/" + str(solution_id) + "/v" + str(solution_version)
    )
    append_solution_identifier: SolutionHeader = {
        "user_agent_extra": solution_identifier
    }
    return append_solution_identifier


V = TypeVar("V")


def safe_list_get(list_: Sequence[V], idx: int, default: V) -> V:
    try:
        return list_[idx]
    except IndexError:
        return default


def validate_and_encode_json(data: Any) -> str:
    """
    Validates if input is valid JSON and returns base64 encoded version if valid.

    Args:
        data: Input to validate as JSON

    Returns:
        str: Base64 encoded JSON string if valid

    Raises:
        json.JSONDecodeError: If input is not valid JSON
        ValidationError: If input is above specified length
    """
    # Validate JSON by attempting to parse it
    if isinstance(data, str):
        json_obj = json.loads(data)
    else:
        json_obj = json.loads(json.dumps(data))

    # Convert back to string and encode as base64
    json_str = json.dumps(json_obj)
    enc = base64.b64encode(json_str.encode()).decode()
    if len(enc) > MAX_EXTENDED_DATA:
        raise ValidationError(
            f"Maximum length of encoded extended data is {MAX_EXTENDED_DATA}"
        )
    return enc


def add_extended_data(rec: dict[str, Any], accp_opj: dict[str, Any]) -> None:
    """
    Adds extended data to the record.

    Args:
        rec: Record to add extended data to
        extended_data: Extended data to add to the record
    """
    extended_data = accp_opj.get("extendedData", None)
    if extended_data:
        rec["extended_data"] = validate_and_encode_json(extended_data)
    else:
        rec["extended_data"] = ""
    return
