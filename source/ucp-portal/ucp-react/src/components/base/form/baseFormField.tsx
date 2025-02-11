// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { FormField, NonCancelableCustomEvent } from '@cloudscape-design/components';
import { ReactNode } from 'react';

export interface BaseFormFieldProps<T> {
    label?: string;
    description?: string;
    info?: ReactNode;
    stretch?: boolean;
    constraintText?: string;
    errorText?: string;
    children?: ReactNode;

    required?: boolean;
    autoFocus?: boolean;
    readOnly?: boolean;

    disabled?: boolean;

    onChange?: (change: NonCancelableCustomEvent<T>) => void;
    invalid?: boolean;
}

export default function BaseFormField<T>({ label, stretch, constraintText, errorText, children }: BaseFormFieldProps<T>) {
    return (
        <>
            <FormField label={label} stretch={stretch} constraintText={constraintText} errorText={errorText}>
                {children}
            </FormField>
        </>
    );
}
