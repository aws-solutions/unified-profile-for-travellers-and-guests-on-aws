// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

export interface GetErrorsRequest {
    page: number;
    pageSize: number;
}

export interface DeleteErrorRequest {
    errorId: string;
}
