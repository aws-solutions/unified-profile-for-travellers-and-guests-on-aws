// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { AsyncEventType } from '../../models/async.ts';
import { ConfigResponse, Domain, DomainConfig } from '../../models/config.ts';

export function returnConfig(txId: string, domainName: string, numberOfProfiles: number, multipleDomains: boolean): ConfigResponse {
    const responseDomain: Domain = {
        customerProfileDomain: domainName,
        numberOfObjects: 0,
        numberOfProfiles: numberOfProfiles,
        matchingEnabled: false,
        cacheMode: 0,
        created: '',
        lastUpdated: '',
        mappings: null,
        integrations: null,
        needsMappingUpdate: false,
    };

    const mockDomainsConfig: DomainConfig = {
        domains: [responseDomain],
    };
    const secondaryDomain: Domain = {
        customerProfileDomain: 'secondary_domain',
        numberOfObjects: 0,
        numberOfProfiles: numberOfProfiles,
        matchingEnabled: false,
        cacheMode: 0,
        created: '',
        lastUpdated: '',
        mappings: null,
        integrations: null,
        needsMappingUpdate: false,
    };

    if (multipleDomains) {
        mockDomainsConfig.domains.push(secondaryDomain);
    }
    return {
        TxID: txId,
        profiles: [],
        ingestionErrors: [],
        dataValidation: null,
        totalErrors: 0,
        totalMatchPairs: 0,
        config: mockDomainsConfig,
        matches: [],
        matchPairs: [],
        error: { code: '', msg: '', type: '' },
        pagination: null,
        connectors: [],
        awsResources: {
            glueRoleArn: '',
            tahConnectorBucketPolicy: '',
            S3Buckets: {
                CONNECT_PROFILE_SOURCE_BUCKET: '',
                S3_AIR_BOOKING: '',
                S3_CLICKSTREAM: '',
                S3_CSI: '',
                S3_GUEST_PROFILE: '',
                S3_HOTEL_BOOKING: '',
                S3_PAX_PROFILE: '',
                S3_STAY_REVENUE: '',
                REGION: '',
            },
            jobs: [],
        },
        mergeResponse: '',
        portalConfig: { hyperlinkMappings: [] },
        ruleSets: [],
        profileMappings: [],
        accpRecords: [],
        asyncEvent: {
            item_id: '',
            item_type: AsyncEventType.NULL,
            status: '',
            lastUpdated: new Date(),
            progress: 0,
        },
        paginationMetadata: {
            airBookingRecords: { total_records: 0 },
            airLoyaltyRecords: { total_records: 0 },
            ancillaryServiceRecords: { total_records: 0 },
            clickstreamRecords: { total_records: 0 },
            customerServiceInteractionRecords: { total_records: 0 },
            emailHistoryRecords: { total_records: 0 },
            hotelBookingRecords: { total_records: 0 },
            hotelLoyaltyRecords: { total_records: 0 },
            hotelStayRecords: { total_records: 0 },
            lotaltyTxRecords: { total_records: 0 },
            phoneHistoryRecords: { total_records: 0 },
        },
        privacySearchResults: [],
        privacySearchResult: {
            connectId: '',
            domainName: '',
            searchDate: new Date(),
            status: '',
            locationResults: {},
            errorMessage: '',
            totalResultsFound: 0,
        },
        privacyPurgeStatus: {
            isPurgeRunning: false,
        },
        interactionHistory: [],
        domainSetting: {
            promptConfig: { value: '', isActive: false },
            matchConfig: { value: '', isActive: false },
        },
        profileSummary: '',
    };
}
