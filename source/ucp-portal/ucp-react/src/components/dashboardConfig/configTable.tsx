// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Header, TableProps } from '@cloudscape-design/components';
import BaseCRUDTable from '../base/baseCrudTable';
import { HyperLinkMapping } from '../../models/config';
import { DeleteConfigButton } from './deleteConfigButton';
import { selectHyperLinkMappings } from '../../store/configSlice';
import { useSelector } from 'react-redux';

interface ConfigTableProps {
    isFetching: boolean;
}

export const ConfigTable = (props: ConfigTableProps) => {
    const tableItems = useSelector(selectHyperLinkMappings);

    const columnDefs: TableProps.ColumnDefinition<HyperLinkMapping>[] = [
        {
            id: 'dataRecord',
            header: 'Data Record',
            cell: config => config.accpObject,
            ariaLabel: () => 'dataRecord',
        },
        {
            id: 'dataField',
            header: 'Data Field',
            cell: config => config.fieldName,
            ariaLabel: () => 'dataField',
        },
        {
            id: 'urlTemplate',
            header: 'URL Template',
            cell: config => config.hyperlinkTemplate,
            ariaLabel: () => 'urlTemplate',
        },
        {
            id: 'action',
            header: 'Action',
            cell: config => <DeleteConfigButton configData={tableItems} configToDelete={config} />,
            ariaLabel: () => 'action',
        },
    ];

    const header = <Header variant="h1">Dashboard Configuration</Header>;

    return (
        <BaseCRUDTable<HyperLinkMapping>
            columnDefinitions={columnDefs}
            items={tableItems}
            header={header}
            preferences={{}}
            isLoading={props.isFetching}
            tableVariant="container"
            isFilterable={false}
        />
    );
};
