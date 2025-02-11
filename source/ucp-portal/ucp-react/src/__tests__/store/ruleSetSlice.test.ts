// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { expect, test } from 'vitest';
import reducer, {
    setEditableRuleSet,
    clearEditableRuleSet,
    setSkipCondition,
    setMatchFilterCondition,
    addSkipCondition,
    addMatchFilterCondition,
    deleteSkipCondition,
    deleteMatchFilterCondition,
    addRule,
    deleteRule,
    setSelectedRuleSet,
    setIsEditPage,
} from '../../store/ruleSetSlice';
import { ConditionType, IndexNormalizationSettings, RuleOperation, SkipCondition, ValueType } from '../../models/config';
import { DisplayRule, MatchFilterCondition } from '../../models/ruleSet';

const initialState = {
    editableRuleSet: [],
    allRuleSets: [],
    allRuleSetsCache: [],
    selectedRuleSet: 'active',
    selectedRuleSetCache: "active",
    isEditPage: false,
    profileMappings: [],
};

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
const matchFilterCondition2: MatchFilterCondition = {
    index: 3,
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
        matchFilterConditions: [matchFilterCondition, matchFilterCondition2],
    },
    {
        index: 1,
        name: '',
        description: '',
        skipConditions: [skipCondition, skipCondition2],
        matchFilterConditions: [],
    },
];

const fullState = {
    editableRuleSet: testRules,
    allRuleSets: [],
    allRuleSetsCache: [],
    selectedRuleSet: 'active',
    selectedRuleSetCache: "active",
    isEditPage: false,
    profileMappings: [],
};

test('Should Return Initial State', () => {
    expect(reducer(undefined, { type: undefined })).toEqual(initialState);
});

test('Should set a edit rule set to full state', () => {
    expect(reducer(initialState, setEditableRuleSet(testRules))).toEqual(fullState);
});

test('Should clear editable rule set', () => {
    expect(reducer(fullState, clearEditableRuleSet())).toEqual(initialState);
});

test('Should set a skip condition field to a different value', () => {
    const input = {
        condition: {
            incomingObjectField: 'new_field',
        } as Partial<SkipCondition>,
        ruleIndex: 0,
        conditionIndex: 0,
    };
    const updatedCondition: SkipCondition = {
        index: 0,
        conditionType: ConditionType.SKIP,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'new_field',
        op: RuleOperation.RULE_OP_EQUALS,
        incomingObjectType2: 'dummy_type2',
        incomingObjectField2: 'dummy_field2',
        indexNormalization: defaultIndexNormalization,
    };
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [updatedCondition, skipCondition2],
            matchFilterConditions: [matchFilterCondition, matchFilterCondition2],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];

    expect(reducer(fullState, setSkipCondition(input))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should set a match filter condition field to a different value', () => {
    const input = {
        condition: {
            incomingObjectField: 'new_field',
            existingObjectField: 'new_field2',
        } as Partial<MatchFilterCondition>,
        ruleIndex: 0,
        conditionIndex: 2,
    };
    const updatedCondition: MatchFilterCondition = {
        index: 2,
        conditionType: ConditionType.MATCH,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'new_field',
        existingObjectType: 'dummy_type2',
        existingObjectField: 'new_field2',
        filterConditionEnabled: true,
        filterConditionValue: '10',
        indexNormalization: defaultIndexNormalization,
    };
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [updatedCondition, matchFilterCondition2],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];

    expect(reducer(fullState, setMatchFilterCondition(input))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should add a new skip condition', () => {
    const newCondition: SkipCondition = {
        index: 2,
        conditionType: ConditionType.SKIP,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        skipConditionValue: { valueType: ValueType.STRING, stringValue: 'dummy_value' },
        op: RuleOperation.RULE_OP_EQUALS_VALUE,
        indexNormalization: defaultIndexNormalization,
    };
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [matchFilterCondition, matchFilterCondition2],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2, newCondition],
            matchFilterConditions: [],
        },
    ];
    expect(reducer(fullState, addSkipCondition({ condition: newCondition, ruleIndex: 1 }))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should add a new skip condition', () => {
    const newCondition: MatchFilterCondition = {
        index: 4,
        conditionType: ConditionType.MATCH,
        incomingObjectType: '',
        incomingObjectField: '',
        existingObjectType: '',
        existingObjectField: '',
        filterConditionEnabled: false,
        filterConditionValue: '0',
        indexNormalization: defaultIndexNormalization,
    };
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [matchFilterCondition, matchFilterCondition2, newCondition],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];
    expect(reducer(fullState, addMatchFilterCondition({ condition: newCondition, ruleIndex: 0 }))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should delete a skip condition', () => {
    const newSkipCondition2: SkipCondition = {
        index: 0,
        conditionType: ConditionType.SKIP,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        skipConditionValue: { valueType: ValueType.STRING, stringValue: 'dummy_value' },
        op: RuleOperation.RULE_OP_EQUALS_VALUE,
        indexNormalization: defaultIndexNormalization,
    };
    const newMatchFilterCondition: MatchFilterCondition = {
        index: 1,
        conditionType: ConditionType.MATCH,
        incomingObjectType: 'dummy_type',
        incomingObjectField: 'dummy_field',
        existingObjectType: 'dummy_type2',
        existingObjectField: 'dummy_field2',
        filterConditionEnabled: true,
        filterConditionValue: '10',
        indexNormalization: defaultIndexNormalization,
    };
    const newMatchFilterCondition2: MatchFilterCondition = {
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
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [newSkipCondition2],
            matchFilterConditions: [newMatchFilterCondition, newMatchFilterCondition2],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];
    expect(reducer(fullState, deleteSkipCondition({ ruleIndex: 0, conditionIndex: 0 }))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should delete a match filter condition', () => {
    const newMatchFilterCondition2: MatchFilterCondition = {
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
    const updatedTestRules: DisplayRule[] = [
        {
            index: 0,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [newMatchFilterCondition2],
        },
        {
            index: 1,
            name: '',
            description: '',
            skipConditions: [skipCondition, skipCondition2],
            matchFilterConditions: [],
        },
    ];
    expect(reducer(fullState, deleteMatchFilterCondition({ ruleIndex: 0, conditionIndex: 2 }))).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should add a new rule', () => {
    const newRule: DisplayRule = {
        index: 2,
        name: '',
        description: '',
        skipConditions: [],
        matchFilterConditions: [],
    };
    const updatedTestRules = [...testRules, newRule];
    expect(reducer(fullState, addRule())).toEqual({
        ...fullState,
        editableRuleSet: updatedTestRules,
    });
});

test('Should delete a rule', () => {
    const updatedTestRules: DisplayRule = {
        index: 0,
        name: '',
        description: '',
        skipConditions: [skipCondition, skipCondition2],
        matchFilterConditions: [],
    };
    expect(reducer(fullState, deleteRule({ ruleIndex: 0 }))).toEqual({
        ...fullState,
        editableRuleSet: [updatedTestRules],
    });
});

test('Should update selected rule set', () => {
    expect(reducer(initialState, setSelectedRuleSet('new_rule_set'))).toEqual({
        ...initialState,
        selectedRuleSet: 'new_rule_set',
    });
});

test('Should set isEditPage', () => {
    expect(reducer(initialState, setIsEditPage(true))).toEqual({
        ...initialState,
        isEditPage: true,
    });
});
