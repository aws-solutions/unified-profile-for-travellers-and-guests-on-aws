// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Select, SelectProps } from '@cloudscape-design/components';
import BaseFormField, { BaseFormFieldProps } from './baseFormField';

export interface BaseSelectProps extends BaseFormFieldProps<SelectProps.ChangeDetail> {
    selectedOption: SelectProps.Option | null;
    options: SelectProps.Options;
    loadingText?: string;
    filteringType?: SelectProps.FilteringType;
    filteringPlaceholder?: string;
    placeholder?: string;
    triggerVariant?: SelectProps.TriggerVariant;
}

export default function BaseSelect({
    label,
    selectedOption,
    options,
    loadingText,
    stretch,
    errorText,
    onChange,
    disabled,
    autoFocus,
    required,
    constraintText,
    placeholder = '',
    filteringType = 'none',
    filteringPlaceholder = '',
    triggerVariant = 'label',
}: BaseSelectProps) {
    const NO_OPTIONS = 'No Options';

    const selectElement = (
        <Select
            selectedOption={selectedOption}
            options={options}
            loadingText={loadingText}
            onChange={onChange}
            disabled={disabled}
            autoFocus={autoFocus}
            ariaRequired={required}
            selectedAriaLabel={`${label}-selected`}
            filteringType={filteringType}
            placeholder={placeholder}
            empty={NO_OPTIONS}
            filteringPlaceholder={filteringPlaceholder}
            triggerVariant={triggerVariant}
        />
    );

    return (
        <>
            {label ? (
                <BaseFormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                    {selectElement}
                </BaseFormField>
            ) : (
                selectElement
            )}
        </>
    );
}
