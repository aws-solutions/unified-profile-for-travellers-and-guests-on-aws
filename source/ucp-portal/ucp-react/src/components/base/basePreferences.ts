// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { CollectionPreferencesProps, NonCancelableCustomEvent } from '@cloudscape-design/components';
import { useState } from 'react';
import { PageSize } from '../../constants';

export interface State {
    preferences: CollectionPreferencesProps.Preferences;
    onConfirm: ({ detail }: NonCancelableCustomEvent<CollectionPreferencesProps.Preferences>) => void;
}

interface basePreferencesProps {
    pageSize?: number;
    specialpreferences?: CollectionPreferencesProps.Preferences;
    handleConfirm?: ({ detail }: NonCancelableCustomEvent<CollectionPreferencesProps.Preferences>) => void;
}

export function useBasePreferences(props: basePreferencesProps): State {
    const [defaultPreferences, setDefaultPreferences] = useState<CollectionPreferencesProps.Preferences>({
        pageSize: props.pageSize ?? PageSize.PAGE_SIZE_10,
    });

    const preferences = props.specialpreferences ?? defaultPreferences;

    /**
     *
     */
    function onConfirm(event: NonCancelableCustomEvent<CollectionPreferencesProps.Preferences>) {
        return props.handleConfirm ?? setDefaultPreferences(event.detail);
    }
    return {
        preferences,
        onConfirm,
    };
}
