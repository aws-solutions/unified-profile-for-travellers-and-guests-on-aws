# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import os
import random
import uuid
from logging import Logger, basicConfig, getLogger
from time import sleep, time

from botocore.exceptions import ClientError
from ucp_s3_excise_queue_processor.mutex.checkInThread import \
    DynamoDbMutexCheckInThread
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexConfig import \
    DynamoDbMutexConfig
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexException import \
    FailedLockAcquisitionError
from ucp_s3_excise_queue_processor.mutex.dynamoDbQueryHelper import (
    ExpressionAttributeNames, ExpressionAttributeValues, TableColumns)
from ucp_s3_excise_queue_processor.mutex.mutexKey import MutexKey
from ucp_s3_excise_queue_processor.mutex.mutexLock import MutexLock

# Setup Logger
LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")
basicConfig(format='%(asctime)s [%(name)s] %(message)s') # NOSONAR (python:S4792) log config is safe in this context
logger: Logger = getLogger(__name__)
logger.setLevel(LOG_LEVEL)

class DynamoDbMutexClient():
    config: DynamoDbMutexConfig

    def __init__(self, config: DynamoDbMutexConfig) -> None:
        self.config = config
        self.check_in_thread: DynamoDbMutexCheckInThread | None = None
        DynamoDbMutexCheckInThread.reap_dangling_check_in_threads()

    def try_acquire_lock(self, key: MutexKey, lease_duration: float | None = None) -> MutexLock | None:
        """
        Attempt to acquire a lock based on the provided key for the provided duration.
        If a Lock is unable to be secured, will return `None`

        key: Resource to attempt to lock

        lease_duration: Time, in seconds, to reserve the lock
        """
        try:
            return self.acquire_lock(key, lease_duration)
        except Exception:
            return None

    def acquire_lock(self, key: MutexKey, lease_duration: float | None = None) -> MutexLock:
        """
        Attempt to acquire a lock based on the provided key for the provided duration.
        If a Lock is unable to be secured, will throw a FailedLockAcquisitionException

        key: Resource to attempt to lock

        lease_duration: Time, in seconds, to reserve the lock
        """
        logger.debug(f"[{self.config.client_name}] Attempting to acquire a new lock table: {self.config.table_name} for: {key} for {self.config.default_lease_duration if lease_duration is None else lease_duration}s")
        lease_duration = self.config.default_lease_duration if lease_duration is None else lease_duration
                
        current_rvn: uuid.UUID | None = None
        start_time: float = time()

        while(True):
            try:
                existing_lock: MutexLock = self.get_lock(key)
                if (existing_lock is None):
                    # No lock found in DDB, place a lock on {key}   
                    lock: MutexLock = self.put_mutex_lock(key, lease_duration)
                    logger.debug(f"[{self.config.client_name}] Acquired lock on {key}")
                    
                    if (self.config.auto_renew):
                        self.check_in_thread = DynamoDbMutexCheckInThread(self, lock).begin()
                    return lock
                
                # Lock found; check RecordValueNumber
                logger.debug(f"[{self.config.client_name}] Lock exists, check if lock is expired")
                if (current_rvn is None):
                    current_rvn = existing_lock.rvn
                else:
                    # The same client as before still has the lock
                    if (current_rvn == existing_lock.rvn):
                        # Check to see if the lock is expired
                        if (existing_lock.is_expired()):
                            # Lease is expired from previous client
                            lock = self.upsert_mutex_lock(existing_lock, key, lease_duration)
                            if (self.config.auto_renew):
                                self.check_in_thread = DynamoDbMutexCheckInThread(self, lock).begin()                            
                            logger.debug(f'[{self.config.client_name}] Acquired lock on expired existing: {key}')
                            return lock
                    else:
                        # A new client claimed the lock in between sleeps
                        current_rvn = existing_lock.rvn
            except ClientError as e:
                if e.response['Error']['Code'] == 'ConditionalCheckFailedException':
                    if (e.response['Error']['Message'] == 'The conditional request failed'):
                        logger.debug(f"[{self.config.client_name}] Another client secured lock; sleeping, then re-try", exc_info=1)
                    else:
                        logger.error(f"Unknown error occured: {e}", exc_info=True)
                else:
                    logger.error(f"[{self.config.client_name}] SDK Error Occurred: {e}", exc_info=True)
                    raise FailedLockAcquisitionError("SDK Error Occured")
            except Exception as e:
                logger.error(f"[{self.config.client_name}] Unknown Error Occured: {e}", exc_info=True)
                raise FailedLockAcquisitionError("Unknown Error Occured")
            
            # Sleep for the preconfigured refreshWaitTime, plus up to 2 additional seconds to mitigate
            # multiple clients in lock step
            wait_time = self.config.refresh_wait_time + (random.random() * 2) # NOSONAR (python:S2245) randomness is just used to add jitter, pseudorandom is safe here
            logger.debug(f"[{self.config.client_name}] Lock not expired; waiting {wait_time} seconds")
            sleep(wait_time)
            
            if (time() - start_time > self.config.max_timeout):
                logger.error(f"[{self.config.client_name}] Maximum timeout of {self.config.max_timeout} exceeded. Unable to secure a Lock")
                raise FailedLockAcquisitionError("Wait timeout for lock acquisition exceeded")

    def get_lock(self, key: MutexKey) -> MutexLock | None:
        logger.debug(f"[{self.config.client_name}] Getting lock from Dynamo")

        response = self.config.dynamo_db_client.get_item(
            TableName=self.config.table_name,
            Key=self.create_key_parameter_query(key),
            ConsistentRead=True
        )
        if ('Item' in response):
            item = response['Item']
            sk = item['sk']['S'] if 'sk' in item else None
            return MutexLock(MutexKey(item['pk']['S'], sk), float(item['duration']['N']), uuid.UUID(item['rvn']['S']), float(item['lastCheckin']['N']))
        
        return None

    def renew_lock(self, lock: MutexLock) -> MutexLock:
        logger.debug(f"[{self.config.client_name}] Attempting to renew lock")
        current_time: float = time()
        new_rvn: uuid.UUID = uuid.uuid4()
        self.config.dynamo_db_client.update_item(
            TableName=self.config.table_name,
            Key=self.create_key_parameter_query(lock.key),
            UpdateExpression='SET ' +
                f'{TableColumns.rvn} = {ExpressionAttributeValues.new_rvn},' +
                f'{TableColumns.last_checkin} = {ExpressionAttributeValues.new_last_checkin},' +
                f'{ExpressionAttributeNames.ttl_attr_name} = {ExpressionAttributeValues.new_ttl}',
            ExpressionAttributeNames={
                ExpressionAttributeNames.ttl_attr_name: self.config.ttl_name
            },
            ExpressionAttributeValues={
                ExpressionAttributeValues.new_rvn: {
                    'S': str(new_rvn)
                },
                ExpressionAttributeValues.new_last_checkin: {
                    'N': f'{current_time}'
                },
                ExpressionAttributeValues.new_ttl: {
                    'N': f'{current_time + lock.lease_duration}'
                },
                ExpressionAttributeValues.existing_rvn: {
                    'S': str(lock.rvn)
                }
            },
            ConditionExpression=f'{TableColumns.rvn} = {ExpressionAttributeValues.existing_rvn}'
        )
        return MutexLock(lock.key, lock.lease_duration, new_rvn, current_time)

    def release_lock(self, lock: MutexLock):
        if (lock is None):
            logger.warning(f"[{self.config.client_name}] Lock is None, cannot release")
            return None
        logger.debug(f"[{self.config.client_name}] Releasing lock on: {lock.key}")
        
        # Signal check-in thread to stop
        thread_lock: MutexLock | None = self.stop_check_in_thread()

        # Delete Lock from DDB
        try:        
            self.config.dynamo_db_client.delete_item(
                TableName=self.config.table_name,
                Key=self.create_key_parameter_query(lock.key),
                ExpressionAttributeValues={
                    ExpressionAttributeValues.existing_rvn: {
                        'S': str(lock.rvn if thread_lock is None else thread_lock.rvn)
                    }
                },
                ConditionExpression=f'{TableColumns.rvn} = {ExpressionAttributeValues.existing_rvn}'
            )
        except Exception:
            # Can ignore this exception; it will either get re-used by
            # another client, or TTL'd by DDB
            logger.warning("Error removing lock from DDB; Can ignore error; The lock will be re-used by another client, or will get TTL'd by DDB", exc_info=True)

        return None

    def put_mutex_lock(self, key: MutexKey, lease_duration: float):
        current_time = time()
        rvn: uuid.UUID = uuid.uuid4()
        item = self.create_key_parameter_query(key)
        item[TableColumns.rvn] = {
            'S': str(rvn)
        }
        item[TableColumns.duration] = {
            'N': f'{lease_duration}'
        }
        item[TableColumns.last_checkin] = {
            'N': f'{current_time}'
        }
        item[self.config.ttl_name] = {
            'N': f'{current_time + lease_duration}'
        }
        self.config.dynamo_db_client.put_item(
            TableName=self.config.table_name,
            Item=item,
            ConditionExpression=f'attribute_not_exists({self.config.pk_name})'
        )

        return MutexLock(key, lease_duration, rvn, current_time)

    def upsert_mutex_lock(self, existing_lock: MutexLock, key: MutexKey, lease_duration: float):
        current_time: float = time()
        rvn: uuid.UUID = uuid.uuid4()
        self.config.dynamo_db_client.update_item(
            TableName=self.config.table_name,
            Key=self.create_key_parameter_query(key),
            UpdateExpression=f'SET ' +
                f'{TableColumns.rvn} = {ExpressionAttributeValues.new_rvn}, ' +
                f'{ExpressionAttributeNames.duration_attr_name} = {ExpressionAttributeValues.new_duration}, ' +
                f'{TableColumns.last_checkin} = {ExpressionAttributeValues.new_last_checkin}, ' +
                f'{ExpressionAttributeNames.ttl_attr_name} = {ExpressionAttributeValues.new_ttl}',
            ExpressionAttributeNames={
                ExpressionAttributeNames.duration_attr_name: TableColumns.duration,
                ExpressionAttributeNames.ttl_attr_name: self.config.ttl_name
            },
            ExpressionAttributeValues={
                ExpressionAttributeValues.existing_rvn: {
                    'S': str(existing_lock.rvn)
                },
                ExpressionAttributeValues.existing_last_checkin: {
                    'N': f'{existing_lock.last_check_in_time}'
                },
                ExpressionAttributeValues.existing_ttl: {
                    'N': f'{existing_lock.last_check_in_time + existing_lock.lease_duration}'
                },
                ExpressionAttributeValues.new_rvn: {
                    'S': str(rvn)
                },
                ExpressionAttributeValues.new_duration: {
                    'N': f'{lease_duration}'
                },
                ExpressionAttributeValues.new_last_checkin: {
                    'N': f'{current_time}'
                },
                ExpressionAttributeValues.new_ttl: {
                    'N': f'{current_time + lease_duration}'
                }
            },
            ConditionExpression=
                f'{TableColumns.rvn} = {ExpressionAttributeValues.existing_rvn} ' +
                f'AND {TableColumns.last_checkin} = {ExpressionAttributeValues.existing_last_checkin} '
                f'AND {ExpressionAttributeNames.ttl_attr_name} = {ExpressionAttributeValues.existing_ttl}'
        )

        return MutexLock(key, lease_duration, rvn, current_time)

    def create_key_parameter_query(self, key: MutexKey):
        item = {
            self.config.pk_name: {
                'S': key.partition_key
            }
        }

        if (key.sort_key is not None):
            item[self.config.sk_name] = {
                'S': key.sort_key
            }
        
        return item
    
    def stop_check_in_thread(self) -> MutexLock | None:
        if (self.check_in_thread is not None):
            return self.check_in_thread.terminate()



