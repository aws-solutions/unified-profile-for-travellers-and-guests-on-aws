// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, ExpandableSection, TableProps } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Common } from '../../../models/common';
import { EmailHistory } from '../../../models/traveller';
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

export default function EmailHistorySection({ profileData, paginationMetadata }: SectionProp) {
    const hyperlinkMappings: Map<string, string> = portalConfigToMap(useSelector(selectPortalConfig).email_history);

    const dispatch = useDispatch();

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<EmailHistory>[] = [
        {
            id: 'confidenceEmailHistory',
            header: 'Confidence',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() =>
                        onClickViewDialog(item, DialogType.HISTORY, dispatch, profileData.connectId, Common.OBJECT_TYPE_EMAIL_HISTORY)
                    }
                    ariaLabel="openConfidenceEmailHistory"
                >
                    {item.overallConfidenceScore}
                </Button>
            ),
            ariaLabel: () => 'confidenceEmailHistory',
        },
        {
            id: 'email',
            header: 'Email',
            cell: ({ address }) => cellAsLinkOrValue(hyperlinkMappings, 'address', address),
            ariaLabel: () => 'email',
        },
        {
            id: 'lastUpdatedEmailHistory',
            header: 'Last Updated',
            cell: ({ lastUpdated }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated', new Date(lastUpdated).toLocaleString()),
            ariaLabel: () => 'lastUpdatedEmailHistory',
        },
        {
            id: 'lastUpdatedByEmailHistory',
            header: 'Last Updated By',
            cell: ({ lastUpdatedBy }) => cellAsLinkOrValue(hyperlinkMappings, 'last_updated_by', lastUpdatedBy),
            ariaLabel: () => 'lastUpdatedByEmailHistory',
        },
        {
            id: 'modalEmailHistory',
            header: 'View Record',
            cell: item => (
                <Button
                    variant="inline-link"
                    onClick={() => onClickViewDialog(item, DialogType.RECORD, dispatch)}
                    ariaLabel="openEmailHistoryRecord"
                >
                    Open
                </Button>
            ),
            ariaLabel: () => 'viewRecordEmailHistory',
        },
    ];

    const paginationParam = useSelector(selectPaginationParam);
    const paginationIndex = paginationParam.objects.indexOf(Common.OBJECT_TYPE_EMAIL_HISTORY);
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

    return !Array.isArray(profileData.emailHistoryRecords) || profileData.emailHistoryRecords.length === 0 ? (
        <></>
    ) : (
        <>
            <ExpandableSection variant="container" headerText="Email History Records">
                <BaseCRUDTable<EmailHistory>
                    isFilterable={false}
                    columnDefinitions={columnDefinitions}
                    items={profileData.emailHistoryRecords}
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
