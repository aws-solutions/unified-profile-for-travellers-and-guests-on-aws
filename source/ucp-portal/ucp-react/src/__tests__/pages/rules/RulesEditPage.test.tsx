// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { act } from '@testing-library/react';
import { ConditionType, IndexNormalizationSettings, RuleOperation, SkipCondition, ValueType } from '../../../models/config';
import { ROUTES } from '../../../models/constants.ts';
import { DisplayRule, MatchFilterCondition } from '../../../models/ruleSet.ts';
import { renderAppContent } from '../../test-utils.tsx';

it('Opens rules edit page', async () => {
    const defaultIndexNormalization: IndexNormalizationSettings = {
        lowercase: true,
        trim: true,
        removeSpaces: false,
        removeNonNumeric: false,
        removeNonAlphaNumeric: false,
        removeLeadingZeros: false,
    };

    const skipCondition: SkipCondition = {
        index: 0,
        conditionType: ConditionType.SKIP,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        op: RuleOperation.RULE_OP_EQUALS,
        incomingObjectType2: 'dummy_type2',
        incomingObjectField2: 'dummy_field2',
        indexNormalization: defaultIndexNormalization,
    };
    const skipCondition2: SkipCondition = {
        index: 1,
        conditionType: ConditionType.SKIP,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        skipConditionValue: { valueType: ValueType.STRING, stringValue: 'dummy_value' },
        op: RuleOperation.RULE_OP_EQUALS_VALUE,
        indexNormalization: defaultIndexNormalization,
    };
    const matchFilterCondition: MatchFilterCondition = {
        index: 2,
        conditionType: ConditionType.MATCH,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        existingObjectType: 'dummy_type2',
        existingObjectField: 'dummy_field2',
        filterConditionEnabled: true,
        filterConditionValue: '10',
        indexNormalization: defaultIndexNormalization,
    };
    const testRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [matchFilterCondition],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];

    await act(async () => {
        renderAppContent({
            initialRoute: `/${ROUTES.RULES_EDIT}`,
            preloadedState: {
                ruleSet: {
                    allRuleSets: [],
                    allRuleSetsCache: [],
                    editableRuleSet: testRules,
                    selectedRuleSet: 'active',
                    selectedRuleSetCache: 'active',
                    isEditPage: true,
                    profileMappings: [],
                },
            },
        });
    });
});
