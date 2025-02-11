# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import base64
import json
import os
import re
import traceback
import uuid
from datetime import datetime

import boto3
from botocore.config import Config
from tah_lib import (air_bookingTransform, clickstreamTransform, common,
                     customer_service_interactionTransform,
                     guest_profileTransform, hotel_bookingTransform,
                     hotel_stayTransform, pax_profileTransform)
from tah_lib.common import (AIR_BOOKING_RECORDS_KEY, AIR_LOYALTY_RECORDS_KEY,
                            AIR_LOYALTY_TX_RECORDS_KEY,
                            AIR_PROFILE_RECORDS_KEY,
                            ALTERNATE_PROFILE_ID_RECORDS_KEY,
                            ANCILLARY_RECORDS_KEY, CLICKSTREAM_RECORDS_KEY,
                            COMMON_EMAIL_RECORDS_KEY, COMMON_PHONE_RECORDS_KEY,
                            CONVERSATION_RECORDS_KEY,
                            GUEST_LOYALTY_RECORDS_KEY,
                            GUEST_PROFILE_RECORDS_KEY,
                            HOTEL_BOOKING_RECORDS_KEY,
                            HOTEL_LOYALTY_RECORDS_KEY,
                            HOTEL_LOYALTY_TX_RECORDS_KEY,
                            HOTEL_STAY_REVENUE_ITEMS_RECORDS_KEY)

TAH_REGION = os.getenv("TAH_REGION")
if TAH_REGION == "":
    # getting region for codeBuild project
    TAH_REGION = os.getenv("AWS_REGION")
DEAD_LETTER_QUEUE_URL = os.getenv("DEAD_LETTER_QUEUE_URL")
FIREHOSE_BACKUP_STREAM = os.getenv("FIREHOSE_BACKUP_STREAM")
ENABLE_BACKUP = os.getenv("ENABLE_BACKUP", "false")

METRICS_SOLUTION_ID = os.getenv("METRICS_SOLUTION_ID")
METRICS_SOLUTION_VERSION = os.getenv("METRICS_SOLUTION_VERSION")
MAX_DOMAIN_NAME_LENGTH = 26
MIN_DOMAIN_NAME_LENGTH = 3

append_solution_identifier = common.build_solution_header(METRICS_SOLUTION_ID, METRICS_SOLUTION_VERSION)
config = Config(
    region_name=TAH_REGION,
    **append_solution_identifier,
)
kinesis_client = boto3.client("kinesis", config=config)
firehose_client = boto3.client("firehose", config=config)
s3 = boto3.client("s3", config=config)
output_stream = os.getenv("OUTPUT_STREAM")

OBJECT_TYPE_HOTEL_BOOKING = "hotel_booking"
OBJECT_TYPE_PAX_PROFILE = "pax_profile"
OBJECT_TYPE_AIR_BOOKING = "air_booking"
OBJECT_TYPE_GUEST_PROFILE = "guest_profile"
OBJECT_TYPE_HOTEL_STAY = "hotel_stay"
OBJECT_TYPE_CLICKSTREAM = "clickstream"
OBJECT_TYPE_CSI = "customer_service_interaction"

