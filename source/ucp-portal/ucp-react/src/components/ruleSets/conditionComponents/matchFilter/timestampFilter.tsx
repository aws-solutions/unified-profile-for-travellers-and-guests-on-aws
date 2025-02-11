// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Checkbox, CheckboxProps, Input, InputProps, SpaceBetween } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import { setMatchFilterCondition, selectIsEditPage } from '../../../../store/ruleSetSlice';

interface TimestampFilterProps {
    checkBoxDisabled: boolean;
    filterEnabled: boolean;
    filterValue: string;
    ruleIndex: number;
    conditionIndex: number;
}
export default function TimestampFilter({ checkBoxDisabled, filterEnabled, filterValue, ruleIndex, conditionIndex }: TimestampFilterProps) {
    const dispatch = useDispatch();
    const isEditPage = useSelector(selectIsEditPage);

    const onCheckClick = ({ detail }: { detail: CheckboxProps.ChangeDetail }) => {
        dispatch(setMatchFilterCondition({ condition: { filterConditionEnabled: detail.checked }, ruleIndex, conditionIndex }));
    };

    const onInputChange = ({ detail }: { detail: InputProps.ChangeDetail }) => {
        dispatch(
            setMatchFilterCondition({
                condition: { filterConditionValue: detail.value },
                ruleIndex,
                conditionIndex,
            }),
        );
    };
    return (
        <SpaceBetween direction="horizontal" size="s">
            <Checkbox onChange={onCheckClick} checked={filterEnabled} disabled={!isEditPage || checkBoxDisabled}>
                Timestamp Condition Enabled?
            </Checkbox>
            {filterEnabled && (
                <>
                    <Input
                        disabled={!isEditPage || checkBoxDisabled}
                        onChange={onInputChange}
                        value={filterValue}
                        type="number"
                        placeholder="timestamp delta"
                    />
                    <Box>minutes</Box>
                </>
            )}
        </SpaceBetween>
    );
}
