// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { HotelStay } from '../../../models/traveller';
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

export default function HotelStaySection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).hotel_stay_revenue_items);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<HotelStay>[] = [
        {
            id: 'confidenceHotelStay',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_STAY_REVENUE)
                    }
                    ariaLabel="openConfidenceHotelStay"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceHotelStay',
        },
        {
            id: 'startDateHotelStay',
            header: 'Stay Date',
            cell: ({ startDate }) => cellAsLinkOrValue(hyperlinkMappings, 'start_date', new Date(startDate).toLocaleDateString()),
            ariaLabel: () => 'startDateHotelStay',
        },
        {
            id: 'hotelCodeHotelStay',
            header: 'Property',
            cell: ({ hotelCode }) => cellAsLinkOrValue(hyperlinkMappings, 'hotel_code', hotelCode),
            ariaLabel: () => 'hotelCodeHotelStay',
        },
        {
            id: 'typeHotelStay',
            header: 'Charge',
            cell: ({ type }) => cellAsLinkOrValue(hyperlinkMappings, 'type', type),
            ariaLabel: () => 'typeHotelStay',
        },
        {
            id: 'amountHotelStay',
            header: 'Amount',
            cell: ({ amount, currencyCode }) => amount + ' ' + currencyCode,
            ariaLabel: () => 'amountHotelStay',
        },
        {
            id: 'dateHotelStay',
            header: 'Charge Date',
            cell: ({ date }) => cellAsLinkOrValue(hyperlinkMappings, 'date', new Date(date).toLocaleString()),
            ariaLabel: () => 'chargeDateHotelStay',
        },
        {
            id: 'lastUpdatedHotelStay',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'lastUpdatedHotelStay',
        },
        {
            id: 'modalHotelStay',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openHotelStayRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordHotelStay',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_STAY_REVENUE);
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

    return !Array.isArray(profileData.hotelStayRecords) || profileData.hotelStayRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Hotel Stay Records">
                <BaseCRUDTable<HotelStay>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.hotelStayRecords}
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
