// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Select, SelectProps } from '@cloudscape-design/components';
import { RuleOperation, ValueType } from '../../../../models/config';
import { selectIsEditPage, setSkipCondition } from '../../../../store/ruleSetSlice';
import { useDispatch, useSelector } from 'react-redux';

interface OperationSelectProps {
    op: RuleOperation;
    ruleIndex: number;
    conditionIndex: number;
    disabled: boolean;
}

export default function OperationSelect({ op, ruleIndex, conditionIndex, disabled }: OperationSelectProps) {
    const dispatch = useDispatch();
    const isEditPage = useSelector(selectIsEditPage);
    const selectedOption: SelectProps.Option = op
        ? {
              label: op,
              value: op,
          }
        : { label: '', value: '' };
    const selectOptions = [
        { label: 'Equals', value: RuleOperation.RULE_OP_EQUALS },
        { label: 'Equals Value', value: RuleOperation.RULE_OP_EQUALS_VALUE },
        { label: 'Not Equals', value: RuleOperation.RULE_OP_NOT_EQUALS },
        {
            label: 'Not Equals Value',
            value: RuleOperation.RULE_OP_NOT_EQUALS_VALUE,
        },
        {
            label: 'Regexp',
            value: RuleOperation.RULE_OP_MATCHES_REGEXP,
        },
    ];
    const onSelectChange = ({ detail }: { detail: SelectProps.ChangeDetail }) => {
        dispatch(
            setSkipCondition({
                condition: {
                    op: detail.selectedOption.value as RuleOperation,
                    incomingObjectType2: '',
                    incomingObjectField2: '',
                    skipConditionValue: { valueType: ValueType.STRING, stringValue: '' },
                },
                ruleIndex,
                conditionIndex,
            }),
        );
    };
    return <Select selectedOption={selectedOption} onChange={onSelectChange} options={selectOptions} disabled={disabled} />;
}
