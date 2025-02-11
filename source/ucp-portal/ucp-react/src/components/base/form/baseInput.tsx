// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Input, InputProps } from '@cloudscape-design/components';
import BaseFormField, { BaseFormFieldProps } from './baseFormField';

interface BaseInputProps extends BaseFormFieldProps<InputProps.ChangeDetail> {
    value: string;
    placeHolder?: string;
    inputMode?: InputProps.InputMode;
    type?: InputProps.Type;
    autoComplete?: boolean | string;
}

/**
 *
 * @param root0
 * @param root0.label
 * @param root0.value
 * @param root0.placeHolder
 * @param root0.disabled
 * @param root0.inputMode
 * @param root0.type
 * @param root0.stretch
 * @param root0.constraintText
 * @param root0.errorText
 * @param root0.onChange
 * @param root0.required
 * @param root0.autoComplete
 * @param root0.autoFocus
 * @param root0.readOnly
 */
export default function BaseInput({
    label,
    value,
    placeHolder,
    disabled = false,
    inputMode = 'text',
    type,
    stretch = false,
    constraintText,
    errorText,
    onChange,
    required,
    autoComplete,
    autoFocus,
    readOnly,
}: BaseInputProps) {
    return (
        <>
            <BaseFormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                <Input
                    readOnly={readOnly}
                    value={value}
                    onChange={onChange}
                    inputMode={inputMode}
                    type={type}
                    placeholder={placeHolder}
                    ariaRequired={required}
                    ariaLabel={label}
                    autoComplete={autoComplete}
                    autoFocus={autoFocus}
                    disabled={disabled}
                />
            </BaseFormField>
        </>
    );
}
