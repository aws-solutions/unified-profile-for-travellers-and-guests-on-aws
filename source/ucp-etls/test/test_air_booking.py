import unittest
import json
from air_booking.air_bookingTransform import buildObjectRecord
from air_booking.testData.expectedOutput import pnr_5m71q4
from air_booking.testData.expectedOutput import pnr_pt4ube

business_object = 'air_booking'
data_path = business_object + '/testData/'

def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    
    actual = buildObjectRecord(data)
    return actual

class TestAirBooking(unittest.TestCase):
    maxDiff = None
    def test_transformation_success(self):
        actual = test_transformation(data_path + '5M71Q4.json')
        expected = pnr_5m71q4
        self.assertEqual(actual, expected)
    def test_transformation_missing_field(self):
        # missing modelVersion
        actual = test_transformation(data_path + 'PT4UBE.json')
        expected = pnr_pt4ube
        self.assertEqual(actual, expected)
