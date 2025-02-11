# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import unittest
import json
import random
import boto3
from tah_lib.air_bookingTransform import build_object_record as airBookingTransform
from tah_lib.guest_profileTransform import build_object_record as guestPofileTransform
from tah_lib.hotel_bookingTransform import build_object_record as hotelBookingTransform
from tah_lib.hotel_stayTransform import build_object_record as hotelStayTransform
from tah_lib.pax_profileTransform import build_object_record as paxProfileTransform
from tah_lib.customer_service_interactionTransform import build_object_record as customerServiceInteractionTransform

objects = [{'path': '../test_data/air_booking/',
            'main_accp_rec': 'air_booking_recs',
            'transform': airBookingTransform
            },
           {'path': '../test_data/guest_profile/',
            'main_accp_rec': 'guest_profile_recs',
            'transform': guestPofileTransform
            },
           {'path': '../test_data/hotel_booking/',
            'main_accp_rec': 'hotel_booking_recs',
            'transform': hotelBookingTransform
            },
           {'path': '../test_data/hotel_stay/',
            'main_accp_rec': 'hotel_stay_revenue_items',
            'transform': hotelStayTransform
            },
           {'path': '../test_data/pax_profile/',
            'main_accp_rec': 'air_profile_recs',
            'transform': paxProfileTransform
            },
           {'path': '../test_data/customer_service_interaction/',
            'main_accp_rec': 'conversation_recs',
            'transform': customerServiceInteractionTransform
            }]

sqsClient = boto3.client('sqs')


def loadTestRecord(data_file, tf, queueUrl=""):
    f = open(data_file)
    first_line = f.readline().strip()
    try:
        data = json.loads(first_line)
    except Exception as e:
        raise Exception("Invalid Business object in file", data_file, ". Error: ", e)
    f.close()
    return tf(data, queueUrl, "test_tx_id")


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class Test(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_success(self):
        for obj in objects:
            for rec in ["data1", "data2"]:
                actual = loadTestRecord(
                    obj['path'] + rec + '.jsonl', obj['transform'])
                expected = loadExpectedRecord(
                    obj['path'] + rec + '_expected.json')
                self.assertEqual(actual, expected)
                self.assertGreaterEqual(
                    len(actual[obj['main_accp_rec']]), 1, "should have at least 1 object of type " + obj['main_accp_rec'])
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0].get("traveller_id", ""), "", "Missing traveller_id for " + obj['path'])
                self.assertIsNot(actual[obj['main_accp_rec']]
                                 [0].get("model_version", ""), "", "Missing model_version for " + obj['path'])
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0].get("object_type", ""), "", "Missing object_type for " + obj['path'])
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0].get("last_updated", ""), "", "Missing last_updated for " + obj['path'])
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0].get("accp_object_id", ""), "", "Missing accp_object_id for " + obj['path'])

    def testInvalidRecord(self):
        qName = 'ucp-transformer-test-queue-' + str(random.randint(0, 10000))
        print("Creating queue ", qName)
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        print("Successfully created ", qUrl)
        for obj in objects:
            loadTestRecord(obj['path'] + 'invalid.jsonl',
                           obj['transform'], qUrl)
            print(obj['path'], " reading from error queue ", qUrl)
            response = sqsClient.receive_message(QueueUrl=qUrl)
            self.assertEqual(len(response["Messages"]), 1)
        print("Deleting queue ", qUrl)
        res = sqsClient.delete_queue(QueueUrl=qUrl)
