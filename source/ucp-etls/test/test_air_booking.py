import unittest
import json
from tah_lib.air_bookingTransform import buildObjectRecord

data_path = 'test_data/air_booking/'


def loadTestRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return buildObjectRecord(data)


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class TestAirBooking(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_success(self):
        for rec in ["data1", "data2"]:
            actual = loadTestRecord(data_path + rec + '.json')
            expected = loadExpectedRecord(data_path + rec + '_expected.json')
            self.assertEqual(actual, expected)
            self.assertIsNot(actual["air_booking_recs"][0]["traveller_id"], "")
            self.assertIsNot(actual["air_booking_recs"]
                             [0]["model_version"], "")
            self.assertIsNot(actual["air_booking_recs"][0]["object_type"], "")
            self.assertIsNot(actual["air_booking_recs"][0]["last_updated"], "")

    def test_transformation_no_pax_id(self):
        # testing that a unique ID is generated if pax don't have Ids
        booking = loadTestRecord(data_path + 'bookingNoPaxId.json')
        self.assertIsNot(booking["air_booking_recs"][0]["traveller_id"], "")
