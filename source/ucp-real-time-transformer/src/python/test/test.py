# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import os
import random
import time
import unittest
import uuid
from datetime import datetime

import boto3
from index import (BUSINESS_OBJECT_CONFIG_MAP, backup_in_s3,
                   biz_object_to_accp_records, build_partition_key,
                   build_s3_key, decode_raw_data, generate_tx_id,
                   handle_with_env, wrap_accp_record)

sqsClient = boto3.client('sqs')
TAH_REGION = os.getenv("TAH_REGION")
if TAH_REGION == "":
    # getting region for codeBuild project
    TAH_REGION = os.getenv("AWS_REGION")

kinesis_client = boto3.client("kinesis", TAH_REGION)
s3 = boto3.client("s3", TAH_REGION)
firehose = boto3.client("firehose", TAH_REGION)
iam = boto3.client("iam", TAH_REGION)

# {  "objectType": "air_booking", "modelVersion": "1.0", "domain": "test_domain",   "timestamp": "2029-10-02T08:07:00.000Z",  "uid": "abcd", "data": {"objectVersion":11,"modelVersion":"1.0","id":"3EM4A0","externalIds":null,"lastUpdatedOn":"2021-06-26T03:27:47.008394Z","createdOn":"2021-05-26T03:27:47.008394Z","lastUpdatedBy":"Melba Swift","createdBy":"Joan Homenick","creationChannelId":"ota-anycompany","price":{"priceItems":[{"segmenRank":0,"passengerIndex":0,"fareClass":"F","price":{"total":4533.8896,"base":4440.92,"fees":[{"amount":16.5,"type":"PASSENGER FACILITY CHARGE (UNITED STATES)"},{"amount":17.2,"type":"SEGMENT TAX (UNITED STATES)"},{"amount":48.07,"type":"TRANSPORTATION TAX (UNITED STATES)"},{"amount":11.2,"type":"US SECURITY FEE (UNITED STATES)"}]}}],"grandTotal":4533.8896},"itinerary":{"from":"LAS","to":"CDG","departureDate":"2021-09-02","departureTime":"11:16","arrivalDate":"2021-09-03","arrivalTime":"00:09","duration":"12h53m0s","segments":[{"rank":0,"from":"LAS","to":"CDG","departureDate":"2021-09-02","departureTime":"11:16","arrivalDate":"2021-09-03","arrivalTime":"00:09","flightNumber":"LF8387","inventory":{"totalSeats":0,"inventoryByFareClass":null},"status":""}],"inventory":{"totalSeats":0,"inventoryByFareClass":null},"status":"canceled"},"return":{"from":"CDG","to":"LAS","departureDate":"2021-08-17","departureTime":"07:31","arrivalDate":"2021-08-17","arrivalTime":"19:12","duration":"11h41m0s","segments":[{"rank":0,"from":"CDG","to":"LAS","departureDate":"2021-08-17","departureTime":"07:31","arrivalDate":"2021-08-17","arrivalTime":"19:12","flightNumber":"LF5576","inventory":{"totalSeats":0,"inventoryByFareClass":null},"status":""}],"inventory":{"totalSeats":0,"inventoryByFareClass":null},"status":""},"passengerInfo":{"passengers":[{"modelVersion":"1.0","id":"9653778983","lastUpdatedOn":"2022-01-15T19:03:08.700006Z","createdOn":"2021-10-15T19:03:08.700006Z","lastUpdatedBy":"Kristina Wehner","createdBy":"George Weimann","isBooker":true,"emails":[],"phones":[{"type":"mobile","number":"7510661749","primary":true,"countryCode":12},{"type":"mobile","number":"(581)886-5851","primary":false,"countryCode":33},{"type":"mobile","number":"3081909371","primary":false,"countryCode":91}],"addresses":[],"honorific":"Lord","firstName":"Cheyanne","middleName":"Turner","lastName":"McKenzie","gender":"male","pronoun":"they","dateOfBirth":"1988-01-16","language":{"code":"kn","name":""},"nationality":{"code":"SD","name":""},"jobTitle":"Specialist","parentCompany":"Russell Investments","loyaltyPrograms":[],"identityProofs":null}]},"paymentInformation":{"paymentType":"bank_account","ccInfo":{"token":"","cardType":"","cardExp":"","cardCvv":"","expiration":"","name":"","address":{"type":"","line1":"","line2":"","line3":"","line4":"","city":"","state":{"code":"","name":""},"province":{"code":"","name":""},"postalCode":"","country":{"code":"","name":""},"primary":false}},"routingNumber":"390545393","accountNumber":"828416127497","voucherID":"","address":{"type":"business","line1":"343 Avenue port","line2":"more content line 2","line3":"more content line 3","line4":"more content line 4","city":"Phoenix","state":{"code":"GA","name":"Indiana"},"province":{"code":"","name":""},"postalCode":"77699","country":{"code":"VU","name":""},"primary":false}},"promocode":"IKUE14Z24M","email":"","phone":"","status":"confirmed"}}
TEST_RECORD = "eyAgIm9iamVjdFR5cGUiOiAiYWlyX2Jvb2tpbmciLCAibW9kZWxWZXJzaW9uIjogIjEuMCIsICJkb21haW4iOiAidGVzdF9kb21haW4iLCAgICJ0aW1lc3RhbXAiOiAiMjAyOS0xMC0wMlQwODowNzowMC4wMDBaIiwgICJ1aWQiOiAiYWJjZCIsICJkYXRhIjogeyJvYmplY3RWZXJzaW9uIjoxMSwibW9kZWxWZXJzaW9uIjoiMS4wIiwiaWQiOiIzRU00QTAiLCJleHRlcm5hbElkcyI6bnVsbCwibGFzdFVwZGF0ZWRPbiI6IjIwMjEtMDYtMjZUMDM6Mjc6NDcuMDA4Mzk0WiIsImNyZWF0ZWRPbiI6IjIwMjEtMDUtMjZUMDM6Mjc6NDcuMDA4Mzk0WiIsImxhc3RVcGRhdGVkQnkiOiJNZWxiYSBTd2lmdCIsImNyZWF0ZWRCeSI6IkpvYW4gSG9tZW5pY2siLCJjcmVhdGlvbkNoYW5uZWxJZCI6Im90YS1leHBlZGlhIiwicHJpY2UiOnsicHJpY2VJdGVtcyI6W3sic2VnbWVuUmFuayI6MCwicGFzc2VuZ2VySW5kZXgiOjAsImZhcmVDbGFzcyI6IkYiLCJwcmljZSI6eyJ0b3RhbCI6NDUzMy44ODk2LCJiYXNlIjo0NDQwLjkyLCJmZWVzIjpbeyJhbW91bnQiOjE2LjUsInR5cGUiOiJQQVNTRU5HRVIgRkFDSUxJVFkgQ0hBUkdFIChVTklURUQgU1RBVEVTKSJ9LHsiYW1vdW50IjoxNy4yLCJ0eXBlIjoiU0VHTUVOVCBUQVggKFVOSVRFRCBTVEFURVMpIn0seyJhbW91bnQiOjQ4LjA3LCJ0eXBlIjoiVFJBTlNQT1JUQVRJT04gVEFYIChVTklURUQgU1RBVEVTKSJ9LHsiYW1vdW50IjoxMS4yLCJ0eXBlIjoiVVMgU0VDVVJJVFkgRkVFIChVTklURUQgU1RBVEVTKSJ9XX19XSwiZ3JhbmRUb3RhbCI6NDUzMy44ODk2fSwiaXRpbmVyYXJ5Ijp7ImZyb20iOiJMQVMiLCJ0byI6IkNERyIsImRlcGFydHVyZURhdGUiOiIyMDIxLTA5LTAyIiwiZGVwYXJ0dXJlVGltZSI6IjExOjE2IiwiYXJyaXZhbERhdGUiOiIyMDIxLTA5LTAzIiwiYXJyaXZhbFRpbWUiOiIwMDowOSIsImR1cmF0aW9uIjoiMTJoNTNtMHMiLCJzZWdtZW50cyI6W3sicmFuayI6MCwiZnJvbSI6IkxBUyIsInRvIjoiQ0RHIiwiZGVwYXJ0dXJlRGF0ZSI6IjIwMjEtMDktMDIiLCJkZXBhcnR1cmVUaW1lIjoiMTE6MTYiLCJhcnJpdmFsRGF0ZSI6IjIwMjEtMDktMDMiLCJhcnJpdmFsVGltZSI6IjAwOjA5IiwiZmxpZ2h0TnVtYmVyIjoiTEY4Mzg3IiwiaW52ZW50b3J5Ijp7InRvdGFsU2VhdHMiOjAsImludmVudG9yeUJ5RmFyZUNsYXNzIjpudWxsfSwic3RhdHVzIjoiIn1dLCJpbnZlbnRvcnkiOnsidG90YWxTZWF0cyI6MCwiaW52ZW50b3J5QnlGYXJlQ2xhc3MiOm51bGx9LCJzdGF0dXMiOiJjYW5jZWxlZCJ9LCJyZXR1cm4iOnsiZnJvbSI6IkNERyIsInRvIjoiTEFTIiwiZGVwYXJ0dXJlRGF0ZSI6IjIwMjEtMDgtMTciLCJkZXBhcnR1cmVUaW1lIjoiMDc6MzEiLCJhcnJpdmFsRGF0ZSI6IjIwMjEtMDgtMTciLCJhcnJpdmFsVGltZSI6IjE5OjEyIiwiZHVyYXRpb24iOiIxMWg0MW0wcyIsInNlZ21lbnRzIjpbeyJyYW5rIjowLCJmcm9tIjoiQ0RHIiwidG8iOiJMQVMiLCJkZXBhcnR1cmVEYXRlIjoiMjAyMS0wOC0xNyIsImRlcGFydHVyZVRpbWUiOiIwNzozMSIsImFycml2YWxEYXRlIjoiMjAyMS0wOC0xNyIsImFycml2YWxUaW1lIjoiMTk6MTIiLCJmbGlnaHROdW1iZXIiOiJMRjU1NzYiLCJpbnZlbnRvcnkiOnsidG90YWxTZWF0cyI6MCwiaW52ZW50b3J5QnlGYXJlQ2xhc3MiOm51bGx9LCJzdGF0dXMiOiIifV0sImludmVudG9yeSI6eyJ0b3RhbFNlYXRzIjowLCJpbnZlbnRvcnlCeUZhcmVDbGFzcyI6bnVsbH0sInN0YXR1cyI6IiJ9LCJwYXNzZW5nZXJJbmZvIjp7InBhc3NlbmdlcnMiOlt7Im1vZGVsVmVyc2lvbiI6IjEuMCIsImlkIjoiOTY1Mzc3ODk4MyIsImxhc3RVcGRhdGVkT24iOiIyMDIyLTAxLTE1VDE5OjAzOjA4LjcwMDAwNloiLCJjcmVhdGVkT24iOiIyMDIxLTEwLTE1VDE5OjAzOjA4LjcwMDAwNloiLCJsYXN0VXBkYXRlZEJ5IjoiS3Jpc3RpbmEgV2VobmVyIiwiY3JlYXRlZEJ5IjoiR2VvcmdlIFdlaW1hbm4iLCJpc0Jvb2tlciI6dHJ1ZSwiZW1haWxzIjpbXSwicGhvbmVzIjpbeyJ0eXBlIjoibW9iaWxlIiwibnVtYmVyIjoiNzUxMDY2MTc0OSIsInByaW1hcnkiOnRydWUsImNvdW50cnlDb2RlIjoxMn0seyJ0eXBlIjoibW9iaWxlIiwibnVtYmVyIjoiKDU4MSk4ODYtNTg1MSIsInByaW1hcnkiOmZhbHNlLCJjb3VudHJ5Q29kZSI6MzN9LHsidHlwZSI6Im1vYmlsZSIsIm51bWJlciI6IjMwODE5MDkzNzEiLCJwcmltYXJ5IjpmYWxzZSwiY291bnRyeUNvZGUiOjkxfV0sImFkZHJlc3NlcyI6W10sImhvbm9yaWZpYyI6IkxvcmQiLCJmaXJzdE5hbWUiOiJDaGV5YW5uZSIsIm1pZGRsZU5hbWUiOiJUdXJuZXIiLCJsYXN0TmFtZSI6Ik1jS2VuemllIiwiZ2VuZGVyIjoibWFsZSIsInByb25vdW4iOiJ0aGV5IiwiZGF0ZU9mQmlydGgiOiIxOTg4LTAxLTE2IiwibGFuZ3VhZ2UiOnsiY29kZSI6ImtuIiwibmFtZSI6IiJ9LCJuYXRpb25hbGl0eSI6eyJjb2RlIjoiU0QiLCJuYW1lIjoiIn0sImpvYlRpdGxlIjoiU3BlY2lhbGlzdCIsInBhcmVudENvbXBhbnkiOiJSdXNzZWxsIEludmVzdG1lbnRzIiwibG95YWx0eVByb2dyYW1zIjpbXSwiaWRlbnRpdHlQcm9vZnMiOm51bGx9XX0sInBheW1lbnRJbmZvcm1hdGlvbiI6eyJwYXltZW50VHlwZSI6ImJhbmtfYWNjb3VudCIsImNjSW5mbyI6eyJ0b2tlbiI6IiIsImNhcmRUeXBlIjoiIiwiY2FyZEV4cCI6IiIsImNhcmRDdnYiOiIiLCJleHBpcmF0aW9uIjoiIiwibmFtZSI6IiIsImFkZHJlc3MiOnsidHlwZSI6IiIsImxpbmUxIjoiIiwibGluZTIiOiIiLCJsaW5lMyI6IiIsImxpbmU0IjoiIiwiY2l0eSI6IiIsInN0YXRlIjp7ImNvZGUiOiIiLCJuYW1lIjoiIn0sInByb3ZpbmNlIjp7ImNvZGUiOiIiLCJuYW1lIjoiIn0sInBvc3RhbENvZGUiOiIiLCJjb3VudHJ5Ijp7ImNvZGUiOiIiLCJuYW1lIjoiIn0sInByaW1hcnkiOmZhbHNlfX0sInJvdXRpbmdOdW1iZXIiOiIzOTA1NDUzOTMiLCJhY2NvdW50TnVtYmVyIjoiODI4NDE2MTI3NDk3Iiwidm91Y2hlcklEIjoiIiwiYWRkcmVzcyI6eyJ0eXBlIjoiYnVzaW5lc3MiLCJsaW5lMSI6IjM0MyBBdmVudWUgcG9ydCIsImxpbmUyIjoibW9yZSBjb250ZW50IGxpbmUgMiIsImxpbmUzIjoibW9yZSBjb250ZW50IGxpbmUgMyIsImxpbmU0IjoibW9yZSBjb250ZW50IGxpbmUgNCIsImNpdHkiOiJQaG9lbml4Iiwic3RhdGUiOnsiY29kZSI6IkdBIiwibmFtZSI6IkluZGlhbmEifSwicHJvdmluY2UiOnsiY29kZSI6IiIsIm5hbWUiOiIifSwicG9zdGFsQ29kZSI6Ijc3Njk5IiwiY291bnRyeSI6eyJjb2RlIjoiVlUiLCJuYW1lIjoiIn0sInByaW1hcnkiOmZhbHNlfX0sInByb21vY29kZSI6IklLVUUxNFoyNE0iLCJlbWFpbCI6IiIsInBob25lIjoiIiwic3RhdHVzIjoiY29uZmlybWVkIn19"


