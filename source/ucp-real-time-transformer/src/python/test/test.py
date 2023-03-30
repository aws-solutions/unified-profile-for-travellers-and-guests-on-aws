import unittest
import json
import random
import os
import boto3
from index import handleWithEnv, parseBizObjectToAccpRecords, BUSINESS_OBJECT_CONFIG_MAP

sqsClient = boto3.client('sqs')
TAH_REGION = os.getenv("TAH_REGION")
kinesis_client = boto3.client("kinesis", TAH_REGION)


class Test(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def testErrorQueue(self):
        qName = 'ucp-transformer-test-queue-' + str(random.randint(0, 10000))
        print("Creating queue ", qName)
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        event = {'Records': [{}]}
        handleWithEnv(event, kinesis_client, "dummy_stream", qName)
        response = sqsClient.receive_message(QueueUrl=qUrl)
        self.assertEqual(len(response["Messages"]), 1)
        print("Deleting queue ", qUrl)
        res = sqsClient.delete_queue(QueueUrl=qUrl)

    def testTransformationToAccp(self):
        qName = 'ucp-transformer-test-queue-' + str(random.randint(0, 10000))
        print("Creating queue ", qName)
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        for obj in [{"type": "hotel_booking", "count": 24}, {"type": "clickstream", "count": 1}]:
            f = open("../../../test_data/"+obj["type"]+"/data1.json")
            data = json.load(f)
            f.close()
            bizObj = data
            bizObjectConfig = BUSINESS_OBJECT_CONFIG_MAP[obj["type"]]
            accpRecords = parseBizObjectToAccpRecords(
                bizObj, bizObjectConfig, qUrl)
            self.assertEqual(len(accpRecords), obj["count"])
        res = sqsClient.delete_queue(QueueUrl=qUrl)
