# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import unittest
import json
import boto3
from tah_lib.air_bookingTransform import build_object_record as airBookingTransform
from tah_lib.guest_profileTransform import build_object_record as guestPofileTransform
from tah_lib.hotel_bookingTransform import build_object_record as hotelBookingTransform
from tah_lib.hotel_stayTransform import build_object_record as hotelStayTransform
from tah_lib.pax_profileTransform import build_object_record as paxProfileTransform
from tah_lib.clickstreamTransform import build_object_record as clickstreamTransform
from tah_lib.customer_service_interactionTransform import build_object_record as customerServiceInteractionTransform

data_path_hotel_booking = '../test_data/hotel_booking/'
data_path_clickstream = '../test_data/clickstream/'
data_path_guest_profile = '../test_data/guest_profile/'
data_paths = [data_path_hotel_booking, data_path_clickstream, data_path_guest_profile]
transforms = [hotelBookingTransform, clickstreamTransform, guestPofileTransform]
sqsClient = boto3.client('sqs')


def loadTestRecords(data_file, transform, queueUrl=""):
    f = open(data_file)
    #iterate on all lines
    recs = []
    for line in f:
        data = json.loads(line)
        recs.append(transform(data, queueUrl, "test_tx_id"))
    f.close()
    return recs


def loadExpectedRecords(data_file):
    f = open(data_file)
    recs = []
    for line in f:
        data = json.loads(line)
        recs.append(data)
    f.close()
    return recs


class TestAirBooking(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_biz_objects_customer_1(self):
        for i, path in enumerate(data_paths):
            print("[customer1] test_biz_objects: ", path)
            records = loadTestRecords(path + 'customer_1.jsonl', transforms[i])
            expected = loadExpectedRecords(path + 'customer_1_expected.jsonl')
            for i in range(len(records)):
                self.assertDictEqual(records[i], expected[i])
