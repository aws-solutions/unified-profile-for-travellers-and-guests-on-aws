// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ContentLayout, Header } from '@cloudscape-design/components';
import { ErrorsTable } from '../../components/errors/errorsTable.tsx';

export const ErrorsPage = () => {
    return (
        <ContentLayout data-testid="errorPage" header={<Header variant="h1"></Header>}>
            <ErrorsTable />
        </ContentLayout>
    );
};
