// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { Clickstream } from '../../../models/traveller';
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

export default function ClickstreamSection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).clickstream);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<Clickstream>[] = [
        {
            id: 'confidenceClickstream',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_CLICKSTREAM)
                    }
                    ariaLabel="openConfidenceClickstream"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceClickstream',
        },
        {
            id: 'eventTypeClickstream',
            header: 'Event Type',
            cell: ({ eventType }) => cellAsLinkOrValue(hyperlinkMappings, 'event_type', eventType),
            ariaLabel: () => 'eventTypeClickstream',
        },
        {
            id: 'eventTimestampClickstream',
            header: 'Recieved',
            cell: ({ eventTimestamp }) =>
                cellAsLinkOrValue(hyperlinkMappings, 'event_timestamp', new Date(eventTimestamp).toLocaleString()),
            ariaLabel: () => 'eventTimestampClickstream',
        },
        {
            id: 'modalClickstream',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openClickstreamRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordClickstream',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_CLICKSTREAM);
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

    return !Array.isArray(profileData.clickstreamRecords) || profileData.clickstreamRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Web and Mobile Activity">
                <BaseCRUDTable<Clickstream>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.clickstreamRecords}
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
