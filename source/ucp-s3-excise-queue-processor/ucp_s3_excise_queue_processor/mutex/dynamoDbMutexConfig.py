# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import boto3

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from mypy_boto3_dynamodb.client import DynamoDBClient
else:
    DynamoDBClient = object

class DynamoDbMutexConfig():    
    def __init__(self, # type: ignore
                 client_name: str = '',
                 pk_name: str = 'pk',
                 sk_name: str = 'sk',
                 ttl_name: str = 'ttl',
                 dynamo_db_client = None,
                 table_name: str = "distributedMutex",
                 default_lease_duration: float = 30,
                 auto_renew: bool = False,
                 refresh_wait_time: float = 1,
                 max_timeout: float = 60) -> None:
        """
        client_name: Name of the Mutex Client

        pk_name: Name of the DynamoDB PartitionKey, defaults to 'pk'

        sk_name: Name of the DynamoDB SortKey, defaults to 'sk'

        ttl_name: Name of the DynamoDB attribute configured for TTL

        dynamo_db_client: Provide a customized boto3 client

        table_name: Name of the DynamoDB Table, defaults to 'distributedMutex'

        default_lease_duration: Length of time to reserve a lock, defaults to 30 seconds

        auto_renew: Have the client attempt to auto renew the lock lease

        refresh_wait_time: Amount of time to wait for a lock to expire or be released, defaults to 1 second
          The client will wait for up to 2 additional seconds -> refreshWaitTime + (random() * 2)
          in order to mitigate multiple clients that happen to be in lock step with one another

        max_timeout: The maximum amount of total time to acquire a lock, returning None if this limit is reached, defaults to 60 seconds

        """
        self.client_name = client_name
        self.pk_name: str = pk_name
        self.sk_name: str = sk_name
        self.ttl_name: str = ttl_name
        self.dynamo_db_client: DynamoDBClient = boto3.client('dynamodb') if dynamo_db_client is None else dynamo_db_client
        self.table_name: str = table_name
        self.default_lease_duration: float = default_lease_duration
        self.refresh_wait_time: float = refresh_wait_time
        self.max_timeout: float = max_timeout
        self.auto_renew: bool = auto_renew

    