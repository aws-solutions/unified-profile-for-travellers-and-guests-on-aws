// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useDispatch, useSelector } from 'react-redux';
import { selectAccpRecords } from '../../../store/configSlice';
import { Select, SelectProps } from '@cloudscape-design/components';
import { setMatchFilterCondition, setSkipCondition } from '../../../store/ruleSetSlice';
import { ConditionType } from '../../../models/config';
import { MatchFilterCondition } from '../../../models/ruleSet';

interface ObjectTypeSelectProps {
    objectType: string;
    objectField: string;
    objectTypeValue?: string;
    otherObjectTypeValue?: string;
    ruleIndex: number;
    conditionIndex: number;
    conditionType: ConditionType;
    disabled: boolean;
}

export const ObjectTypeSelect = ({
    objectType,
    objectField,
    objectTypeValue,
    otherObjectTypeValue,
    ruleIndex,
    conditionIndex,
    conditionType,
    disabled,
}: ObjectTypeSelectProps) => {
    const dispatch = useDispatch();
    const accpRecords = useSelector(selectAccpRecords);
    const accpRecordsWithProfile = [
        {
            name: '_profile',
            struct: {},
        },
        ...accpRecords,
    ];

    const options: SelectProps.Options = accpRecordsWithProfile.map(accpRecord => {
        return { label: accpRecord.name, value: accpRecord.name };
    });

    const selectedObjectType = objectTypeValue ? ({ label: objectTypeValue, value: objectTypeValue } as SelectProps.Option) : null;

    const onSelectChange = ({ detail }: { detail: SelectProps.ChangeDetail }) => {
        if (conditionType === ConditionType.SKIP) {
            dispatch(
                setSkipCondition({
                    condition: { [objectType]: detail.selectedOption.value!, [objectField]: '' },
                    ruleIndex,
                    conditionIndex,
                }),
            );
        } else if (conditionType === ConditionType.MATCH) {
            const condition: Partial<MatchFilterCondition> = { [objectType]: detail.selectedOption.value!, [objectField]: '' };
            if (otherObjectTypeValue === detail.selectedOption.value) {
                condition.filterConditionEnabled = false;
            }
            dispatch(
                setMatchFilterCondition({
                    condition,
                    ruleIndex,
                    conditionIndex,
                }),
            );
        }
    };

    return <Select selectedOption={selectedObjectType} options={options} onChange={onSelectChange} disabled={disabled} />;
};