class Test(unittest.TestCase):
    unittest.TestCase.maxDiff = None
    os.environ['METRICS_SOLUTION_ID'] = 'SO0244'
    os.environ['METRICS_SOLUTION_VERSION'] = '0.0.0'

    def testTransformationToAccp(self):
        qName = 'ucp-transformer-test-queue-' + str(random.randint(0, 10000))
        print("Creating queue ", qName)
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        failures = []
        for obj in [
            {"type": "hotel_booking", "count": {'total': 8, 'hotel_booking': 6, 'alternate_profile_id': 2}},
            {"type": "air_booking", "count": {'total': 35, 'air_booking': 9, 'email_history': 3, 'phone_history': 6, 'air_loyalty': 3, 'ancillary_service': 10, 'alternate_profile_id': 4}},
            {"type": "hotel_stay", "count": {'total': 8, 'hotel_stay_revenue_items': 6, 'alternate_profile_id': 2}},
            {"type": "clickstream", "count": {'total': 5, 'clickstream': 1, 'alternate_profile_id': 4}},
            {"type": "guest_profile", "count": {'total': 6, 'guest_profile': 1, 'email_history': 1, 'alternate_profile_id': 4}},
            {"type": "pax_profile", "count": {'total': 3, 'pax_profile': 1, 'email_history': 1, 'phone_history': 1}},
            {"type": "customer_service_interaction", "count": {'total': 1, 'customer_service_interaction': 1}}
            ]:

            f = open("../../../test_data/"+obj["type"]+"/data1.jsonl")
            first_line = f.readline().strip()
            data = json.loads(first_line)
            f.close()
            bizObj = data
            accpRecords = biz_object_to_accp_records("test_tx",
                                                     bizObj, obj["type"], qUrl)
            # count accp records by object type
            accpCount = {"total": len(accpRecords)}
            for accpRecord in accpRecords:
                value = accpRecord["object_type"]
                if value in accpCount:
                    accpCount[value] += 1
                else:
                    accpCount[value] = 1
            print("Object type ", obj["type"])
            print("Accp count ", accpCount)
            print("expected count ", obj["count"])
            # iterate on object keys and compare values
            for key in obj["count"]:
                ct = accpCount[key]
                exp = obj["count"][key]
                if ct != exp:
                    failures.append(f"Transformation to ACCP failed for object type {obj['type']}: {key} should be {exp}, but was {ct}")

            response = sqsClient.receive_message(QueueUrl=qUrl)
            if "Messages" in response:
                failures.append(f"Expected 0 messages in queue, but found {len(response['Messages'])}")
        res = sqsClient.delete_queue(QueueUrl=qUrl)
        if failures:
            self.fail("\n".join(failures))

    def testRun(self):
        qName = 'ucp-transformer-test-queue-1-' + str(uuid.uuid4())
        bName = 'ucp-backup-bucket-' + str(uuid.uuid4())
        sName = 'ucp-test-stream-' + str(uuid.uuid4())
        dsName = 'ucp-test-firehose-' + str(uuid.uuid4())
        domain="test_domain"

        print("I.Creating queue, S3 bucket and Firehose stream", qName)
        if TAH_REGION == "us-east-1":
            bucket = s3.create_bucket(Bucket=bName)
        else:
            bucket = s3.create_bucket(Bucket=bName, CreateBucketConfiguration={'LocationConstraint': TAH_REGION})
        
        bucket_arn="arn:aws:s3:::" + bName
        s3.put_bucket_policy(
            Bucket=bName,
            Policy='{"Version": "2012-10-17", "Statement": [{ "Sid": "id-1","Effect": "Allow","Principal": {"Service": "firehose.amazonaws.com"}, "Action": ["s3:PutObject"], "Resource":["'+bucket_arn+'","'+bucket_arn+'/*"]}]}'
            )

        print("I.1 create sqs queue")
        q = sqsClient.create_queue(QueueName=qName)
        print("I.2 create kinesis stream")
        kinesis_client.create_stream(StreamName=sName, ShardCount=1)
        print("I.3create firehose role")
        role_create_res = create_firehose_role(bucket_arn)
        role_arn = role_create_res["roleArn"]
        policy_arn = role_create_res["policyArn"]
        role_name = role_create_res["roleName"]
        
        print("I.4 Waiting 10 s for firehose role to propagate")
        time.sleep(10)
        S3Config={
                'RoleARN': role_arn,
                'BucketARN': bucket_arn,
                'DynamicPartitioningConfiguration': {
                    'Enabled': True,
                },
                'Prefix':  'data/domainname=!{partitionKeyFromQuery:domainname}/',
                'ErrorOutputPrefix': 'error/!{firehose:error-output-type}/',
                'BufferingHints': {
                    'IntervalInSeconds': 60
                },
                'ProcessingConfiguration': {
                    'Enabled': True,
                    'Processors': [
                        {
                            'Type': 'MetadataExtraction',
                            'Parameters': [
                                {
                                    'ParameterName': 'MetadataExtractionQuery',
                                    'ParameterValue': '{domainname: .domain}'
                                },
                                {
                                    'ParameterName': 'JsonParsingEngine',
                                    'ParameterValue': 'JQ-1.6'
                                },
                            ]
                        },
                        {
                            'Type': 'AppendDelimiterToRecord',
                            'Parameters': [
                                {
                                    'ParameterName': 'Delimiter',
                                    'ParameterValue': '\\n'
                                },
                            ]
                        },
                    ],
                },
            }
        print("Firehose S3 Config:", S3Config)
        print("I.5 Create delivery stream")
        firehose.create_delivery_stream(DeliveryStreamName=dsName,
            DeliveryStreamEncryptionConfigurationInput={
             'KeyType': 'AWS_OWNED_CMK'
            },
            ExtendedS3DestinationConfiguration=S3Config)
        print("I.6 Waiting for both streams creation")
        isCreated = False
        firehoIsCreated = False
        while (not isCreated or not firehoIsCreated):
            res = kinesis_client.describe_stream(StreamName=sName)
            print(res)
            isCreated = (res['StreamDescription']['StreamStatus'] == 'ACTIVE')
            res = firehose.describe_delivery_stream(DeliveryStreamName=dsName)
            print(res)
            firehoIsCreated = (res['DeliveryStreamDescription']['DeliveryStreamStatus'] == 'ACTIVE')
            time.sleep(5)

        print("II. Sending data")
        qUrl = q["QueueUrl"]
        event = {'Records': [{"kinesis": {"data": TEST_RECORD}}]}
        handle_with_env(event, kinesis_client,firehose, sName, qName, dsName, 'true')
        print("II.1 Checking SQS queue")
        response = sqsClient.receive_message(QueueUrl=qUrl)
        print("SQS errorrs: ", response)
        self.assertEqual("Messages" not in response, True)

        print("II.3 Checking S3 Backup")
        attemps=0
        s3Res = s3.list_objects_v2(Bucket=bName)
        while s3Res["KeyCount"] == 0 and attemps < 60:
            print("Waiting for backup data (10s poll), attemp: ", attemps)
            s3Res = s3.list_objects_v2(Bucket=bName)
            print("s3 response:", s3Res)
            attemps+=1
            time.sleep(10)
        
        self.assertEqual(s3Res["KeyCount"], 1)
        self.assertEqual(s3Res["Contents"][0]["Key"].startswith("data/domainname=" + domain + "/"), 1)

        print("III. Cleaning up")
        print("III.1 Deleting queue ", qUrl)
        sqsClient.delete_queue(QueueUrl=qUrl)
        print("III.2 Deleting bucket ", bName)
        for key in s3Res["Contents"]:
            s3.delete_object(Bucket=bName, Key=key["Key"])
        s3.delete_bucket(Bucket=bName)
        print("III.3 Deleting kinesis stream")
        kinesis_client.delete_stream(StreamName=sName)
        print("III.3 Deleting firehose stream role")
        delete_firehose_role(role_name, policy_arn)
        print("III.4 Deleting firehose stream")
        firehose.delete_delivery_stream(DeliveryStreamName=dsName)

    def testErrorQueue(self):
        qName = 'ucp-transformer-test-queue-2-' + str(random.randint(0, 10000))
        bName = 'ucp-backup-bucket-' + str(random.randint(0, 10000))
        print("Creating queue and bucket", qName)
        if TAH_REGION == "us-east-1":
            bucket = s3.create_bucket(Bucket=bName)
        else:
            bucket = s3.create_bucket(Bucket=bName, CreateBucketConfiguration={
                                  'LocationConstraint': TAH_REGION})
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        event = {'Records': [{"kinesis": {"key": "invalid message format"}}]}
        handle_with_env(event, kinesis_client,firehose,"dummy_stream", qName, "", "true")
        #handle_with_env(event, kinesis_client, s3,"dummy_stream", qName, bName)
        response = sqsClient.receive_message(QueueUrl=qUrl)
        self.assertEqual(len(response["Messages"]), 1)

        print("Deleting queue ", qUrl)
        sqsClient.delete_queue(QueueUrl=qUrl)
        print("Deleting bucket ", bName)
        s3.delete_bucket(Bucket=bName)

    def testUniqueID(self):
        txid = generate_tx_id()
        self.assertEqual(len(txid), 36)
    
    def testBuildPartitionKey(self):
        accp_record1 = {"key": "value", }
        accp_record2 = {"key": "value", "traveller_id": "123456789"}
        self.assertEqual(build_partition_key("txid", accp_record2), "123456789")
        self.assertEqual(len(build_partition_key("txid",accp_record1)), 36)

    def testBuildS3Key(self):
        now = datetime.now()
        key = build_s3_key(now, "123456789")
        # Extract year, month, day, and hour from the current date and time
        year = str(now.year)
        month = str(now.month).zfill(2)  # Zero padding if single digit
        day = str(now.day).zfill(2)
        hour = str(now.hour).zfill(2)
        formatted_string = f"{year}/{month}/{day}/{hour}"
        self.assertEqual(key, formatted_string + "/123456789.json")

    def testdecode_raw_data(self):
        raw_data = TEST_RECORD
        json_data = decode_raw_data("abcd", raw_data)
        print(json_data)
        self.assertEqual(json_data.get("objectType"), "air_booking")
        self.assertEqual(json_data.get("modelVersion"), "1.0")
        self.assertEqual(json_data.get("domain"), "test_domain")
        self.assertEqual(json_data.get("uid"), "abcd")
        self.assertEqual(json_data.get("timestamp"),
                         "2029-10-02T08:07:00.000Z")
    
    def test_wrap_accp_record(self):
        json_data = {
            "objectType": "test",
            "modelVersion": "1.0",
            "domain": "test_domain",
        }
        accp_record = {"key": "value"}
        res = wrap_accp_record("abcd", accp_record, json_data)
        self.assertEqual(res.get("transactionId"), "abcd")
        self.assertEqual(res.get("modelVersion"), "1.0")
        self.assertEqual(res.get("domain"), "test_domain")
        self.assertEqual(res.get("data"), [accp_record])


