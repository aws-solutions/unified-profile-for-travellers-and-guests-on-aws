// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import TravelerInfoInput from '../../components/profile/search/travelerInfoInput';
import ProfilesTable from '../../components/profile/search/profilesTable';

export default function ProfileSearch() {
    return (
        <>
            <TravelerInfoInput />
            <br />
            <hr />
            <br />
            <ProfilesTable />
        </>
    );
}
