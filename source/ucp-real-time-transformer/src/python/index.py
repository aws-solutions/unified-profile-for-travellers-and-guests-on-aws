
import boto3
import os
import base64
import json
from tah_lib import air_bookingTransform, hotel_stayTransform, hotel_bookingTransform, clickstreamTransform, pax_profileTransform, guest_profileTransform, common

TAH_REGION = os.getenv("TAH_REGION")
DEAD_LETTER_QUEUE_URL = os.getenv("DEAD_LETTER_QUEUE_URL")

kinesis_client = boto3.client("kinesis", TAH_REGION)
output_stream = os.getenv("output_stream")

OBJECT_TYPE_HOTEL_BOOKING = "hotel_booking"
OBJECT_TYPE_PAX_PROFILE = "pax_profile"
OBJECT_TYPE_AIR_BOOKING = "air_booking"
OBJECT_TYPE_GUEST_PROFILE = "guest_profile"
OBJECT_TYPE_HOTEL_STAY = "hotel_stay"
OBJECT_TYPE_CLICKSTREAM = "clickstream"

BUSINESS_OBJECT_CONFIG_MAP = {
    OBJECT_TYPE_HOTEL_BOOKING: {
        "transformer": hotel_bookingTransform.buildObjectRecord,
        "keys": ["hotel_booking_recs", "common_email_recs", "common_phone_recs", "hotel_loyalty"]
    },
    OBJECT_TYPE_PAX_PROFILE: {
        "transformer": pax_profileTransform.buildObjectRecord,
        "keys": ["air_profile_recs", "common_email_recs", "common_phone_recs", "air_loyalty_recs"]
    },
    OBJECT_TYPE_AIR_BOOKING: {
        "transformer": air_bookingTransform.buildObjectRecord,
        "keys": ["air_booking_recs", "common_email_recs", "common_phone_recs", "air_loyalty_recs"]
    },
    OBJECT_TYPE_GUEST_PROFILE: {
        "transformer": guest_profileTransform.buildObjectRecord,
        "keys": ["guest_profile_recs", "common_email_recs", "common_phone_recs", "hotel_loyalty_recs"]
    },
    OBJECT_TYPE_HOTEL_STAY: {
        "transformer": hotel_stayTransform.buildObjectRecord,
        "keys": ["hotel_stay_revenue_items"]
    },
    OBJECT_TYPE_CLICKSTREAM: {
        "transformer": clickstreamTransform.buildObjectRecord,
        "keys": []
    }
}


def handler(event, context):
    return handleWithEnv(event, kinesis_client, output_stream, DEAD_LETTER_QUEUE_URL)


def handleWithEnv(event, kinesisClient, outStream, dlqUrl):
    print(f'Received ', len(event["Records"]),
          ' reccords from Kinesis stream')
    for record in event["Records"]:
        objectType = ""
        try:
            print("0-Decoding Kineisis record")
            # Decode the kinesis data from base64 to utf-8
            kinesisRecord = record.get("kinesis", "")
            if kinesisRecord == "":
                raise Exception("Unknown kinesis record format")

            dataRaw = kinesisRecord.get("data", "")
            if dataRaw == "":
                raise Exception("Unknown kinesis record format")

            data = base64.b64decode(dataRaw).decode("utf-8")
            json_data = json.loads(data)

            # mandatory metadata
            objectType = json_data.get("objectType", "")
            modelVersion = json_data.get("modelVersion", "")
            bizObj = json_data.get("data", "")
            domain = json_data.get("domain", "")
            if domain == "":
                raise Exception(
                    "No Amazon connect profile domain name provided")

            print("1-Ingesting object type ", objectType,
                  " version ", modelVersion, " in domain ", domain)

            print("2-Getting object configuration for object type ", objectType)
            bizObjectConfig = BUSINESS_OBJECT_CONFIG_MAP[objectType]
            if bizObjectConfig is None:
                raise Exception("Unknown object type", objectType)
            print("Object Type configuration found: ", bizObjectConfig)

            print("3-Transforming Business object to ACCP Records ")
            accpRecords = parseBizObjectToAccpRecords(
                bizObj, bizObjectConfig, dlqUrl)

            partition_key = kinesisRecord.get("partitionKey", "")
            newRecord = {
                "objectType": objectType,
                "modelVersion": modelVersion,
                "domain": domain,
                "data": accpRecords,
            }

            kinesisClient.put_record(
                StreamName=outStream,
                Data=json.dumps(newRecord, ensure_ascii=False),
                PartitionKey=partition_key
            )
            print(f'Successfully sent record to output stream.')
        except Exception as ex:
            if objectType == "":
                objectType = "unknown"
            common.logErrorToSQS(ex, record, dlqUrl, objectType)


def parseBizObjectToAccpRecords(bizObj, bizObjectConfig, dlqUrl):
    accpRecords = []
    output = bizObjectConfig["transformer"](bizObj, dlqUrl)
    # tranfomers can output either 1 ACCP record or multiple list of ACCP records for each sub-key. for ex:
    # clickstream object => accp_rec
    # air booking object => {sub_key1: [accp_rec_1, accp_rec_2], subkey2: ["accp_rec_..."]}
    # the code below handle both cases
    print("4-Aggregating ACCP Records")
    if len(bizObjectConfig["keys"]) > 0:
        print("output has ", len(bizObjectConfig["keys"]),
              " sub-keys (1 to many transformation)")
        for key in bizObjectConfig["keys"]:
            for accpRec in output.get(key, []):
                accpRecords.append(accpRec)
    else:
        print("output has no sub key (1 to 1 tranfromation)")
        accpRecords.append(output)
    return accpRecords
