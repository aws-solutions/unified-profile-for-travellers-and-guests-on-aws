// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Header, TableProps } from '@cloudscape-design/components';
import { skipToken } from '@reduxjs/toolkit/query';
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';
import BaseCRUDTable from '../../components/base/baseCrudTable';
import { useBasePreferences } from '../../components/base/basePreferences';
import { PurgeProfileConfirmationDialog } from '../../components/privacy/purgeProfileConfirmationDialog';
import { AsyncEventStatus } from '../../models/async';
import { AsyncEventResponse } from '../../models/config';
import { ROUTES } from '../../models/constants';
import { PrivacyProfileLocationsTableItem, PrivacyResultsLocationsTableTestId } from '../../models/privacy';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import { useGetProfileSearchResultQuery, usePurgeProfileMutation } from '../../store/privacyApiSlice';
import { selectPrivacySearchResult } from '../../store/privacySlice';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';

export default function PrivacyResults() {
    const navigate = useNavigate();

    const { connectId } = useParams();
    const privacySearchResults: PrivacyProfileLocationsTableItem[] = useSelector(selectPrivacySearchResult);
    const { preferences } = useBasePreferences({
        pageSize: 20,
    });

    const [isPurgeProfileConfirmationDialogVisible, setIsPurgeProfileConfirmationDialogVisible] = useState<boolean>(false);

    const [purgeProfileTrigger, purgeProfileTriggerResult] = usePurgeProfileMutation();
    const { isFetching } = useGetProfileSearchResultQuery(connectId ? { connectId: connectId } : skipToken);

    const onPurgeButtonClicked = () => {
        purgeProfileTrigger({ createPrivacyPurgeRq: { connectIds: [connectId!] } });
        setIsPurgeProfileConfirmationDialogVisible(false);
    };

    const asyncRunStatus = useRef<AsyncEventResponse>();
    const { data: asyncData } = useGetAsyncStatusQuery(
        purgeProfileTriggerResult.isSuccess &&
            purgeProfileTriggerResult.data &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? {
                  id: purgeProfileTriggerResult.data.item_type,
                  useCase: purgeProfileTriggerResult.data.item_id,
              }
            : skipToken,
        {
            pollingInterval: 2000,
        },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (asyncRunStatus.current?.status === AsyncEventStatus.EVENT_STATUS_SUCCESS) {
            navigate(`/${ROUTES.PRIVACY}`);
        }
    }, [asyncData]);

    const columns: TableProps.ColumnDefinition<PrivacyProfileLocationsTableItem>[] = [
        {
            id: 'source',
            cell: item => item.source,
            header: 'Source',
        },
        {
            id: 'path',
            cell: item => item.path,
            header: 'Path',
        },
    ];

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserPerformPurge = (userAppAccess & Permissions.PrivacyDataPurgePermission) === Permissions.PrivacyDataPurgePermission;

    return (
        <>
            <Box variant="h1" padding={{ top: 'm' }}></Box>
            <BaseCRUDTable<PrivacyProfileLocationsTableItem>
                columnDefinitions={columns}
                header={
                    <Header
                        actions={
                            <Button
                                disabled={!(connectId && canUserPerformPurge)}
                                variant="primary"
                                onClick={_e => setIsPurgeProfileConfirmationDialogVisible(true)}
                            >
                                Purge Profile
                            </Button>
                        }
                    >
                        Search results for: {connectId}
                    </Header>
                }
                items={privacySearchResults}
                isLoading={isFetching}
                loadingText="Loading Profile Storage Locations"
                preferences={preferences}
                tableVariant="embedded"
                data-testid={PrivacyResultsLocationsTableTestId}
            ></BaseCRUDTable>
            <PurgeProfileConfirmationDialog
                headerText="Are you sure you want to purge this profile?"
                isVisible={isPurgeProfileConfirmationDialogVisible}
                onDismiss={() => setIsPurgeProfileConfirmationDialogVisible(false)}
                onConfirm={() => onPurgeButtonClicked()}
            />
        </>
    );
}
