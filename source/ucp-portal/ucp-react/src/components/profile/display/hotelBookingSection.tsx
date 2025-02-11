// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { HotelBooking } from '../../../models/traveller';
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

export default function HotelBookingSection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).hotel_booking);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<HotelBooking>[] = [
        {
            id: 'confidenceHotelBooking',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_HOTEL_BOOKING)
                    }
                    ariaLabel="openConfidenceHotelBooking"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceHotelBooking',
        },
        {
            id: 'bookingIdHotelBooking',
            header: 'Booking Id',
            cell: ({ bookingId }) => cellAsLinkOrValue(hyperlinkMappings, 'booking_id', bookingId),
            ariaLabel: () => 'bookingIdHotelBooking',
        },
        {
            id: 'hotelCodeHotelBooking',
            header: 'Property',
            cell: ({ hotelCode }) => cellAsLinkOrValue(hyperlinkMappings, 'hotel_code', hotelCode),
            ariaLabel: () => 'hotelCodeHotelBooking',
        },
        {
            id: 'checkInDateHotelBooking',
            header: 'Check-In',
            cell: ({ checkInDate }) => cellAsLinkOrValue(hyperlinkMappings, 'check_in_date', new Date(checkInDate).toLocaleDateString()),
            ariaLabel: () => 'checkInDateHotelBooking',
        },
        {
            id: 'numNightsHotelBooking',
            header: '#Nights',
            cell: ({ numNights }) => cellAsLinkOrValue(hyperlinkMappings, 'n_nights', numNights.toString()),
            ariaLabel: () => 'numNightsHotelBooking',
        },
        {
            id: 'numGuestsHotelBooking',
            header: '#Guests',
            cell: ({ numGuests }) => cellAsLinkOrValue(hyperlinkMappings, 'n_guests', numGuests.toString()),
            ariaLabel: () => 'numGuestsHotelBooking',
        },
        {
            id: 'roomTypeCodeHotelBooking',
            header: 'Room Type',
            cell: ({ roomTypeCode }) => cellAsLinkOrValue(hyperlinkMappings, 'room_type_code', roomTypeCode),
            ariaLabel: () => 'roomTypeCodeHotelBooking',
        },
        {
            id: 'ratePlanCodeHotelBooking',
            header: 'Rate Plan',
            cell: ({ ratePlanCode }) => cellAsLinkOrValue(hyperlinkMappings, 'rate_plan_code', ratePlanCode),
            ariaLabel: () => 'ratePlanCodeHotelBooking',
        },
        {
            id: 'statusHotelBooking',
            header: 'Status',
            cell: ({ status }) => cellAsLinkOrValue(hyperlinkMappings, 'status', status),
            ariaLabel: () => 'statusHotelBooking',
        },
        {
            id: 'lastUpdatedHotelBooking',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'lastUpdatedHotelBooking',
        },
        {
            id: 'modalHotelBooking',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openHotelBookingRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordHotelBooking',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_HOTEL_BOOKING);
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

    return !Array.isArray(profileData.hotelBookingRecords) || profileData.hotelBookingRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Hotel Booking Records">
                <BaseCRUDTable<HotelBooking>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.hotelBookingRecords}
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
