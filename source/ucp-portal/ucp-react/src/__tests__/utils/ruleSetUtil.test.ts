// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import {
    ConditionType,
    FilterCondition,
    IndexNormalizationSettings,
    MatchCondition,
    Rule,
    RuleOperation,
    SkipCondition,
    ValueType,
} from '../../models/config.ts';
import { DisplayRule, MatchFilterCondition } from '../../models/ruleSet.ts';
import {
    displayRulesToRules,
    getInitialMatchFilterCondition,
    getInitialSkipCondition,
    rulesToDisplayRules,
    validateCacheRules,
    validateRules,
} from '../../utils/ruleSetUtil.ts';

describe('Rule Set Util Functions', async () => {
    const defaultIndexNormalization: IndexNormalizationSettings = {
        lowercase: true,
        trim: true,
        removeSpaces: false,
        removeNonNumeric: false,
        removeNonAlphaNumeric: false,
        removeLeadingZeros: false,
    };
    it('getInitialSkipCondition | success', async () => {
        // arrange
        const conditionIndex = 1;

        // act
        const skipCondition: SkipCondition = getInitialSkipCondition(conditionIndex);

        // assert
        expect(skipCondition.index).toBe(conditionIndex);
        expect(skipCondition.conditionType).toBe(ConditionType.SKIP);
        expect(JSON.stringify(skipCondition.indexNormalization)).toBe(JSON.stringify(defaultIndexNormalization));
    });

    it('getInitialMatchFilterCondition | success', async () => {
        // arrange
        const conditionIndex = 1;

        // act
        const matchFilterCondition: MatchFilterCondition = getInitialMatchFilterCondition(conditionIndex);

        // assert
        expect(matchFilterCondition.index).toBe(conditionIndex);
        expect(matchFilterCondition.conditionType).toBe(ConditionType.MATCH);
        expect(JSON.stringify(matchFilterCondition.indexNormalization)).toBe(JSON.stringify(defaultIndexNormalization));
        expect(matchFilterCondition.filterConditionEnabled).toBe(false);
    });

    it('validateRules | successfully validate one of each condition', async () => {
        // arrange
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
            indexNormalization: defaultIndexNormalization,
            filterConditionEnabled: true,
            filterConditionValue: '10',
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [skipCondition, skipCondition2],
                matchFilterConditions: [matchFilterCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(true);
        expect(validationResult.issue).toBe(undefined);
    });

    it('validateRules | failure | at least one match filter condition', async () => {
        // arrange
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
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [skipCondition],
                matchFilterConditions: [],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('At least one match condition is required');
    });

    it('validateRules skip condition | failure | missing incoming object type', async () => {
        // arrange
        const emptyIncomingTypeCondition: SkipCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: '',
            incomingObjectField: '',
            op: RuleOperation.RULE_OP_EQUALS,
            incomingObjectType2: 'dummy_type2',
            incomingObjectField2: 'dummy_field2',
            indexNormalization: defaultIndexNormalization,
        };
        const defaultMatchFilterCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [emptyIncomingTypeCondition],
                matchFilterConditions: [defaultMatchFilterCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Incoming object type or field should not be empty');
    });

    it('validateRules skip condition | failure | missing op field', async () => {
        // arrange
        const missingOpCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            incomingObjectType2: '',
            incomingObjectField2: '',
            indexNormalization: defaultIndexNormalization,
        };
        const defaultMatchFilterCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [missingOpCondition as SkipCondition],
                matchFilterConditions: [defaultMatchFilterCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Skip condition operation should not be empty');
    });

    it('validateRules skip condition | failure | missing second incoming object type/field', async () => {
        // arrange
        const emptySecondIncomingTypeCondition: SkipCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            op: RuleOperation.RULE_OP_EQUALS,
            incomingObjectType2: '',
            incomingObjectField2: '',
            indexNormalization: defaultIndexNormalization,
        };
        const defaultMatchFilterCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [emptySecondIncomingTypeCondition],
                matchFilterConditions: [defaultMatchFilterCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Second incoming object type or field should not be empty');
    });

    it('validateRules skip condition | failure | skipConditionValue empty', async () => {
        // arrange
        const skipValueEmptyCondition: SkipCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            op: RuleOperation.RULE_OP_EQUALS_VALUE,
            skipConditionValue: { valueType: ValueType.STRING, stringValue: '' },
            indexNormalization: defaultIndexNormalization,
        };
        const defaultMatchFilterCondition: MatchFilterCondition = {
            index: 1,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [skipValueEmptyCondition],
                matchFilterConditions: [defaultMatchFilterCondition],
            },
        ];
        const rulesCache: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [skipValueEmptyCondition],
                matchFilterConditions: [],
            },
        ];

        // act
        const validationResult = validateRules(rules);
        const validationResultCache = validateCacheRules(rulesCache);

        // assert
        expect(validationResult.isValid).toBe(true);
        expect(validationResultCache.isValid).toBe(true);
    });

    it('validateRules match filter condition | failure | missing incoming object type', async () => {
        // arrange
        const emptyIncomingTypeCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: '',
            incomingObjectField: '',
            existingObjectType: 'dummy_type',
            existingObjectField: 'dummy_field',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [],
                matchFilterConditions: [emptyIncomingTypeCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Incoming object type or field should not be empty');
    });

    it('validateRules match filter condition | failure | existing object type empty', async () => {
        // arrange
        const emptyExistingTypeCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: '',
            existingObjectField: '',
            filterConditionEnabled: false,
            filterConditionValue: '10',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [],
                matchFilterConditions: [emptyExistingTypeCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Existing object type or field should not be empty');
    });

    it('validateRules match filter condition | failure | filter condition value should not be empty', async () => {
        // arrange
        const emptyFilterValueCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_field',
            existingObjectField: 'dummy_type',
            filterConditionEnabled: true,
            filterConditionValue: '',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [],
                matchFilterConditions: [emptyFilterValueCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Filter condition value should not be empty');
    });

    it('validateRules match filter condition | failure | filter condition value should not be less than 0', async () => {
        // arrange
        const emptyFilterValueCondition: MatchFilterCondition = {
            index: 0,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_field',
            existingObjectField: 'dummy_type',
            filterConditionEnabled: true,
            filterConditionValue: '-1',
            indexNormalization: defaultIndexNormalization,
        };
        const rules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [],
                matchFilterConditions: [emptyFilterValueCondition],
            },
        ];

        // act
        const validationResult = validateRules(rules);

        // assert
        expect(validationResult.isValid).toBe(false);
        expect(validationResult.issue).toBe('Filter condition value should not be less than 0');
    });

    it('rulesToDisplayRules | success | rules of each type translated to display rules', async () => {
        // arrange
        const skipCondition: SkipCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            op: RuleOperation.RULE_OP_EQUALS_VALUE,
            skipConditionValue: { valueType: ValueType.STRING, stringValue: 'dummy_value' },
            indexNormalization: defaultIndexNormalization,
        };
        const matchCondition: MatchCondition = {
            index: 1,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            indexNormalization: defaultIndexNormalization,
        };
        const filterCondition: FilterCondition = {
            index: 2,
            conditionType: ConditionType.FILTER,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            op: RuleOperation.RULE_OP_WITHIN_SECONDS,
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            filterConditionValue: { valueType: ValueType.INT, intValue: 10 },
            indexNormalization: defaultIndexNormalization,
        };
        const matchFilterCondition: MatchFilterCondition = {
            index: 1,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            indexNormalization: defaultIndexNormalization,
            filterConditionEnabled: true,
            filterConditionValue: '10',
        };
        const rules: Rule[] = [
            {
                index: 0,
                name: '',
                description: '',
                conditions: [skipCondition, matchCondition, filterCondition],
            },
        ];

        // act
        const translateResults = rulesToDisplayRules(rules);

        // assert
        expect(translateResults.length).toBe(1);
        expect(translateResults[0].skipConditions.length).toBe(1);
        expect(translateResults[0].matchFilterConditions.length).toBe(1);
        expect(JSON.stringify(translateResults[0].skipConditions[0])).toBe(JSON.stringify(skipCondition));
        expect(JSON.stringify(translateResults[0].matchFilterConditions[0])).toBe(JSON.stringify(matchFilterCondition));
    });

    it('displayRulesToRules | success | display rules of each type translated to rules', async () => {
        // arrange
        const skipCondition: SkipCondition = {
            index: 0,
            conditionType: ConditionType.SKIP,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            op: RuleOperation.RULE_OP_EQUALS_VALUE,
            skipConditionValue: { valueType: ValueType.STRING, stringValue: 'dummy_value' },
            indexNormalization: defaultIndexNormalization,
        };
        const matchCondition: MatchCondition = {
            index: 1,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            indexNormalization: defaultIndexNormalization,
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
        };
        const filterCondition: FilterCondition = {
            index: 2,
            conditionType: ConditionType.FILTER,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'timestamp',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'timestamp',
            op: RuleOperation.RULE_OP_WITHIN_SECONDS,
            filterConditionValue: { valueType: ValueType.INT, intValue: 10 },
            indexNormalization: defaultIndexNormalization,
        };
        const matchFilterCondition: MatchFilterCondition = {
            index: 1,
            conditionType: ConditionType.MATCH,
            incomingObjectType: 'dummy_type',
            incomingObjectField: 'dummy_field',
            existingObjectType: 'dummy_type2',
            existingObjectField: 'dummy_field2',
            indexNormalization: defaultIndexNormalization,
            filterConditionEnabled: true,
            filterConditionValue: '10',
        };
        const rules: Rule[] = [
            {
                index: 0,
                name: '',
                description: '',
                conditions: [skipCondition, matchCondition, filterCondition],
            },
        ];
        const displayRules: DisplayRule[] = [
            {
                index: 0,
                name: '',
                description: '',
                skipConditions: [skipCondition],
                matchFilterConditions: [matchFilterCondition],
            },
        ];

        // act
        const translateResults = displayRulesToRules(displayRules);

        // assert
        expect(translateResults.length).toBe(1);
        expect(translateResults[0].conditions.length).toBe(3);
        const conditions = translateResults[0].conditions;
        expect(conditions[0].conditionType).toBe(ConditionType.SKIP);
        expect(conditions[0].index).toBe(0);
        expect(JSON.stringify(conditions[0])).toBe(JSON.stringify(skipCondition));

        expect(conditions[1].conditionType).toBe(ConditionType.MATCH);
        expect(conditions[1].index).toBe(1);
        expect(JSON.stringify(conditions[1])).toBe(JSON.stringify(matchCondition));

        expect(conditions[2].conditionType).toBe(ConditionType.FILTER);
        expect(conditions[2].index).toBe(2);
        expect(JSON.stringify(conditions[2])).toBe(JSON.stringify(filterCondition));
    });
});
