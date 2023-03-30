import unittest
import json
import random
import boto3
from tah_lib.air_bookingTransform import buildObjectRecord as airBookingTransform
from tah_lib.guest_profileTransform import buildObjectRecord as guestPofileTransform
from tah_lib.hotel_bookingTransform import buildObjectRecord as hotelBookingTransform
from tah_lib.hotel_stayTransform import buildObjectRecord as hotelStayTransform
from tah_lib.pax_profileTransform import buildObjectRecord as paxProfileTransform

objects = [{'path': 'test_data/air_booking/',
            'main_accp_rec': 'air_booking_recs',
            'transform': airBookingTransform
            },
           {'path': 'test_data/guest_profile/',
            'main_accp_rec': 'guest_profile_recs',
            'transform': guestPofileTransform
            },
           {'path': 'test_data/hotel_booking/',
            'main_accp_rec': 'hotel_booking_recs',
            'transform': hotelBookingTransform
            },
           {'path': 'test_data/hotel_stay/',
            'main_accp_rec': 'hotel_stay_revenue_items',
            'transform': hotelStayTransform
            },
           {'path': 'test_data/pax_profile/',
            'main_accp_rec': 'air_profile_recs',
            'transform': paxProfileTransform
            }]

sqsClient = boto3.client('sqs')


def loadTestRecord(data_file, tf, queueUrl=""):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return tf(data, queueUrl)


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
                    obj['path'] + rec + '.json', obj['transform'])
                expected = loadExpectedRecord(
                    obj['path'] + rec + '_expected.json')
                self.assertEqual(actual, expected)
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0]["traveller_id"], "")
                self.assertIsNot(actual[obj['main_accp_rec']]
                                 [0]["model_version"], "")
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0]["object_type"], "")
                self.assertIsNot(
                    actual[obj['main_accp_rec']][0]["last_updated"], "")

    def testInvalidRecord(self):
        qName = 'ucp-transformer-test-queue-' + str(random.randint(0, 10000))
        print("Creating queue ", qName)
        q = sqsClient.create_queue(QueueName=qName)
        qUrl = q["QueueUrl"]
        print("Successfully created ", qUrl)
        for obj in objects:
            loadTestRecord(obj['path'] + 'invalid.json',
                           obj['transform'], qUrl)
            print(obj['path'], " reading from error queue ", qUrl)
            response = sqsClient.receive_message(QueueUrl=qUrl)
            self.assertEqual(len(response["Messages"]), 1)
        print("Deleting queue ", qUrl)
        res = sqsClient.delete_queue(QueueUrl=qUrl)
