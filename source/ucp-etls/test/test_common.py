import unittest
from unittest.mock import MagicMock
from tah_lib.common import getJobPredicates

class TestCommon(unittest.TestCase):
    def test_getJobPredicates_withBookmark(self):
        # set up dynamodb mock
        mock_client = MagicMock()
        mock_response = {'Item': {'bookmark': {'S': '2023/04/12/23'}}}
        mock_client.get_item.return_value = mock_response
        # test getJobPredicates
        predicates = getJobPredicates(mock_client, 'test_table', 'test_obj', 'test_domain')
        pdp = predicates['pdp']
        cpp = predicates['cpp']
        self.assertEqual(pdp, "year>='2023' and month>='04' and day>='12'")
        self.assertEqual(cpp, "year>='2023' and month>='04'")

    def test_getJobPredicates_withoutBookmark(self):
        # set up dynamodb mock
        mock_client = MagicMock()
        mock_response = {}
        mock_client.get_item.return_value = mock_response
        # test getJobPredicates
        predicates = getJobPredicates(mock_client, 'test_table', 'test_obj', 'test_domain')
        pdp = predicates['pdp']
        cpp = predicates['cpp']
        self.assertEqual(pdp, None)
        self.assertEqual(cpp, None)
