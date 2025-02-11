// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import {
    Box,
    Button,
    ButtonDropdown,
    ButtonDropdownProps,
    Container,
    FormField,
    Header,
    Input,
    NonCancelableCustomEvent,
    Popover,
    SpaceBetween,
    StatusIndicator,
    StatusIndicatorProps,
    TableProps,
} from '@cloudscape-design/components';
import { BaseChangeDetail } from '@cloudscape-design/components/input/interfaces';
import { ClickDetail } from '@cloudscape-design/components/internal/events';
import { skipToken } from '@reduxjs/toolkit/query';
import React, { ReactNode, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import BaseCRUDTable, { CustomPropsSortingType } from '../../components/base/baseCrudTable';
import { useBasePreferences } from '../../components/base/basePreferences';
import { PurgeProfileConfirmationDialog } from '../../components/privacy/purgeProfileConfirmationDialog';
import { AsyncEventStatus } from '../../models/async';
import { AsyncEventResponse } from '../../models/config';
import {
    PrivacyPurgeRequest,
    PrivacyPurgeStatus,
    PrivacySearchRequest,
    PrivacySearchStatus,
    PrivacySearchesTableItem,
    SearchContainerTestId,
    SearchResultsContainerTestId,
    SearchResultsTestId,
} from '../../models/privacy';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import {
    useCreateSearchMutation,
    useDeleteSearchesMutation,
    useGetPurgeIsRunningQuery,
    useListSearchesQuery,
    usePurgeProfileMutation,
} from '../../store/privacyApiSlice';
import { clearSearchResult, selectPrivacySearchResults } from '../../store/privacySlice';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

export default function PrivacyScreen() {
    const dispatch = useDispatch();

    const { isFetching, refetch: refetchPrivacySearches } = useListSearchesQuery();
    const [createPrivacySearch, createSearchResponse] = useCreateSearchMutation();
    const [deleteSearches, { isSuccess: isDeleteSearchesSuccess }] = useDeleteSearchesMutation();
    const [purgeProfileTrigger, purgeProfileTriggerResult] = usePurgeProfileMutation();

    const privacySearchResults = useSelector(selectPrivacySearchResults);

    const [inputValue, setInputValue] = useState<string>('');
    const [selectedItems, setSelectedItems] = useState<PrivacySearchesTableItem[]>([]);
    const [isPurgeProfileConfirmationDialogVisible, setIsPurgeProfileConfirmationDialogVisible] = useState<boolean>(false);

    const { preferences } = useBasePreferences({
        pageSize: 50,
    });

    const onSearchInputChange = ({ detail }: NonCancelableCustomEvent<BaseChangeDetail>): void => {
        setInputValue(detail.value.trim());
    };

    const onSearchButtonClicked = (_e: CustomEvent<ClickDetail>): void => {
        asyncRunStatus.current = undefined;
        const privacySearchRq: PrivacySearchRequest[] = inputValue
            .trim()
            .split(',')
            .map(id => {
                return { connectId: id.trim() };
            });
        createPrivacySearch({
            privacySearchRq,
        });
    };

    const onSelectionChanged = (e: PrivacySearchesTableItem[]): void => {
        setSelectedItems(e);
    };

    const onGridActionsButtonClicked = (e: CustomEvent<ButtonDropdownProps.ItemClickDetails>): void => {
        const { detail } = e;
        if (detail.id === 'repeatSearch') {
            onRepeatSearchButtonClicked();
        } else if (detail.id === 'deleteSearch') {
            onDeleteSearchButtonClicked();
        }
    };

    const onDeleteSearchButtonClicked = (): void => {
        const connectIds: string[] = selectedItems.map<string>(({ connectId }) => connectId);
        deleteSearches(connectIds);
        setSelectedItems([]);
    };

    const onRepeatSearchButtonClicked = (): void => {
        asyncRunStatus.current = undefined;
        const privacySearchRq: PrivacySearchRequest[] = selectedItems.map<PrivacySearchRequest>(({ connectId }) => ({ connectId }));
        createPrivacySearch({
            privacySearchRq,
        });
        setSelectedItems([]);
    };

    const onPurgeButtonClicked = async (): Promise<void> => {
        const createPrivacyPurgeRq: PrivacyPurgeRequest = {
            connectIds: selectedItems.map(({ connectId }) => connectId),
        };
        purgeProfileTrigger({ createPrivacyPurgeRq });
        purgeRunStatus.current = { isPurgeRunning: true };
        setSelectedItems([]);
        setIsPurgeProfileConfirmationDialogVisible(false);
    };

    const purgeRunStatus = useRef<PrivacyPurgeStatus>();
    const { data: isPurgeRunningData } = useGetPurgeIsRunningQuery(null, {
        pollingInterval: 2000,
        skip: !(purgeProfileTriggerResult.isSuccess && purgeProfileTriggerResult.data && purgeRunStatus.current?.isPurgeRunning),
    });

    useEffect(() => {
        purgeRunStatus.current = isPurgeRunningData;
    }, [isPurgeRunningData]);

    const asyncRunStatus = useRef<AsyncEventResponse>();
    const { data: asyncData } = useGetAsyncStatusQuery(
        createSearchResponse.isSuccess &&
            createSearchResponse.data &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? { id: createSearchResponse.data.item_type, useCase: createSearchResponse.data.item_id }
            : skipToken,
        {
            pollingInterval: 2000,
        },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (asyncRunStatus.current?.status === AsyncEventStatus.EVENT_STATUS_SUCCESS) {
            setInputValue('');
        }
    }, [asyncData]);

    const isEnabled =
        createSearchResponse.isUninitialized ||
        createSearchResponse.isError ||
        (createSearchResponse.isSuccess &&
            (asyncRunStatus.current?.status === AsyncEventStatus.EVENT_STATUS_SUCCESS ||
                asyncRunStatus.current?.status === AsyncEventStatus.EVENT_STATUS_FAILED));

    useEffect(() => {
        if (isDeleteSearchesSuccess) {
            setSelectedItems([]);
        }
    }, [isDeleteSearchesSuccess]);

    const getStatus = (status: string, errorMessage?: string | undefined): ReactNode => {
        //  Set Defaults
        let type: StatusIndicatorProps.Type = 'warning';
        let text: string = 'Unknown Status';
        let wrapper = <></>;

        //  'status' is keyed by privacy search event status:
        if (status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_SUCCESS) {
            type = 'success';
            text = 'Search Results Ready';
        } else if (status === PrivacySearchStatus.PRIVACY_STATUS_PURGE_SUCCESS) {
            type = 'success';
            text = 'Purge Completed';
        } else if (status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_RUNNING) {
            type = 'in-progress';
            text = 'Searching';
        } else if (status === PrivacySearchStatus.PRIVACY_STATUS_PURGE_RUNNING) {
            type = 'in-progress';
            text = 'Purging';
        } else if (
            status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_FAILED ||
            status === PrivacySearchStatus.PRIVACY_STATUS_PURGE_FAILED
        ) {
            type = 'error';
            text = `${status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_FAILED ? 'Search' : 'Purge'} Failed`;
            wrapper = (
                <Popover triggerType="text" content={`${errorMessage && errorMessage !== '' ? errorMessage : 'Unknown Error'}`}></Popover>
            );
        } else if (status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_INVOKED) {
            type = 'pending';
            text = 'Search Queued';
        } else if (status === PrivacySearchStatus.PRIVACY_STATUS_PURGE_INVOKED) {
            type = 'pending';
            text = 'Purge Queued';
        }

        return React.createElement(wrapper.type, wrapper.props, [<StatusIndicator type={type}>{text}</StatusIndicator>]);
    };

    const columnDefinitions: TableProps.ColumnDefinition<PrivacySearchesTableItem>[] = [
        {
            id: 'connectId',
            header: 'Connect ID',
            cell: ({ connectId, status }) =>
                status === PrivacySearchStatus.PRIVACY_STATUS_SEARCH_SUCCESS ? (
                    <Link to={connectId} onClick={_e => dispatch(clearSearchResult())}>
                        {connectId}
                    </Link>
                ) : (
                    connectId
                ),
            sortingField: 'connectId',
            ariaLabel: () => 'connectId',
        },
        {
            id: 'status',
            header: 'Status',
            cell: ({ status, errorMessage }) => getStatus(status, errorMessage),
            sortingField: 'status',
            ariaLabel: () => 'status',
        },
        {
            id: 'totalResultsFound',
            header: 'Results Found',
            cell: privacySearchesTableItem => privacySearchesTableItem.totalResultsFound,
            sortingField: 'totalResultsFound',
            ariaLabel: () => 'totalResultsFound',
        },
        {
            id: 'searchDate',
            header: 'Search Date',
            cell: ({ searchDate }) => `${searchDate}`,
            sortingField: 'searchDate',
            ariaLabel: () => 'searchDate',
        },
    ];

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserPerformPrivacySearch =
        (userAppAccess & Permissions.CreatePrivacySearchPermission) === Permissions.CreatePrivacySearchPermission;

    const canUserDeletePrivacySearch =
        (userAppAccess & Permissions.DeletePrivacySearchPermission) === Permissions.DeletePrivacySearchPermission;

    const canUserPerformPurge = (userAppAccess & Permissions.PrivacyDataPurgePermission) === Permissions.PrivacyDataPurgePermission;

    const resultsTableHeader = (
        <Header
            counter={`(${selectedItems.length === 0 ? '' : selectedItems.length + '/'}${privacySearchResults.length})`}
            actions={
                <SpaceBetween direction="horizontal" size="xs">
                    <Button iconName="refresh" onClick={refetchPrivacySearches}></Button>
                    <ButtonDropdown
                        items={[
                            {
                                text: 'Repeat Search',
                                id: 'repeatSearch',
                                disabled: selectedItems.length === 0 || !canUserPerformPrivacySearch,
                                iconName: 'redo',
                            },
                            {
                                text: 'Delete Search',
                                id: 'deleteSearch',
                                disabled: selectedItems.length === 0 || !canUserDeletePrivacySearch,
                                iconName: 'remove',
                            },
                        ]}
                        mainAction={{
                            text: 'Purge Profile',
                            onClick: _e => setIsPurgeProfileConfirmationDialogVisible(true),
                            disabled: selectedItems.length === 0 || !canUserPerformPurge,
                        }}
                        variant="primary"
                        disabled={!(selectedItems.length !== 0 && canUserPerformPrivacySearch && canUserPerformPrivacySearch)}
                        onItemClick={e => onGridActionsButtonClicked(e)}
                    >
                        Actions
                    </ButtonDropdown>
                </SpaceBetween>
            }
        >
            Previous Searches
        </Header>
    );

    const sortingProps: CustomPropsSortingType<PrivacySearchesTableItem> = {
        defaultSortState: {
            sortingColumn: {
                sortingField: 'searchDate',
            },
            isDescending: true,
        },
    };

    return (
        <>
            <Box variant="h1" padding={{ top: 'm' }}></Box>
            <SpaceBetween size="xs" direction="vertical">
                <Container variant="default" header={<Header>Search for Profile Data</Header>} data-testid={SearchContainerTestId}>
                    <FormField
                        description="Enter a single Connect ID, or a comma separated list of Connect IDs to search for"
                        secondaryControl={
                            <Button
                                variant="primary"
                                disabled={!(isEnabled && canUserPerformPrivacySearch)}
                                onClick={(e: CustomEvent<ClickDetail>) => onSearchButtonClicked(e)}
                            >
                                Create Profile Search
                            </Button>
                        }
                    >
                        <Input
                            value={inputValue}
                            placeholder="id1, id2, id3"
                            onChange={e => onSearchInputChange(e)}
                            type="search"
                            disabled={!canUserPerformPrivacySearch}
                        ></Input>
                    </FormField>
                </Container>
                <Container data-testid={SearchResultsContainerTestId}>
                    <BaseCRUDTable<PrivacySearchesTableItem>
                        tableVariant="embedded"
                        columnDefinitions={columnDefinitions}
                        isLoading={isFetching}
                        header={resultsTableHeader}
                        customPropsSorting={sortingProps}
                        loadingText="Loading Profile Data Searches..."
                        items={privacySearchResults}
                        selectedItems={selectedItems}
                        onSelectionChange={e => onSelectionChanged(e)}
                        selectionType="multi"
                        preferences={preferences}
                        data-testid={SearchResultsTestId}
                    ></BaseCRUDTable>
                </Container>
            </SpaceBetween>
            <PurgeProfileConfirmationDialog
                headerText="Are you sure you want to purge the selected profiles?"
                isVisible={isPurgeProfileConfirmationDialogVisible}
                onDismiss={() => setIsPurgeProfileConfirmationDialogVisible(false)}
                onConfirm={() => onPurgeButtonClicked()}
                selectedProfiles={selectedItems.length > 0 ? selectedItems.map(item => item.connectId) : undefined}
            />
        </>
    );
}
