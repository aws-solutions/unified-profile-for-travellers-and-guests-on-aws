// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, Header, Icon, Link, Popover, SpaceBetween, TableProps } from '@cloudscape-design/components';
import { skipToken } from '@reduxjs/toolkit/query';
import { Auth } from 'aws-amplify';
import { useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { AsyncEventStatus } from '../../models/async';
import { AsyncEventResponse } from '../../models/config';
import { IconName } from '../../models/iconName';
import { MatchDiff, MatchPair } from '../../models/match';
import { PaginationOptions } from '../../models/pagination';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import { selectCurrentDomain } from '../../store/domainSlice';
import { useGetMatchesQuery, usePerformMergeMutation } from '../../store/matchesApiSlice';
import { selectMatchPairs, selectMatchTableItems, selectTotalMatchCount, setProfileDataDiff } from '../../store/matchesSlice';
import { buildMergeRequest, getBadge } from '../../utils/matchUtils';
import BaseCRUDTable, { CustomPropsPaginationType } from '../base/baseCrudTable';
import { useBasePreferences } from '../base/basePreferences';

export default function MatchTable() {
    const dispatch = useDispatch();

    const selectedDomain = useSelector(selectCurrentDomain).customerProfileDomain;
    const matchPairs: MatchPair[] = useSelector(selectMatchPairs);
    const matchCount: number = useSelector(selectTotalMatchCount);
    const tableItems: MatchDiff[] = useSelector(selectMatchTableItems);

    const [currentPageIndex, setCurrentPageIndex] = useState<number>(1);
    const [pagesCount, setPagesCount] = useState<number>();
    const [selectedItems, setSelectedItems] = useState<MatchDiff[]>([]);
    const onPageChange = (newPage: number): void => {
        if (selectedItems.length > 0) {
            if (confirm('You have unsaved changes. Are you sure you want to change pages?')) {
                setCurrentPageIndex(newPage);
                setSelectedItems([]);
            }
        } else {
            setCurrentPageIndex(newPage);
        }
    };

    const pagination: CustomPropsPaginationType = {
        onPageChange: onPageChange,
        currentPageIndex,
        pagesCount,
        openEnded: tableItems.length === 10,
    };

    const paginationArgs: PaginationOptions = {
        page: currentPageIndex - 1,
        pageSize: 10,
    };

    const { isFetching: isMatchFetching, refetch: refetchMatches } = useGetMatchesQuery(paginationArgs);

    const onSelectionChanged = (e: MatchDiff[]): void => {
        setSelectedItems(e);
    };

    useEffect(() => {
        setPagesCount(Math.ceil(matchCount / 10));
    }, [matchCount]);

    const onClickViewDialog = (item: MatchDiff) => {
        dispatch(setProfileDataDiff(item));
    };

    const [performMergeTrigger, { data: performMergeData, isSuccess: isPerformMergeSuccess }] = usePerformMergeMutation();
    const asyncRunStatus = useRef<AsyncEventResponse>();
    const { data: asyncData, isSuccess: isAsyncSuccess } = useGetAsyncStatusQuery(
        isPerformMergeSuccess &&
            performMergeData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED &&
            performMergeData.asyncEvent.item_type !== '' 
            ? { id: performMergeData.asyncEvent.item_type, useCase: performMergeData.asyncEvent.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (isPerformMergeSuccess || (isAsyncSuccess && asyncData.status === AsyncEventStatus.EVENT_STATUS_SUCCESS)) {
            setSelectedItems([]);
        }
    }, [asyncData, isPerformMergeSuccess]);

    let operatorCognitoId: string = '';
    Auth.currentUserInfo().then(user => {
        if (user && user.attributes && user.attributes.sub) operatorCognitoId = user.attributes.sub;
    });
    const onClickMarkOrMerge = (fp: boolean) => {
        if (confirm('Are you sure you want to perform this operation?')) {
            performMergeTrigger(buildMergeRequest(selectedItems, fp, operatorCognitoId));
        }
    };

    const actions = (
        <SpaceBetween direction="horizontal" size="s">
            <Button
                variant="icon"
                iconName={IconName.REFRESH}
                onClick={() => {
                    refetchMatches();
                }}
                ariaLabel="refetchMatchesButton"
            />
            <Button
                disabled={selectedItems.length === 0}
                variant="normal"
                iconName={IconName.FLAG}
                onClick={() => onClickMarkOrMerge(true)}
                ariaLabel="markFalseButton"
            >
                Mark as False Positives
            </Button>
            <Button
                disabled={selectedItems.length === 0}
                variant="primary"
                onClick={() => onClickMarkOrMerge(false)}
                ariaLabel="mergeButton"
            >
                Merge
            </Button>
        </SpaceBetween>
    );

    const header = (
        <Header counter={`(${selectedItems.length === 0 ? '' : selectedItems.length + '/'}${matchCount})`} variant="h2" actions={actions}>
            AI Matches{' '}
            <Popover
                dismissButton={false}
                position="right"
                size="medium"
                triggerType="text"
                content={
                    'Match count may take up to 6 hours to update. Only matches for the currently selected domain are enabled: ' +
                    selectedDomain
                }
            >
                <Icon name={IconName.STATUS_INFO} />
            </Popover>
        </Header>
    );

    const { preferences } = useBasePreferences({ pageSize: 10 });

    const lineBreak = <br />;
    const columnDefinitions: TableProps.ColumnDefinition<MatchDiff>[] = [
        {
            id: 'domain',
            header: 'Domain',
            cell: ({ Domain }) => Domain,
            ariaLabel: () => 'domain',
        },
        {
            id: 'viewProfiles',
            header: 'View Profiles',
            cell: ({ Id }) => (
                <>
                    <Link external href={`/profile/${matchPairs[Id].domain_sourceProfileId.split('_').pop()}`}>
                        Source
                    </Link>
                    {lineBreak}
                    <Link external href={`/profile/${matchPairs[Id].targetProfileID}`}>
                        Target
                    </Link>
                </>
            ),
            ariaLabel: () => 'viewProfiles',
        },
        {
            id: 'aiConfidenceScore',
            header: 'Score',
            cell: ({ Score }) => Score,
            ariaLabel: () => 'aiConfidenceScore',
        },
        {
            id: 'firstName',
            header: 'First Name',
            cell: ({ ProfileAttributes }) => getBadge(ProfileAttributes.FirstName, lineBreak),
            ariaLabel: () => 'firstName',
        },
        {
            id: 'lastName',
            header: 'Last Name',
            cell: ({ ProfileAttributes }) => getBadge(ProfileAttributes.LastName, lineBreak),
            ariaLabel: () => 'lastName',
        },
        {
            id: 'birthDate',
            header: 'Birth Date',
            cell: ({ ProfileAttributes }) => getBadge(ProfileAttributes.BirthDate, lineBreak),
            ariaLabel: () => 'birthDate',
        },
        {
            id: 'gender',
            header: 'Gender',
            cell: ({ ProfileAttributes }) => getBadge(ProfileAttributes.Gender, lineBreak),
            ariaLabel: () => 'gender',
        },
        {
            id: 'emailId',
            header: 'Email',
            cell: ({ ContactInfoAttributes }) => getBadge(ContactInfoAttributes.EmailAddress, lineBreak),
            ariaLabel: () => 'emailId',
        },
        {
            id: 'attrDiffModal',
            header: 'View Diff',
            cell: item => (
                <Button variant="inline-link" onClick={() => onClickViewDialog(item)} ariaLabel="openAttrDiffButton">
                    Open
                </Button>
            ),
            ariaLabel: () => 'attrDiffModal',
        },
    ];

    return (
        <BaseCRUDTable<MatchDiff>
            columnDefinitions={columnDefinitions}
            items={tableItems}
            preferences={preferences}
            isFilterable={false}
            tableVariant="container"
            isLoading={isMatchFetching}
            header={header}
            data-testid="ai-matches-table"
            customPropsPagination={pagination}
            selectionType="multi"
            selectedItems={selectedItems}
            onSelectionChange={e => onSelectionChanged(e)}
            isItemDisabled={item => item.Domain != selectedDomain}
        />
    );
}
