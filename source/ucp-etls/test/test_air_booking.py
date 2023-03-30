import unittest
import json
import boto3
from tah_lib.air_bookingTransform import buildObjectRecord

data_path = 'test_data/air_booking/'
sqsClient = boto3.client('sqs')


def loadTestRecord(data_file, queueUrl=""):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return buildObjectRecord(data, queueUrl)


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class TestAirBooking(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_no_pax_id(self):
        # testing that a unique ID is generated if pax don't have Ids
        booking = loadTestRecord(data_path + 'bookingNoPaxId.json')
        self.assertIsNot(booking["air_booking_recs"][0]["traveller_id"], "")
