// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// build an array of breadcrumb items, one for each element of the given path
import { BreadcrumbGroupProps } from '@cloudscape-design/components';

export const createBreadcrumbs = (path: string): BreadcrumbGroupProps.Item[] => {
    if (path === '/') {
        return [{ text: 'Home', href: '' }];
    }

    const pathElements: string[] = path.split('/');

    return pathElements.map((currentElement, index) => {
        const capitalizedElement = capitalizeFirstLetter(currentElement);
        const previousPathElementsPlusCurrent = pathElements.slice(0, index + 1);
        const href = `${previousPathElementsPlusCurrent.join('/')}`;
        return { text: getLabelForPathElement(capitalizedElement), href };
    });
};

function capitalizeFirstLetter(element: string): string {
    if (element.length === 0) return element;
    return element.charAt(0).toUpperCase() + element.slice(1);
}

const pathLabels: Record<string, string> = {
    '': 'Home',
    Settings: 'Domain Settings',
    Matches: 'AI Matches',
};

function getLabelForPathElement(pathElement: string): string {
    const pathLabel = pathLabels[pathElement];
    if (pathLabel) return pathLabel;

    // 'Details' is supposed to be used for the uuids that are part of the route
    if (/^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(pathElement)) return 'Details';

    return pathElement;
}
