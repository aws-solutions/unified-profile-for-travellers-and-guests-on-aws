// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { CalendarProps, DatePicker } from '@cloudscape-design/components';
import { DATE_FORMAT } from '../../../models/constants';
import BaseFormField, { BaseFormFieldProps } from './baseFormField';

export interface BaseDatePickerProps extends BaseFormFieldProps<CalendarProps.ChangeDetail> {
    value: string;
    isDateEnabled?: CalendarProps.IsDateEnabledFunction;
    placeholder?: string;
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
 * @param root0.readOnly
 * @param root0.required
 * @param root0.constraintText
 * @param root0.isDateEnabled
 * @param root0.placeholder
 */
export default function BaseDatePicker({
    label,
    stretch,
    errorText,
    value,
    onChange,
    disabled,
    autoFocus,
    readOnly,
    required,
    constraintText = '',
    isDateEnabled = () => true,
    placeholder = `${DATE_FORMAT}`,
}: BaseDatePickerProps) {
    const TODAY = 'Today';
    const NEXT_MONTH = 'Next month';
    const PREVIOUS_MONTH = 'Previous month';
    return (
        <>
            <BaseFormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                <DatePicker
                    disabled={disabled}
                    autoFocus={autoFocus}
                    readOnly={readOnly}
                    ariaRequired={required}
                    onChange={onChange}
                    value={value}
                    placeholder={placeholder}
                    todayAriaLabel={TODAY}
                    nextMonthAriaLabel={NEXT_MONTH}
                    previousMonthAriaLabel={PREVIOUS_MONTH}
                    isDateEnabled={isDateEnabled}
                />
            </BaseFormField>
        </>
    );
}
