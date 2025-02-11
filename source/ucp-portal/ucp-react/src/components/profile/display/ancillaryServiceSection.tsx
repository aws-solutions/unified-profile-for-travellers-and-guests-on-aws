// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { AncillaryService } from '../../../models/traveller';
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

export default function AncillaryServiceSection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).ancillary_service);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<AncillaryService>[] = [
        {
            id: 'confidenceAncillaryService',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_ANCILLARY)
                    }
                    ariaLabel="openConfidenceAncillaryService"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceAncillaryService',
        },
        {
            id: 'ancillaryType',
            header: 'Ancillary Type',
            cell: ({ ancillaryType }) => cellAsLinkOrValue(hyperlinkMappings, 'ancillary_type', ancillaryType),
            ariaLabel: () => 'ancillaryType',
        },
        {
            id: 'ancillaryBookingId',
            header: 'Booking Id',
            cell: ({ bookingId }) => cellAsLinkOrValue(hyperlinkMappings, 'booking_id', bookingId),
            ariaLabel: () => 'ancillaryBookingId',
        },
        {
            id: 'ancillarySelection',
            header: 'Selection',
            cell: ({ baggageType, seatNumber }) => baggageType || seatNumber || 'other',
            ariaLabel: () => 'ancillarySelection',
        },
        {
            id: 'ancillaryQuantity',
            header: 'Quantity',
            cell: ({ quantity }) => cellAsLinkOrValue(hyperlinkMappings, 'quantity', quantity.toString()),
            ariaLabel: () => 'ancillaryQuantity',
        },
        {
            id: 'ancillaryPrice',
            header: 'Price',
            cell: ({ price }) => cellAsLinkOrValue(hyperlinkMappings, 'price', price.toString()),
            ariaLabel: () => 'ancillaryPrice',
        },
        {
            id: 'ancillaryLastUpdated',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'ancillaryLastUpdated',
        },
        {
            id: 'ancillaryViewRecord',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openAncillaryRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'ancillaryViewRecord',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_ANCILLARY);
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

    return !Array.isArray(profileData.ancillaryServiceRecords) || profileData.ancillaryServiceRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Ancillary Service Records">
                <BaseCRUDTable<AncillaryService>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.ancillaryServiceRecords}
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
