// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { CollectionPreferences, NonCancelableCustomEvent } from '@cloudscape-design/components';
import { CollectionPreferencesProps } from '@cloudscape-design/components/collection-preferences';
import React from 'react';

type BaseCollectionPreferencesProps = {
    onConfirm: ({ detail }: NonCancelableCustomEvent<CollectionPreferencesProps.Preferences>) => void;
    preferences: CollectionPreferencesProps.Preferences;
    visibleContentOptions?: ReadonlyArray<CollectionPreferencesProps.VisibleContentOption>;
    setVisibleContentOptions?: React.Dispatch<React.SetStateAction<CollectionPreferencesProps.VisibleContentOption[]>>;
    type: string;
    title?: string;
    confirmLabel?: string;
    cancelLabel?: string;
};

export default function BaseCollectionPreferences({
    onConfirm,
    preferences,
    visibleContentOptions,
    setVisibleContentOptions,
    type,
    title,
    confirmLabel,
    cancelLabel,
}: BaseCollectionPreferencesProps) {
    const visibleContentOptionsGroup: CollectionPreferencesProps.VisibleContentOptionsGroup = {
        label: 'Columns',
        options: visibleContentOptions ?? [],
    };

    const visibleContentPreference: CollectionPreferencesProps.VisibleContentPreference = {
        title: 'Select Visible Columns',
        options: [visibleContentOptionsGroup],
    };

    return (
        <CollectionPreferences
            onConfirm={onConfirm}
            title={title ?? 'Preferences'}
            confirmLabel={confirmLabel ?? 'Confirm'}
            cancelLabel={cancelLabel ?? 'Cancel'}
            preferences={preferences}
            pageSizePreference={{
                title: 'Select page size',
                options: [
                    { value: 10, label: `10 ${type}s` },
                    { value: 20, label: `20 ${type}s` },
                    { value: 30, label: `30 ${type}s` },
                    { value: 50, label: `50 ${type}s` },
                    { value: 100, label: `100 ${type}s` },
                ],
            }}
            visibleContentPreference={visibleContentOptions && visibleContentPreference}
        />
    );
}
