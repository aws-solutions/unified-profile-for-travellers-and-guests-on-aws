// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { HotelLoyalty } from '../../../models/traveller';
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

export default function HotelLoyaltySection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).hotel_loyalty);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<HotelLoyalty>[] = [
        {
            id: 'confidenceHotelLoyalty',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_HOTEL_LOYALTY)
                    }
                    ariaLabel="openConfidenceHotelLoyalty"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceHotelLoyalty',
        },
        {
            id: 'programNameHotelLoyalty',
            header: 'Program Name',
            cell: ({ programName }) => cellAsLinkOrValue(hyperlinkMappings, 'program_name', programName),
            ariaLabel: () => 'programNameHotelLoyalty',
        },
        {
            id: 'joinedHotelLoyalty',
            header: 'Joined',
            cell: ({ joined }) => cellAsLinkOrValue(hyperlinkMappings, 'joined', new Date(joined).toLocaleDateString()),
            ariaLabel: () => 'joinedHotelLoyalty',
        },
        {
            id: 'loyaltyIdHotelLoyalty',
            header: 'Loyalty Id',
            cell: ({ loyaltyId }) => cellAsLinkOrValue(hyperlinkMappings, 'id', loyaltyId),
            ariaLabel: () => 'loyaltyIdHotelLoyalty',
        },
        {
            id: 'pointsHotelLoyalty',
            header: 'Points',
            cell: ({ points }) => cellAsLinkOrValue(hyperlinkMappings, 'points', points),
            ariaLabel: () => 'pointsHotelLoyalty',
        },
        {
            id: 'levelHotelLoyalty',
            header: 'Level',
            cell: ({ level }) => cellAsLinkOrValue(hyperlinkMappings, 'level', level),
            ariaLabel: () => 'levelHotelLoyalty',
        },
        {
            id: 'pointsToNextLevelHotelLoyalty',
            header: 'Points To Next Level',
            cell: ({ pointsToNextLevel }) => cellAsLinkOrValue(hyperlinkMappings, 'points_to_next_level', pointsToNextLevel),
            ariaLabel: () => 'pointsToNextLevelHotelLoyalty',
        },
        {
            id: 'lastUpdatedHotelLoyalty',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'lastUpdatedHotelLoyalty',
        },
        {
            id: 'modalHotelLoyalty',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openHotelLoyaltyRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordHotelLoyalty',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_HOTEL_LOYALTY);
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

    return !Array.isArray(profileData.hotelLoyaltyRecords) || profileData.hotelLoyaltyRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Hotel Loyalty Records">
                <BaseCRUDTable<HotelLoyalty>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.hotelLoyaltyRecords}
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
