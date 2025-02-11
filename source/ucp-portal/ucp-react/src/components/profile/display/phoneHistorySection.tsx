// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { PhoneHistory } from '../../../models/traveller';
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

export default function PhoneHistorySection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).phone_history);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<PhoneHistory>[] = [
        {
            id: 'confidencePhoneHistory',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_PHONE_HISTORY)
                    }
                    ariaLabel="openConfidencePhoneHistory"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidencePhoneHistory',
        },
        {
            id: 'phoneHistoryType',
            header: 'Type',
            cell: ({ type }) => cellAsLinkOrValue(hyperlinkMappings, 'type', type),
            ariaLabel: () => 'phoneHistoryType',
        },
        {
            id: 'phoneHistoryCountryCode',
            header: 'Country Code',
            cell: ({ countryCode }) => cellAsLinkOrValue(hyperlinkMappings, 'country_code', countryCode),
            ariaLabel: () => 'phoneHistoryCountryCode',
        },
        {
            id: 'phoneHistoryPhoneNumber',
            header: 'Phone',
            cell: ({ number }) => cellAsLinkOrValue(hyperlinkMappings, 'number', number),
            ariaLabel: () => 'phoneHistoryPhoneNumber',
        },
        {
            id: 'phoneHistoryLastUpdated',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'phoneHistoryLastUpdated',
        },
        {
            id: 'phoneHistoryLastUpdatedBy',
            header: 'Last Updated By',
            cell: ({ lastUpdatedBy }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated_by', lastUpdatedBy),
            ariaLabel: () => 'phoneHistoryLastUpdatedBy',
        },
        {
            id: 'phoneHistoryViewRecord',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openPhoneHistoryRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'phoneHistoryViewRecord',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_PHONE_HISTORY);
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

    return !Array.isArray(profileData.phoneHistoryRecords) || profileData.phoneHistoryRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Phone History Records">
                <BaseCRUDTable<PhoneHistory>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.phoneHistoryRecords}
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
