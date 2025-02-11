// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { useEffect, useRef, useState } from 'react';
import { DomainCreateRequest } from '../../../models/domain.ts';
import { Container, Form, FormField, Hotspot, Input } from '@cloudscape-design/components';
import Header from '@cloudscape-design/components/header';
import SpaceBetween from '@cloudscape-design/components/space-between';
import Button from '@cloudscape-design/components/button';
import { useNavigate } from 'react-router-dom';
import { useCreateDomainMutation, useLazyGetDomainQuery } from '../../../store/domainApiSlice.ts';
import { setDomainValueStorage } from '../../../utils/localStorageUtil.ts';
import { AsyncEventResponse } from '../../../models/config.ts';
import { useGetAsyncStatusQuery } from '../../../store/asyncApiSlice.ts';
import { skipToken } from '@reduxjs/toolkit/query';
import { ROUTES } from '../../../models/constants.ts';
import { AsyncEventStatus } from '../../../models/async.ts';
import { useSelector } from 'react-redux';
import { selectAppAccessPermission } from '../../../store/userSlice.ts';
import { Permissions } from '../../../constants.ts';

type DomainCreateFormState = {
    domain: string;
    version: string;
    portfolioId?: string;
    validationErrorMessage?: string;
    formFieldErrors: {
        name?: string;
    };
};

export const DomainCreateForm = () => {
    const navigate = useNavigate();
    const [isAsyncRunning, setIsAsyncRunning] = useState<boolean>(false);

    const [createDomainMutation, { data: createDomainData, error: createDomainError, isSuccess: isCreateDomainSuccess }] =
        useCreateDomainMutation();

    const asyncRunStatus = useRef<AsyncEventResponse>();
    const {
        data: asyncData,
        error: asyncError,
        isSuccess: isAsyncSuccess,
    } = useGetAsyncStatusQuery(
        isCreateDomainSuccess &&
            createDomainData &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_SUCCESS &&
            asyncRunStatus.current?.status !== AsyncEventStatus.EVENT_STATUS_FAILED
            ? { id: createDomainData.item_type, useCase: createDomainData.item_id }
            : skipToken,
        { pollingInterval: 2000 },
    );

    const [getDomainTrigger] = useLazyGetDomainQuery();

    const initialState: DomainCreateFormState = {
        domain: '',
        version: '',
        portfolioId: undefined,
        validationErrorMessage: undefined,
        formFieldErrors: {},
    };
    const [state, setState] = useState<DomainCreateFormState>(initialState);

    useEffect(() => {
        asyncRunStatus.current = asyncData;
        if (isAsyncSuccess) {
            setIsAsyncRunning(true);
            if (asyncData.status === AsyncEventStatus.EVENT_STATUS_SUCCESS) {
                setDomainValueStorage(state.domain);
                getDomainTrigger(state.domain);
                navigate(`/${ROUTES.SETTINGS}`);
            } else if (asyncData.status === AsyncEventStatus.EVENT_STATUS_FAILED) {
                setIsAsyncRunning(false);
            }
        }
    }, [asyncData]);

    const submit = async () => {
        if (isFormValid(state)) {
            const createRequest: DomainCreateRequest = {
                domain: {
                    customerProfileDomain: state.domain,
                },
            };
            asyncRunStatus.current = undefined;
            createDomainMutation({ createRequest });
        } else {
            setState(state => ({
                ...state,
                validationErrorMessage: 'Domain name has incorrect format. Only lowercase letters, numbers, in snake case allowed.',
                formFieldErrors: {
                    name: state.domain ? undefined : 'This field is required',
                },
            }));
        }
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserCreateDomain = (userAppAccess & Permissions.CreateDomainPermission) === Permissions.CreateDomainPermission;

    return (
        <Form
            header={
                <Hotspot hotspotId={'create-item-hotspot'} side={'left'}>
                    <Header variant="h1" description="Enter the name of the domain you want to create.">
                        Create Domain
                    </Header>
                </Hotspot>
            }
            actions={
                <SpaceBetween direction="horizontal" size="xs">
                    <Button onClick={() => navigate('/')} disabled={isAsyncRunning}>
                        Cancel
                    </Button>
                    <Hotspot hotspotId={'create-button'} side={'right'}>
                        <Button
                            disabled={isAsyncRunning || !canUserCreateDomain}
                            variant={'primary'}
                            onClick={submit}
                            data-testid="domainCreateButton"
                        >
                            Create
                        </Button>
                    </Hotspot>
                </SpaceBetween>
            }
            errorText={state.validationErrorMessage}
            data-testid="domainCreateForm"
        >
            <Container>
                <SpaceBetween size={'m'}>
                    <FormField label="Domain Name" stretch={true} errorText={state.formFieldErrors.name}>
                        <Input
                            value={state.domain}
                            onChange={({ detail }) => {
                                return setState(state => ({
                                    ...state,
                                    domain: detail.value,
                                }));
                            }}
                            data-testid="domainCreateInput"
                            disabled={!canUserCreateDomain}
                        />
                    </FormField>
                </SpaceBetween>
            </Container>
        </Form>
    );
};

// validate according to your business/data needs
function isFormValid(state: DomainCreateFormState) {
    return state.domain && isSnakeCaseLower(state.domain);
}

function isSnakeCaseLower(input: string): boolean {
    return /^[a-z0-9_]+$/.test(input);
}
