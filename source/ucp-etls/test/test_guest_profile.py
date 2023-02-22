import unittest
import json
from guest_profile.guest_profileTransform import buildObjectRecord
from guest_profile.testData.expectedOutput import happy_path

business_object = 'guest_profile'
data_path = business_object + '/testData/'


def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    
    actual = buildObjectRecord(data)
    return actual


class TestGuest_Profile(unittest.TestCase):
    def test_transformation_success(self):
        actual = test_transformation(data_path + 'rawData.json')
        expected = happy_path
        self.assertEqual(actual, expected)

    def test_transformation_missing_field(self):
        # missing modelVersion
        actual = test_transformation(data_path + 'missingData.json')
        expected = "'id'"
        self.assertEqual(actual['error'], expected)
