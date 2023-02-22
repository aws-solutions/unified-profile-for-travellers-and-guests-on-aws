import unittest
import json
from hotel_stay.hotel_stayTransform import buildObjectRecord

business_object = 'hotel_stay'
data_path = business_object + '/testData/'

def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()

    actual = buildObjectRecord(data)
    return actual

class TestHotelStay(unittest.TestCase):
    def test_transformation_success(self):
        actual = test_transformation(data_path + 'rawData.json')
        print(actual)
        self.assertIsNotNone(actual['data'])
