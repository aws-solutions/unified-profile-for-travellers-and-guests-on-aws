// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { useCollection, UseCollectionOptions } from '@cloudscape-design/collection-hooks';
import { Button, HeaderProps, NonCancelableCustomEvent } from '@cloudscape-design/components';
import { CollectionPreferencesProps } from '@cloudscape-design/components/collection-preferences';
import Pagination, { PaginationProps } from '@cloudscape-design/components/pagination';
import Table, { TableProps } from '@cloudscape-design/components/table';
import TextFilter from '@cloudscape-design/components/text-filter';
import EmptyState from './emptyState';

export interface CustomPropsSortingType<T> {
    defaultSortState?: TableProps.SortingState<T>;
    sortingDescending?: boolean;
    sortingColumn?: TableProps.SortingColumn<T>;
    onSortingChange?: (isDescending: boolean | undefined, sortingColumn: TableProps.SortingColumn<T> | undefined) => void;
}

export interface CustomPropsPaginationType {
    onPageChange?: (currentPage: number) => void;
    currentPageIndex?: number;
    pagesCount?: number;
    openEnded?: boolean;
}

interface TableDataProps<T> {
    items: T[];
    selectedItems?: T[];
    ariaLabels?: TableProps.AriaLabels<T>;
    columnDefinitions: TableProps.ColumnDefinition<T>[];
    preferences: CollectionPreferencesProps.Preferences;
    collectionPreferences?: React.ReactNode;
    useCollectionConfiguration?: UseCollectionOptions<T>;
}

interface TableCustomizationProps {
    tableVariant?: TableProps.Variant;
    isFilterable?: boolean;
    isLoading?: boolean;
    loadingText?: string;
    wrapLines?: boolean;
    stripedRows?: boolean;
    stickyHeader?: boolean;
    selectionType?: TableProps.SelectionType;
    stickyColumns?: TableProps.StickyColumns;
}

interface TableCallbackProps<T> {
    onSelectionChange?: (selectedItems: T[]) => void;
    submitEdit?: TableProps.SubmitEditFunction<T>;
    onEditCancel?: () => void;
    isItemDisabled?: (item: T) => boolean;
}

interface BaseCrudTableProps<T> extends HeaderProps, TableDataProps<T>, TableCustomizationProps, TableCallbackProps<T> {
    header?: React.ReactNode;
    customPropsSorting?: CustomPropsSortingType<T>;
    customPropsPagination?: CustomPropsPaginationType;
    'data-testid'?: string;
}

/**
 * TableView Template
 * @param props
 * @param props.columnDefinitions
 * @param props.items
 * @param props.header
 * @param props.selectedItems
 * @param props.onSelectionChange
 * @param props.preferences
 * @param props.collectionPreferences
 * @param props.submitEdit
 * @param props.ariaLabels
 * @param props.onEditCancel
 * @param props.tableVariant
 * @param props.isItemDisabled
 * @param props.isFilterable
 * @param props.stickyHeader
 * @param props.selectionType
 * @param props.stickyColumns
 * @param props.isLoading
 * @param props.customPropsSorting
 * @param props.customPropsPagination
 */
