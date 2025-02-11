// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { AirBooking } from '../../../models/traveller';
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

export default function AirBookingSection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).air_booking);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});

    const columnDefinitions: TableProps.ColumnDefinition<AirBooking>[] = [
        {
            id: 'confidenceAirBooking',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_AIR_BOOKING)
                    }
                    ariaLabel="openConfidenceAirBooking"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceAirBooking',
        },
        {
            id: 'bookingId',
            header: 'Booking Id',
            cell: ({ bookingId }) => cellAsLinkOrValue(hyperlinkMappings, 'booking_id', bookingId),
            ariaLabel: () => 'airBookingId',
        },
        {
            id: 'fromLoc',
            header: 'From',
            cell: ({ from }) => cellAsLinkOrValue(hyperlinkMappings, 'from', from),
            ariaLabel: () => 'airBookingFromLoc',
        },
        {
            id: 'toLoc',
            header: 'To',
            cell: ({ to }) => cellAsLinkOrValue(hyperlinkMappings, 'to', to),
            ariaLabel: () => 'airBookingToLoc',
        },
        {
            id: 'departureDate',
            header: 'Departure Date',
            cell: ({ departureDate }) => cellAsLinkOrValue(hyperlinkMappings, 'departure_date', departureDate),
            ariaLabel: () => 'airBookingDepartureDate',
        },
        {
            id: 'departureTime',
            header: 'Departure Time',
            cell: ({ departureTime }) => cellAsLinkOrValue(hyperlinkMappings, 'departure_time', departureTime),
            ariaLabel: () => 'airBookingDepartureTime',
        },
        {
            id: 'status',
            header: 'Status',
            cell: ({ status }) => cellAsLinkOrValue(hyperlinkMappings, 'status', status),
            ariaLabel: () => 'airBookingStatus',
        },
        {
            id: 'lastUpdated',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'airBookingLastUpdated',
        },
        {
            id: 'modal',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openAirBookingRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'airBookingViewRecord',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_AIR_BOOKING);
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

    return !Array.isArray(profileData.airBookingRecords) || profileData.airBookingRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Air Booking Records">
                <BaseCRUDTable<AirBooking>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.airBookingRecords}
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
