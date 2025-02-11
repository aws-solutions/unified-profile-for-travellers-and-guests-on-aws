# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import copy
import json
import os
import random
import threading
import unittest
from logging import Logger, basicConfig, getLogger
from time import time
# Optional in test code; required in prod code
from typing import TYPE_CHECKING, Any

import awswrangler as wr
import boto3
from pandas import DataFrame
from ucp_s3_excise_queue_processor.index import handler_with_env
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexClient import \
    DynamoDbMutexClient
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexConfig import \
    DynamoDbMutexConfig

if TYPE_CHECKING:
    from mypy_boto3_dynamodb import TableExistsWaiter
    from mypy_boto3_dynamodb.client import DynamoDBClient
    from mypy_boto3_dynamodb.type_defs import CreateTableOutputTypeDef
    from mypy_boto3_glue.client import GlueClient
    from mypy_boto3_s3.client import S3Client
    from mypy_boto3_s3.service_resource import Object, S3ServiceResource
else:
    DynamoDBClient = object
    GlueClient = object
    S3Client = object
    S3ServiceResource = object
    CreateTableOutputTypeDef = type
    TableExistsWaiter = object
    Object = object

TAH_REGION: str | None = os.getenv("TAH_REGION")
if TAH_REGION == None or TAH_REGION == "":
    # Get Region for CodeBuild
    TAH_REGION = os.getenv("AWS_REGION")
TEST_DATA_FOLDER: str | None = os.getenv("TEST_DATA_FOLDER")
GLUE_SCHEMA_PATH_ASYNC: str | None = os.getenv("GLUE_SCHEMA_PATH_ASYNC")

s3Client: S3Client = boto3.client('s3', region_name=TAH_REGION)
s3Resource: S3ServiceResource = boto3.resource('s3', region_name=TAH_REGION)
glueClient: GlueClient = boto3.client('glue', region_name=TAH_REGION)
dynamoClient: DynamoDBClient = boto3.client('dynamodb', region_name=TAH_REGION)

basicConfig(format='%(asctime)s %(message)s')
logger: Logger = getLogger(__name__)
logger.setLevel("INFO")

