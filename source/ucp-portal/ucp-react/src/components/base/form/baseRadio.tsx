// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { RadioGroup, RadioGroupProps } from '@cloudscape-design/components';
import BaseFormField, { BaseFormFieldProps } from './baseFormField';

interface BaseRadioProps extends BaseFormFieldProps<RadioGroupProps.ChangeDetail> {
    value: string | null;
    items: Array<RadioGroupProps.RadioButtonDefinition>;
}

export default function BaseRadio({ label, stretch, constraintText, errorText, value, items, onChange }: BaseRadioProps) {
    return (
        <>
            <BaseFormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                <RadioGroup value={value} items={items} onChange={onChange} />
            </BaseFormField>
        </>
    );
}
