# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import unittest
import json
import boto3
from tah_lib.air_bookingTransform import build_object_record

data_path = '../test_data/air_booking/'
sqsClient = boto3.client('sqs')


def loadTestRecord(data_file, queueUrl=""):
    f = open(data_file)
    first_line = f.readline().strip()
    data = json.loads(first_line)
    f.close()
    return build_object_record(data, queueUrl, "test_tx_id")


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class TestAirBooking(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_no_pax_id(self):
        # testing that a unique ID is generated if pax don't have Ids
        booking = loadTestRecord(data_path + 'bookingNoPaxId.jsonl')
        self.assertIsNot(booking["air_booking_recs"][0]["traveller_id"], "")

    def test_transformation_air_booking_with_issue(self):
        print("test_transformation_air_booking_with_issue")
        # testing that a unique ID is generated if pax don't have Ids
        booking = loadTestRecord(data_path + 'airBookingWithIssue1.jsonl')
        print("result: ", booking)
        self.assertIsNot(booking["air_booking_recs"][0]["traveller_id"], "")


class TestAirBookingCustomInputs(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_no_pax_id(self):
        # testing that a unique ID is generated if pax don't have Ids
        booking = loadTestRecord(data_path + 'bookingNoPaxId.jsonl')
        self.assertIsNot(booking["air_booking_recs"][0]["traveller_id"], "")

