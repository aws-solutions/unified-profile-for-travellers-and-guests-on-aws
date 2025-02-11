# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import os
import typing
from logging import Logger, basicConfig, getLogger
from typing import TYPE_CHECKING, TypedDict
from urllib.parse import ParseResult, urlparse

import awswrangler as wr
import boto3
from awswrangler.typing import _S3WriteDataReturnValue
from botocore.config import Config
from botocore.exceptions import ClientError
from pandas import DataFrame
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexClient import (
    DynamoDbMutexClient, DynamoDbMutexConfig, MutexKey)
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexException import \
    FailedLockAcquisitionError
from ucp_s3_excise_queue_processor.mutex.lockContext import LockContext

if TYPE_CHECKING:
    from mypy_boto3_dynamodb.client import DynamoDBClient
    from mypy_boto3_s3.client import S3Client
    from mypy_boto3_s3.service_resource import Bucket, S3ServiceResource
else:
    S3Client = object
    S3ServiceResource = object
    Bucket = object
    DynamoDBClient = object

# Metrics Environment Variables
TAH_REGION: str | None = os.getenv("TAH_REGION")
if TAH_REGION == "" or TAH_REGION is None:
    # getting region for codeBuild project
    TAH_REGION = os.getenv("AWS_REGION")
METRICS_SOLUTION_ID: str | None = os.getenv("METRICS_SOLUTION_ID")
METRICS_SOLUTION_VERSION: str | None = os.getenv("METRICS_SOLUTION_VERSION")

# Function Handler Environment Variables
GLUE_DATABASE: str | None = os.getenv("GLUE_DATABASE")
GLUE_TABLE: str | None = os.getenv("GLUE_TABLE")
S3_ROOT_PATH: str | None = os.getenv("S3_ROOT_PATH")
DISTRIBUTED_MUTEX_TABLE_NAME: str | None = os.getenv("DISTRIBUTED_MUTEX_TABLE")
DISTRIBUTED_MUTEX_TABLE_PK_NAME: str | None = os.getenv("DISTRIBUTED_MUTEX_TABLE_PK_NAME")
DISTRIBUTED_MUTEX_TTL_ATTRIBUTE_NAME: str | None = os.getenv("TTL_ATTRIBUTE_NAME")
PRIVACY_RESULTS_TABLE_NAME: str | None = os.getenv("PRIVACY_RESULTS_TABLE_NAME")
PRIVACY_RESULTS_PK_NAME: str | None = os.getenv("PRIVACY_RESULTS_PK_NAME")
PRIVACY_RESULTS_SK_NAME: str | None = os.getenv("PRIVACY_RESULTS_SK_NAME")

# Setup Logger
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
basicConfig(format='%(asctime)s %(message)s') # NOSONAR (python:S4792) log config is safe in this context
logger: Logger = getLogger(__name__)
logger.setLevel(LOG_LEVEL)

# Setup Metrics Collection; AWS Clients/Resources
solution_identifier: dict[str, str] = {"user_agent_extra": 'AWSSOLUTION/' + str(METRICS_SOLUTION_ID) + '/v' + str(METRICS_SOLUTION_VERSION)}
config = Config(
    region_name=TAH_REGION,
    **solution_identifier,
)

# Setup boto3 clients / resources
s3_resource: S3ServiceResource = boto3.resource("s3", config=config)
s3_client: S3Client = boto3.client("s3", config=config)
dynamo_client: DynamoDBClient = boto3.client("dynamodb", config=config)

class RecordBody(TypedDict):
    s3Path: str
    connectIds: list[str]
    domainName: str
    txId: str

