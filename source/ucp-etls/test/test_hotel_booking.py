import unittest
import json
from hotel_booking.hotel_bookingTransform import buildObjectRecord
from hotel_booking.testData.expectedOutput import happy_path

business_object = 'hotel_booking'
data_path = business_object + '/testData/'

def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    
    actual = buildObjectRecord(data)
    return actual

class TestHotel_Booking(unittest.TestCase):
    def test_transformation_success(self):
        actual = test_transformation(data_path + 'rawData.json')
        expected = happy_path
        self.assertEqual(actual, expected)
    def test_transformation_missing_field(self):
        # missing modelVersion
        actual = test_transformation(data_path + 'missingData.json')
        expected = "'id'"
        self.assertEqual(actual['error'], expected)
