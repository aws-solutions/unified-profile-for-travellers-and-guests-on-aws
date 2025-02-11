// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Modal, TableProps, Tabs, TabsProps } from '@cloudscape-design/components';
import { useDispatch, useSelector } from 'react-redux';
import BaseCRUDTable from '../base/baseCrudTable';
import { DataDiff, MatchDiff } from '../../models/match';
import { AddressAtrributes, ContactInfoAttributes, ProfileAttributes } from '../../models/match';
import { selectProfileDataDiff, setProfileDataDiff } from '../../store/matchesSlice';
import { getBadge } from '../../utils/matchUtils';
import { useState } from 'react';

interface DiffModalProps {
    matches: ProfileAttributes | ContactInfoAttributes | AddressAtrributes;
}

function DiffModal({ matches }: DiffModalProps) {
    const tableItem: DataDiff[] = [];
    if (matches === undefined) return <></>;
    for (const [key, value] of Object.entries(matches as object)) {
        if (key !== 'Score' && key !== 'id') tableItem.push(value);
    }

    const columnDefinitions: TableProps.ColumnDefinition<DataDiff>[] = [
        {
            id: 'fieldName',
            header: 'Field Name',
            cell: ({ fieldName }) => fieldName.split('.').pop(),
            ariaLabel: () => 'fieldName',
        },
        {
            id: 'difference',
            header: 'Difference',
            cell: item => getBadge(item, <>{' vs '}</>),
            ariaLabel: () => 'difference',
        },
    ];

    return (
        <BaseCRUDTable<DataDiff>
            isFilterable={false}
            columnDefinitions={columnDefinitions}
            items={tableItem}
            preferences={{}}
            tableVariant="embedded"
            stripedRows={true}
        />
    );
}

export default function MatchDifferenceDialog() {
    const dispatch = useDispatch();
    const dataDiffToShow: MatchDiff = useSelector(selectProfileDataDiff);
    const isVisible = dataDiffToShow === undefined || Object.keys(dataDiffToShow).length === 0 ? false : true;

    const [tabValue, setTabValue] = useState('profileFields');

    const tabs: TabsProps.Tab[] = [
        {
            label: 'Profile Fields',
            id: 'profileFields',
            content: <DiffModal matches={dataDiffToShow.ProfileAttributes} />,
        },
        {
            label: 'Contact Info',
            id: 'contactInfo',
            content: <DiffModal matches={dataDiffToShow.ContactInfoAttributes} />,
        },
    ];

    // Show additional tabs if they have atleast one diff
    if (isVisible && Object.keys(dataDiffToShow.AddressAtrributes).length > 0) {
        tabs.push({
            label: 'Address',
            id: 'addressInfo',
            content: <DiffModal matches={dataDiffToShow.AddressAtrributes} />,
        });
    }
    if (isVisible && Object.keys(dataDiffToShow.BillingAddress).length > 0) {
        tabs.push({
            label: 'Billing Address',
            id: 'BillingAddress',
            content: <DiffModal matches={dataDiffToShow.BillingAddress} />,
        });
    }
    if (isVisible && Object.keys(dataDiffToShow.MailingAddress).length > 0) {
        tabs.push({
            label: 'Mailing Address',
            id: 'MailingAddress',
            content: <DiffModal matches={dataDiffToShow.MailingAddress} />,
        });
    }
    if (isVisible && Object.keys(dataDiffToShow.ShippingAddress).length > 0) {
        tabs.push({
            label: 'Shipping Address',
            id: 'ShippingAddress',
            content: <DiffModal matches={dataDiffToShow.ShippingAddress} />,
        });
    }

    const onDismiss = (): void => {
        setTabValue('profileFields');
        dispatch(setProfileDataDiff({} as MatchDiff));
    };

    return (
        <Modal onDismiss={onDismiss} visible={isVisible} header={'Detailed differences'} size="large">
            <Tabs tabs={tabs} activeTabId={tabValue} onChange={({ detail }) => setTabValue(detail.activeTabId)} />
        </Modal>
    );
}
