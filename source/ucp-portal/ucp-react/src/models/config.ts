// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { AsyncEventType } from './async';
import { PaginationMetadata } from './pagination';
import { PrivacyPurgeStatus, PrivacySearchResponse } from './privacy';
import {
    AirBooking,
    AirLoyalty,
    AncillaryService,
    Clickstream,
    CustomerServiceInteraction,
    EmailHistory,
    HotelBooking,
    HotelLoyalty,
    HotelStay,
    MergeContext,
    LoyaltyTx,
    PhoneHistory,
} from './traveller';
import { MatchPair } from './match';

export interface TxIdResponse {
    TxID: string;
}

export interface ConfigResponse extends TxIdResponse {
    profiles: ProfileResponse[];
    ingestionErrors: IngestionError[];
    dataValidation: null;
    totalErrors: number;
    totalMatchPairs: number;
    config: DomainConfig;
    matches: ProfileMatches[];
    matchPairs: MatchPair[];
    error: { code: string; msg: string; type: string };
    pagination: null;
    connectors: Connector[];
    awsResources: AWSResources;
    mergeResponse: string;
    portalConfig: PortalConfig;
    ruleSets: RuleSet[];
    profileMappings: string[];
    accpRecords: AccpRecord[];
    asyncEvent: AsyncEventResponse;
    paginationMetadata: PaginationMetadata;
    privacySearchResults: PrivacySearchResponse[];
    privacySearchResult: PrivacySearchResponse;
    privacyPurgeStatus: PrivacyPurgeStatus;
    interactionHistory: MergeContext[];
    domainSetting: DomainSetting;
    profileSummary: string;
}

interface Address {
    Address1: string;
    Address2: string;
    Address3: string;
    Address4: string;
    City: string;
    State: string;
    Province: string;
    PostalCode: string;
    Country: string;
}

export interface ProfileMatches {
    id: string;
    birthDate: string;
    confidence: string;
    email: string;
    firstName: string;
    lastName: string;
    phone: string;
}

export interface ProfileResponse {
    modelVersion: string;
    errors: string[];
    lastUpdated: string; //time.Time
    lastUpdatedBy: string;
    domain: string;
    mergedIn: string;

    connectId: string;
    travellerId: string;
    pssId: string;
    gdsId: string;
    pmsId: string;
    crsId: string;

    honorific: string;
    firstName: string;
    middleName: string;
    lastName: string;
    gender: string;
    pronoun: string;
    birthDate: string; //time.Time
    jobTitle: string;
    companyName: string;

    phoneNumber: string;
    mobilePhoneNumber: string;
    homePhoneNumber: string;
    businessPhoneNumber: string;
    personalEmailAddress: string;
    businessEmailAddress: string;
    nationalityCode: string;
    nationalityName: string;
    languageCode: string;
    languageName: string;

    homeAddress: Address;
    businessAddress: Address;
    mailingAddress: Address;
    billingAddress: Address;

    airBookingRecords: AirBooking[];
    airLoyaltyRecords: AirLoyalty[];
    ancillaryServiceRecords: AncillaryService[];
    emailHistoryRecords: EmailHistory[];
    phoneHistoryRecords: PhoneHistory[];
    hotelBookingRecords: HotelBooking[];
    hotelLoyaltyRecords: HotelLoyalty[];
    hotelStayRecords: HotelStay[];
    customerServiceInteractionRecords: CustomerServiceInteraction[];
    clickstreamRecords: Clickstream[];
    parsingErrors: string[];
    lotaltyTxRecords: LoyaltyTx[];
}

export interface IngestionError {
    error_type: string;
    error_id: string;
    category: string;
    message: string;
    domain: string;
    businessObjectType: string;
    accpRecordType: string;
    reccord: string;
    travelerId: string;
    timestamp: Date;
    transactionId: string;
}

export interface DomainConfig {
    domains: Domain[];
}

export interface Connector {
    name: string;
    status: ConnectorStatus;
    linkedDomains: string[];
}

