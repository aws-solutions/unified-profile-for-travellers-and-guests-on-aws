import unittest
import json
from tah_lib.pax_profileTransform import buildObjectRecord

data_path = 'test_data/pax_profile/'


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


class TestPaxProfile(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_success(self):
        for rec in ["data1", "data2"]:
            actual = loadTestRecord(data_path + rec + '.json')
            expected = loadExpectedRecord(data_path + rec + '_expected.json')
            self.assertEqual(actual, expected)
