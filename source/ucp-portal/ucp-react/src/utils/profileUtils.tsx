// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button, Link } from '@cloudscape-design/components';
import { CancelableEventHandler, ClickDetail } from '@cloudscape-design/components/internal/events';
import { Dispatch, SetStateAction } from 'react';
import { AnyAction } from 'redux';
import { CustomPropsPaginationType } from '../components/base/baseCrudTable';
import { DEFAULT_ACCP_PAGE_SIZE } from '../models/common';
import { ConfigResponse, HyperLinkMapping, ProfileResponse } from '../models/config';
import { MultiPaginationProp } from '../models/pagination';
import { MergeContext, MergeType, mergeTypePrefix } from '../models/traveller';
import { DialogType, ProfileTableItem, RecordType, setDialogRecord, setDialogType, setPaginationParam } from '../store/profileSlice';

export function loadProfiles(searchResults: ConfigResponse | null) {
    const profiles: ProfileTableItem[] = [];
    if (searchResults !== null) {
        const fetchedProfiles: ProfileResponse[] = searchResults.profiles;
        for (const profile of fetchedProfiles) {
            const currProfile: ProfileTableItem = {
                connectId: profile.connectId,
                travellerId: profile.travellerId,
                email: profile.personalEmailAddress,
                firstName: profile.firstName,
                lastName: profile.lastName,
                phone: profile.phoneNumber,
                addressAddress1: '',
                addressAddress2: '',
                addressCity: '',
                addressState: '',
                addressProvince: '',
                addressPostalCode: '',
                addressCountry: '',
                addressType: '',
            };
            profiles.push(currProfile);
        }
    }
    return profiles;
}

export function cellAsLinkOrValue(hyperlinkMappings: Map<string, string>, attribute: string, attributeValue: string) {
    if (hyperlinkMappings.has(attribute))
        return (
            <Link external href={hyperlinkMappings.get(attribute)}>
                {attributeValue}
            </Link>
        );
    return attributeValue;
}

export function portalConfigToMap(portalConfig: HyperLinkMapping[]) {
    const hyperlinkMappings: Map<string, string> = new Map();
    for (const config of portalConfig) {
        hyperlinkMappings.set(config.fieldName, config.hyperlinkTemplate);
    }
    return hyperlinkMappings;
}

export function buildPaginatonProps(
    setCurrentPageIndex: Dispatch<SetStateAction<number>>,
    paginationParam: MultiPaginationProp,
    paginationIndex: number,
    currentPageIndex: number,
    pagesCount: number,
    dispatch: Dispatch<AnyAction>,
): CustomPropsPaginationType {
    return {
        onPageChange: currentPage => {
            setCurrentPageIndex(currentPage);
            const updatedCurrentPages = [...paginationParam.pages];
            updatedCurrentPages[paginationIndex] = currentPage - 1;
            dispatch(setPaginationParam({ ...paginationParam, pages: updatedCurrentPages }));
        },
        currentPageIndex,
        pagesCount,
    };
}

export function updatePaginationOnPreferenceChange(
    paginationParam: MultiPaginationProp,
    paginationIndex: number,
    pageSize: number | undefined,
    dispatch: Dispatch<AnyAction>,
) {
    const updatedPageSizes = [...paginationParam.pageSizes];
    updatedPageSizes[paginationIndex] = pageSize ?? DEFAULT_ACCP_PAGE_SIZE;
    // If page size is changed, reset the current page to 0
    const updatedCurrentPages = [...paginationParam.pages];
    updatedCurrentPages[paginationIndex] = 0;
    dispatch(setPaginationParam({ ...paginationParam, pageSizes: updatedPageSizes, pages: updatedCurrentPages }));
}

export const onClickViewDialog = (
    record: RecordType,
    dialogType: DialogType,
    dispatch: Dispatch<AnyAction>,
    connectId?: string,
    interactionType?: string,
) => {
    const clonedRecord = JSON.parse(JSON.stringify(record));
    for (const [key, value] of Object.entries(clonedRecord)) {
        if (value === null || value === undefined) {
            delete clonedRecord[key];
        }
        clonedRecord[key] = String(value);
        if (clonedRecord[key] === '') {
            delete clonedRecord[key];
        }
    }
    if (interactionType) {
        clonedRecord.interactionType = interactionType;
    }
    if (connectId) {
        clonedRecord.connectId = connectId;
    }
    dispatch(setDialogType(dialogType));
    dispatch(setDialogRecord(clonedRecord));
};

export function getEventTypeDisplayValue(
    item: MergeContext,
    record: RecordType,
    canUserUnmerge: boolean,
    onClick: CancelableEventHandler<ClickDetail>,
) {
    const mergeTypes: string[] = [MergeType.RULE, MergeType.AI, MergeType.MANUAL];
    if (mergeTypes.includes(item.mergeType.toLowerCase())) {
        if (item.toMergeConnectId != record.connectId && canUserUnmerge) {
            return (
                <div>
                    <span>{mergeTypePrefix + item.mergeType.toUpperCase()}</span>
                    <Button variant="inline-link" onClick={onClick} ariaLabel="mergeUndo">
                        {' '}
                        (Undo){' '}
                    </Button>
                </div>
            );
        }
        return mergeTypePrefix + item.mergeType.toUpperCase();
    }

    return item.mergeType.toUpperCase();
}
