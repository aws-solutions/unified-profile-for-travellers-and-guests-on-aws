// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export interface PaginationOptions {
    page: number;
    pageSize: number;
    objectType?: string;
}

export interface MultiPaginationProp {
    objects: string[];
    pages: number[];
    pageSizes: number[];
}

export interface Metadata {
    total_records: number;
}

export interface PaginationMetadata {
    airBookingRecords: Metadata;
    airLoyaltyRecords: Metadata;
    clickstreamRecords: Metadata;
    emailHistoryRecords: Metadata;
    hotelBookingRecords: Metadata;
    hotelLoyaltyRecords: Metadata;
    hotelStayRecords: Metadata;
    phoneHistoryRecords: Metadata;
    customerServiceInteractionRecords: Metadata;
    lotaltyTxRecords: Metadata;
    ancillaryServiceRecords: Metadata;
}
