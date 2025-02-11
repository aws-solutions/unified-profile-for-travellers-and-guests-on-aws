// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ExpandableSection, TableProps } from '@cloudscape-design/components';
import { ProfileResponse } from '../../../models/config';
import BaseCRUDTable from '../../base/baseCrudTable';
import DefaultCollectionPreferences from '../../base/defaultCollectionPreferences';
import { useBasePreferences } from '../../base/basePreferences';

interface ParsingErrorProps {
    profileData: ProfileResponse;
}

export default function ParsingErrorSection({ profileData }: ParsingErrorProps) {
    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<string>[] = [
        {
            id: 'parsingErrorCol',
            header: 'Errors',
            cell: item => item,
            ariaLabel: () => 'parsingErrorCol',
        },
    ];

    return !Array.isArray(profileData.parsingErrors) || profileData.parsingErrors.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Parsing Errors">
                <BaseCRUDTable<string>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.parsingErrors}
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
