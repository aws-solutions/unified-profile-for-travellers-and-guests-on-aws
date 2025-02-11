// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useDispatch, useSelector } from 'react-redux';
import { selectAccpRecords, selectDataRecordInput, setDataFieldInput, setDataRecordInput } from '../../store/configSlice';
import { Select, SelectProps } from '@cloudscape-design/components';

export const DataRecordDropdown = () => {
    const dispatch = useDispatch();
    const accpRecords = useSelector(selectAccpRecords);
    const selectedDataRecord = useSelector(selectDataRecordInput);
    const options: SelectProps.Options = accpRecords.map(accpRecord => {
        return { label: accpRecord.name, value: accpRecord.name };
    });

    return (
        <Select
            selectedOption={selectedDataRecord}
            options={options}
            onChange={({ detail }) => {
                dispatch(setDataRecordInput(detail.selectedOption));
                dispatch(setDataFieldInput(null));
            }}
        />
    );
};
