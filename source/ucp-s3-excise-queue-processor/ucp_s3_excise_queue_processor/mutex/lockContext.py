# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexClient import DynamoDbMutexClient
from ucp_s3_excise_queue_processor.mutex.mutexKey import MutexKey
from ucp_s3_excise_queue_processor.mutex.mutexLock import MutexLock


class LockContext():
    def __init__(self,
                 mutex_client: DynamoDbMutexClient,
                 mutex_key: MutexKey,
                 lease_duration: float | None = None):
        self.mutex_client: DynamoDbMutexClient = mutex_client
        self.mutex_key: MutexKey = mutex_key
        self.lease_duration: float = mutex_client.config.default_lease_duration if lease_duration is None else lease_duration
        self.lock: MutexLock | None = None

    def __enter__(self):        
        self.lock = self.mutex_client.acquire_lock(self.mutex_key, self.lease_duration)
        return self.lock

    def __exit__(self, exc_type, exc_value, traceback):
        self.mutex_client.release_lock(self.lock)
        if (exc_type is not None):
            # Some exception occurred in the body of the context
            logging.exception(f"[LockContext] An exception of type: {exc_type} occurred: {exc_value}")
        # Do not suppress any exception that occurs
        return False