export default function BaseCRUDTable<T>({
    columnDefinitions,
    items,
    header,
    selectedItems,
    onSelectionChange = () => {},
    preferences,
    collectionPreferences,
    submitEdit,
    ariaLabels,
    onEditCancel,
    selectionType,
    stickyColumns = {},
    tableVariant = 'full-page',
    isFilterable = true,
    stickyHeader = true,
    isItemDisabled = () => false,
    isLoading = false,
    loadingText = 'Loading Resources',
    wrapLines = false,
    useCollectionConfiguration,
    customPropsSorting,
    customPropsPagination,
    stripedRows = false,
    'data-testid': dataTestId,
}: BaseCrudTableProps<T>) {
    const {
        items: updatedItems,
        actions,
        filteredItemsCount,
        collectionProps,
        filterProps,
        paginationProps,
    } = useCollection(items, {
        filtering: {
            empty: <EmptyState title="Nothing to display" subtitle="" action={<></>} />,
            noMatch: (
                <EmptyState
                    title="No matches"
                    subtitle="We can't find a match."
                    action={
                        <Button ariaLabel="clearFilter" onClick={() => actions.setFiltering('')}>
                            Clear filter
                        </Button>
                    }
                />
            ),
        },
        pagination: { pageSize: preferences.pageSize },
        sorting: { defaultState: customPropsSorting?.defaultSortState },
        selection: {},
        ...useCollectionConfiguration,
    });
    const handleSelectionChange = ({ detail }: NonCancelableCustomEvent<TableProps.SelectionChangeDetail<T>>): void => {
        onSelectionChange(detail.selectedItems);
    };

    const paginatedTableCheck =
        (customPropsPagination?.pagesCount && customPropsPagination.pagesCount > 1) || customPropsPagination?.openEnded;
    const nonPaginatedTableCheck = preferences.pageSize && preferences.pageSize < items.length;
    const showPagination = paginatedTableCheck || nonPaginatedTableCheck;

    if (customPropsSorting) {
        if (customPropsSorting.sortingColumn) collectionProps.sortingColumn = customPropsSorting.sortingColumn;
        if (customPropsSorting.sortingDescending) collectionProps.sortingDescending = customPropsSorting.sortingDescending;
        if (customPropsSorting.onSortingChange) {
            collectionProps.onSortingChange = ({ detail }: NonCancelableCustomEvent<TableProps.SortingState<T>>): void => {
                customPropsSorting.onSortingChange!(detail.isDescending, detail.sortingColumn);
            };
        }
    }

    if (customPropsPagination) {
        if (customPropsPagination.currentPageIndex) paginationProps.currentPageIndex = customPropsPagination.currentPageIndex;
        if (customPropsPagination.pagesCount) paginationProps.pagesCount = customPropsPagination.pagesCount;
        if (customPropsPagination.onPageChange) {
            paginationProps.onChange = ({ detail }: NonCancelableCustomEvent<PaginationProps.ChangeDetail>) => {
                customPropsPagination.onPageChange!(detail.currentPageIndex);
            };
        }
    }

    return (
        <Table<T>
            data-testid={dataTestId}
            ariaLabels={ariaLabels}
            selectedItems={selectedItems}
            empty={collectionProps.empty}
            sortingColumn={collectionProps.sortingColumn}
            sortingDescending={collectionProps.sortingDescending}
            onSortingChange={collectionProps.onSortingChange}
            onSelectionChange={handleSelectionChange}
            columnDefinitions={columnDefinitions}
            items={updatedItems}
            isItemDisabled={isItemDisabled}
            variant={tableVariant}
            loadingText={loadingText}
            loading={isLoading}
            wrapLines={wrapLines}
            selectionType={selectionType}
            filter={
                isFilterable &&
                !paginatedTableCheck && (
                    <TextFilter
                        {...filterProps}
                        countText={filteredItemsCount + ' matches.'}
                        filteringPlaceholder="Filter results"
                        filteringAriaLabel="text-filter"
                        filteringClearAriaLabel="filter-clear"
                    />
                )
            }
            pagination={
                showPagination && (
                    <Pagination
                        {...paginationProps}
                        openEnd={customPropsPagination?.openEnded}
                        ariaLabels={{
                            nextPageLabel: 'Next page',
                            previousPageLabel: 'Previous page',
                            pageLabel: pageNumber => `Page ${pageNumber} of all pages`,
                        }}
                    />
                )
            }
            submitEdit={submitEdit}
            preferences={collectionPreferences}
            stickyHeader={stickyHeader}
            header={header}
            onEditCancel={onEditCancel}
            stickyColumns={stickyColumns}
            stripedRows={stripedRows}
        />
    );
}
