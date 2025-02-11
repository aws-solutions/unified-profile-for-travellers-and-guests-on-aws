// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, Button, Header, SpaceBetween } from '@cloudscape-design/components';
import { ClickDetail } from '@cloudscape-design/components/internal/events/index';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import { ProfileResponse } from '../../../models/config.ts';
import { ROUTES } from '../../../models/constants.ts';
import { addNotification } from '../../../store/notificationsSlice.ts';
import { useCreateSearchMutation } from '../../../store/privacyApiSlice.ts';
import ValueWithLabel from '../../base/form/baseValueWithLabel.tsx';
import { AsyncEventStatus } from '../../../models/async.ts';
import { useLazyGetProfilesSummaryQuery } from '../../../store/profileApiSlice.ts';
import { useEffect, useState } from 'react';
import { useGetPromptConfigQuery } from '../../../store/domainApiSlice.ts';
import { selectIsPromptActive } from '../../../store/domainSlice.ts';
import { IconName } from '../../../models/iconName.ts';
import { selectProfileSummary } from '../../../store/profileSlice.ts';

interface BasicInformationProp {
    profileData: ProfileResponse;
}

function buildFullName(profileData: ProfileResponse) {
    return (
        (profileData.honorific ? profileData.honorific + ' ' : '') +
        (profileData.firstName ? profileData.firstName + ' ' : '') +
        (profileData.middleName ? profileData.middleName + ' ' : '') +
        (profileData.lastName ? profileData.lastName : '')
    );
}

export default function BasicInformation({ profileData }: BasicInformationProp) {
    const [triggerPrivacySearch, { isLoading }] = useCreateSearchMutation();
    const dispatch = useDispatch();

    useGetPromptConfigQuery();
    const isSummaryGenerationActive = useSelector(selectIsPromptActive);
    const [triggerGetSummary] = useLazyGetProfilesSummaryQuery();
    const summaryValue = useSelector(selectProfileSummary);

    const searchForDataLocationButtonClick = async (_: CustomEvent<ClickDetail>) => {
        const response = await triggerPrivacySearch({ privacySearchRq: [{ connectId: profileData.connectId }] }).unwrap();
        if (response.status === AsyncEventStatus.EVENT_STATUS_INVOKED) {
            dispatch(
                addNotification({
                    id: response.item_type,
                    type: 'success',
                    content: (
                        <>
                            Created Privacy Search. Navigate to the <Link to={`${ROUTES.PRIVACY}`}>privacy screen</Link> to check results.
                        </>
                    ),
                }),
            );
        }
    };

    const headerActions = (
        <SpaceBetween direction="horizontal" size="xs">
            {isSummaryGenerationActive && (
                <Button onClick={() => triggerGetSummary(profileData.connectId)} iconName={IconName.GEN_AI}>
                    Generate summary
                </Button>
            )}
            <Button disabled={isLoading} onClick={searchForDataLocationButtonClick}>
                Search for Data Locations
            </Button>
        </SpaceBetween>
    );
    return (
        <>
            <Header variant="h1" actions={headerActions}>
                <Box variant="h1" padding={{ top: 'xl', bottom: 'xxl' }}>
                    {buildFullName(profileData)}
                </Box>
            </Header>
            <SpaceBetween direction="horizontal" size="l">
                <Box padding={{ right: 'xxxl' }}>
                    <ValueWithLabel label={'Company Name'} value={profileData.companyName} />
                </Box>
                <Box padding={{ right: 'xxxl' }}>
                    <ValueWithLabel label={'Job Title'} value={profileData.jobTitle} />
                </Box>
                <Box padding={{ right: 'xxxl' }}>
                    <ValueWithLabel label={'Business Email'} value={profileData.businessEmailAddress} />
                </Box>
                <Box padding={{ right: 'xxxl' }}>
                    <ValueWithLabel label={'Personal Email'} value={profileData.personalEmailAddress} />
                </Box>
                <Box padding={{ right: 'xxxl' }}>
                    <ValueWithLabel label={'Phone'} value={profileData.phoneNumber} />
                </Box>
            </SpaceBetween>
            {summaryValue && (
                <Box>
                    <br />
                    <ValueWithLabel label={'Traveler summary'} value={''} />
                    {summaryValue}
                </Box>
            )}
        </>
    );
}
