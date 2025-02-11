// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { LoyaltyTx } from '../../../models/traveller';
import { DialogType, selectPaginationParam, selectPortalConfig } from '../../../store/profileSlice';
import {
    buildPaginatonProps,
    cellAsLinkOrValue,
    onClickViewDialog,
    portalConfigToMap,
    updatePaginationOnPreferenceChange,
} from '../../../utils/profileUtils';
import BaseCRUDTable, { CustomPropsPaginationType } from '../../base/baseCrudTable';
import { useBasePreferences } from '../../base/basePreferences';
import DefaultCollectionPreferences from '../../base/defaultCollectionPreferences';
import { SectionProp } from './displayHome';

export default function LoyaltyTxSection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).loyalty_transaction);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<LoyaltyTx>[] = [
        {
            id: 'confidenceLoyaltyTx',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_LOYALTY_TRANSACTION)
                    }
                    ariaLabel="openConfidenceLoyaltyTx"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceLoyaltyTx',
        },
        {
            id: 'programNameLoyaltyTx',
            header: 'Program',
            cell: ({ programName }) => cellAsLinkOrValue(hyperlinkMappings, 'program_name', programName),
            ariaLabel: () => 'programNameLoyaltyTx',
        },
        {
            id: 'accpObjectIDLoyaltyTx',
            header: 'Tx ID',
            cell: ({ accpObjectID }) => cellAsLinkOrValue(hyperlinkMappings, 'accp_object_id', accpObjectID),
            ariaLabel: () => 'accpObjectIDLoyaltyTx',
        },
        {
            id: 'pointsChangeLoyaltyTx',
            header: 'Points Change',
            cell: ({ pointsOffset, pointUnit }) => pointsOffset + ' ' + pointUnit,
            ariaLabel: () => 'pointsChangeLoyaltyTx',
        },
        {
            id: 'categoryLoyaltyTx',
            header: 'Category',
            cell: ({ category }) => cellAsLinkOrValue(hyperlinkMappings, 'category', category),
            ariaLabel: () => 'categoryLoyaltyTx',
        },
        {
            id: 'sourceLoyaltyTx',
            header: 'Source',
            cell: ({ source }) => cellAsLinkOrValue(hyperlinkMappings, 'source', source),
            ariaLabel: () => 'sourceLoyaltyTx',
        },
        {
            id: 'orderNumberLoyaltyTx',
            header: 'Booking ID',
            cell: ({ orderNumber }) => cellAsLinkOrValue(hyperlinkMappings, 'order_number', orderNumber),
            ariaLabel: () => 'orderNumberLoyaltyTx',
        },
        {
            id: 'lastUpdatedLoyaltyTx',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'lastUpdatedLoyaltyTx',
        },
        {
            id: 'modalLoyaltyTx',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openLoyaltyTxRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordLoyaltyTx',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_LOYALTY_TRANSACTION);
    const [currentPageIndex, setCurrentPageIndex] = useState<number>(paginationParam.pages[paginationIndex]);
    const [pagesCount, setPagesCount] = useState<number>(
        Math.ceil(paginationMetadata.total_records / paginationParam.pageSizes[paginationIndex]),
    );

    const paginationProps: CustomPropsPaginationType = buildPaginatonProps(
        setCurrentPageIndex,
        paginationParam,
        paginationIndex,
        currentPageIndex,
        pagesCount,
        dispatch,
    );

    useEffect(() => {
        updatePaginationOnPreferenceChange(paginationParam, paginationIndex, preferences.pageSize, dispatch);
    }, [preferences]);

    useEffect(() => {
        setCurrentPageIndex(0);
        setPagesCount(Math.ceil(paginationMetadata.total_records / paginationParam.pageSizes[paginationIndex]));
    }, [paginationParam.pageSizes]);

    return !Array.isArray(profileData.lotaltyTxRecords) || profileData.lotaltyTxRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Loyalty Transactions">
                <BaseCRUDTable<LoyaltyTx>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.lotaltyTxRecords}
                    preferences={preferences}
                    tableVariant="embedded"
                    stripedRows={true}
                    customPropsPagination={paginationProps}
                    collectionPreferences={<DefaultCollectionPreferences preferences={preferences} onConfirm={onConfirm} />}
                />
            </ExpandableSection>
            <br />
        </>
    );
}
