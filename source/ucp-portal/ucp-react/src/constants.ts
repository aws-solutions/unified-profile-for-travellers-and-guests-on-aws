// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import * as permissions from '../../../appPermissions.json';

export enum BaseCrudTableSelectionType {
    SINGLE = 'single',
    MULTI = 'multi',
}

export enum PageSize {
    PAGE_SIZE_10 = 10,
    PAGE_SIZE_20 = 20,
    PAGE_SIZE_30 = 30,
}

export enum Permissions {
    AdminPermission = 4294967295, // max uint32 value, indicating all permissions
    SearchProfilePermission = 1 << permissions.SearchProfilePermission,
    DeleteProfilePermission = 1 << permissions.DeleteProfilePermission,
    MergeProfilePermission = 1 << permissions.MergeProfilePermission,
    UnmergeProfilePermission = 1 << permissions.UnmergeProfilePermission,
    CreateDomainPermission = 1 << permissions.CreateDomainPermission,
    DeleteDomainPermission = 1 << permissions.DeleteDomainPermission,
    ConfigureGenAiPermission = 1 << permissions.ConfigureGenAiPermission,
    SaveHyperlinkPermission = 1 << permissions.SaveHyperlinkPermission,
    SaveRuleSetPermission = 1 << permissions.SaveRuleSetPermission,
    ActivateRuleSetPermission = 1 << permissions.ActivateRuleSetPermission,
    RunGlueJobsPermission = 1 << permissions.RunGlueJobsPermission,
    ClearAllErrorsPermission = 1 << permissions.ClearAllErrorsPermission,
    RebuildCachePermission = 1 << permissions.RebuildCachePermission,
    IndustryConnectorPermission = 1 << permissions.IndustryConnectorPermission,
    ListPrivacySearchPermission = 1 << permissions.ListPrivacySearchPermission,
    GetPrivacySearchPermission = 1 << permissions.GetPrivacySearchPermission,
    CreatePrivacySearchPermission = 1 << permissions.CreatePrivacySearchPermission,
    DeletePrivacySearchPermission = 1 << permissions.DeletePrivacySearchPermission,
    PrivacyDataPurgePermission = 1 << permissions.PrivacyDataPurgePermission,
    ListRuleSetPermission = 1<< permissions.ListRuleSetPermission
}
