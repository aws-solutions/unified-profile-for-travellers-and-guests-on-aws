// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { BaseCondition, Rule, SkipCondition } from './config';

export interface SaveRuleSetRequest {
    rules: Rule[];
}

export interface DisplayRule {
    index: number;
    name: string;
    description: string;
    skipConditions: SkipCondition[];
    matchFilterConditions: MatchFilterCondition[];
}

export interface MatchFilterCondition extends BaseCondition {
    existingObjectType: string;
    existingObjectField: string;
    filterConditionEnabled: boolean;
    filterConditionValue: string;
}
