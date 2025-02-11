// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { Box, Button, Container, Header, SpaceBetween } from '@cloudscape-design/components';
import { ConditionType, RuleOperation, SkipCondition } from '../../../models/config';
import OperationSelect from '../conditionComponents/skip/operationSelect';
import { ObjectTypeSelect } from '../conditionComponents/objectTypeSelect';
import { ObjectFieldAutosuggest } from '../conditionComponents/objectFieldAutosuggest';
import ObjectValueInput from '../conditionComponents/skip/objectValueInput';
import { IconName } from '../../../models/iconName';
import { useDispatch } from 'react-redux';
import { deleteSkipCondition } from '../../../store/ruleSetSlice';

interface SkipConditionProps {
    condition: SkipCondition;
    ruleIndex: number;
    disabled: boolean;
}

const SkipConditionComparison: React.FC<SkipConditionProps> = ({ condition, ruleIndex, disabled }) => {
    const dispatch = useDispatch();
    const conditionIndex = condition.index;
    const keySuffix = `_rule${ruleIndex}_condition${conditionIndex}`;
    const conditionFields: React.ReactNode[] = [
        <Box key={`conditionIndex${keySuffix}`}>{conditionIndex}.</Box>,
        <ObjectTypeSelect
            objectType="incomingObjectType"
            objectField="incomingObjectField"
            objectTypeValue={condition.incomingObjectType}
            ruleIndex={ruleIndex}
            conditionIndex={conditionIndex}
            conditionType={ConditionType.SKIP}
            key={`incomingObjectType${keySuffix}`}
            disabled={disabled}
        />,
        <ObjectFieldAutosuggest
            objectField="incomingObjectField"
            objectTypeValue={condition.incomingObjectType}
            objectFieldValue={condition.incomingObjectField}
            ruleIndex={ruleIndex}
            conditionIndex={conditionIndex}
            conditionType={ConditionType.SKIP}
            key={`incomingObjectField${keySuffix}`}
            disabled={disabled}
        />,
        <OperationSelect
            op={condition.op}
            ruleIndex={ruleIndex}
            conditionIndex={conditionIndex}
            key={`opField${keySuffix}`}
            disabled={disabled}
        />,
    ];

    // Adding additional fields based on the operation type
    if (condition.op === 'equals' || condition.op === 'not_equals') {
        conditionFields.push(
            <ObjectTypeSelect
                objectType="incomingObjectType2"
                objectField="incomingObjectField2"
                objectTypeValue={condition.incomingObjectType2}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.SKIP}
                key={`incomingObjectType2${keySuffix}`}
                disabled={disabled}
            />,
            <ObjectFieldAutosuggest
                objectField="incomingObjectField2"
                objectTypeValue={condition.incomingObjectType2!}
                objectFieldValue={condition.incomingObjectField2!}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                conditionType={ConditionType.SKIP}
                key={`incomingObjectField2${keySuffix}`}
                disabled={disabled}
            />,
        );
    } else if (
        condition.op === RuleOperation.RULE_OP_EQUALS_VALUE ||
        condition.op === RuleOperation.RULE_OP_NOT_EQUALS_VALUE ||
        condition.op === RuleOperation.RULE_OP_MATCHES_REGEXP
    ) {
        conditionFields.push(
            <ObjectValueInput
                objectValue={condition.skipConditionValue!}
                ruleIndex={ruleIndex}
                conditionIndex={conditionIndex}
                key={`valueInput${keySuffix}`}
                disabled={disabled}
            />,
        );
    }

    // Add delete button if not disabled
    if (!disabled) {
        const onDeleteConditionButtonClick = () => {
            dispatch(deleteSkipCondition({ ruleIndex, conditionIndex }));
        };
        conditionFields.push(
            <Button
                iconName={IconName.REMOVE}
                variant="icon"
                onClick={onDeleteConditionButtonClick}
                key={`deleteConditionButton${keySuffix}`}
            />,
        );
    }

    return (
        <SpaceBetween direction="horizontal" size="xs" key={`spaceBetween${keySuffix}`}>
            {conditionFields}
        </SpaceBetween>
    );
};

export default SkipConditionComparison;
