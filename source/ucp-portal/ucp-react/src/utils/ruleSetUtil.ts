// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import {
    Condition,
    ConditionType,
    FilterCondition,
    IndexNormalizationSettings,
    MatchCondition,
    Rule,
    RuleOperation,
    SkipCondition,
    ValueType,
} from '../models/config.ts';
import { TIMESTAMP } from '../models/constants.ts';
import { DisplayRule, MatchFilterCondition } from '../models/ruleSet.ts';

const defaultIndexNormalization: IndexNormalizationSettings = {
    lowercase: true,
    trim: true,
    removeSpaces: false,
    removeNonNumeric: false,
    removeNonAlphaNumeric: false,
    removeLeadingZeros: false,
};

export function getInitialSkipCondition(conditionIndex: number): SkipCondition {
    return {
        index: conditionIndex,
        conditionType: ConditionType.SKIP,
        incomingObjectType: '',
        incomingObjectField: '',
        op: RuleOperation.RULE_OP_EQUALS,
        indexNormalization: defaultIndexNormalization,
    } as SkipCondition;
}

export function getInitialMatchFilterCondition(conditionIndex: number): MatchFilterCondition {
    return {
        index: conditionIndex,
        conditionType: ConditionType.MATCH,
        incomingObjectType: '',
        incomingObjectField: '',
        existingObjectType: '',
        existingObjectField: '',
        filterConditionEnabled: false,
        filterConditionValue: '0',
        indexNormalization: defaultIndexNormalization,
    } as MatchFilterCondition;
}

type ValidationResult = { isValid: boolean; issue?: string };

export function validateRules(rules: DisplayRule[]): ValidationResult {
    for (const rule of rules) {
        const skipConditions = rule.skipConditions;
        const matchFilterConditions = rule.matchFilterConditions;
        let indexCheck = 0;
        if (matchFilterConditions.length === 0) {
            return {
                isValid: false,
                issue: 'At least one match condition is required',
            };
        }
        for (const condition of skipConditions) {
            if (condition.index !== indexCheck) {
                return {
                    isValid: false,
                    issue: 'Index mismatch found for rule: ' + rule.index + ', skip condition: ' + condition.index,
                };
            }
            const skipValidateResult = validateSkipCondition(condition);
            if (!skipValidateResult.isValid) {
                return skipValidateResult;
            }
            indexCheck++;
        }
        for (const condition of matchFilterConditions) {
            if (condition.index !== indexCheck) {
                return {
                    isValid: false,
                    issue: 'Index mismatch found for rule: ' + rule.index + ', match condition: ' + condition.index,
                };
            }
            const matchFilterValidateResult = validateMatchFilterCondition(condition);
            if (!matchFilterValidateResult.isValid) {
                return matchFilterValidateResult;
            }
            indexCheck++;
        }
    }
    return { isValid: true };
}

export function validateCacheRules(rules: DisplayRule[]): ValidationResult {
    for (const rule of rules) {
        const skipConditions = rule.skipConditions;
        const matchFilterConditions = rule.matchFilterConditions;
        let indexCheck = 0;
        if (matchFilterConditions.length !== 0) {
            return {
                isValid: false,
                issue: 'No Match conditions on cache rule set',
            };
        }
        for (const condition of skipConditions) {
            if (condition.index !== indexCheck) {
                return {
                    isValid: false,
                    issue: 'Index mismatch found for rule: ' + rule.index + ', skip condition: ' + condition.index,
                };
            }
            const skipValidateResult = validateSkipCondition(condition);
            if (!skipValidateResult.isValid) {
                return skipValidateResult;
            }
            indexCheck++;
        }
    }
    return { isValid: true };
}

function validateSkipCondition(skipCondition: SkipCondition): ValidationResult {
    if (!skipCondition.incomingObjectType || !skipCondition.incomingObjectField) {
        return {
            isValid: false,
            issue: 'Incoming object type or field should not be empty',
        };
    }
    if (!skipCondition.op) {
        return {
            isValid: false,
            issue: 'Skip condition operation should not be empty',
        };
    }
    if (
        skipCondition.op === RuleOperation.RULE_OP_EQUALS_VALUE ||
        skipCondition.op === RuleOperation.RULE_OP_NOT_EQUALS_VALUE ||
        skipCondition.op === RuleOperation.RULE_OP_MATCHES_REGEXP
    ) {
        return { isValid: true };
    } else if (skipCondition.op === RuleOperation.RULE_OP_EQUALS || skipCondition.op === RuleOperation.RULE_OP_NOT_EQUALS) {
        if (!skipCondition.incomingObjectType2 || !skipCondition.incomingObjectField2) {
            return {
                isValid: false,
                issue: 'Second incoming object type or field should not be empty',
            };
        }
    } else {
        return {
            isValid: false,
            issue: 'Unsupported operation found: ' + skipCondition.op,
        };
    }
    return { isValid: true };
}