def handler(event, context) -> dict[str, typing.Any]:
    logger.debug(f'Lambda Event: {event}')
    logger.debug(f'Lambda Context: {context}')
    if S3_ROOT_PATH is None:
        raise ValueError("S3_ROOT_PATH must be specified")
    if GLUE_DATABASE is None:
        raise ValueError("GLUE_DATABASE must be specified")
    if GLUE_TABLE is None:
        raise ValueError("GLUE_TABLE must be specified")
    if DISTRIBUTED_MUTEX_TABLE_NAME is None:
        raise ValueError("DISTRIBUTED_MUTEX_TABLE must be specified")
    if DISTRIBUTED_MUTEX_TABLE_PK_NAME is None:
        raise ValueError("DISTRIBUTED_MUTEX_TABLE_PK_NAME must be specified")
    if DISTRIBUTED_MUTEX_TTL_ATTRIBUTE_NAME is None:
        raise ValueError("TTL_ATTRIBUTE_NAME must be specified")
    if PRIVACY_RESULTS_TABLE_NAME is None:
        raise ValueError("PRIVACY_RESULTS_TABLE_NAME must be specified")
    if PRIVACY_RESULTS_PK_NAME is None:
        raise ValueError("PRIVACY_RESULTS_PK_NAME must be specified")
    if PRIVACY_RESULTS_SK_NAME is None:
        raise ValueError("PRIVACY_RESULTS_SK_NAME must be specified")
    # Setup distributed mutex
    mutex = DynamoDbMutexClient(
        DynamoDbMutexConfig(
            client_name="GDPR_S3_EXCISE_QUEUE_PROCESSOR",
            table_name=DISTRIBUTED_MUTEX_TABLE_NAME,
            pk_name=DISTRIBUTED_MUTEX_TABLE_PK_NAME,
            ttl_name=DISTRIBUTED_MUTEX_TTL_ATTRIBUTE_NAME
        )
    )

    return handler_with_env(event=event,
                            glue_database=GLUE_DATABASE,
                            glue_table=GLUE_TABLE,
                            s3_root_path=S3_ROOT_PATH,
                            privacy_results_table_name=PRIVACY_RESULTS_TABLE_NAME,
                            privacy_results_pk_name=PRIVACY_RESULTS_PK_NAME,
                            privacy_results_sk_name=PRIVACY_RESULTS_SK_NAME,
                            mutex=mutex)

def handler_with_env(event,
                     glue_database: str,
                     glue_table: str,
                     s3_root_path: str,
                     privacy_results_table_name: str,
                     privacy_results_pk_name: str,
                     privacy_results_sk_name: str,
                     mutex: DynamoDbMutexClient):
    for record in event['Records']:
        logger.debug(f'Record: {record}')

        payload = record['body']
        logger.debug(f'Payload: {str(payload)}')
        
        body: RecordBody = json.loads(payload)
        logger.debug(f'Body: {body}')   

        response = process_record(body=body, glue_database=glue_database, glue_table=glue_table, s3_root_path=s3_root_path, mutex=mutex)

        # Update PrivacySearch DDB with PK: domainName, SK: connectId|s3Path with a status = purge_success
        for connect_id in body['connectIds']:
            logger.info(f'[tx-{body["txId"]}] Updating PrivacyResults table connectId: {connect_id}')
            dynamo_client.update_item(
                TableName=privacy_results_table_name,
                Key={
                    privacy_results_pk_name: {
                        'S': body['domainName']
                    },
                    privacy_results_sk_name: {
                        'S': f'{connect_id}|{body["s3Path"]}'
                    }
                },
                UpdateExpression=f'SET #status = :new_status, #txId = :new_txId',
                ExpressionAttributeNames={
                    '#status': 'status',
                    '#txId': 'txId'
                },
                ExpressionAttributeValues={
                    ':new_status': {
                        'S': 'purge_success'
                    },
                    ':new_txId': {
                        'S': body['txId']
                    }
                }
            )

    return response


