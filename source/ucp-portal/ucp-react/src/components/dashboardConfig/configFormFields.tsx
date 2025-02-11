// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ColumnLayout, Container, Header, SpaceBetween } from '@cloudscape-design/components';
import { AddConfigButton } from './addConfigButton';
import { ClearFieldsButton } from './clearFieldsButton';
import { DataRecordDropdown } from './dataRecordDropdown';
import { DataFieldDropdown } from './dataFieldDropdown';
import { UrlTemplateInput } from './urlTemplateInput';

export const ConfigFormFields = () => {
    const headerActions = (
        <SpaceBetween direction="horizontal" size={'xs'}>
            <ClearFieldsButton />
            <AddConfigButton />
        </SpaceBetween>
    );

    const configFormFieldsHeader = (
        <Header variant="h2" actions={headerActions}>
            Add New Config
        </Header>
    );

    return (
        <Container header={configFormFieldsHeader}>
            <ColumnLayout columns={3}>
                <DataRecordDropdown />
                <DataFieldDropdown />
                <UrlTemplateInput />
            </ColumnLayout>
        </Container>
    );
};
