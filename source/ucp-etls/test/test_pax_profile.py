import unittest
import json
from pax_profile.pax_profileTransform import buildObjectRecord
from pax_profile.testData.expectedOutput import success, multipleHistoryRecs

business_object = 'pax_profile'
data_path = business_object + '/testData/'

def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()

    actual = buildObjectRecord(data)
    return actual

class TestPaxProfile(unittest.TestCase):
    def test_transformation_success(self):
        actual = test_transformation(data_path + '2476172448.json')
        expected = success
        self.assertEqual(expected, actual)
    def test_transformation_preferred_email(self):
        actual = test_transformation(data_path + '5456783918.json')
        expected = multipleHistoryRecs
        self.assertEqual(expected, actual)
    def test_transformation_missing_field(self):
        actual = test_transformation(data_path + 'missing_id.json')
        expected = "'id'"
        self.assertEqual(str(actual['error']), expected)