# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import os
import random
import threading
import unittest
from logging import Logger, basicConfig, getLogger
from time import sleep

import boto3
from botocore.config import Config
from testcontainers.core.container import DockerContainer
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexClient import (
    DynamoDbMutexClient, DynamoDbMutexConfig, MutexKey, MutexLock)
from ucp_s3_excise_queue_processor.mutex.dynamoDbMutexException import \
    FailedLockAcquisitionError
from ucp_s3_excise_queue_processor.mutex.lockContext import LockContext

TAH_REGION: str | None = os.getenv("TAH_REGION")
if TAH_REGION == None or TAH_REGION == "":
    # Get Region for CodeBuild
    TAH_REGION = os.getenv("AWS_REGION")

# Setup Logger
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
basicConfig(format='%(asctime)s %(message)s')
logger: Logger = getLogger(__name__)
logger.setLevel(LOG_LEVEL)

class DynamoDbContainer(DockerContainer):
    def __init__(self, **kwargs) -> None:
        super().__init__("amazon/dynamodb-local:2.2.1", **kwargs)
        self.with_exposed_ports(8000)


class LogEntry():
    def __init__(self, client_idx: int, action: str) -> None:
        self.client_idx: int = client_idx
        self.action: str = action

class Test_Mutex(unittest.TestCase):
    ddbClient = None
    waiter = None
    ddbContainer: DynamoDbContainer
    mutex: DynamoDbMutexClient

    def createMutexClient(self, client_name: str = "", autoRenew: bool = False, maxTimeout: int = 60):
        return DynamoDbMutexClient(DynamoDbMutexConfig(client_name=client_name, dynamo_db_client=self.ddbClient, auto_renew=autoRenew, max_timeout=maxTimeout))
    
    @classmethod
    def setUpClass(self) -> None:
        self.ddbContainer = DynamoDbContainer()
        self.ddbContainer.start()

        self.ddbClient = boto3.client('dynamodb',
            config=Config(max_pool_connections=50),
            region_name=TAH_REGION,
            endpoint_url=f"http://{self.ddbContainer.get_container_host_ip()}:{self.ddbContainer.get_exposed_port(8000)}"
        )
        self.waiter = self.ddbClient.get_waiter('table_exists')
        response = self.ddbClient.create_table(
            AttributeDefinitions=[
                {
                    'AttributeName': 'pk',
                    'AttributeType': 'S'
                }
            ],
            KeySchema=[
                {
                    'AttributeName': 'pk',
                    'KeyType': 'HASH',
                }
            ],
            BillingMode='PAY_PER_REQUEST',
            TableName='distributedMutex'
        )
        logger.info(f"[setUpClass] Waiting for table creation")
        self.waiter.wait(TableName='distributedMutex', WaiterConfig={'Delay': 5, 'MaxAttempts': 10})
        self.mutex = DynamoDbMutexClient(DynamoDbMutexConfig(dynamo_db_client=self.ddbClient, table_name=response['TableDescription']['TableName']))

    @classmethod
    def tearDownClass(self) -> None:
        logger.info("[tearDownClass] Tests completed; tearing down testcontainers")
        self.ddbContainer.stop()

    def testAcquireAndReleaseLock(self) -> None:
        lock: MutexLock | None = self.mutex.try_acquire_lock(MutexKey("test_pk"), 10)
        self.assertIsNotNone(lock)

        # Sleep briefly to simulate a slightly longer running operation
        sleep(3)

        lock = self.mutex.release_lock(lock)
        self.assertIsNone(lock)


    def testAcquireSameLock(self) -> None:
        # Client1 acquires lock
        lock: MutexLock | None = self.mutex.try_acquire_lock(MutexKey("test_pk"), 10)
        self.assertIsNotNone(lock)

        # Client1 performs its operations, and 5 seconds later, releases the lock
        threading.Timer(interval=3, function=self.mutex.release_lock, kwargs={'lock': lock}).start()

        # Client 2 attempts to acquire lock, must wait until C1 releases
        newLock: MutexLock | None = self.mutex.try_acquire_lock(MutexKey("test_pk"), 10)
        self.assertIsNotNone(newLock)
        self.mutex.release_lock(newLock)
        # The two lock objects returned should have different RVN values, indicating
        # that each client was successful in acquiring a lock on the same resource
        self.assertNotEqual(lock.rvn, newLock.rvn)

    def testLockTimeout(self) -> None:
        client: DynamoDbMutexClient = self.createMutexClient(client_name="longDuration", maxTimeout=60)   
        timeoutClient: DynamoDbMutexClient = self.createMutexClient(client_name="timeoutClient", maxTimeout=5)
        key = MutexKey("test_pk")

        lock: MutexLock = client.try_acquire_lock(key, 20)
        self.assertIsNotNone(lock)

        # client performs its operations, and 10 seconds later, releases the lock
        threading.Timer(interval=10, function=client.release_lock, args=(lock,))

        # timeoutClient attempts to get the Lock, but should timeout and return None
        noneLock: MutexLock = timeoutClient.try_acquire_lock(key)
        self.assertIsNone(noneLock)

    def testAcquireLockWithThreads(self) -> None:
        locks: list[MutexLock] = [None] * 3
        def acquireLock(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}')
            locks[idx] = client.try_acquire_lock(key, 60)
            sleep(5)
            client.release_lock(locks[idx])

        
        c1 = threading.Thread(target=acquireLock, args=[MutexKey("test_pk"), 0])
        c2 = threading.Thread(target=acquireLock, args=[MutexKey("test_pk"), 1])
        c3 = threading.Thread(target=acquireLock, args=[MutexKey("test_pk"), 2])
        c1.start()
        c2.start()
        c3.start()
        c1.join()
        c2.join()
        c3.join()
        
        self.assertIsNotNone(locks[0])
        self.assertIsNotNone(locks[1])
        self.assertIsNotNone(locks[2])

        self.assertNotEqual(locks[0].key, locks[1].key)
        self.assertNotEqual(locks[1].key, locks[2].key)
        
    def testAcquireLockWithThreadExceptions(self) -> None:
        locks: list[MutexLock] = [None] * 3
        
        def acquireLockWithException(key: MutexKey, idx: int):
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}')
            locks[idx] = client.try_acquire_lock(key, 10)
            sleep(5)
            raise Exception(f"[{self._testMethodName}] Some exception occured; cannot release lock")
        
        def acquireLock(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}')
            locks[idx] = client.try_acquire_lock(key, 3)
            sleep(5)
            client.release_lock(locks[idx])

        c1 = threading.Thread(name="exceptionThread", target=acquireLockWithException, args=[MutexKey("test_pk"), 0])
        c2 = threading.Thread(target=acquireLock, args=(MutexKey("test_pk"), 1))
        c1.start()
        
        # Ensure that c1 has a head start so that it assuredly acquires the lock
        sleep(3)
        c2.start()

        c1.join()
        c2.join()     

    @unittest.skip("Fails frequently on CodeBuild; skipping until multiprocessing package can be implemented")
    def testAutoRenew(self) -> None:
        maxTests = 4
        threads: list[threading.Thread] = [None] * maxTests
        expected_locking_results: list[str] = [""] * maxTests
        log_entry_list: list[LogEntry] = []
        def acquireLock(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}', True)
            lock: MutexLock | None = client.try_acquire_lock(key, 4)
            expected_locking_results[idx] += "a"
            log_entry_list.append(LogEntry(idx, "secured lock"))
            sleep(6)
            client.release_lock(lock)
            expected_locking_results[idx] += "r"
            log_entry_list.append(LogEntry(idx, "released lock"))

        key = MutexKey("pk")
        for i in range(0, maxTests):
            threads[i] = threading.Thread(name=f'c{i}Thread', target=acquireLock, args=(key, i,))
        
        for thread in threads:
            thread.start()

        for thread in threads:
            thread.join()

        for r in expected_locking_results:
            self.assertEqual(r, "ar")

        self.assertEqual(len(log_entry_list), len(threads) * 2)
        for idx, log_entry in enumerate(log_entry_list):
            if (idx % 2 == 0):
                # Even entry, should be "securing a lock"
                self.assertEqual("secured lock", log_entry.action)
                if (idx != 0):
                    self.assertNotEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)
            else:
                # Odd entry, should be releasing or erroring
                self.assertTrue("released lock" == log_entry.action)
                self.assertEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)



    def testReapExistingThreads(self) -> None:
        locks: list[MutexLock] = [None] * 3
        def acquireLockWithException(key: MutexKey, idx: int):
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}', True)
            locks[idx] = client.try_acquire_lock(key, 3)
            sleep(1)
            raise Exception(f"[{self._testMethodName}] Some exception occured; cannot release lock")
        
        def acquireLock(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(f'{idx}', True)
            locks[idx] = client.try_acquire_lock(key, 10)
            sleep(5)
            client.release_lock(locks[idx])

        c1 = threading.Thread(name="exceptionThread", target=acquireLockWithException, args=[MutexKey("test_pk"), 0])
        c2 = threading.Thread(target=acquireLock, args=(MutexKey("test_pk"), 1))
        
        # Start and complete c1; it should raise an exception, and the checkin thread should be active
        c1.start()
        c1.join()
        
        # When c2 starts, it should identify and terminate all checkin threads
        # This is necessary because lambda may re-use the existing context; if that's
        # the case, then any background threads that were spawned will be paused and then resumed,
        # potentially causing an issue with a persistent lock
        c2.start()
        c2.join()
        
    def testLoadWithThreading(self) -> None:
        max_tests = 50
        threads: list[threading.Thread] = [None] * max_tests
        expected_locking_results: list[str] = [""] * max_tests
        log_entry_list: list[LogEntry] = []

        def successfulAcquire(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(client_name=str(idx), maxTimeout=240)
            lock: MutexLock | None = client.try_acquire_lock(key, 4)
            if (lock is not None):
                expected_locking_results[idx] += "a"
                log_entry_list.append(LogEntry(idx, "secured lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} secured lock")
                sleep(2)
                log_entry_list.append(LogEntry(idx, "released lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} released lock.")
                client.release_lock(lock)
                expected_locking_results[idx] += "r"

        def exceptionAcquire(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(client_name=str(idx), maxTimeout=240)
            lock: MutexLock | None = client.try_acquire_lock(key, 4)
            if (lock is not None):
                expected_locking_results[idx] += "a"
                log_entry_list.append(LogEntry(idx, "secured lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} secured lock")
                sleep(2)
                log_entry_list.append(LogEntry(idx, "erroring out"))
                logger.info(f"[{self._testMethodName}] Client {idx} erroring out.")
                raise Exception(f"[{self._testMethodName}] Exception occurred; unable to release lock")

        key = MutexKey("pk_test")
        for i in range(0, max_tests):
            rnd: float = random.random()
            if (rnd < 0.1):
                threads[i] = threading.Thread(name=f"Test_exception_{i}", target=exceptionAcquire, args=(key, i))
            else:
                threads[i] = threading.Thread(name=f"Test_success_{i}", target=successfulAcquire, args=(key, i))

        for t in threads:
            t.start()

        for t in threads:
            t.join()

        # Ensure that each client represented by a thread was able to acquire a lock, and those that raise an exception could not explicitly release it
        for idx, expected_result in enumerate(expected_locking_results):
            if ("exception" in threads[idx].name):
                # Threads that raise an exception, should only have acquired ("a") a lock
                self.assertEqual(expected_result, "a")
            else:
                # Threads that do not raise an exception should have acquired ("a") and released ("r") a lock
                self.assertEqual(expected_result, "ar")

        # Ensure that order is preserved, and no client could acquire a lock before it being released
        # The `logging_list` should have a length of `max_tests * 2`
        # Each even entry [0, 2, 4,...] should be a client securing a lock
        # Each odd entry [1, 3, 5,...] should be a client releasing or erroring;
        # Each odd entry should match [i-1]'s client ID, e.g. logging_list[1]'s client idx should match logging_list[0]'s client idx
        self.assertEqual(len(log_entry_list), max_tests * 2)
        for idx, log_entry in enumerate(log_entry_list):
            if (idx % 2 == 0):
                # Even entry, should be "securing a lock"
                self.assertEqual("secured lock", log_entry.action)
                if (idx != 0):
                    self.assertNotEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)
            else:
                # Odd entry, should be releasing or erroring
                self.assertTrue("released lock" == log_entry.action or "erroring out" == log_entry.action)
                self.assertEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)


    def testWithLockContext(self) -> None:
        with LockContext(self.mutex, MutexKey("pk"), 60) as lock:
            self.assertIsNotNone(lock)

    def testWithLockContextException(self) -> None:
        c0: DynamoDbMutexClient = self.createMutexClient("0")
        c1: DynamoDbMutexClient = self.createMutexClient("1", False, 5)

        key = MutexKey("pk_test")
        lock: MutexLock = c0.try_acquire_lock(key, 60)
        logger.info(f"[{self._testMethodName}] C0 Acquired Lock")
        self.assertIsNotNone(lock)

        with self.assertRaises(FailedLockAcquisitionError):
            with LockContext(c1, key, 10):
                logger.info(f"[{self._testMethodName}] C1 Acquired Lock")

        c0.release_lock(lock)


    def testWithLockContextWaitTimeout(self) -> None:
        expectedLockingResults: list[str] = [""] * 2
        def acquireLock(key: MutexKey, idx: int, maxTimeout: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(client_name=str(idx), maxTimeout=maxTimeout)
            with LockContext(client, key, 10):
                logger.info(f"[{self._testMethodName}] Client {idx} acquired lock")
                expectedLockingResults[idx] += "a"
                sleep(5)
                logger.info(f"[{self._testMethodName}] Client {idx} releasing lock")
                expectedLockingResults[idx] += "r"

        key = MutexKey("pk_test")

        # Create a thread simulating a client that will wait for 60 seconds to acquire a lock
        t0 = threading.Thread(name="Thread1", target=acquireLock, args=(key, 0, 60))

        # Create a thread simulating a client that will only wait for 2 seconds to acquire a lock
        t1 = threading.Thread(name="ShortWaitThread", target=acquireLock, args=(key, 1, 2))

        t0.start()
        # Sleep to ensure that t1 acquires the lock
        sleep(1)
        t1.start()

        # t2 should throw an error because its wait timeout was exceeded
        t0.join()
        t1.join()

        self.assertEqual(expectedLockingResults[0], "ar")
        self.assertEqual(expectedLockingResults[1], "")


    def testLoadContextWithThreading(self) -> None:
        max_tests = 50
        threads: list[threading.Thread] = [None] * max_tests
        expected_locking_results: list[str] = [""] * max_tests
        log_entry_list: list[LogEntry] = []

        def successfulAcquire(key: MutexKey, idx: int) -> None:
            client: DynamoDbMutexClient = self.createMutexClient(client_name=str(idx), maxTimeout=240)
            with LockContext(client, key, 4):
                expected_locking_results[idx] += "a"
                log_entry_list.append(LogEntry(idx, "secured lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} secured lock")
                sleep(2)
                log_entry_list.append(LogEntry(idx, "released lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} released lock.")
                expected_locking_results[idx] += "r"

        def exceptionAcquire(key: MutexKey, idx: int):
            client: DynamoDbMutexClient = self.createMutexClient(client_name=str(idx), maxTimeout=240)
            with LockContext(client, key, 4):
                expected_locking_results[idx] += "a"
                log_entry_list.append(LogEntry(idx, "secured lock"))
                logger.info(f"[{self._testMethodName}] Client {idx} secured lock")
                sleep(2)
                log_entry_list.append(LogEntry(idx, "erroring out"))
                logger.info(f"[{self._testMethodName}] Client {idx} erroring out")
                raise Exception(f"[{self._testMethodName}] Exception occurred while holding lock")
            
        key = MutexKey("pk_test")
        for i in range(0, max_tests):
            rnd: float = random.random()
            if (rnd < 0.1):
                threads[i] = threading.Thread(name=f"Test_exception_{i}", target=exceptionAcquire, args=(key, i))
            else:
                threads[i] = threading.Thread(name=f"Test_success_{i}", target=successfulAcquire, args=(key, i))
        
        for t in threads:
            t.start()

        for t in threads:
            t.join()

        for idx, expectedResult in enumerate(expected_locking_results):
            if ("exception" in threads[idx].name):
                self.assertEqual(expectedResult, "a")
            else:
                self.assertEqual(expectedResult, "ar")

        # Ensure that order is preserved, and no client could acquire a lock before it being released
        # The `logging_list` should have a length of `max_tests * 2`
        # Each even entry [0, 2, 4,...] should be a client securing a lock
        # Each odd entry [1, 3, 5,...] should be a client releasing or erroring;
        # Each odd entry should match [i-1]'s client ID, e.g. logging_list[1]'s client idx should match logging_list[0]'s client idx
        self.assertEqual(len(log_entry_list), max_tests * 2)
        for idx, log_entry in enumerate(log_entry_list):
            if (idx % 2 == 0):
                # Even entry, should be "securing a lock"
                self.assertEqual("secured lock", log_entry.action)
                if (idx != 0):
                    self.assertNotEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)
            else:
                # Odd entry, should be releasing or erroring
                self.assertTrue("released lock" == log_entry.action or "erroring out" == log_entry.action)
                self.assertEqual(log_entry.client_idx, log_entry_list[idx-1].client_idx)