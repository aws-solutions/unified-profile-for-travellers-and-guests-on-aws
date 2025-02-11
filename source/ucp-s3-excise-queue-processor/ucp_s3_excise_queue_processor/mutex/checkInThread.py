# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import os
import threading
from logging import Logger, basicConfig, getLogger
from time import sleep

from ucp_s3_excise_queue_processor.mutex.mutexLock import MutexLock

# Setup Logger
LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")
basicConfig(format='%(asctime)s [%(name)s] %(message)s') # NOSONAR (python:S4792) log config is safe in this context
logger: Logger = getLogger(__name__)
logger.setLevel(LOG_LEVEL)

class DynamoDbMutexCheckInThread(threading.Thread):
    lock: MutexLock | None = None
    def __init__(self, mutex_client, lock: MutexLock) -> None:
        super().__init__(target=self.check, name="lockCheckIn", daemon=True, args=(mutex_client, lock,))
        self.event = threading.Event()
    
    def begin(self):
        self.start()
        return self

    def terminate(self) -> MutexLock:
        self.event.set()
        return self.lock

    def check(self, client, lock) -> None:
        logger.info(f"[{client.config.client_name}] Initializing check-in thread")
        self.lock = lock
        while(self.event.is_set() == False):
            self.lock = client.renew_lock(self.lock)
            sleep(1)

    """
    Locate and terminate any threads named "lockCheckin". This is to ensure that
    for warm start lambdas, any zombie threads that were unable to be terminated
    from a previous execution, possibly due to an unhandled exception or other 
    catastophic event, are able to be terminated and not hold onto a lock
    """
    @staticmethod
    def reap_dangling_check_in_threads() -> None:
        for t in threading.enumerate():
            if (t.name == "lockCheckIn" and isinstance(t, DynamoDbMutexCheckInThread)):
                logger.info(f"[ThreadReaper] Terminating dangling checkin thread: {t.native_id}")
                t.terminate()