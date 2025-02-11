// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ColumnLayout, Container, Header, StatusIndicator } from '@cloudscape-design/components';
import { Domain } from '../../models/config';

interface OverviewContainersProps {
    selectedDomain: Domain;
    isLoading: boolean;
}

export const OverviewContainers = (props: OverviewContainersProps) => {
    return (
        <ColumnLayout columns={2}>
            <Container data-testid="numObjectsContainer" header={<Header variant="h2">Number of Objects</Header>}>
                {props.isLoading ? <StatusIndicator type="loading" /> : props.selectedDomain.numberOfObjects}
            </Container>

            <Container data-testid="numProfilesContainer" header={<Header variant="h2">Number of Profiles</Header>}>
                {props.isLoading ? <StatusIndicator type="loading" /> : props.selectedDomain.numberOfProfiles}
            </Container>
        </ColumnLayout>
    );
};
