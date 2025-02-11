// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, ExpandableSectionProps, SpaceBetween } from '@cloudscape-design/components';
import { SkipCondition } from '../../models/config';
import SkipConditionContainer from './conditionContainers/skipConditionContainer';
import MatchFilterConditionContainer from './conditionContainers/matchFilterConditionContainer';
import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { deleteRule, selectIsEditPage } from '../../store/ruleSetSlice';
import { IconName } from '../../models/iconName';
import { DisplayRule, MatchFilterCondition } from '../../models/ruleSet';

interface RuleContainerProps {
    ruleIndex: number;
    rule: DisplayRule;
    suppressMatch: boolean;
}

export default function RuleContainer({ ruleIndex, rule, suppressMatch }: RuleContainerProps) {
    const dispatch = useDispatch();
    const headerText = 'Rule ' + ruleIndex;
    const skipConditions: SkipCondition[] = rule.skipConditions;
    const matchFilterConditions: MatchFilterCondition[] = rule.matchFilterConditions;
    const isEditPage = useSelector(selectIsEditPage);

    const [headerDescription, setHeaderDescription] = useState<string>('');
    const onExpandChange = ({ detail }: { detail: ExpandableSectionProps.ChangeDetail }) => {
        if (detail.expanded) {
            setHeaderDescription('');
        } else {
            setHeaderDescription(skipConditions.length + ' Skip Condition(s) and ' + matchFilterConditions.length + ' Match Condition(s)');
        }
    };

    const onDeleteRuleButtonClick = () => {
        dispatch(deleteRule({ ruleIndex }));
    };
    const deleteRuleButton = isEditPage ? (
        <Button iconName={IconName.REMOVE} onClick={onDeleteRuleButtonClick}>
            Delete Rule
        </Button>
    ) : null;
    return (
        <ExpandableSection
            headerText={headerText}
            headerDescription={headerDescription}
            headerActions={deleteRuleButton}
            defaultExpanded
            variant="container"
            onChange={onExpandChange}
        >
            <SpaceBetween size="s">
                <SkipConditionContainer
                    conditions={skipConditions}
                    ruleIndex={ruleIndex}
                    addConditionIndex={skipConditions.length}
                    disabled={!isEditPage}
                />
                {!suppressMatch && (
                    <MatchFilterConditionContainer
                        matchFilterConditions={matchFilterConditions}
                        ruleIndex={ruleIndex}
                        addConditionIndex={skipConditions.length + matchFilterConditions.length}
                        disabled={!isEditPage}
                    />
                )}
            </SpaceBetween>
        </ExpandableSection>
    );
}
