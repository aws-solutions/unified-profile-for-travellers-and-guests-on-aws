// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ExpandableSection, TableProps } from '@cloudscape-design/components';
import { ProfileMatches } from '../../../models/config';
import BaseCRUDTable from '../../base/baseCrudTable';
import DefaultCollectionPreferences from '../../base/defaultCollectionPreferences';
import { useBasePreferences } from '../../base/basePreferences';

interface ProfileMatchesProps {
    profileMatches: ProfileMatches[];
}

export default function ProfileMatchesSection({ profileMatches }: ProfileMatchesProps) {
    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<ProfileMatches>[] = [
        {
            id: 'profileMatchId',
            header: 'ID',
            cell: ({ id }) => id,
            ariaLabel: () => 'profileMatchId',
        },
        {
            id: 'profileMatchFirstName',
            header: 'First Name',
            cell: ({ firstName }) => firstName,
            ariaLabel: () => 'profileMatchFirstName',
        },
        {
            id: 'profileMatchLastName',
            header: 'Last Name',
            cell: ({ lastName }) => lastName,
            ariaLabel: () => 'profileMatchLastName',
        },
        {
            id: 'profileMatchConfidence',
            header: 'Confidence',
            cell: ({ confidence }) => confidence,
            ariaLabel: () => 'profileMatchConfidence',
        },
    ];

    return !Array.isArray(profileMatches) || profileMatches.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Matching Profiles">
                <BaseCRUDTable<ProfileMatches>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileMatches}
                    preferences={preferences}
                    tableVariant="embedded"
                    stripedRows={true}
                    collectionPreferences={<DefaultCollectionPreferences preferences={preferences} onConfirm={onConfirm} />}
                />
            </ExpandableSection>
            <br />
        </>
    );
}
