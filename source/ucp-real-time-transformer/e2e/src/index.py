
import boto3
import os
import base64
import json
from tah_lib import air_bookingTransform, hotel_stayTransform, hotel_bookingTransform, clickstreamTransform, pax_profileTransform, guest_profileTransform

TAH_REGION = os.getenv("TAH_REGION")
kinesis_client = boto3.client("kinesis", TAH_REGION)
sqs_resource = boto3.client("sqs", TAH_REGION)

output_stream = os.getenv("output_stream")
dlq_name = os.getenv("dlqname")

def handler(event, context):
  for record in event["Records"]:
    try:
      # Decode the kinesis data from base64 to utf-8
      kinesisRecord = record.get("kinesis", "")
      
      if kinesisRecord == "":
        put_response = sqs_resource.send_message(
          QueueUrl=dlq_name,
          MessageBody="Unable to get kinesis field from record"
        )
        print(put_response)
        continue

      dataRaw = kinesisRecord.get("data", "")
      if dataRaw == "":
        put_response = sqs_resource.send_message(
          QueueUrl=dlq_name,
          MessageBody="Unable to get data field from kinesis record"
        )
        print(put_response)
        continue

      data = base64.b64decode(dataRaw).decode("utf-8")
      output = data
      json_data = json.loads(data)

      objectType = json_data.get("objectType", "")
      inletData = json_data.get("data", "")

      if objectType == "hotel_stay":
        # Apply hotel_stay_revenue transform
        output = hotel_stayTransform.buildObjectRecord(inletData)
      elif objectType == "hotel_booking":
        # Apply hotel_booking transform
        output = hotel_bookingTransform.buildObjectRecord(inletData)
      elif objectType == "pax_profile":
        # Apply pax_profile transform
        output = pax_profileTransform.buildObjectRecord(inletData)
      elif objectType == "air_booking":
        # Apply air_booking transform
        output = air_bookingTransform.buildObjectRecord(inletData)
      elif objectType == "guest_profile":
        # Apply air_booking transform
        output = guest_profileTransform.buildObjectRecord(inletData)
      elif objectType == "clickstream":
        # Apply clickstream transform
        output = clickstreamTransform.buildObjectRecord(inletData)
      else:
        put_response = sqs_resource.send_message(
          QueueUrl=dlq_name,
          MessageBody="Unable to identify the objectType of the unpacked data"
        )
        continue
    
      partition_key = kinesisRecord.get("partitionKey", "")

      try:
          put_response = kinesis_client.put_record(
              StreamName=output_stream,
              Data=json.dumps(output),
              PartitionKey=partition_key
          )
          print(f'Successfully sent record to output stream.')
          print(put_response)
      except Exception as e:
        print(f'Failed to send record to output stream. Error: {e}')
    except Exception as ex:
      put_response = sqs_resource.send_message(
          QueueUrl=dlq_name,
          MessageBody="Exception thrown in script execution. Error: {ex}"
        )
      print(put_response)
