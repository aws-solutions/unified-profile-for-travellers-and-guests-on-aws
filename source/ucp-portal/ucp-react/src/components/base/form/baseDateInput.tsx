// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { DateInput, InputProps } from '@cloudscape-design/components';
import { DATE_FORMAT } from '../../../models/constants';
import BaseFormField, { BaseFormFieldProps } from './baseFormField';

export interface BaseDateInputProps extends BaseFormFieldProps<InputProps.ChangeDetail> {
    value: string;
}

/**
 *
 * @param root0
 * @param root0.label
 * @param root0.stretch
 * @param root0.errorText
 * @param root0.value
 * @param root0.onChange
 * @param root0.disabled
 * @param root0.autoFocus
 * @param root0.invalid
 * @param root0.readOnly
 * @param root0.required
 * @param root0.constraintText
 */
export default function BaseDateInput({
    label,
    stretch,
    errorText,
    value,
    onChange,
    disabled,
    autoFocus,
    invalid,
    readOnly,
    required,
    constraintText = '',
}: BaseDateInputProps) {
    return (
        <>
            <BaseFormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                <DateInput
                    onChange={onChange}
                    value={value}
                    placeholder={`${DATE_FORMAT}`}
                    disabled={disabled}
                    invalid={invalid}
                    readOnly={readOnly}
                    ariaRequired={required}
                    autoFocus={autoFocus}
                />
            </BaseFormField>
        </>
    );
}
