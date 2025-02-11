// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Box, ColumnLayout } from '@cloudscape-design/components';
import BaseButton from '../base/form/baseButton';
import { IconName } from '../../models/iconName';
import { ConfigResponse } from '../../models/config';
import { HTTPS_PREFIX, S3_BUCKET_URL } from '../../models/constants';

interface OverviewDataRepositoryProps {
    configData?: ConfigResponse;
    isDisabled: boolean;
}

export const OverviewDataRepositories = (props: OverviewDataRepositoryProps) => {
    const s3Buckets = props.configData?.awsResources.S3Buckets;
    const region = s3Buckets?.REGION ?? 'us-east-1';

    return (
        <>
            <Box variant="h1">Data Repositories</Box>
            <ColumnLayout columns={8}>
                <BaseButton
                    title="Traveller Profile Records"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.CONNECT_PROFILE_SOURCE_BUCKET}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Air Booking"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_AIR_BOOKING}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Clickstream"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_CLICKSTREAM}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Customer Service Interactions"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_CSI}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Guest Profiles"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_GUEST_PROFILE}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Hotel Booking"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_HOTEL_BOOKING}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Passenger Profiles"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_PAX_PROFILE}`}
                    disabled={props.isDisabled}
                    isExternal
                />
                <BaseButton
                    title="Stay Revenue"
                    iconName={IconName.USER_PROFILE_ACTIVE}
                    href={`${HTTPS_PREFIX}${region}${S3_BUCKET_URL}${s3Buckets?.S3_STAY_REVENUE}`}
                    disabled={props.isDisabled}
                    isExternal
                />
            </ColumnLayout>
        </>
    );
};
