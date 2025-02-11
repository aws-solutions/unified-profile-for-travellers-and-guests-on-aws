// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { selectCurrentDomain } from '../../store/domainSlice';
import { useSelector } from 'react-redux';
import { useDeleteDomainMutation } from '../../store/domainApiSlice';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import { Dispatch, SetStateAction, useEffect, useRef } from 'react';
import { AsyncEventResponse } from '../../models/config';
import { skipToken } from '@reduxjs/toolkit/query';
import { useNavigate } from 'react-router-dom';
import { AsyncEventStatus } from '../../models/async';

interface DeleteDomainButtonProps {
    setIsModalVisible: Dispatch<SetStateAction<boolean>>;
    setIsAsyncRunning: Dispatch<SetStateAction<boolean>>;
}

export const DeleteDomainButton = (props: DeleteDomainButtonProps) => {
    const navigate = useNavigate();
    const selectedDomain = useSelector(selectCurrentDomain);
    const [deleteDomainTrigger, { data: deleteDomainData, error: deleteDomainError, isSuccess: isDeleteDomainSuccess }] =
        useDeleteDomainMutation();

    const asyncRunStatus = useRef<AsyncEventResponse>();

    const {
        data: asyncData,
        error: asyncError,
        isSuccess: isAsyncSuccess,
    } = useGetAsyncStatusQuery(
        isDeleteDomainSuccess &&
            deleteDomainData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? { id: deleteDomainData.item_type, useCase: deleteDomainData.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (isAsyncSuccess) {
            props.setIsAsyncRunning(true);
            if (asyncData.status === AsyncEventStatus.EVENT_STATUS_SUCCESS || asyncData.status === AsyncEventStatus.EVENT_STATUS_FAILED) {
                navigate('/');
            }
        }
    }, [asyncData]);

    const onDeleteDomainButtonClick = (): void => {
        props.setIsModalVisible(false);
        asyncRunStatus.current = undefined;
        deleteDomainTrigger(selectedDomain.customerProfileDomain);
    };

    return (
        <Button data-testid="confirmDeleteDomainButton" variant="primary" onClick={onDeleteDomainButtonClick}>
            Delete Domain
        </Button>
    );
};