def process_record(body: RecordBody, glue_database: str, glue_table: str, s3_root_path: str, mutex: DynamoDbMutexClient):

    logger.info(f'[tx-{body["txId"]}] Removing {body["connectIds"]} from {body["s3Path"]}')
    tx_id = body['txId']

    try:
        # Try to obtain a lock on the body['s3Path'] object
        with LockContext(mutex, MutexKey(body['s3Path'])):
            try:
                df: DataFrame = wr.s3.read_parquet(path=body['s3Path'], dataset=False)
            except wr.exceptions.NoFilesFound:
                logger.info(f'[tx-{tx_id}] No files found at {body["s3Path"]} for {body["connectIds"]}')
                return format_response(status_code=200, message="No files found at S3Path")
            except Exception as e:
                logger.exception(f'[tx-{tx_id}] Error reading parquet f: {e}')
                raise
            logger.debug(f'[tx-{tx_id}] Number of records before mutation: {len(df.index)} for {body["connectIds"]}')
            
            # Remove the records with the specified connectIds from the sqs message body
            query: str = f"connectid in {str(body['connectIds'])}"        
            mutated_df: DataFrame = df.drop(df.query(query).index)
            logger.debug(f'[tx-{tx_id}] Number of records after mutation: {len(mutated_df.index)}')
            
            # Get the table types from the Glue Catalog; this is needed to determine the correct Parquet file format to write to S3
            try:
                dtypes: dict[str, str] | None = wr.catalog.get_table_types(database=glue_database, table=glue_table)
                if (dtypes is None):
                    logger.error(f"[tx-{tx_id}] Unable to retrieve data types from Glue Tables; will not be able to write parquet")
                    raise ValueError("Unable to retrieve data types from glue tables")
            except Exception as e:
                logger.exception(f'[tx-{tx_id}] Error retrieving glue table types: {e}')
                raise
            
            # Add the 'domainnmame' to the dataframe; this is not in the dataset by default
            # This is needed to allow the wr.s3.to_parquet() function to partition the data by 'domainname'
            mutated_df['domainname'] = mutated_df['domain']

            # Parse the source s3Path, i.e. s3://bucket/prefix/object.ext to
            # delete the original object key and all its versions, and to
            # ensure the new file written is put back to where the old one came from
            parsed_s3_url: ParseResult = urlparse(body['s3Path'])
            bucket: Bucket = s3_resource.Bucket(parsed_s3_url.netloc)
            bucket.object_versions.filter(Prefix=parsed_s3_url.path.strip("/")).delete()
            if (len(mutated_df.index) > 0):
                base_path: str=f'{parsed_s3_url.scheme}://{parsed_s3_url.netloc}/{s3_root_path}'
                try:                
                    result: _S3WriteDataReturnValue = wr.s3.to_parquet(
                        df=mutated_df,                  
                        path=base_path,
                        dataset=True,
                        partition_cols=['domainname'],
                        database=glue_database,
                        table=glue_table,
                        compression='gzip',
                        mode='append',
                        dtype=dtypes
                    )

                    # Move uploaded object back to the original key name
                    logger.info(f"[tx-{tx_id}] Successfully removed {len(df.index) - len(mutated_df.index)} records from S3Path: {body['s3Path']} for {body['connectIds']}")
                    logger.debug(f"[tx-{tx_id}] result.paths: {result['paths']}")
                    uploaded_s3_url: ParseResult = urlparse(result['paths'][0])
                    s3_client.copy_object(CopySource=f'{uploaded_s3_url.netloc}{uploaded_s3_url.path}', Bucket=parsed_s3_url.netloc, Key=parsed_s3_url.path.strip("/"))
                    bucket.object_versions.filter(Prefix=uploaded_s3_url.path.strip("/")).delete()
                except Exception as e:
                    logger.exception(f'[tx-{tx_id}] Error writing to S3: {e}', exc_info=True)
                    raise
    except FailedLockAcquisitionError as e:
        logger.exception(f"[tx-{tx_id}] Unable to acquire lock on {body['s3Path']}")
        raise
    except Exception as e:
        logger.exception(f"[tx-{tx_id}] Error occured: {e}")
        raise
        
    return format_response(status_code=200, message=f"[tx-{tx_id}] Successfully removed specified records from S3")

def format_response(status_code: int, message: str, error: Exception | None = None) -> dict[str, typing.Any]:
    msg: dict[str, typing.Any] = {
        "statusCode": status_code,
        "body": {
            "message": message,
            "error": error
        }
    }

    return msg