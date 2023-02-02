import unittest
import json
from clickstreamTransform import buildObjectRecord
from expectedOutput import happy_path

def test_transformation(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    
    actual = buildObjectRecord(data)
    return actual

class TestClickstream(unittest.TestCase):
    def test_transformation_success(self):
        actual = test_transformation('rawData.json')
        expected = happy_path
        self.assertEqual(actual, expected)
    def test_transformation_missing_field(self):
        # missing modelVersion
        actual = test_transformation('missingData.json')
        expected = "'modelVersion'"
        self.assertEqual(actual['error'], expected)