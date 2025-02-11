// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Container, Header, SpaceBetween } from '@cloudscape-design/components';
import { ObjectFieldAutosuggest } from '../conditionComponents/objectFieldAutosuggest';
import { ObjectTypeSelect } from '../conditionComponents/objectTypeSelect';
import { ReactNode } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { addMatchFilterCondition, deleteMatchFilterCondition, selectIsEditPage } from '../../../store/ruleSetSlice';
import { IconName } from '../../../models/iconName';
import { MatchFilterCondition } from '../../../models/ruleSet';
import { getInitialMatchFilterCondition } from '../../../utils/ruleSetUtil';
import TimestampFilter from '../conditionComponents/matchFilter/timestampFilter';
import { ConditionType } from '../../../models/config';

interface MatchConditionContainerProps {
    matchFilterConditions: MatchFilterCondition[];
    ruleIndex: number;
    addConditionIndex: number;
    disabled: boolean;
}

export default function MatchFilterConditionContainer({
    matchFilterConditions,
    ruleIndex,
    addConditionIndex,
    disabled,
}: MatchConditionContainerProps) {
    const dispatch = useDispatch();
    const headerText = 'Match Conditions';
    const header = <Header variant="h2">{headerText}</Header>;
    const conditionComponents: ReactNode[] = [];
    const isEditPage = useSelector(selectIsEditPage);
    for (const condition of matchFilterConditions) {
        const conditionFields: ReactNode[] = [];
        const conditionIndex = condition.index;
        const keySuffix = '_rule' + ruleIndex + '_condition' + conditionIndex;
        const filterCheckBoxDisabled =
            condition.incomingObjectType === '' ||
            condition.existingObjectType === '' ||
            condition.incomingObjectType === condition.existingObjectType;
        conditionFields.push(
            <Box key={'conditionIndex' + keySuffix}>{conditionIndex}.</Box>,
            <ObjectTypeSelect
                objectType="incomingObjectType"
                objectField="incomingObjectField"
                objectTypeValue={condition.incomingObjectType}
                otherObjectTypeValue={condition.existingObjectType}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.MATCH}
                key={'incomingObjectType' + keySuffix}
                disabled={disabled}
            />,
            <ObjectFieldAutosuggest
                objectField="incomingObjectField"
                objectTypeValue={condition.incomingObjectType}
                objectFieldValue={condition.incomingObjectField}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.MATCH}
                key={'incomingObjectField' + keySuffix}
                disabled={disabled}
            />,
            <Box key={'equalOp' + keySuffix}>Equals</Box>,
            <ObjectTypeSelect
                objectType="existingObjectType"
                objectField="existingObjectField"
                objectTypeValue={condition.existingObjectType}
                otherObjectTypeValue={condition.incomingObjectType}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.MATCH}
                key={'existingObjectType' + keySuffix}
                disabled={disabled}
            />,
            <ObjectFieldAutosuggest
                objectField="existingObjectField"
                objectTypeValue={condition.existingObjectType}
                objectFieldValue={condition.existingObjectField}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.MATCH}
                key={'existingObjectField' + keySuffix}
                disabled={disabled}
            />,
            <TimestampFilter
                checkBoxDisabled={filterCheckBoxDisabled}
                filterEnabled={condition.filterConditionEnabled}
                filterValue={condition.filterConditionValue}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                key={'timestampFilter' + keySuffix}
            />,
        );
        if (isEditPage) {
            const onDeleteConditionButtonClick = () => {
                dispatch(deleteMatchFilterCondition({ ruleIndex, conditionIndex }));
            };
            conditionFields.push(
                <Button
                    iconName={IconName.REMOVE}
                    variant="icon"
                    onClick={onDeleteConditionButtonClick}
                    key={'deleteConditionButton' + keySuffix}
                />,
            );
        }
        conditionComponents.push(
            <SpaceBetween direction="horizontal" size="xs" key={'spaceBetween' + keySuffix}>
                {conditionFields}
            </SpaceBetween>,
        );
    }
    if (isEditPage) {
        const onAddConditionButtonClick = () => {
            const newSkipCondition = getInitialMatchFilterCondition(addConditionIndex);
            dispatch(addMatchFilterCondition({ condition: newSkipCondition, ruleIndex }));
        };
        conditionComponents.push(
            <Button
                onClick={onAddConditionButtonClick}
                iconName={IconName.ADD}
                iconAlign="left"
                key={'addMatchConditionButton_rule' + ruleIndex}
            >
                Add New Match Condition
            </Button>,
        );
    }
    return (
        <Container header={header}>
            <SpaceBetween size="s">{conditionComponents}</SpaceBetween>
        </Container>
    );
}
