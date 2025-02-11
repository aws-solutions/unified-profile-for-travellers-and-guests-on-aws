// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box } from '@cloudscape-design/components';
import { useSelector } from 'react-redux';
import { ConfigResponse, ProfileResponse } from '../../../models/config.ts';
import { Metadata, MultiPaginationProp } from '../../../models/pagination.ts';
import { useGetPortalConfigQuery } from '../../../store/configApiSlice.ts';
import { useGetProfilesByIdQuery } from '../../../store/profileApiSlice.ts';
import { selectPaginationParam, selectViewProfile } from '../../../store/profileSlice.ts';
import AirBookingSection from './airBookingSection';
import AirLoyaltySection from './airLoyaltySection';
import AncillaryServiceSection from './ancillaryServiceSection';
import BasicInformation from './basicInformation.tsx';
import ClickstreamSection from './clickstreamSection';
import CustomerServiceSection from './customerServiceSection';
import EmailHistorySection from './emailHistorySection';
import HotelBookingSection from './hotelBookingSection';
import HotelLoyaltySection from './hotelLoyaltySection';
import HotelStaySection from './hotelStaySection';
import LoyaltyTxSection from './loyaltyTxSection';
import ParsingErrorSection from './parsingErrorSection';
import PhoneHistorySection from './phoneHistorySection';
import ProfileMatchesSection from './profileMatchesSection';
import ViewRecordDialog from './viewDialog.tsx';

interface DisplayHomeProp {
    profileId: string;
}

export interface SectionProp {
    profileData: ProfileResponse;
    paginationMetadata: Metadata;
}

export default function DisplayHome({ profileId }: DisplayHomeProp) {
    const paginationParams: MultiPaginationProp = useSelector(selectPaginationParam);
    const response: ConfigResponse = useSelector(selectViewProfile);

    const { error: getProfilesByIdQueryError } = useGetProfilesByIdQuery({ id: profileId, params: paginationParams });
    useGetPortalConfigQuery();

    if (getProfilesByIdQueryError) {
        return (
            <Box variant="h1" textAlign="center" margin="xxxl">
                Profile Not Found.
            </Box>
        );
    }
    if (response.profiles === undefined || response.profiles.length === 0) {
        return <></>;
    }
    const profileData = response.profiles[0];

    return (
        <>
            <BasicInformation profileData={profileData} />
            <br />
            <br />
            <AirBookingSection profileData={profileData} paginationMetadata={response.paginationMetadata.airBookingRecords} />
            <AirLoyaltySection profileData={profileData} paginationMetadata={response.paginationMetadata.airLoyaltyRecords} />
            <AncillaryServiceSection profileData={profileData} paginationMetadata={response.paginationMetadata.ancillaryServiceRecords} />
            <EmailHistorySection profileData={profileData} paginationMetadata={response.paginationMetadata.emailHistoryRecords} />
            <PhoneHistorySection profileData={profileData} paginationMetadata={response.paginationMetadata.phoneHistoryRecords} />
            <HotelBookingSection profileData={profileData} paginationMetadata={response.paginationMetadata.hotelBookingRecords} />
            <HotelLoyaltySection profileData={profileData} paginationMetadata={response.paginationMetadata.hotelLoyaltyRecords} />
            <HotelStaySection profileData={profileData} paginationMetadata={response.paginationMetadata.hotelStayRecords} />
            <LoyaltyTxSection profileData={profileData} paginationMetadata={response.paginationMetadata.lotaltyTxRecords} />
            <CustomerServiceSection
                profileData={profileData}
                paginationMetadata={response.paginationMetadata.customerServiceInteractionRecords}
            />
            <ClickstreamSection profileData={profileData} paginationMetadata={response.paginationMetadata.clickstreamRecords} />
            <ProfileMatchesSection profileMatches={response.matches} />
            <ParsingErrorSection profileData={profileData} />
            <ViewRecordDialog />
        </>
    );
}