def create_firehose_role(bucketArn):
    iam = boto3.client("iam")
    assume_role_policy_document = json.dumps({
        "Version": "2012-10-17",
        "Statement": [
            {
            "Effect": "Allow",
            "Principal": {
                "Service": "firehose.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
            }
        ]
    })
    role_name="ucpRealTimeFirehoseTestRole" + str(random.randint(0, 10000))
    role_create_res = iam.create_role(
        RoleName = role_name,
        AssumeRolePolicyDocument = assume_role_policy_document
    )
    print("Role: ",role_create_res)
    policy_document = {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": ["s3:PutObject",
                    "s3:PutObjectAcl",
                    "s3:ListBucket",
                ],
                "Resource": [bucketArn,bucketArn+"/*"]
            }
        ]
    }
    policy_create_res = iam.create_policy(
        PolicyName='ucpRealTimeFirehoseTestPolicy' + str(random.randint(0, 10000)),
        PolicyDocument=json.dumps(policy_document)
    )
    attach_res = iam.attach_role_policy(
        RoleName=role_create_res["Role"]["RoleName"],
        PolicyArn=policy_create_res["Policy"]["Arn"]
    )

    return {"roleArn": role_create_res["Role"]["Arn"], "roleName":role_name, "policyArn": policy_create_res["Policy"]["Arn"]}


def delete_firehose_role(role_name, policy_arn):
    iam = boto3.client("iam")

    iam.detach_role_policy(
        RoleName=role_name,
        PolicyArn=policy_arn
    )
    iam.delete_policy(
        PolicyArn = policy_arn,
    )

    iam.delete_role(
        RoleName = role_name,
    )