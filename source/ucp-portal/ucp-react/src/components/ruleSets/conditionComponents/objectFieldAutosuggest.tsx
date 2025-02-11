// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useDispatch, useSelector } from 'react-redux';
import { selectAccpRecords } from '../../../store/configSlice';
import { Autosuggest, AutosuggestProps } from '@cloudscape-design/components';
import { selectIsEditPage, selectProfileMappings, setMatchFilterCondition, setSkipCondition } from '../../../store/ruleSetSlice';
import { ConditionType } from '../../../models/config';

interface ObjectFieldAutosuggestProps {
    objectField: string;
    objectTypeValue: string;
    objectFieldValue: string;
    ruleIndex: number;
    conditionIndex: number;
    conditionType: ConditionType;
    disabled: boolean;
}

export const ObjectFieldAutosuggest = ({
    objectField,
    objectTypeValue,
    objectFieldValue,
    ruleIndex,
    conditionIndex,
    conditionType,
    disabled,
}: ObjectFieldAutosuggestProps) => {
    const dispatch = useDispatch();
    const accpRecords = useSelector(selectAccpRecords);
    const profileMappingsRaw = useSelector(selectProfileMappings);
    const profileMappings = profileMappingsRaw.map(mapping => ({ label: mapping, value: mapping }));
    const options = [];
    if (objectTypeValue !== '_profile') {
        const accpRecordFields = accpRecords.find(accpRecord => accpRecord.name === objectTypeValue)?.objectFields ?? [];
        for (const dataField of accpRecordFields) {
            options.push({ label: dataField, value: dataField });
        }
    } else {
        options.push(...profileMappings);
    }

    const onAutosuggestChange = ({ detail }: { detail: AutosuggestProps.ChangeDetail }) => {
        if (conditionType === ConditionType.SKIP) {
            dispatch(
                setSkipCondition({
                    condition: { [objectField]: detail.value },
                    ruleIndex,
                    conditionIndex,
                }),
            );
        } else if (conditionType === ConditionType.MATCH) {
            dispatch(
                setMatchFilterCondition({
                    condition: { [objectField]: detail.value },
                    ruleIndex,
                    conditionIndex,
                }),
            );
        }
    };

    return <Autosuggest value={objectFieldValue} options={options} onChange={onAutosuggestChange} disabled={disabled} />;
};
