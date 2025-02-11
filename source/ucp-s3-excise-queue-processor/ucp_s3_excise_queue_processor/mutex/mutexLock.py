# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

from time import time
from uuid import UUID

import math

from ucp_s3_excise_queue_processor.mutex.mutexKey import MutexKey


class MutexLock():
    def __init__(self,
                 key: MutexKey,
                 lease_duration: float,
                 rvn: UUID,
                 last_check_in_time: float) -> None:
        self.key: MutexKey = key
        self.rvn: UUID = rvn
        self.lease_duration: float = 60.0 if math.isnan(lease_duration) else lease_duration
        self.last_check_in_time: float = time() if math.isnan(last_check_in_time) else last_check_in_time

    def is_expired(self) -> bool:
        current_time: float = time()
        return current_time - self.last_check_in_time > self.lease_duration