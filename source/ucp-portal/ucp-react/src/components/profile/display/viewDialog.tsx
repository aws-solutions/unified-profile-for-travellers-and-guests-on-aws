// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, ColumnLayout, Container, Header, Link, Modal, TableProps } from '@cloudscape-design/components';
import { skipToken } from '@reduxjs/toolkit/query';
import { Buffer } from 'buffer';
import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Permissions } from '../../../constants';
import { AsyncEventStatus } from '../../../models/async';
import { AsyncEventResponse } from '../../../models/config';
import { ConversationItem, CustomerServiceInteraction, MergeContext, MergeType } from '../../../models/traveller';
import { useGetAsyncStatusQuery } from '../../../store/asyncApiSlice';
import { useGetInteractionHistoryQuery, usePerformUnmergeMutation } from '../../../store/profileApiSlice';
import {
    DialogType,
    RecordType,
    selectDialogRecord,
    selectDialogType,
    selectInteractionHistory,
    setDialogType,
} from '../../../store/profileSlice';
import { selectAppAccessPermission } from '../../../store/userSlice';
import { getEventTypeDisplayValue } from '../../../utils/profileUtils';
import BaseCRUDTable from '../../base/baseCrudTable';
import { useBasePreferences } from '../../base/basePreferences';
import DefaultCollectionPreferences from '../../base/defaultCollectionPreferences';

interface ViewRecordProp<T> {
    record: T;
}

interface ViewRecordItem {
    key: string;
    value: string;
}

interface DialogProps {
    record: RecordType;
}

function TabularRecord<T>({ record }: ViewRecordProp<T>) {
    const tableItem: ViewRecordItem[] = [];

    for (const [key, value] of Object.entries(record as object)) {
        tableItem.push({
            key: key,
            value: value,
        });
    }

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<ViewRecordItem>[] = [
        {
            id: 'key',
            header: 'Key',
            cell: ({ key }) => key,
            ariaLabel: () => 'key',
        },
        {
            id: 'value',
            header: 'Value',
            cell: ({ value }) => value,
            ariaLabel: () => 'value',
        },
    ];

    return (
        <BaseCRUDTable<ViewRecordItem>
            isFilterable={false}
            columnDefinitions={columnDefinitions}
            items={tableItem}
            preferences={preferences}
            tableVariant="embedded"
            stripedRows={true}
            collectionPreferences={<DefaultCollectionPreferences preferences={preferences} onConfirm={onConfirm} />}
        />
    );
}

function CustomerServiceInteractionDialog({ record }: DialogProps) {
    const interaction = record as CustomerServiceInteraction;
    const bufferData = JSON.parse(Buffer.from(interaction.conversation, 'base64').toString('binary'));
    const conversation: ConversationItem[] = bufferData.items;
    conversation.sort((a, b) => {
        return new Date(a.start_time).getTime() - new Date(b.start_time).getTime();
    });
    const elements = [];
    for (let i = 0; i < conversation.length; i++) {
        const message = (
            <div key={i}>
                <Box float="left">
                    <Container fitHeight={true} header={<Header>{conversation[i].from}</Header>}>
                        <p>{new Date(conversation[i].start_time).toLocaleString()}</p>
                        <Box margin={'xxxs'}>{conversation[i].content}</Box>
                    </Container>
                </Box>
            </div>
        );
        if (conversation[i].from === conversation[0].from) {
            // left align the message
            elements.push(message);
            elements.push(<div key={i + 'alt2'}></div>);
        } else {
            // right align the message
            elements.push(<div key={i + 'alt1'}></div>);
            elements.push(message);
        }
    }
    return <ColumnLayout columns={2}>{elements}</ColumnLayout>;
}

