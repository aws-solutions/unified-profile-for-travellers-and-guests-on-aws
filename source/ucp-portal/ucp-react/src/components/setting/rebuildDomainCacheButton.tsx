// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { useRebuildCacheMutation } from '../../store/cacheApiSlice';
import { Dispatch, SetStateAction, useEffect, useRef } from 'react';
import { CacheMode } from '../../models/domain';
import { AsyncEventResponse } from '../../models/config';
import { useGetAsyncStatusQuery } from '../../store/asyncApiSlice';
import { AsyncEventStatus } from '../../models/async';
import { skipToken } from '@reduxjs/toolkit/query';

interface RebuildDomainCacheButtonProps {
    setIsModalVisible: Dispatch<SetStateAction<boolean>>;
    setCpChecked: Dispatch<SetStateAction<boolean>>;
    setDynamoChecked: Dispatch<SetStateAction<boolean>>;
    cacheMode: CacheMode;
}

export const RebuildDomainCacheButton = ({
    setIsModalVisible,
    setCpChecked,
    setDynamoChecked,
    cacheMode,
}: RebuildDomainCacheButtonProps) => {
    const [rebuildCacheTrigger, { data: rebuildData, isSuccess: isRebuildSuccess }] = useRebuildCacheMutation();

    const asyncRunStatus = useRef<AsyncEventResponse>();

    const { data: asyncData } = useGetAsyncStatusQuery(
        isRebuildSuccess &&
            rebuildData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? { id: rebuildData.item_type, useCase: rebuildData.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
    }, [asyncData]);

    const onRebuildDomainCacheButtonClick = (): void => {
        if (cacheMode !== CacheMode.NONE) {
            asyncRunStatus.current = undefined;
            rebuildCacheTrigger(cacheMode);
        }
        setCpChecked(false);
        setDynamoChecked(false);
        setIsModalVisible(false);
    };

    return (
        <Button data-testid="rebuildCacheButton" variant="primary" onClick={onRebuildDomainCacheButtonClick}>
            Rebuild Cache
        </Button>
    );
};