BUSINESS_OBJECT_CONFIG_MAP = {
    OBJECT_TYPE_HOTEL_BOOKING: {
        "transformer": hotel_bookingTransform.build_object_record,
        "keys": [HOTEL_BOOKING_RECORDS_KEY, COMMON_EMAIL_RECORDS_KEY, COMMON_PHONE_RECORDS_KEY, GUEST_LOYALTY_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    },
    OBJECT_TYPE_PAX_PROFILE: {
        "transformer": pax_profileTransform.build_object_record,
        "keys": [AIR_PROFILE_RECORDS_KEY, COMMON_EMAIL_RECORDS_KEY, COMMON_PHONE_RECORDS_KEY, AIR_LOYALTY_RECORDS_KEY, AIR_LOYALTY_TX_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    },
    OBJECT_TYPE_AIR_BOOKING: {
        "transformer": air_bookingTransform.build_object_record,
        "keys": [AIR_BOOKING_RECORDS_KEY, COMMON_EMAIL_RECORDS_KEY, COMMON_PHONE_RECORDS_KEY, AIR_LOYALTY_RECORDS_KEY, ANCILLARY_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    },
    OBJECT_TYPE_GUEST_PROFILE: {
        "transformer": guest_profileTransform.build_object_record,
        "keys": [GUEST_PROFILE_RECORDS_KEY, COMMON_EMAIL_RECORDS_KEY, COMMON_PHONE_RECORDS_KEY, HOTEL_LOYALTY_RECORDS_KEY, HOTEL_LOYALTY_TX_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    },
    OBJECT_TYPE_HOTEL_STAY: {
        "transformer": hotel_stayTransform.build_object_record,
        "keys": [HOTEL_STAY_REVENUE_ITEMS_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    },
    OBJECT_TYPE_CSI: {
        "transformer": customer_service_interactionTransform.build_object_record,
        "keys": [CONVERSATION_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    }
    ,
    OBJECT_TYPE_CLICKSTREAM: {
        "transformer": clickstreamTransform.build_object_record,
        "keys": [CLICKSTREAM_RECORDS_KEY, ALTERNATE_PROFILE_ID_RECORDS_KEY]
    }
}


class IngestionException(Exception):
    """Raised when there is an error during ingestion"""
    pass


def handler(event, context):
    print("Lambda context: ", context)
    return handle_with_env(event, kinesis_client, firehose_client, output_stream, DEAD_LETTER_QUEUE_URL, FIREHOSE_BACKUP_STREAM, ENABLE_BACKUP)

def handle_with_env(event, kinesis_client, firehose_client, out_stream, dlq_url, delivery_stream_name, enable_backup):
    tx_id = generate_tx_id()
    print(tx_id, f'Received ', len(event["Records"]),
          ' reccords from Kinesis stream')
    for record in event["Records"]:
        object_type = ""
        try:
            print(tx_id, "0 Extracting Kinesis record")
            # Decode the kinesis data from base64 to utf-8
            kinesis_record = record.get("kinesis", "")
            if kinesis_record == "":
                raise IngestionException("Unknown kinesis record format")

            data_raw = kinesis_record.get("data", "")
            if data_raw == "":
                raise IngestionException("Unknown kinesis record format")

            json_data = decode_raw_data(tx_id, data_raw)
            if enable_backup == 'true':
                backup_in_s3(tx_id, firehose_client, delivery_stream_name, json_data)
            accp_records = biz_object_to_accp_records(
                tx_id, json_data.get("data", ""), json_data.get("objectType", ""), dlq_url)
            if len(accp_records) == 0:
                object_type = json_data.get("objectType", "")
                raise IngestionException("Transformation resulted in 0 profile objects")
            print(f'{tx_id} 5-Ingesting {len(accp_records)} records in downstream stream')
            for accp_record in accp_records:
                kinesis_client.put_record(
                    StreamName=out_stream,
                    Data=json.dumps(wrap_accp_record(
                        tx_id, accp_record, json_data), ensure_ascii=False),
                    PartitionKey=kinesis_record.get(
                        "partitionKey", build_partition_key(tx_id, accp_record))
                )
            print(tx_id, 'Successfully sent record to output stream.')
        except Exception as ex:
            if object_type == "":
                object_type = "unknown"
            print(tx_id, "Error occured while ingesting record to downstream stream", ex)
            common.log_error_to_sqs(ex, record, dlq_url, object_type, tx_id)


def build_partition_key(tx_id, accp_record):
    pid = accp_record.get("traveller_id",str(uuid.uuid4()))
    print(tx_id, "[build_partition_key] Partition key is", pid)
    return pid

def decode_raw_data(tx_id, data_raw):
    print(tx_id, "1-Decoding raw data")
    data = base64.b64decode(data_raw).decode("utf-8")
    json_data = json.loads(data)
    return json_data


def wrap_accp_record(tx_id, accp_record, json_data):
    print(tx_id, "The mode is", json_data.get("mode", ""))
    return {
        "transactionId": tx_id,
        "objectType": json_data.get("objectType", ""),
        "modelVersion": json_data.get("modelVersion", ""),
        "domain": validate_domain_name(json_data.get("domain", "")),
        "mode": json_data.get("mode", ""),
        "mergeModeProfileId": json_data.get("mergeModeProfileId", ""),
        "partialModeOptions": json_data.get("partialModeOptions", {"fields": []}),
        "data": [accp_record],
    }


def validate_domain_name(name):
    # Check if the name length is within the valid range
    if len(name) > MAX_DOMAIN_NAME_LENGTH or len(name) < MIN_DOMAIN_NAME_LENGTH:
        raise IngestionException(f"Domain name length should be between {MIN_DOMAIN_NAME_LENGTH} and {MAX_DOMAIN_NAME_LENGTH} chars")

    # Check if the name is in snake case
    snake_case_pattern = r'^[a-z0-9]+(?:_[a-z0-9]+)*$'
    if not re.match(snake_case_pattern, name):
        raise IngestionException("Invalid domain name (must be snake case between 3 and 63 chars)")

    return name

def backup_in_s3(tx_id, firehose_client, delivery_stream_name, json_data):
    try:
        print(tx_id, "2-Backup in S3 through firehose stream ", delivery_stream_name)
        # store records in S3 as backup
        # we use the evelop timestamp to support replay of events from the S3 backup
        timstamp = datetime.now()
        uid = str(uuid.uuid4())
        if "timestamp" in json_data:
            timstampstr = json_data.get("timestamp")
            timstamp = datetime.strptime(timstampstr, "%Y-%m-%dT%H:%M:%S.%fZ")
        if "uid" in json_data:
            uid = json_data.get("uid")
        json_data["timestamp"] = timstamp.strftime("%Y-%m-%dT%H:%M:%S.%fZ")
        json_data["uid"] = uid
        firehose_client.put_record(
            DeliveryStreamName=delivery_stream_name,
            Record={'Data': json.dumps(json_data)}
        )
    except Exception as ex:
        traceback.print_exc()
        print(tx_id, "Error backing up record in S3:", ex)


def biz_object_to_accp_records(tx_id, biz_obj, object_type, dlq_url):
    print(tx_id, "4-Transforming Business object to ACCP Records ")
    accp_records = []
    if object_type not in BUSINESS_OBJECT_CONFIG_MAP:
        raise IngestionException("Unknown object type", object_type)
    biz_object_config = BUSINESS_OBJECT_CONFIG_MAP[object_type]
    print(tx_id, "Object Type configuration found: ", biz_object_config)
    output = biz_object_config["transformer"](biz_obj, dlq_url, tx_id)
    # tranfomers can output either 1 ACCP record or multiple list of ACCP records for each sub-key. for ex:
    # clickstream object => accp_rec
    # air booking object => {sub_key1: [accp_rec_1, accp_rec_2], subkey2: ["accp_rec_..."]}
    # the code below handle both cases
    print(tx_id, "Aggregating ACCP Records")
    if len(biz_object_config["keys"]) > 0:
        print(tx_id, "-output has ", len(biz_object_config["keys"]),
              " sub-keys (1 to many transformation)")
        for key in biz_object_config["keys"]:
            for accp_rec in output.get(key, []):
                accp_records.append(accp_rec)
    else:
        print(tx_id, "output has no sub key (1 to 1 tranformation)")
        accp_records.append(output)
    return accp_records


def generate_tx_id():
    return str(uuid.uuid4())


def build_s3_key(timestamp, uid):
    # build s3 key useing current timestamp in a YYYY/MM/DD/HH format and partition key
    # and sequence number of the record
    # Extract year, month, day, and hour from the current date and time
    year = str(timestamp.year)
    month = str(timestamp.month).zfill(2)  # Zero padding if single digit
    day = str(timestamp.day).zfill(2)
    hour = str(timestamp.hour).zfill(2)
    # Build the string using the extracted values
    formatted_string = f"{year}/{month}/{day}/{hour}"
    return formatted_string + "/" + uid + ".json"
