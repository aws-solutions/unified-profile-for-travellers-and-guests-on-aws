// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { CollectionPreferencesProps, NonCancelableCustomEvent } from '@cloudscape-design/components';
import BaseCollectionPreferences from '../base/page/baseCollectionPreferences';

interface DefaultCollectionPreferencesProps {
    preferences: CollectionPreferencesProps.Preferences;
    onConfirm: ({ detail }: NonCancelableCustomEvent<CollectionPreferencesProps.Preferences>) => void;
}
export default function DefaultCollectionPreferences({ preferences, onConfirm }: DefaultCollectionPreferencesProps) {
    return <BaseCollectionPreferences onConfirm={onConfirm} preferences={preferences} type={'item'} />;
}
