// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Badge } from '@cloudscape-design/components';
import {
    AddressAtrributes,
    AddressAttributesList,
    BillingAddressAttributesList,
    ContactInfoAttributes,
    DataDiff,
    MailingAddressAttributesList,
    MatchDiff,
    MatchPair,
    ProfileAttributes,
    ShippingAddressAttributesList,
    contactInfoAttributesList,
    profileAttributesList,
} from '../models/match';
import { MergeContext, MergeRq, MergeType } from '../models/traveller';

/**
 * converts match pairs to match tableItems
 * @param matchPairs - list of matches
 * @returns tableItems
 */
export function convertMatchPairsToTableItems(matchPairs: MatchPair[]): MatchDiff[] {
    const diffPairs: MatchDiff[] = [];
    for (let i = 0; i < matchPairs.length; i++) {
        if (!matchPairs[i].mergeInProgress) {
            const diffPair = {} as MatchDiff;
            diffPair.ProfileAttributes = {} as ProfileAttributes;
            diffPair.ContactInfoAttributes = {} as ContactInfoAttributes;
            diffPair.AddressAtrributes = {} as AddressAtrributes;
            diffPair.BillingAddress = {} as AddressAtrributes;
            diffPair.MailingAddress = {} as AddressAtrributes;
            diffPair.ShippingAddress = {} as AddressAtrributes;

            const currentMatchPair = matchPairs[i];
            diffPair.Id = i;
            diffPair.Score = parseFloat(currentMatchPair.score).toFixed(4);
            diffPair.Domain = currentMatchPair.domain;
            const sourceId = currentMatchPair.domain_sourceProfileId.split('_').pop();
            if (sourceId === undefined) {
                continue;
            }
            diffPair.SourceProfileId = sourceId;
            diffPair.TargetProfileId = currentMatchPair.targetProfileID;
            if (currentMatchPair.profileDataDiff !== undefined) {
                for (const diff of currentMatchPair.profileDataDiff) {
                    if (profileAttributesList.includes(diff.fieldName)) {
                        diffPair.ProfileAttributes[diff.fieldName as keyof ProfileAttributes] = diff;
                    } else if (contactInfoAttributesList.includes(diff.fieldName)) {
                        diffPair.ContactInfoAttributes[diff.fieldName as keyof ContactInfoAttributes] = diff;
                    } else if (AddressAttributesList.includes(diff.fieldName) && (diff.sourceValue !== '' || diff.targetValue !== '')) {
                        diffPair.AddressAtrributes[diff.fieldName.split('.')[1] as keyof AddressAtrributes] = diff;
                    } else if (
                        BillingAddressAttributesList.includes(diff.fieldName) &&
                        (diff.sourceValue !== '' || diff.targetValue !== '')
                    ) {
                        diffPair.BillingAddress[diff.fieldName.split('.')[1] as keyof AddressAtrributes] = diff;
                    } else if (
                        MailingAddressAttributesList.includes(diff.fieldName) &&
                        (diff.sourceValue !== '' || diff.targetValue !== '')
                    ) {
                        diffPair.MailingAddress[diff.fieldName.split('.')[1] as keyof AddressAtrributes] = diff;
                    } else if (
                        ShippingAddressAttributesList.includes(diff.fieldName) &&
                        (diff.sourceValue !== '' || diff.targetValue !== '')
                    ) {
                        diffPair.ShippingAddress[diff.fieldName.split('.')[1] as keyof AddressAtrributes] = diff;
                    }
                }
            }
            diffPairs.push(diffPair);
        }
    }
    return diffPairs;
}

/**
 * Compares the source and target profile values for similarities/differences
 * @param dataDiffObj contains source and target profile values
 * @param breaker can be line break (<br>) or <>{" vs "}</>
 * @returns <Badge> object
 */
export function getBadge(dataDiffObj: DataDiff, breaker: JSX.Element): JSX.Element {
    if (!dataDiffObj || (dataDiffObj.sourceValue === '' && dataDiffObj.targetValue === ''))
        return <Badge color="grey">n/a</Badge>; // If both profiles have missing values, return a grey badge
    else if (dataDiffObj.sourceValue === '')
        // If source profile value is missing
        return (
            <>
                <Badge color="grey">n/a</Badge>
                {breaker}
                <Badge color="blue">{dataDiffObj.targetValue}</Badge>
            </>
        );
    else if (dataDiffObj.targetValue === '')
        // If target profile value is missing
        return (
            <>
                <Badge color="blue">{dataDiffObj.sourceValue}</Badge>
                {breaker}
                <Badge color="grey">n/a</Badge>
            </>
        );
    else if (dataDiffObj.sourceValue === dataDiffObj.targetValue)
        // If both profiles have same values, return a green badge
        return <Badge color="green">{dataDiffObj.sourceValue}</Badge>;

    // If both profiles have different values, return two red badge
    return (
        <>
            <Badge color="red">{dataDiffObj.sourceValue}</Badge>
            {breaker}
            <Badge color="red">{dataDiffObj.targetValue}</Badge>
        </>
    );
}

/**
 * Builds merge request body
 * @param selectedMatchTableItems list of match items
 * @param falsePositive to mark as false positives or not
 * @returns body for merge api
 */
export function buildMergeRequest(selectedMatchTableItems: MatchDiff[], falsePositive: boolean, operatorCognitoId: string): MergeRq[] {
    const merges: MergeRq[] = [];
    for (let i = 0; i < selectedMatchTableItems.length; i++) {
        const sourceProfileId = selectedMatchTableItems[i].SourceProfileId;
        const targetProfileId = selectedMatchTableItems[i].TargetProfileId;
        merges.push({
            falsePositive: falsePositive,
            source: sourceProfileId,
            target: targetProfileId,
            mergeContext: {
                confidenceUpdateFactor: parseFloat(selectedMatchTableItems[i].Score),
                mergeType: MergeType.AI,
                operatorId: operatorCognitoId,
            } as MergeContext,
        });
    }
    return merges;
}

export enum CheckBoxNumber {
    None,
    First,
    Second,
}
