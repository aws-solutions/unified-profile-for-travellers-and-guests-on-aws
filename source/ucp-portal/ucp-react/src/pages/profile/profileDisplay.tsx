// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import DisplayHome from '../../components/profile/display/displayHome';
import { useParams } from 'react-router';

export default function ProfileSearch() {
    const { profileId } = useParams();
    return <DisplayHome profileId={profileId ?? ''} />;
}
