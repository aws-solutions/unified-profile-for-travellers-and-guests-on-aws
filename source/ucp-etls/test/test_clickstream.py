# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import json
import unittest

from tah_lib.clickstreamTransform import build_object_record

data_path = '../test_data/clickstream/'


def loadTestRecord(data_file):
    f = open(data_file)
    first_line = f.readline().strip()
    data = json.loads(first_line)
    f.close()
    return build_object_record(data, "", "test_tx_id")


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


class TestClickstream(unittest.TestCase):
    unittest.TestCase.maxDiff = None

    def test_transformation_success(self):
        for rec in ["data1", "data2"]:
            actual = loadTestRecord(data_path + rec + '.jsonl')
            expected = loadExpectedRecord(data_path + rec + '_expected.json')
            self.assertEqual(actual, expected)
            self.assertIsNot(actual["clickstream_recs"][0]["traveller_id"], "")
            self.assertIsNot(actual["clickstream_recs"][0]["model_version"], "")
            self.assertIsNot(actual["clickstream_recs"][0]["object_type"], "")
            self.assertIsNot(actual["clickstream_recs"][0]["last_updated"], "")
            self.assertIsNot(actual["clickstream_recs"][0]["accp_object_id"], "")
    
    def test_full(self):
        actual = loadTestRecord(data_path + 'full.jsonl')
        expected = loadExpectedRecord(data_path + 'full_expected.jsonl')
        self.assertEqual(actual, expected)
