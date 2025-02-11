// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { loadProfiles } from '../../utils/profileUtils.tsx';
import { ConfigResponse } from '../../models/config.ts';
import { ProfileTableItem } from '../../store/profileSlice.ts';

it('extracts profiles info from query response', async () => {
    // arrange
    const profileOne = {
        connectId: 'connectIdOne',
        travellerId: 'travellerIdOne',
        personalEmailAddress: 'personalEmailAddressOne',
        firstName: 'firstNameOne',
        lastName: 'lastNameOne',
        extraAttr: '',
    } as any;

    const profileTwo = {
        connectId: 'connectIdTwo',
        travellerId: 'travellerIdTwo',
        personalEmailAddress: 'personalEmailAddressTwo',
        firstName: 'firstNameTwo',
        lastName: 'lastNameTwo',
        extraAttr: '',
    } as any;

    const profileResponse: ConfigResponse = {
        profiles: [profileOne, profileTwo],
    } as any;

    // act
    const results: ProfileTableItem[] = loadProfiles(profileResponse);

    // assert
    expect(results.length).toBe(2);

    expect(Object.keys(results[0]).length).toBe(14);
    expect(results[0].connectId).toBe(profileOne.connectId);
    expect(results[0].travellerId).toBe(profileOne.travellerId);

    expect(Object.keys(results[0]).length).toBe(14);
    expect(results[1].connectId).toBe(profileTwo.connectId);
    expect(results[1].travellerId).toBe(profileTwo.travellerId);
});
