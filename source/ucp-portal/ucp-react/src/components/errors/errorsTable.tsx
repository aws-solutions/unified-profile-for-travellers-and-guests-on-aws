// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, Header, Icon, Popover, SpaceBetween, TableProps } from '@cloudscape-design/components';
import BaseCRUDTable, { CustomPropsPaginationType } from '../base/baseCrudTable';
import { useEffect, useState } from 'react';
import { IngestionError } from '../../models/config';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import { useGetAllErrorsQuery } from '../../store/errorApiSlice';
import { GetErrorsRequest } from '../../models/error';
import { DeleteErrorButton } from './deleteErrorButton';
import { IconName } from '../../models/iconName';
import { ErrorsConfirmationModal } from './errorsConfirmationModal';
import { ViewErrorBodyButton } from './viewErrorBodyButton';
import { ErrorBodyModal } from './errorBodyModal';
import { useSelector } from 'react-redux';
import { selectAppAccessPermission } from '../../store/userSlice';
import { Permissions } from '../../constants';
dayjs.extend(utc);

export const ErrorsTable = () => {
    const [isConfirmationModalVisible, setIsConfirmationModalVisible] = useState<boolean>(false);
    const [isErrorBodyModalVisible, setIsErrorBodyModalVisible] = useState<boolean>(false);
    const [errorBody, setErrorBody] = useState<string>('');
    const [currentPageIndex, setCurrentPageIndex] = useState<number>(1);
    const [tableItems, setTableItems] = useState<IngestionError[]>([]);
    const [pagesCount, setPagesCount] = useState<number>(1);
    const getErrorsRequest: GetErrorsRequest = {
        page: currentPageIndex - 1,
        pageSize: 10,
    };
    const { data: errorData, isFetching: getErrorsFetching, refetch: refetchErrors } = useGetAllErrorsQuery(getErrorsRequest);

    useEffect(() => {
        if (errorData !== undefined) {
            if (errorData.ingestionErrors.length !== 0 || currentPageIndex === 1) {
                setTableItems(errorData.ingestionErrors);
                setPagesCount(Math.ceil(errorData.totalErrors / 10));
            } else {
                setCurrentPageIndex(currentPageIndex - 1);
            }
        }
    }, [errorData]);

    const pagination: CustomPropsPaginationType = {
        onPageChange: newPage => setCurrentPageIndex(newPage),
        currentPageIndex,
        pagesCount,
        openEnded: true,
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserClearErrors = (userAppAccess & Permissions.ClearAllErrorsPermission) === Permissions.ClearAllErrorsPermission;

    const columnDefs: TableProps.ColumnDefinition<IngestionError>[] = [
        {
            id: 'timestamp',
            header: 'Timestamp',
            cell: ingestionError => dayjs.utc(ingestionError.timestamp).format('YYYY-MM-DD hh:mm:ss'),
            ariaLabel: () => 'timestamp',
        },
        {
            id: 'errorType',
            header: 'Error Type',
            cell: ingestionError => ingestionError.error_type,
            ariaLabel: () => 'errorType',
        },
        {
            id: 'businessObject',
            header: 'Business Object',
            cell: ingestionError => ingestionError.businessObjectType || 'N/A',
            ariaLabel: () => 'businessObject',
        },
        {
            id: 'message',
            header: 'Message',
            cell: ingestionError => ingestionError.message,
            ariaLabel: () => 'message',
            maxWidth: '500px',
        },
        {
            id: 'action',
            header: 'Action',
            cell: ingestionError => (
                <SpaceBetween direction="horizontal" size="xs">
                    <ViewErrorBodyButton
                        errorBody={ingestionError.reccord}
                        setIsErrorBodyModalVisible={setIsErrorBodyModalVisible}
                        setErrorBody={setErrorBody}
                    />
                    {canUserClearErrors && <DeleteErrorButton errorId={ingestionError.error_id} />}
                </SpaceBetween>
            ),
            ariaLabel: () => 'action',
        },
    ];

    const counterText = '(' + errorData?.totalErrors + ')';
    const actions = (
        <SpaceBetween direction="horizontal" size="s">
            <Button
                variant="icon"
                iconName={IconName.REFRESH}
                onClick={() => {
                    refetchErrors();
                }}
            />
            <Button
                disabled={tableItems.length === 0 || !canUserClearErrors}
                variant="primary"
                onClick={() => setIsConfirmationModalVisible(true)}
                iconName={IconName.REMOVE}
            >
                Clear All Errors
            </Button>
        </SpaceBetween>
    );

    const header = (
        <Header variant="h1" actions={actions} counter={counterText}>
            Errors{' '}
            <Popover
                dismissButton={false}
                position="top"
                size="small"
                triggerType="text"
                content="Error count may take up to 6 hours to update"
            >
                <Icon name={IconName.STATUS_INFO} />
            </Popover>
        </Header>
    );

    return (
        <>
            <BaseCRUDTable<IngestionError>
                columnDefinitions={columnDefs}
                items={tableItems}
                header={header}
                preferences={{}}
                isLoading={getErrorsFetching}
                tableVariant="container"
                customPropsPagination={pagination}
                isFilterable={false}
                wrapLines={true}
                data-testid="errorsTable"
            />
            <ErrorsConfirmationModal
                isConfirmationModalVisible={isConfirmationModalVisible}
                setIsConfirmationModalVisible={setIsConfirmationModalVisible}
                refetch={refetchErrors}
            />
            <ErrorBodyModal
                isErrorBodyModalVisible={isErrorBodyModalVisible}
                setIsErrorBodyModalVisible={setIsErrorBodyModalVisible}
                errorBody={errorBody}
            />
        </>
    );
};
