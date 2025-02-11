// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { InputProps, NonCancelableCustomEvent } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import BaseInput from '../../base/form/baseInput';

type StateType = any;

interface GenericInputProps {
    actionCreator: (value: string) => { payload: string; type: string };
    selector: (state: StateType) => string;
    labelName: string;
}

export default function GenericInput({ actionCreator, selector, labelName }: GenericInputProps) {
    const dispatch = useDispatch();
    const inputValue = useSelector(selector);

    const onChange = ({ detail }: NonCancelableCustomEvent<InputProps.ChangeDetail>): void => {
        dispatch(actionCreator(detail.value));
    };

    return <BaseInput label={labelName} value={inputValue} onChange={onChange} />;
}
