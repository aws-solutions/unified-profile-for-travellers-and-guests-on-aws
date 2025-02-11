# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import base64
import json
import unittest
from datetime import datetime
from unittest.mock import MagicMock, patch

from tah_lib.common import (
    ValidationError,
    add_extended_data,
    get_job_predicates,
    send_solution_metrics,
    set_timestamp,
    validate_and_encode_json,
)


class TestCommon(unittest.TestCase):
    def test_get_job_predicates_withBookmark(self):
        # set up dynamodb mock
        mock_client = MagicMock()
        bookmarkDate = "2023/04/12/23"
        mock_response = {"Item": {"bookmark": {"S": bookmarkDate}}}
        mock_client.get_item.return_value = mock_response
        # test get_job_predicates
        predicates = get_job_predicates(
            mock_client, "test_table", "test_obj", "test_domain"
        )
        pdp = predicates["pdp"]
        cpp = predicates["cpp"]
        # same month, same day
        self.assertTrue(testPredicate(pdp, bookmarkDate))
        self.assertTrue(testPredicate(cpp, bookmarkDate))
        # same month, later day
        self.assertTrue(testPredicate(pdp, "2023/04/13/23"))
        self.assertTrue(testPredicate(cpp, "2023/04/13/23"))
        # later month, earlier day
        self.assertTrue(testPredicate(pdp, "2023/05/01/23"))
        self.assertTrue(testPredicate(cpp, "2023/05/01/23"))
        # later year, earlier month
        self.assertTrue(testPredicate(pdp, "2024/01/01/23"))
        self.assertTrue(testPredicate(cpp, "2024/01/01/23"))
        # earlier day
        self.assertFalse(testPredicate(pdp, "2023/04/01/23"))
        self.assertFalse(testPredicate(cpp, "2023/04/01/23"))
        # earlier month
        self.assertFalse(testPredicate(pdp, "2023/01/30/23"))
        self.assertFalse(testPredicate(cpp, "2023/01/30/23"))
        # earlier year
        self.assertFalse(testPredicate(pdp, "2022/012/30/23"))
        self.assertFalse(testPredicate(cpp, "2022/012/30/23"))

    def test_get_job_predicates_withoutBookmark(self):
        # set up dynamodb mock
        mock_client = MagicMock()
        mock_response = {}
        mock_client.get_item.return_value = mock_response
        # test get_job_predicates
        predicates = get_job_predicates(
            mock_client, "test_table", "test_obj", "test_domain"
        )
        pdp = predicates["pdp"]
        cpp = predicates["cpp"]
        self.assertEqual(pdp, None)
        self.assertEqual(cpp, None)

    def test_set_timestamp(self):
        # test get_job_predicates
        now = datetime.now()
        rec = {}
        set_timestamp(rec)
        self.assertEqual(
            rec.get("last_updated").split(".")[0], now.strftime("%Y-%m-%dT%H:%M:%S")
        )
        self.assertEqual(rec["last_updated_partition"], now.strftime("%Y-%m-%d-%H"))

        accp_record = {"last_updated": "2023-04-19T00:44:03.570367Z"}
        set_timestamp(accp_record)
        self.assertEqual(accp_record.get("last_updated"), "2023-04-19T00:44:03.570367Z")
        self.assertEqual(accp_record.get("last_updated_partition"), "2023-04-19-00")

        accp_record = {"last_updated": "2023-04-19T00:44:03.000Z"}
        set_timestamp(accp_record)
        self.assertEqual(accp_record.get("last_updated"), "2023-04-19T00:44:03.000Z")
        self.assertEqual(accp_record.get("last_updated_partition"), "2023-04-19-00")

        # test invalid timestamp - no "Z"
        accp_record = {"last_updated": "2023-04-19T00:44:03.000"}
        with self.assertRaises(ValidationError):
            set_timestamp(accp_record)

        # test invalid timestamp - 9 digits of accuracy (only 6 accepted)
        accp_record = {"last_updated": "2023-04-19T00:44:03.123456789Z"}
        with self.assertRaises(ValidationError):
            set_timestamp(accp_record)

        # test invalid timestamp - 2006-01-02 15:04:05.000
        accp_record = {"last_updated": "2023-04-19 00:44:03.123"}
        with self.assertRaises(ValidationError):
            set_timestamp(accp_record)

    # test for send_solution_metrics(solution_id, solution_version, metrics_uuid, send_anonymized_data, payload)
    @patch("requests.post")
    def test_send_solution_metrics(self, mock_post):
        # Mock response
        mock_status_code = 200
        mock_response = {"status": "success"}
        mock_post.return_value.status_code = mock_status_code
        mock_post.return_value.json.return_value = mock_response
        # Actual request
        solution_id = "solution_id"
        solution_version = "1.0.0"
        metrics_uuid = "metrics_uuid"
        send_anonymized_data = "Yes"
        payload = {"service": "SO0244", "status": "success", "records": 100}
        # Test send_solution_metrics
        allow_logging, error_code = send_solution_metrics(
            solution_id, solution_version, metrics_uuid, send_anonymized_data, payload
        )
        self.assertTrue(allow_logging)
        self.assertEqual(error_code, None)

    def test_validate_and_encode_json_with_valid_string(self):
        """Test encoding a valid JSON string"""
        # Arrange
        test_data = '{"name": "test", "value": 123}'
        expected_json = json.loads(test_data)

        # Act
        result = validate_and_encode_json(test_data)
        decoded_result = json.loads(base64.b64decode(result).decode())

        # Assert
        self.assertEqual(decoded_result, expected_json)

    def test_validate_and_encode_json_with_dict(self):
        """Test encoding a Python dictionary"""
        # Arrange
        test_data = {"name": "test", "nested": {"key": "value"}}

        # Act
        result = validate_and_encode_json(test_data)
        decoded_result = json.loads(base64.b64decode(result).decode())

        # Assert
        self.assertEqual(decoded_result, test_data)

    def test_add_extended_data_with_valid_data(self):
        """Test adding extended data to a record"""
        # Arrange
        record = {}
        accp_obj = {"extendedData": {"field1": "value1", "field2": 123}}

        # Act
        add_extended_data(record, accp_obj)

        # Assert
        self.assertIn("extended_data", record)
        decoded_data = json.loads(base64.b64decode(record["extended_data"]).decode())
        self.assertEqual(decoded_data, accp_obj["extendedData"])

    def test_add_extended_data_with_valid_data_string(self):
        """Test adding extended data to a record"""
        # Arrange
        record = {}
        accp_obj = {"extendedData": '{"field1": "value1", "field2": 123}'}
        non_string_data = {"field1": "value1", "field2": 123}

        # Act
        add_extended_data(record, accp_obj)

        # Assert
        self.assertIn("extended_data", record)
        decoded_data = json.loads(base64.b64decode(record["extended_data"]).decode())
        self.assertEqual(decoded_data, non_string_data)

    def test_add_extended_data_with_oversized_data(self):
        """Test adding extremely large extended data to record throws error"""
        # Arrange
        record = {}
        accp_obj = {
            "extendedData": {
                "field1": "x" * 1000000,  # Very large string
            }
        }

        # Act & Assert
        with self.assertRaises(ValidationError):
            add_extended_data(record, accp_obj)


# HELPER FUNCTIONS


# example predicate pattern
# (year > '2023') OR (year = '2023' AND month > '04') OR (year = '2023' AND month = '04' AND day >= '12')
def testPredicate(predicate, date):
    year, month, day, _ = date.split("/")  # used in eval()
    py_formatted_predicate = (
        predicate.replace("AND", "and").replace("OR", "or").replace(" = ", " == ")
    )
    result = eval(py_formatted_predicate)  # evaluate string using available variables
    return result