export enum ConnectorStatus {
    DEPLOYED_WITHOUT_BUCKET = 'deployed_without_bucket',
    DEPLOYED_WITH_BUCKET = 'deployed_with_bucket',
}

export interface PortalConfig {
    hyperlinkMappings: HyperLinkMapping[];
}

export interface HyperLinkMapping {
    accpObject: string;
    fieldName: string;
    hyperlinkTemplate: string;
}

export interface RuleSet {
    name: string;
    rules: Rule[];
    lastActivationTimestamp: Date;
    latestVersion: number;
}

export interface Rule {
    index: number;
    name: string;
    description: string;
    conditions: Condition[];
}

export type Condition = SkipCondition | MatchCondition | FilterCondition;

export interface BaseCondition {
    index: number;
    conditionType: ConditionType;
    incomingObjectType: string;
    incomingObjectField: string;
    indexNormalization: IndexNormalizationSettings;
}

export interface SkipCondition extends BaseCondition {
    incomingObjectType2?: string;
    incomingObjectField2?: string;
    skipConditionValue?: Value;
    op: RuleOperation;
}

export interface MatchCondition extends BaseCondition {
    existingObjectType?: string;
    existingObjectField?: string;
}

export interface FilterCondition extends BaseCondition {
    existingObjectType: string;
    existingObjectField: string;
    op: RuleOperation;
    filterConditionValue: Value;
}

export interface Value {
    valueType: ValueType;
    stringValue?: string;
    intValue?: number;
}

export enum ValueType {
    STRING = 'string',
    INT = 'int',
}

export enum ConditionType {
    SKIP = 'skip',
    MATCH = 'match',
    FILTER = 'filter',
}

export enum RuleOperation {
    RULE_OP_EQUALS = 'equals',
    RULE_OP_EQUALS_VALUE = 'equals_value',
    RULE_OP_NOT_EQUALS = 'not_equals',
    RULE_OP_NOT_EQUALS_VALUE = 'not_equals_value',
    RULE_OP_MATCHES_REGEXP = 'matches_regexp',
    RULE_OP_WITHIN_SECONDS = 'within_seconds',
}

export interface IndexNormalizationSettings {
    lowercase: boolean;
    trim: boolean;
    removeSpaces: boolean;
    removeNonNumeric: boolean;
    removeNonAlphaNumeric: boolean;
    removeLeadingZeros: boolean;
}

export interface Domain {
    customerProfileDomain: string;
    numberOfObjects: number;
    numberOfProfiles: number;
    matchingEnabled: boolean;
    cacheMode: number;
    created: string;
    lastUpdated: string;
    mappings: null;
    integrations: null;
    needsMappingUpdate: boolean;
}

export interface AWSResources {
    glueRoleArn: string;
    tahConnectorBucketPolicy: string;
    S3Buckets: S3Bucket;
    jobs: Job[];
}

export interface Job {
    jobName: string;
    lastRunTime: Date;
    status: string;
}

export interface PromptConfig {
    value: string;
    isActive: boolean;
}

export interface DomainSetting {
    promptConfig: PromptConfig;
    matchConfig: PromptConfig;
}

export interface PromptConfigRq {
    domainSetting: {
        promptConfig: PromptConfig;
        matchConfig: PromptConfig;
    };
}

export interface AccpRecord {
    name: string;
    objectFields: string[];
}

export interface S3Bucket {
    CONNECT_PROFILE_SOURCE_BUCKET: string;
    S3_AIR_BOOKING: string;
    S3_CLICKSTREAM: string;
    S3_CSI: string;
    S3_GUEST_PROFILE: string;
    S3_HOTEL_BOOKING: string;
    S3_PAX_PROFILE: string;
    S3_STAY_REVENUE: string;
    REGION: string;
}

export interface AsyncEventResponse {
    item_type: AsyncEventType;
    item_id: string;
    status: string;
    lastUpdated: Date;
    progress: number;
}

export type AsyncEventResponseWithTxID = AsyncEventResponse & TxIdResponse;
