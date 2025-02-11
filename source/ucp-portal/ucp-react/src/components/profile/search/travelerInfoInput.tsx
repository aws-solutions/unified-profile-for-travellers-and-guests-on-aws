// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as React from 'react';
import { Button, Form, Header, Box, Container, ColumnLayout } from '@cloudscape-design/components';
import { useEffect, useState } from 'react';
import { useGetProfilesByParamQuery, useGetProfilesByIdQuery } from '../../../store/profileApiSlice.ts';
import {
    selectConnectId,
    setConnectId,
    selectTravellerId,
    setTravellerId,
    selectEmailId,
    setEmailId,
    selectLastName,
    setLastName,
    selectPhone,
    setPhone,
    selectAddress1,
    setAddress1,
    selectAddress2,
    setAddress2,
    selectCity,
    setCity,
    selectState,
    setState,
    selectProvince,
    setProvince,
    selectPostalCode,
    setPostalCode,
    selectCountry,
    setCountry,
    setSearchResults,
    ProfileRequest,
} from '../../../store/profileSlice.ts';
import { useDispatch, useSelector } from 'react-redux';
import { OptionDefinition } from '@cloudscape-design/components/internal/components/option/interfaces';
import GenericInput from './genericInput.tsx';
import BaseSelect from '../../base/form/baseSelect.tsx';
import { selectAppAccessPermission } from '../../../store/userSlice.ts';
import { Permissions } from '../../../constants.ts';

export default function TravelerInfoInput() {
    const dispatch = useDispatch();
    const [selectedOption, setSelectedOption] = React.useState<OptionDefinition>({ label: 'Address', value: 'Address' });

    const connectId = useSelector(selectConnectId);
    const travellerId = useSelector(selectTravellerId);
    const emailId = useSelector(selectEmailId);
    const lastName = useSelector(selectLastName);
    const phone = useSelector(selectPhone);
    const address1 = useSelector(selectAddress1);
    const address2 = useSelector(selectAddress2);
    const city = useSelector(selectCity);
    const state = useSelector(selectState);
    const province = useSelector(selectProvince);
    const postalCode = useSelector(selectPostalCode);
    const country = useSelector(selectCountry);
    const [toggle, setToggle] = useState('reset');

    const params: ProfileRequest = {
        lastName: lastName,
        connectId: connectId,
        travellerId: travellerId,
        email: emailId,
        phone: phone,
        addressAddress1: address1,
        addressAddress2: address2,
        addressType: String(selectedOption.value),
        addressCity: city,
        addressState: state,
        addressProvince: province,
        addressPostalCode: postalCode,
        addressCountry: country,
    };

    // Skip token will avoid api calls until search button is clicked i.e. toggled
    const { data: fetchedById } = useGetProfilesByParamQuery(params, {
        skip: toggle !== 'param',
    });

    const { data: fetchedByParam } = useGetProfilesByIdQuery(
        { id: connectId },
        {
            skip: toggle !== 'id',
        },
    );

    useEffect(() => {
        if (fetchedById !== undefined) {
            dispatch(setSearchResults(fetchedById));
            setToggle('reset');
        }
    }, [fetchedById]);

    useEffect(() => {
        if (fetchedByParam !== undefined) {
            dispatch(setSearchResults(fetchedByParam));
            setToggle('reset');
        }
    }, [fetchedByParam]);

    const onClickSubmit = (): void => {
        if (connectId !== '') {
            setToggle('id');
        } else {
            setToggle('param');
        }
    };

    const userAppAccess = useSelector(selectAppAccessPermission);
    const canUserSearchProfile = (userAppAccess & Permissions.SearchProfilePermission) === Permissions.SearchProfilePermission;

    return (
        <>
            <Box variant="h1" padding={{ top: 'm' }}></Box>
            <Container
                header={
                    <Header
                        variant="h1"
                        actions={
                            <Button onClick={onClickSubmit} disabled={!canUserSearchProfile}>
                                Search
                            </Button>
                        }
                    >
                        Search Traveller
                    </Header>
                }
            >
                <Form>
                    <ColumnLayout columns={5} minColumnWidth={0.5}>
                        <GenericInput actionCreator={setConnectId} selector={selectConnectId} labelName="Connect Id" />
                        <GenericInput actionCreator={setTravellerId} selector={selectTravellerId} labelName="Traveller Id" />
                        <GenericInput actionCreator={setEmailId} selector={selectEmailId} labelName="Email" />
                        <GenericInput actionCreator={setLastName} selector={selectLastName} labelName="Last Name" />
                        <GenericInput actionCreator={setPhone} selector={selectPhone} labelName="Phone" />
                    </ColumnLayout>
                </Form>
                <br />

                <Form>
                    <ColumnLayout columns={5} minColumnWidth={0.5}>
                        <BaseSelect
                            selectedOption={selectedOption}
                            onChange={({ detail }) => setSelectedOption(detail.selectedOption)}
                            options={[
                                { label: 'No Address', value: 'NoAddress' },
                                { label: 'Address', value: 'Address' },
                                { label: 'Shipping Address', value: 'ShippingAddress' },
                                { label: 'Billing Address', value: 'BillingAddress' },
                                { label: 'Mailing Address', value: 'MailingAddress' },
                            ]}
                            label="Address Type"
                        />
                    </ColumnLayout>
                </Form>
                <br />

                {selectedOption.value !== 'NoAddress' && (
                    <Form>
                        <ColumnLayout columns={5} minColumnWidth={0.5}>
                            <GenericInput actionCreator={setAddress1} selector={selectAddress1} labelName="Address Line 1" />
                            <GenericInput actionCreator={setAddress2} selector={selectAddress2} labelName="Address Line 2" />
                            <GenericInput actionCreator={setCity} selector={selectCity} labelName="City" />
                            <GenericInput actionCreator={setState} selector={selectState} labelName="State" />
                            <GenericInput actionCreator={setProvince} selector={selectProvince} labelName="Province" />
                            <GenericInput actionCreator={setPostalCode} selector={selectPostalCode} labelName="Postal Code" />
                            <GenericInput actionCreator={setCountry} selector={selectCountry} labelName="Country" />
                        </ColumnLayout>
                    </Form>
                )}
            </Container>
        </>
    );
}
