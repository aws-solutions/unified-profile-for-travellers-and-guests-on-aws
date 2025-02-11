// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import MatchTable from '../../components/aiMatches/matchTable';
import { ContentLayout } from '@cloudscape-design/components';
import MatchDifferenceDialog from '../../components/aiMatches/detailedDiffDialog';

export default function MatchesPage() {
    return (
        <ContentLayout>
            <MatchTable></MatchTable>
            <MatchDifferenceDialog />
        </ContentLayout>
    );
}
