// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useDeleteAllErrorsMutation } from '../../store/errorApiSlice';
import { Dispatch, SetStateAction, useEffect, useRef } from 'react';
import { AsyncEventResponse } from '../../models/config';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import { skipToken } from '@reduxjs/toolkit/query';
import { AsyncEventStatus } from '../../models/async';

export interface DeleteAllErrorsButtonProps {
    refetch: () => void;
    setIsConfirmationModalVisible: Dispatch<SetStateAction<boolean>>;
}

export const DeleteAllErrorsButton = (props: DeleteAllErrorsButtonProps) => {
    const [deleteAllErrorsTrigger, { data: deleteErrorsData, error: deleteErrorsError, isSuccess: isDeleteErrorsSuccess }] =
        useDeleteAllErrorsMutation();

    const asyncRunStatus = useRef<AsyncEventResponse>();

    const { data: asyncData, error: asyncError } = useGetAsyncStatusQuery(
        isDeleteErrorsSuccess &&
            deleteErrorsData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? { id: deleteErrorsData.item_type, useCase: deleteErrorsData.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (asyncData?.status === AsyncEventStatus.EVENT_STATUS_SUCCESS || asyncData?.status === AsyncEventStatus.EVENT_STATUS_FAILED) {
            props.refetch();
        }
    }, [asyncData]);

    const onDeleteAllErrorsButtonClick = async (): Promise<void> => {
        props.setIsConfirmationModalVisible(false);
        asyncRunStatus.current = undefined;
        deleteAllErrorsTrigger();
    };

    return (
        <Button variant="primary" onClick={onDeleteAllErrorsButtonClick}>
            Clear All Errors
        </Button>
    );
};