function validateMatchFilterCondition(matchFilterCondition: MatchFilterCondition): ValidationResult {
    if (!matchFilterCondition.incomingObjectType || !matchFilterCondition.incomingObjectField) {
        return {
            isValid: false,
            issue: 'Incoming object type or field should not be empty',
        };
    }
    if (!matchFilterCondition.existingObjectType || !matchFilterCondition.existingObjectField) {
        return {
            isValid: false,
            issue: 'Existing object type or field should not be empty',
        };
    }
    if (matchFilterCondition.filterConditionEnabled) {
        if (!matchFilterCondition.filterConditionValue) {
            return {
                isValid: false,
                issue: 'Filter condition value should not be empty',
            };
        }
        if (Number.parseInt(matchFilterCondition.filterConditionValue) < 0) {
            return {
                isValid: false,
                issue: 'Filter condition value should not be less than 0',
            };
        }
    }
    return { isValid: true };
}

export function rulesToDisplayRules(rules: Rule[]): DisplayRule[] {
    const displayRules: DisplayRule[] = [];
    for (const rule of rules) {
        const conditions = rule.conditions;
        const skipConditions = conditions.filter(cond => cond.conditionType === ConditionType.SKIP) as SkipCondition[];
        const matchConditions = conditions.filter(cond => cond.conditionType === ConditionType.MATCH) as MatchCondition[];
        const filterConditions = conditions.filter(cond => cond.conditionType === ConditionType.FILTER) as FilterCondition[];

        const matchFilterConditions = [];
        for (const matchCondition of matchConditions) {
            matchFilterConditions.push(getMatchFilterCondition(matchCondition));
        }
        for (const filterCondition of filterConditions) {
            const matchFilterCondition = matchFilterConditions.find(
                cond =>
                    cond.incomingObjectType === filterCondition.incomingObjectType &&
                    cond.existingObjectType === filterCondition.existingObjectType,
            );
            if (matchFilterCondition) {
                matchFilterCondition.filterConditionEnabled = true;

                matchFilterCondition.filterConditionValue = filterCondition.filterConditionValue.intValue
                    ? filterCondition.filterConditionValue.intValue.toString()
                    : '0';
            }
        }
        displayRules.push({ index: rule.index, name: rule.name, description: rule.description, skipConditions, matchFilterConditions });
    }
    return displayRules;
}

function getMatchFilterCondition(matchCondition: MatchCondition): MatchFilterCondition {
    return {
        index: matchCondition.index,
        conditionType: ConditionType.MATCH,
        incomingObjectType: matchCondition.incomingObjectType,
        incomingObjectField: matchCondition.incomingObjectField,
        existingObjectType: matchCondition.existingObjectType ? matchCondition.existingObjectType : matchCondition.incomingObjectType,
        existingObjectField: matchCondition.existingObjectField ? matchCondition.existingObjectField : matchCondition.incomingObjectField,
        indexNormalization: defaultIndexNormalization,
        filterConditionEnabled: false,
    } as MatchFilterCondition;
}

export function displayRulesToRules(displayRules: DisplayRule[]): Rule[] {
    const rules: Rule[] = [];
    for (const displayRule of displayRules) {
        const conditions: Condition[] = [...displayRule.skipConditions];
        const matchConditions = [];
        const filterConditions = [];
        const filterKeys = new Set<string>();
        for (const matchFilterCondition of displayRule.matchFilterConditions) {
            // Extracting Match Condition from MatchFilterCondition
            const matchCondition = {
                index: matchFilterCondition.index,
                conditionType: ConditionType.MATCH,
                incomingObjectType: matchFilterCondition.incomingObjectType,
                incomingObjectField: matchFilterCondition.incomingObjectField,
                indexNormalization: defaultIndexNormalization,
            } as MatchCondition;
            if (
                matchFilterCondition.incomingObjectType !== matchFilterCondition.existingObjectType ||
                matchFilterCondition.incomingObjectField !== matchFilterCondition.existingObjectField
            ) {
                matchCondition.existingObjectType = matchFilterCondition.existingObjectType;
                matchCondition.existingObjectField = matchFilterCondition.existingObjectField;
            }
            matchConditions.push(matchCondition);

            // Extracting Filter Condition from MatchFilterCondition if enabled
            const TEMP_INDEX = -1;
            if (matchFilterCondition.filterConditionEnabled) {
                const newFilterKey = matchFilterCondition.incomingObjectType + matchFilterCondition.existingObjectType;
                if (!filterKeys.has(newFilterKey)) {
                    const filterCondition = {
                        index: TEMP_INDEX,
                        conditionType: ConditionType.FILTER,
                        incomingObjectType: matchFilterCondition.incomingObjectType,
                        incomingObjectField: TIMESTAMP,
                        existingObjectType: matchFilterCondition.existingObjectType,
                        existingObjectField: TIMESTAMP,
                        op: RuleOperation.RULE_OP_WITHIN_SECONDS,
                        filterConditionValue: { valueType: ValueType.INT, intValue: Number(matchFilterCondition.filterConditionValue) },
                        indexNormalization: defaultIndexNormalization,
                    } as FilterCondition;
                    filterConditions.push(filterCondition);
                    filterKeys.add(newFilterKey);
                }
            }
        }
        conditions.push(...matchConditions);

        // Recreates filter condition indexing (order does not matter)
        for (let newIndex = conditions.length; newIndex < conditions.length + filterConditions.length; newIndex++) {
            filterConditions[newIndex - conditions.length].index = newIndex;
        }
        conditions.push(...filterConditions);

        rules.push({ index: displayRule.index, name: displayRule.name, description: displayRule.description, conditions });
    }
    return rules;
}
export type { ValidationResult };

