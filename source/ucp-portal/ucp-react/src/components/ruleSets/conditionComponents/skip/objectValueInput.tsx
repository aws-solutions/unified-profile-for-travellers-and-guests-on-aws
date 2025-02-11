// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Input, InputProps } from '@cloudscape-design/components';
import { selectIsEditPage, setSkipCondition } from '../../../../store/ruleSetSlice';
import { useDispatch, useSelector } from 'react-redux';
import { Value, ValueType } from '../../../../models/config';

interface ObjectValueInputProps {
    objectValue: Value;
    ruleIndex: number;
    conditionIndex: number;
    disabled: boolean;
}

export default function ObjectValueInput({ objectValue, ruleIndex, conditionIndex, disabled }: ObjectValueInputProps) {
    const dispatch = useDispatch();
    const currentValue: string = objectValue.stringValue ? objectValue.stringValue : '';
    const onInputChange = ({ detail }: { detail: InputProps.ChangeDetail }) => {
        dispatch(
            setSkipCondition({
                condition: { skipConditionValue: { valueType: ValueType.STRING, stringValue: detail.value } },
                ruleIndex,
                conditionIndex,
            }),
        );
    };
    return <Input value={currentValue} onChange={onInputChange} disabled={disabled} />;
}
