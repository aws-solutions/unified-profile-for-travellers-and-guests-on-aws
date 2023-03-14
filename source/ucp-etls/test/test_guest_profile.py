import unittest
import json
from tah_lib.guest_profileTransform import buildObjectRecord

data_path = 'test_data/guest_profile/'


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


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class TestGuest_Profile(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_success(self):
        for rec in ["data1", "data2"]:
            actual = loadTestRecord(data_path + rec + '.json')
            expected = loadExpectedRecord(data_path + rec + '_expected.json')
            self.assertEqual(actual, expected)
            self.assertIsNot(actual["guest_profile_recs"]
                             [0]["traveller_id"], "")
            self.assertIsNot(actual["guest_profile_recs"]
                             [0]["model_version"], "")
            self.assertIsNot(actual["guest_profile_recs"]
                             [0]["object_type"], "")
            self.assertIsNot(actual["guest_profile_recs"]
                             [0]["last_updated"], "")
