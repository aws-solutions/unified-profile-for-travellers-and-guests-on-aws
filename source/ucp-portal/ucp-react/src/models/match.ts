// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export interface MatchPair {
    domain_sourceProfileId: string;
    match_targetProfileId: string;
    targetProfileID: string;
    domain: string;
    score: string;
    profileDataDiff: DataDiff[];
    runId: string;
    scoreTargetId: string;
    mergeInProgress: boolean;
}

export interface DataDiff {
    fieldName: string;
    sourceValue: string;
    targetValue: string;
}

export interface MatchDiff {
    Id: number;
    SourceProfileId: string;
    TargetProfileId: string;
    Score: string;
    Domain: string;
    ProfileAttributes: ProfileAttributes;
    ContactInfoAttributes: ContactInfoAttributes;
    AddressAtrributes: AddressAtrributes;
    ShippingAddress: AddressAtrributes;
    MailingAddress: AddressAtrributes;
    BillingAddress: AddressAtrributes;
}

export interface ProfileAttributes {
    BusinessName: DataDiff;
    FirstName: DataDiff;
    LastName: DataDiff;
    BirthDate: DataDiff;
    Gender: DataDiff;
}

export interface ContactInfoAttributes {
    EmailAddress: DataDiff;
    PersonalEmailAddress: DataDiff;
    BusinessEmailAddress: DataDiff;
    PhoneNumber: DataDiff;
    MobilePhoneNumber: DataDiff;
    HomePhoneNumber: DataDiff;
    BusinessPhoneNumber: DataDiff;
}

export interface AddressAtrributes {
    Address1: DataDiff;
    Address2: DataDiff;
    Address3: DataDiff;
    Address4: DataDiff;
    City: DataDiff;
    County: DataDiff;
    State: DataDiff;
    Province: DataDiff;
    Country: DataDiff;
    PostalCode: DataDiff;
}

export const profileAttributesList = ['BusinessName', 'FirstName', 'LastName', 'BirthDate', 'Gender'];
export const contactInfoAttributesList = [
    'EmailAddress',
    'PersonalEmailAddress',
    'BusinessEmailAddress',
    'PhoneNumber',
    'MobilePhoneNumber',
    'HomePhoneNumber',
    'BusinessPhoneNumber',
];
export const AddressAttributesList = [
    'Address.Address1',
    'Address.Address2',
    'Address.Address3',
    'Address.Address4',
    'Address.City',
    'Address.County',
    'Address.State',
    'Address.Province',
    'Address.Country',
    'Address.PostalCode',
];
export const ShippingAddressAttributesList = [
    'ShippingAddress.Address1',
    'ShippingAddress.Address2',
    'ShippingAddress.Address3',
    'ShippingAddress.Address4',
    'ShippingAddress.City',
    'ShippingAddress.County',
    'ShippingAddress.State',
    'ShippingAddress.Province',
    'ShippingAddress.Country',
    'ShippingAddress.PostalCode',
];
export const MailingAddressAttributesList = [
    'MailingAddress.Address1',
    'MailingAddress.Address2',
    'MailingAddress.Address3',
    'MailingAddress.Address4',
    'MailingAddress.City',
    'MailingAddress.County',
    'MailingAddress.State',
    'MailingAddress.Province',
    'MailingAddress.Country',
    'MailingAddress.PostalCode',
];
export const BillingAddressAttributesList = [
    'BillingAddress.Address1',
    'BillingAddress.Address2',
    'BillingAddress.Address3',
    'BillingAddress.Address4',
    'BillingAddress.City',
    'BillingAddress.County',
    'BillingAddress.State',
    'BillingAddress.Province',
    'BillingAddress.Country',
    'BillingAddress.PostalCode',
];