function InteractionHistory({ record }: DialogProps) {
    useGetInteractionHistoryQuery({
        id: record.accpObjectID,
        params: {
            objectType: record.interactionType,
            connectId: record.connectId
        },
    });
    const history = useSelector(selectInteractionHistory);

    const [performUnmergeTrigger, { data: performUnmergeData, error: performUnmergeError, isSuccess: isPerformUnmergeSuccess }] =
        usePerformUnmergeMutation();

    const asyncRunStatus = useRef<AsyncEventResponse>();
    const { data: asyncData, isSuccess: isAsyncSuccess } = useGetAsyncStatusQuery(
        isPerformUnmergeSuccess &&
            performUnmergeData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS
            ? { id: performUnmergeData.asyncEvent.item_type, useCase: performUnmergeData.asyncEvent.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (isAsyncSuccess && asyncData.status === AsyncEventStatus.EVENT_STATUS_SUCCESS) window.location.reload();
    }, [asyncData]);

    const onClickView = (rq: MergeContext) => {
        if (confirm('Are you sure you want to undo this merge? Note: This may unmerge other interactions belonging to the profile.')) {
            asyncRunStatus.current = undefined;
            performUnmergeTrigger({
                interactionToUnmerge: record.accpObjectID,
                interactionType: record.interactionType,
                mergedIntoConnectID: rq.mergeIntoConnectId,
                toUnmergeConnectID: rq.toMergeConnectId,
            });
        }
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserUnmerge = (userAppAccess & Permissions.UnmergeProfilePermission) === Permissions.UnmergeProfilePermission;

    const { preferences, onConfirm } = useBasePreferences({});
    const columnDefinitions: TableProps.ColumnDefinition<MergeContext>[] = [
        {
            id: 'timestamp',
            header: 'Timestamp',
            cell: ({ timestamp }) => {
                const date = new Date(timestamp);
                return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
            },
            ariaLabel: () => 'timestamp',
        },
        {
            id: 'eventType',
            header: 'Event Type',
            cell: item => getEventTypeDisplayValue(item, record, canUserUnmerge, () => onClickView(item)),
            ariaLabel: () => 'eventType',
        },
        {
            id: 'confidenceUpdateFactor',
            header: 'Confidence Update Factor',
            cell: item =>
                item.mergeType == MergeType.UNMERGE
                    ? 'Revert by ' + item.confidenceUpdateFactor.toPrecision(3)
                    : item.confidenceUpdateFactor.toPrecision(3),
            ariaLabel: () => 'confidenceUpdateFactor',
        },
        {
            id: 'mergedFromConnectId',
            header: 'Merged from Profile',
            cell: ({ toMergeConnectId }) => {
                if (toMergeConnectId !== '')
                    return (
                        <Link external href={`/profile/${toMergeConnectId}`}>
                            {toMergeConnectId}
                        </Link>
                    );
            },
            ariaLabel: () => 'mergedIntoConnectId',
        },
        {
            id: 'mergedIntoConnectId',
            header: 'Merged into Profile',
            cell: ({ mergeIntoConnectId }) => {
                if (mergeIntoConnectId !== '')
                    return (
                        <Link external href={`/profile/${mergeIntoConnectId}`}>
                            {mergeIntoConnectId}
                        </Link>
                    );
            },
            ariaLabel: () => 'mergedIntoConnectId',
        },
        {
            id: 'ruleID',
            header: 'Rule ID',
            cell: ({ ruleId }) => ruleId,
            ariaLabel: () => 'ruleID',
        },
        {
            id: 'ruleSetVersion',
            header: 'Ruleset Version',
            cell: ({ ruleSetVersion }) => ruleSetVersion,
            ariaLabel: () => 'ruleSetVersion',
        },
        {
            id: 'operatorId',
            header: 'Operator Id',
            cell: ({ operatorId }) => operatorId,
            ariaLabel: () => 'operatorId',
        },
    ];

    return (
        <BaseCRUDTable<MergeContext>
            isFilterable={false}
            columnDefinitions={columnDefinitions}
            items={history}
            preferences={preferences}
            tableVariant="embedded"
            stripedRows={true}
            collectionPreferences={<DefaultCollectionPreferences preferences={preferences} onConfirm={onConfirm} />}
        />
    );
}

export default function ViewRecordDialog() {
    const dispatch = useDispatch();
    const dialogType = useSelector(selectDialogType);
    const isVisible = dialogType === DialogType.NONE ? false : true;
    const record = useSelector(selectDialogRecord);

    return (
        <Modal
            onDismiss={() => dispatch(setDialogType(DialogType.NONE))}
            visible={isVisible}
            header={dialogType.charAt(0).toUpperCase() + dialogType.slice(1) + ' details'}
            size="large"
        >
            {dialogType === DialogType.RECORD && <TabularRecord record={record} />}
            {dialogType === DialogType.CHAT && record && <CustomerServiceInteractionDialog record={record} />}
            {dialogType === DialogType.HISTORY && record && <InteractionHistory record={record} />}
        </Modal>
    );
}
