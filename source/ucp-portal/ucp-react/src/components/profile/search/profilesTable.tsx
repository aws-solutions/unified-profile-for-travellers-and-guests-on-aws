// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Header, Link, TableProps } from '@cloudscape-design/components';
import { useSelector } from 'react-redux';
import { ProfileTableItem, selectSearchResults } from '../../../store/profileSlice';
import { loadProfiles } from '../../../utils/profileUtils';
import BaseCRUDTable from '../../base/baseCrudTable';
import { useBasePreferences } from '../../base/basePreferences';

export default function ProfilesTable() {
    const searchResults = useSelector(selectSearchResults);
    const profiles: ProfileTableItem[] = loadProfiles(searchResults);

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<ProfileTableItem>[] = [
        {
            id: 'connectId',
            header: 'Connect Id',
            cell: ({ connectId }) => connectId,
            sortingField: 'connectId',
            ariaLabel: () => 'connectId',
        },
        {
            id: 'travellerId',
            header: 'Traveller Id',
            cell: ({ travellerId }) => travellerId,
            ariaLabel: () => 'travellerId',
        },
        {
            id: 'firstName',
            header: 'First Name',
            cell: ({ firstName }) => firstName,
            ariaLabel: () => 'firstName',
        },
        {
            id: 'lastName',
            header: 'Last Name',
            cell: ({ lastName }) => lastName,
            ariaLabel: () => 'lastName',
        },
        {
            id: 'emailId',
            header: 'Email',
            cell: ({ email }) => email,
            ariaLabel: () => 'emailId',
        },
        {
            id: 'view',
            header: 'View',
            cell: ({ connectId }) => (
                //removing the external link to support agent workspace use case
                <Link href={`/profile/${connectId}`}> View profile </Link>
            ),
        },
    ];

    return (
        <BaseCRUDTable<ProfileTableItem>
            columnDefinitions={columnDefinitions}
            items={profiles}
            header={
                <Header counter={`(${profiles.length})`} variant="h3">
                    {' '}
                    Profiles{' '}
                </Header>
            }
            preferences={preferences}
            tableVariant="embedded"
        />
    );
}
