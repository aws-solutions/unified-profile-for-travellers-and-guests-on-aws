// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, Container, Header, SpaceBetween } from '@cloudscape-design/components';
import { SkipCondition } from '../../../models/config';
import { IconName } from '../../../models/iconName';
import { useDispatch } from 'react-redux';
import { addSkipCondition } from '../../../store/ruleSetSlice';
import { getInitialSkipCondition } from '../../../utils/ruleSetUtil';
import SkipConditionComparison from './SkipCondition';

interface SkipConditionContainerProps {
    conditions: SkipCondition[];
    ruleIndex: number;
    addConditionIndex: number;
    disabled: boolean;
}

export default function SkipConditionContainer({ conditions, ruleIndex, addConditionIndex, disabled }: SkipConditionContainerProps) {
    const dispatch = useDispatch();
    const headerText = 'Skip Conditions';
    const header = <Header variant="h2">{headerText}</Header>;

    const conditionComponents = conditions.map(condition => (
        <SkipConditionComparison
            key={`condition_${condition.index}_rule_${ruleIndex}`}
            condition={condition}
            ruleIndex={ruleIndex}
            disabled={disabled}
        />
    ));
    const onAddConditionButtonClick = () => {
        const newSkipCondition = getInitialSkipCondition(addConditionIndex);
        dispatch(addSkipCondition({ condition: newSkipCondition, ruleIndex }));
    };
    if (!disabled) {
        conditionComponents.push(
            <Button
                onClick={onAddConditionButtonClick}
                iconName={IconName.ADD}
                iconAlign="left"
                key={'addSkipConditionButton_rule' + ruleIndex}
            >
                Add New Skip Condition
            </Button>,
        );
    }
    return (
        <Container header={header}>
            <SpaceBetween size="s">{conditionComponents}</SpaceBetween>
        </Container>
    );
}