class Test_TestS3Excise(unittest.TestCase):
    unittest.TestCase.maxDiff = None
    resourceTearDownStack: list = []
    bucketName: str
    glueDatabaseName: str
    glueTableName: str
    s3RootPath: str
    mutexTableName: str
    privacySearchTableName: str
    privacySearchTablePkName: str
    privacySearchTableSkName: str

    # Create template event to build test cases off of
    lambdaEventTemplate: dict = {
            'Records': [
                {
                    'messageId': '19dd0b57-b21e-4ac1-bd88-01bbb068cb78',
                    'receiptHandle': 'MessageReceiptHandle',
                    'body': '{bodyContent}',
                    'attributes': {
                        'ApproximateReceiveCount': '1',
                        'SentTimestamp': '1523232000000',
                        'SenderId': '123456789012',
                        'ApproximateFirstReceiveTimestamp': '1523232000001'
                    },
                    'messageAttributes': {},
                    'md5OfBody': '{{{md5_of_body}}}',
                    'eventSource': 'aws:sqs',
                    'eventSourceARN': 'arn:aws:sqs:us-east-1:123456789012:MyQueue',
                    'awsRegion': 'us-east-1'
                }
            ]
        }

    @classmethod
    def setUpClass(self) -> None:
        logger.info("[setUpClass] Verify env variables")
        logger.info("[setUpClass] TAH_REGION=%s", TAH_REGION)
        if TAH_REGION is None or TAH_REGION == "":
            raise ValueError("TAH_REGION must be specified")
        if TEST_DATA_FOLDER is None or TEST_DATA_FOLDER == "":
            raise ValueError("TEST_DATA_FOLDER must be specified")
        if GLUE_SCHEMA_PATH_ASYNC is None or GLUE_SCHEMA_PATH_ASYNC == "":
            raise ValueError("GLUE_SCHEMA_PATH_ASYNC must be specified")

        logger.info("[setUpClass] Setup Test Resources")
        
        # Arrange
        testPostfix = str(random.randint(10000, 99999))
        self.s3RootPath = "profiles"

        logger.info("[setUpClass] Creating Dynamo Mutex Table")
        self.mutexTableName = 'distributedMutex_' + testPostfix
        reponse: CreateTableOutputTypeDef = dynamoClient.create_table(
            AttributeDefinitions=[
                {
                    'AttributeName': 'pk',
                    'AttributeType': 'S'
                }
            ],
            KeySchema=[
                {
                    'AttributeName': 'pk',
                    'KeyType': 'HASH'
                }
            ],
            BillingMode='PAY_PER_REQUEST',
            TableName=self.mutexTableName,

        )
        self.resourceTearDownStack.append(lambda: dynamoClient.delete_table(TableName=self.mutexTableName))

        logger.info("[setUpClass] Creating Privacy Search Results Table")
        self.privacySearchTableName = 'privacySearchTable_' + testPostfix
        self.privacySearchTablePkName = 'domainName'
        self.privacySearchTableSkName = 'connectId'
        dynamoClient.create_table(
            AttributeDefinitions=[
                {
                    'AttributeName': f'{self.privacySearchTablePkName}',
                    'AttributeType': 'S'
                },
                {
                    'AttributeName': f'{self.privacySearchTableSkName}',
                    'AttributeType': 'S'
                }
            ],
            KeySchema=[
                {
                    'AttributeName': f'{self.privacySearchTablePkName}',
                    'KeyType': 'HASH'
                },
                {
                    'AttributeName': f'{self.privacySearchTableSkName}',
                    'KeyType': 'RANGE'
                }
            ],
            BillingMode='PAY_PER_REQUEST',
            TableName=self.privacySearchTableName
        )
        self.resourceTearDownStack.append(lambda: dynamoClient.delete_table(TableName=self.privacySearchTableName))
        
        logger.info("[setUpClass] Creating Test Output Bucket on region %s", TAH_REGION)
        self.bucketName = "test-excise-from-s3-" + testPostfix
        #us-east-1 is a special case in boto3 https://github.com/boto/boto3/issues/125
        CreateBucketConfiguration={}
        if TAH_REGION == "us-east-1":
            s3Client.create_bucket(Bucket=self.bucketName)
        else:
            s3Client.create_bucket(Bucket=self.bucketName, CreateBucketConfiguration={'LocationConstraint': TAH_REGION})

        self.resourceTearDownStack.append(lambda: s3Client.delete_bucket(Bucket=self.bucketName))

        logger.info("[setUpClass] Uploading test data")
        filePath: str = TEST_DATA_FOLDER + "/change_proc_output/test_change_proc_output.parquet"
        # Upload six copies:
        #   One copy has only a date prefix without any other firehose injected prefixes
        #   One copy with the format that firehose should output
        #   Four copies for use with bulk excise
        #     - Complete excise with array of ids
        #     - Partial excise with array of ids
        #     - Complete excise using an array of threads, each with a single id
        #     - Partial excise using an array of threads, each with a single id
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/test_change_proc.parquet")
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-test_change_proc.parquet")
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_bulk.parquet")
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_bulk_except_one.parquet")
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_threading.parquet")
        s3Resource.Bucket(self.bucketName).upload_file(filePath, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_threading_except_one.parquet")
        self.resourceTearDownStack.append(lambda: s3Resource.Bucket(self.bucketName).object_versions.all().delete())

        logger.info("[setUpClass] Creating Glue Database")
        self.glueDatabaseName = "test_s3_excise_glue_db_" + testPostfix
        glueClient.create_database(DatabaseInput={
            "Name": self.glueDatabaseName
        })
        self.resourceTearDownStack.append(lambda: glueClient.delete_database(Name=self.glueDatabaseName))

        logger.info("[setUpClass] Loading Glue Schema")
        schemaFilePath: str = GLUE_SCHEMA_PATH_ASYNC + "/tah-common-glue-schemas/traveller.glue.json"
        with open(schemaFilePath) as file:
            data = json.load(file)

        glueColumnList: list = []
        for column in data['columns']:
            glueColumnList.append({
                'Name': column['name'],
                'Type': column['type']['inputString']
            })

        logger.info("[setUpClass] Creating Glue Table")
        self.glueTableName = "test_s3_excise_glue_table_" + testPostfix
        glueClient.create_table(DatabaseName=self.glueDatabaseName, TableInput={
            'Name': self.glueTableName,
            'StorageDescriptor': {
                'Location': f"s3://{self.bucketName}/{self.s3RootPath}",
                'Columns': glueColumnList,
                'InputFormat': 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat',
                'OutputFormat': 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat',
                'SerdeInfo': {
                    'SerializationLibrary': 'org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe'
                }
            },
            'Parameters': {
                'classification': 'parquet'
            },
            'PartitionKeys': [
                {
                    'Name': 'domainname',
                    'Type': 'string'
                }
            ],
            'TableType': 'EXTERNAL_TABLE'
        })
        self.resourceTearDownStack.append(lambda: glueClient.delete_table(DatabaseName=self.glueDatabaseName, Name=self.glueTableName))

        logger.info("[setUpClass] Creating Glue Partition")
        glueClient.batch_create_partition(DatabaseName=self.glueDatabaseName, TableName=self.glueTableName, PartitionInputList=[
            {
                'Values': ['dev'],
                'StorageDescriptor': {
                    'Location': f's3://{self.bucketName}/{self.s3RootPath}/domainname=dev',
                    'InputFormat': 'org.apache.hadoop.mapred.TextInputFormat',
                    'OutputFormat': 'org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat',
                    'Compressed': False,
                    'SerdeInfo': {
                        'SerializationLibrary': 'org.openx.data.jsonserde.JsonSerDe'
                    }
                }
            }
        ])

        logger.info("[setUpClass] Ensuring Dynamo Mutex Table exists")
        waiter: TableExistsWaiter = dynamoClient.get_waiter('table_exists')        
        waiter.wait(TableName=self.mutexTableName, WaiterConfig={'Delay': 5, 'MaxAttempts': 10})

        logger.info("[setUpClass] Ensuring Dynamo Privacy Results Table exists")
        waiter: TableExistsWaiter = dynamoClient.get_waiter('table_exists')        
        waiter.wait(TableName=self.privacySearchTableName, WaiterConfig={'Delay': 5, 'MaxAttempts': 10})

    @classmethod
    def tearDownClass(self) -> None:
        logger.info("[tearDownClass] Tearing Down")
        while len(self.resourceTearDownStack) > 0:
            f = self.resourceTearDownStack.pop()
            f()

    def setUp(self) -> None:
        boto3.setup_default_session(region_name=TAH_REGION)

    def callHandlerWithEnv(self, testEvent) -> dict[str, Any]:
        settings = DynamoDbMutexConfig(table_name=self.mutexTableName, dynamo_db_client=boto3.client('dynamodb', region_name=TAH_REGION))
        mutex = DynamoDbMutexClient(settings)
        return handler_with_env(event=testEvent,
                                glue_database=self.glueDatabaseName,
                                glue_table=self.glueTableName,
                                s3_root_path=self.s3RootPath,
                                privacy_results_table_name=self.privacySearchTableName,
                                privacy_results_pk_name=self.privacySearchTablePkName,
                                privacy_results_sk_name=self.privacySearchTableSkName,
                                mutex=mutex)

    def testExceptionReadingParquet(self) -> None:        
        # Create Success Path Test Case with valid s3 path
        successWithFirehosePrefixBodyContent: dict[str, Any] = {
            "s3Path": f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/test_change_proc.parquet",
            "connectIds": ["420bdbf6f581401695e60fa650958e97"],
            "txId": __name__,
            "domainName": "dev"
        }
        # successLambdaEvent = replaceTemplateBodyContent(lambdaEventTemplate, '{"s3Path": "s3://' + bucketName + '/profiles/domainname=dev/2024/02/01/test_change_proc.parquet", "connectIds": ["420bdbf6f581401695e60fa650958e97"]}')
        successWithFirehosePrefixLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(successWithFirehosePrefixBodyContent))
        
        # Call handler without setting a boto3 default region to trigger wr.s3.read_parquet exception
        boto3.setup_default_session(region_name="bogus-region-to-force-exception")
        with self.assertRaises(Exception):
            self.callHandlerWithEnv(successWithFirehosePrefixLambdaEvent)


    def testNoS3PathFound(self) -> None:
        # Create test case with empty s3 path
        noS3PathFoundBodyContent: dict[str, Any] = {
            "s3Path": f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/1999/01/01/file.parquet",
            "connectIds": [
                "12345",
                "23456"
            ],
            "txId": __name__,
            "domainName": "dev"
        }
        noS3PathFoundEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(noS3PathFoundBodyContent))

        response: dict[str, Any] = self.callHandlerWithEnv(noS3PathFoundEvent)
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['body']['message'], "No files found at S3Path")
        self.assertIsNone(response['body']['error'])

    def testExciseObjectWithoutFirehosePrefix(self) -> None:
        # Create Success Path Test Case with valid s3 path
        successNoFirehosePrefixBodyContent: dict[str, Any] = {
            "s3Path": f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/test_change_proc.parquet",
            "connectIds": ["420bdbf6f581401695e60fa650958e97"],
            "txId": __name__,
            "domainName": "dev"
        }
        successNoFirehosePrefixLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(successNoFirehosePrefixBodyContent))
        
        origDf: DataFrame = wr.s3.read_parquet(successNoFirehosePrefixBodyContent['s3Path'], dataset=False)
        origQueryDf: DataFrame = origDf.query(f"connectid in {str(successNoFirehosePrefixBodyContent['connectIds'])}")

        # Act
        startTime: float = time()
        response: dict[str, Any] = self.callHandlerWithEnv(successNoFirehosePrefixLambdaEvent)
        endTime: float = time()
        logger.info(f"[{self._testMethodName}] Time taken (s): {endTime - startTime}")

        # Assert
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['body']['message'], f"[tx-{__name__}] Successfully removed specified records from S3")
        self.assertIsNone(response['body']['error'])

        # Original S3 Path should exist after excising (Test data will have records remaining)
        s3Object: Object = s3Resource.Object(self.bucketName, f"{self.s3RootPath}/domainname=dev/2024/02/01/test_change_proc.parquet")
        self.assertIsNotNone(s3Object.content_length)

        testDf: DataFrame = wr.s3.read_parquet(path=successNoFirehosePrefixBodyContent['s3Path'], dataset=False)
        testQueryDf: DataFrame = testDf.query(f"connectid in {str(successNoFirehosePrefixBodyContent['connectIds'])}")
        self.assertEqual(len(origDf.index) - len(origQueryDf.index), len(testDf.index))
        self.assertEqual(len(testQueryDf.index), 0)

    def testExciseObjectWithFirehosePrefix(self) -> None:
        # Create Success Path Test Case with valid s3 path
        successWithFirehosePrefixBodyContent: dict[str, Any] = {
            "s3Path": f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-test_change_proc.parquet",
            "connectIds": ["420bdbf6f581401695e60fa650958e97"],
            "txId": __name__,
            "domainName": "dev"
        }
        successWithFirehosePrefixLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(successWithFirehosePrefixBodyContent))
        
        origDf: DataFrame = wr.s3.read_parquet(successWithFirehosePrefixBodyContent['s3Path'], dataset=False)
        origQueryDf: DataFrame = origDf.query(f"connectid in {str(successWithFirehosePrefixBodyContent['connectIds'])}")

        # Act
        startTime: float = time()
        response: dict[str, Any] = self.callHandlerWithEnv(successWithFirehosePrefixLambdaEvent)
        endTime: float = time()
        logger.info(f"[{self._testMethodName}] Time taken (s): {endTime - startTime}")

        # Assert
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['body']['message'], f"[tx-{__name__}] Successfully removed specified records from S3")
        self.assertIsNone(response['body']['error'])

        # Original S3 Path should exist after excising (Test data will have records remaining)
        s3Object: Object = s3Resource.Object(self.bucketName, f"{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-test_change_proc.parquet")
        self.assertIsNotNone(s3Object.content_length)

        testDf: DataFrame = wr.s3.read_parquet(path=successWithFirehosePrefixBodyContent['s3Path'], dataset=False)
        testQueryDf: DataFrame = testDf.query(f"connectid in {str(successWithFirehosePrefixBodyContent['connectIds'])}")
        self.assertEqual(len(origDf.index) - len(origQueryDf.index), len(testDf.index))
        self.assertEqual(len(testQueryDf.index), 0)

    # Bulk Tests
    """
        Test File has the following connectids, with the following counts
            connectid                           count
            846a14c82fef40a387bdd37f52ecc368    4
            0778103810c4489d8cab81beba3daa5a    6
            420bdbf6f581401695e60fa650958e97    2
            c89d57008b0b45b8bca6114598d2fd2f    4
            481f571a82ba4c90a12d54565754c456    1
            ce90dc46716c4a35b5ce63a9cfa9f305    2
            e41c852c472d436c8907389ae3c22087    3
    """

    def testExciseAllExceptOneInBulk(self) -> None:
        s3Path: str = f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_bulk_except_one.parquet"
        exciseAllBodyContent: dict[str, Any] = {
            "s3Path": s3Path,
            "connectIds": [
                '846a14c82fef40a387bdd37f52ecc368',
                '420bdbf6f581401695e60fa650958e97',
                'c89d57008b0b45b8bca6114598d2fd2f',
                '481f571a82ba4c90a12d54565754c456',
                'ce90dc46716c4a35b5ce63a9cfa9f305',
                'e41c852c472d436c8907389ae3c22087',
                # Omitting 0778103810c4489d8cab81beba3daa5a, which has 6 records
            ],
            "txId": __name__,
            "domainName": "dev"
        }
        exciseAllLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(exciseAllBodyContent))

        startTime: float = time()
        response: dict[str, Any] = self.callHandlerWithEnv(exciseAllLambdaEvent)
        endTime: float = time()
        logger.info(f"[{self._testMethodName}] Time taken (s): {endTime - startTime}")

        # Assert
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['body']['message'], f"[tx-{__name__}] Successfully removed specified records from S3")
        self.assertIsNone(response['body']['error'])

        testDf: DataFrame = wr.s3.read_parquet(s3Path)
        self.assertIsNotNone(testDf)
        self.assertEqual(6, len(testDf.index))
        self.assertEqual(6, len(testDf.query('connectid == "0778103810c4489d8cab81beba3daa5a"')))

    def testExciseAllInBulk(self) -> None:
        s3Path = f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_bulk.parquet"
        exciseAllBodyContent: dict[str, Any] = {
            "s3Path": s3Path,
            "connectIds": [
                '846a14c82fef40a387bdd37f52ecc368',
                '0778103810c4489d8cab81beba3daa5a',
                '420bdbf6f581401695e60fa650958e97',
                'c89d57008b0b45b8bca6114598d2fd2f',
                '481f571a82ba4c90a12d54565754c456',
                'ce90dc46716c4a35b5ce63a9cfa9f305',
                'e41c852c472d436c8907389ae3c22087'
            ],
            "txId": __name__,
            "domainName": "dev"
        }
        exciseAllLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(exciseAllBodyContent))

        startTime: float = time()
        response: dict[str, Any] = self.callHandlerWithEnv(exciseAllLambdaEvent)
        endTime: float = time()
        logger.info(f"[{self._testMethodName}] Time taken (s): {endTime - startTime}")

        # Assert
        self.assertEqual(response['statusCode'], 200)
        self.assertEqual(response['body']['message'], f"[tx-{__name__}] Successfully removed specified records from S3")
        self.assertIsNone(response['body']['error'])

        with self.assertRaises(expected_exception=wr.exceptions.NoFilesFound):
            wr.s3.read_parquet(path=s3Path, dataset=False)

    def testExciseAllWithThreading(self) -> None:
        s3Path: str = f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_threading.parquet"
        allIds: list[str] = [
            '846a14c82fef40a387bdd37f52ecc368',
            '0778103810c4489d8cab81beba3daa5a',
            '420bdbf6f581401695e60fa650958e97',
            'c89d57008b0b45b8bca6114598d2fd2f',
            '481f571a82ba4c90a12d54565754c456',
            'ce90dc46716c4a35b5ce63a9cfa9f305',
            'e41c852c472d436c8907389ae3c22087'
        ]
        
        threads: list[threading.Thread | None] = [None] * len(allIds)

        for idx, id in enumerate(allIds):
            exciseAllBodyContent: dict[str, Any] = {
                "s3Path": s3Path,
                "connectIds": [id],
                "txId": __name__,
                "domainName": "dev"
            }
            exciseAllLambdaEvent = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(exciseAllBodyContent))
            
            threads[idx] = threading.Thread(name=f"excise_{id}", target=self.callHandlerWithEnv, args=(exciseAllLambdaEvent,))
        
        df: DataFrame = wr.s3.read_parquet(exciseAllBodyContent['s3Path'], dataset=False)
        self.assertIsNotNone(df)
        self.assertEqual(len(df.index), 22)
        
        for t in threads:
            if(isinstance(t, threading.Thread)):
                t.start()

        for t in threads:
            if(isinstance(t, threading.Thread)):
                t.join()

        with self.assertRaises(expected_exception=wr.exceptions.NoFilesFound):
            wr.s3.read_parquet(path=s3Path, dataset=False)

    def testExciseAllExceptOneWithThreading(self) -> None:
        s3Path: str = f"s3://{self.bucketName}/{self.s3RootPath}/domainname=dev/2024/02/01/ucpFirehose2024-15-31-22-11-55-excise_all_threading_except_one.parquet"
        allIds: list[str] = [
            '846a14c82fef40a387bdd37f52ecc368',
            '0778103810c4489d8cab81beba3daa5a',
            '420bdbf6f581401695e60fa650958e97',
            # Omit 'c89d57008b0b45b8bca6114598d2fd2f', which has 4 records
            '481f571a82ba4c90a12d54565754c456',
            'ce90dc46716c4a35b5ce63a9cfa9f305',
            'e41c852c472d436c8907389ae3c22087'
        ]
        
        threads: list[threading.Thread | None] = [None] * len(allIds)

        for idx, id in enumerate(allIds):
            exciseAllBodyContent: dict[str, Any] = {
                "s3Path": s3Path,
                "connectIds": [id],
                "txId": __name__,
                "domainName": "dev"
            }
            exciseAllLambdaEvent: dict = replaceTemplateBodyContent(self.lambdaEventTemplate, json.dumps(exciseAllBodyContent))
            
            threads[idx] = threading.Thread(name=f"excise_{id}", target=self.callHandlerWithEnv, args=(exciseAllLambdaEvent,))
        
        df: DataFrame = wr.s3.read_parquet(exciseAllBodyContent['s3Path'], dataset=False)
        self.assertIsNotNone(df)
        self.assertEqual(len(df.index), 22)
        
        for t in threads:
            if(isinstance(t, threading.Thread)):
                t.start()

        for t in threads:
            if(isinstance(t, threading.Thread)):
                t.join()

        testDf: DataFrame = wr.s3.read_parquet(path=s3Path, dataset=False)
        self.assertIsNotNone(testDf)
        self.assertEqual(4, len(testDf.index))
        self.assertEqual(4, len(testDf.query('connectid == "c89d57008b0b45b8bca6114598d2fd2f"')))

def replaceTemplateBodyContent(template: dict, replaceWith: str):
    clone: dict = copy.deepcopy(template)
    clone['Records'][0]['body'] = replaceWith

    return clone