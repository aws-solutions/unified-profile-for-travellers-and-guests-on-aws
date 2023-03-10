
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

#There is an Access Denied exception here, spent way to long trying 
#to figure out how this works so for now its commented, will try
#to address later
#queue = sqs_resource.get_queue_by_name(QueueName = dlq_name)

def handler(event, context):
  for record in event["Records"]:
    try:
      # Decode the kinesis data from base64 to utf-8
      print(str(record))
      data = base64.b64decode(record["kinesis"]["data"]).decode("utf-8")
      output = data
      json_data = json.loads(data)
      print(json_data)

      if json_data["objectType"] == "hotel_stay":
        # Apply hotel_stay_revenue transform
        output = hotel_stayTransform.buildObjectRecord(json_data["data"])
      elif json_data["objectType"] == "hotel_booking":
        # Apply hotel_booking transform
        output = hotel_bookingTransform.buildObjectRecord(json_data["data"])
      elif json_data["objectType"] == "pax_profile":
        # Apply pax_profile transform
        output = pax_profileTransform.buildObjectRecord(json_data["data"])
      elif json_data["objectType"] == "air_booking":
        # Apply air_booking transform
        output = air_bookingTransform.buildObjectRecord(json_data["data"])
      elif json_data["objectType"] == "guest_profile":
        # Apply air_booking transform
        output = guest_profileTransform.buildObjectRecord(json_data["data"])
      elif json_data["objectType"] == "clickstream":
        # Apply clickstream transform
        output = clickstreamTransform.buildObjectRecord(json_data["data"])
      else:
        raise Exception("Invalid input data")
    
      partition_key = record["kinesis"]['partitionKey']

      try:
          put_response = kinesis_client.put_record(
              StreamName=output_stream,
              Data=json.dumps(output),
              PartitionKey=partition_key
          )
          print(f'Successfully sent record {data} to output stream.')
          print(put_response)
      except Exception as e:
        print(f'Failed to send record {data} to output stream. Error: {e}')
    except Exception as ex:
      print(ex)
      rec = str(record)
      print("Error processing data, record is: " + rec)
      #response = queue.send_message(MessageBody = "Record failed to process. Check Cloudwatch logs. Record: {rec}")
      #print("Response ID:" + response.get('MessageId'))