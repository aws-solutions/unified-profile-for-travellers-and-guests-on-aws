# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

class MutexKey():
    def __init__(self, partition_key: str, sort_key: str | None = None) -> None:
        self.partition_key: str = partition_key
        self.sort_key: str | None = sort_key

    def __str__(self) -> str:
        if (self.sort_key is None):
            return self.partition_key
        return f'{self.partition_key}|{self.sort_key}'
