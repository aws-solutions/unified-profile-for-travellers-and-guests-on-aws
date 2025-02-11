// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { DataDiff, MatchPair } from '../../models/match.ts';
import { buildMergeRequest, convertMatchPairsToTableItems } from '../../utils/matchUtils.tsx';

const sourceID = 'qb1a3db7827e495ehw7365e7e29da0c9';
const targetID = 'pff36408ddg243ff93d60f6b46807445';
const matchPair: MatchPair = {
    domain_sourceProfileId: 'domain_' + sourceID,
    match_targetProfileId: 'match_' + targetID,
    targetProfileID: targetID,
    score: '0.9904272410936669',
    runId: 'latest',
    scoreTargetId: 'match_0.9904272410936669_pff36408ddg243ff93d60f6b46807445_qb1a3db7827e495ehw7365e7e29da0c9',
    mergeInProgress: false,
} as MatchPair;

it('buildMergeRequest | creates correct merge request', async () => {
    // act
    const matchDiff = convertMatchPairsToTableItems([matchPair]);
    const results = buildMergeRequest(matchDiff, false, 'abcd');

    // assert
    expect(results.length).toBe(1);

    expect(results[0].falsePositive).toBe(false);
    expect(results[0].target).toBe(targetID);
    expect(results[0].source).toBe(sourceID);
    expect(results[0].mergeContext.mergeType).toBe('ai');
    expect(results[0].mergeContext.operatorId).toBe('abcd');
});
