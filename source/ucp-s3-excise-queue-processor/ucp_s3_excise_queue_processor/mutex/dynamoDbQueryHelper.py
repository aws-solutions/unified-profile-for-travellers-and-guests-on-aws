# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

class ExpressionAttributeValues():
    existing_rvn: str = ":existingRvn"
    existing_last_checkin: str = ":existingLastCheckin"
    existing_ttl: str = ":existingTtl"

    new_rvn: str = ":newRvn"
    new_duration: str = ":newDuration"
    new_last_checkin: str = ":newLastCheckin"
    new_ttl: str = ":newTtl"

class ExpressionAttributeNames():
    ttl_attr_name: str = "#ttl"
    duration_attr_name: str = "#duration"

class TableColumns():
    rvn: str = "rvn"
    duration: str = "duration"
    last_checkin: str = "lastCheckin"