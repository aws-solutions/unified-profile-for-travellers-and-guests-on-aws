// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Input } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import { selectUrlTemplateInput, setUrlTemplateInput } from '../../store/configSlice';

export const UrlTemplateInput = () => {
    const dispatch = useDispatch();
    const urlTemplateInput = useSelector(selectUrlTemplateInput);
    const isValid = isValidHttpsUrl(urlTemplateInput, true);

    return (
        <Input
            onChange={({ detail }) => dispatch(setUrlTemplateInput(detail.value))}
            value={urlTemplateInput}
            inputMode="url"
            placeholder="https://example.com/{{id}}"
            invalid={!isValid}
        />
    );
};

function isValidHttpsUrl(input: string, allowEmptyUrls: boolean) {
    if (allowEmptyUrls && input === '') {
        return true;
    }
    return /^https:\/\/.*/.test(input);
}
