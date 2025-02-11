// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { CustomerServiceInteraction } from '../../../models/traveller';
import { DialogType, selectPaginationParam, setDialogRecord, setDialogType } from '../../../store/profileSlice';
import { buildPaginatonProps, onClickViewDialog, updatePaginationOnPreferenceChange } from '../../../utils/profileUtils';
import BaseCRUDTable, { CustomPropsPaginationType } from '../../base/baseCrudTable';
import DefaultCollectionPreferences from '../../base/defaultCollectionPreferences';
import { SectionProp } from './displayHome';
import { useBasePreferences } from '../../base/basePreferences';

export default function CustomerServiceSection({ profileData, paginationMetadata }: SectionProp) {
    const dispatch = useDispatch();
    const onClickViewChat = (record: CustomerServiceInteraction) => {
        dispatch(setDialogType(DialogType.CHAT));
        dispatch(setDialogRecord(record));
    };

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<CustomerServiceInteraction>[] = [
        {
            id: 'confidenceCSI',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_CSI)}
                    ariaLabel="openConfidenceCSI"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceCSI',
        },
        {
            id: 'channelCS',
            header: 'Channel',
            cell: ({ channel }) => channel,
            ariaLabel: () => 'channelCS',
        },
        {
            id: 'startTimeCS',
            header: 'Start Time',
            cell: ({ startTime }) => new Date(startTime).toLocaleString(),
            ariaLabel: () => 'startTimeCS',
        },
        {
            id: 'endTimeCS',
            header: 'End Time',
            cell: ({ endTime }) => new Date(endTime).toLocaleString(),
            ariaLabel: () => 'toendTimeLocCS',
        },
        {
            id: 'interactionTypeCS',
            header: 'Interaction Type',
            cell: ({ interactionType }) => interactionType,
            ariaLabel: () => 'interactionTypeCS',
        },
        {
            id: 'sentimentScoreCS',
            header: 'Sentiment Score',
            cell: ({ sentimentScore }) => sentimentScore,
            ariaLabel: () => 'sentimentScoreCS',
        },
        {
            id: 'recordModalCS',
            header: 'View Record',
            cell: item => (
                <Button variant="inline-link" onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}>
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordCS',
        },
        {
            id: 'chatModal',
            header: 'View Chat',
            cell: item => (
                <Button variant="inline-link" onClick={() => onClickViewChat(item)} ariaLabel="openChatRecord">
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewChat',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_CSI);
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

    return !Array.isArray(profileData.customerServiceInteractionRecords) || profileData.customerServiceInteractionRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Customer Service Interaction Records">
                <BaseCRUDTable<CustomerServiceInteraction>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.customerServiceInteractionRecords}
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
