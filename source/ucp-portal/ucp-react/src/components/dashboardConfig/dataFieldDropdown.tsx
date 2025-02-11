// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Select } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import { selectAccpRecords, selectDataFieldInput, selectDataRecordInput, setDataFieldInput } from '../../store/configSlice';

export const DataFieldDropdown = () => {
    const dispatch = useDispatch();
    const accpRecords = useSelector(selectAccpRecords);
    const selectedDataRecord = useSelector(selectDataRecordInput);
    const selectedDataField = useSelector(selectDataFieldInput);
    const accpRecordFields = accpRecords.find(accpRecord => accpRecord.name === selectedDataRecord?.label)?.objectFields ?? [];

    const options = [];
    for (const dataField of accpRecordFields) {
        options.push({ label: dataField, value: dataField });
    }
    const isDisabled = selectedDataRecord === null;

    return (
        <Select
            selectedOption={selectedDataField}
            options={options}
            onChange={({ detail }) => dispatch(setDataFieldInput(detail.selectedOption))}
            disabled={isDisabled}
        />
    );
};
