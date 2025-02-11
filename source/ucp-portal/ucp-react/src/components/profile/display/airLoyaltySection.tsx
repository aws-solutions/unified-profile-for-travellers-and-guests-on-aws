// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { AirLoyalty } from '../../../models/traveller';
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

export default function AirLoyaltySection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).air_loyalty);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<AirLoyalty>[] = [
        {
            id: 'confidenceAirLoyalty',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_AIR_LOYALTY)
                    }
                    ariaLabel="openConfidenceAirLoyalty"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceAirLoyalty',
        },
        {
            id: 'airLoyaltyprogramName',
            header: 'Program Name',
            cell: ({ programName }) => cellAsLinkOrValue(hyperlinkMappings, 'program_name', programName),
            ariaLabel: () => 'airLoyaltyprogramName',
        },
        {
            id: 'airLoyaltyjoined',
            header: 'Joined',
            cell: ({ joined }) => cellAsLinkOrValue(hyperlinkMappings, 'joined', new Date(joined).toLocaleDateString()),
            ariaLabel: () => 'airLoyaltyjoined',
        },
        {
            id: 'airLoyaltyId',
            header: 'Loyalty Id',
            cell: ({ loyaltyId }) => cellAsLinkOrValue(hyperlinkMappings, 'id', loyaltyId),
            ariaLabel: () => 'airLoyaltyId',
        },
        {
            id: 'airLoyaltymiles',
            header: 'Miles',
            cell: ({ miles }) => cellAsLinkOrValue(hyperlinkMappings, 'miles', miles),
            ariaLabel: () => 'airLoyaltymiles',
        },
        {
            id: 'airLoyaltylevel',
            header: 'Level',
            cell: ({ level }) => cellAsLinkOrValue(hyperlinkMappings, 'level', level),
            ariaLabel: () => 'airLoyaltylevel',
        },
        {
            id: 'airLoyaltymilesToNextLevel',
            header: 'Miles To Next Level',
            cell: ({ milesToNextLevel }) => cellAsLinkOrValue(hyperlinkMappings, 'miles_to_next_level', milesToNextLevel),
            ariaLabel: () => 'airLoyaltymilesToNextLevel',
        },
        {
            id: 'airLoyaltylastUpdated',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'airLoyaltylastUpdated',
        },
        {
            id: 'airLoyaltyViewRecord',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openAirLoyaltyRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'airLoyaltyViewRecord',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_AIR_LOYALTY);
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

    return !Array.isArray(profileData.airLoyaltyRecords) || profileData.airLoyaltyRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Air Loyalty Records">
                <BaseCRUDTable<AirLoyalty>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.airLoyaltyRecords}
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
