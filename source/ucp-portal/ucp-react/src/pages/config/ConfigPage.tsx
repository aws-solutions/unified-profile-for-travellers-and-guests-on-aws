// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { ContentLayout, Header } from '@cloudscape-design/components';
import { ConfigTable } from '../../components/dashboardConfig/configTable';
import { useGetPortalConfigQuery } from '../../store/configApiSlice';
import { ConfigFormFields } from '../../components/dashboardConfig/configFormFields';

export const ConfigPage = () => {
    const { isFetching: isConfigFetching } = useGetPortalConfigQuery();

    return (
        <>
            <ContentLayout header={<Header variant="h1">Dashboard Configuration</Header>}>
                <ConfigFormFields />
                <ConfigTable isFetching={isConfigFetching} />
            </ContentLayout>
        </>
    );
};
