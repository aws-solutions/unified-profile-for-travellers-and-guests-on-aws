// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import { v4 } from 'uuid';
import { createBreadcrumbs } from '../../../components/navigation/create-breadcrumbs.ts';

describe('Tests for Breadcrumb component', () => {
    it('generates the Home breadcrumb for the empty path', () => {
        // WHEN
        const result = createBreadcrumbs('/');

        // THEN
        expect(result).toHaveLength(1);
        expect(result[0]).toEqual({ text: 'Home', href: '' });
    });

    it('generates breadcrumbs for multiple path elements', () => {
        // WHEN
        const result = createBreadcrumbs('/projects/create/foo');

        // THEN
        expect(result).toHaveLength(4);
        expect(result[0]).toEqual({ text: 'Home', href: '' });
        expect(result[1]).toEqual({ text: 'Projects', href: '/projects' });
        expect(result[2]).toEqual({ text: 'Create', href: '/projects/create' });
        expect(result[3]).toEqual({ text: 'Foo', href: '/projects/create/foo' });
    });

    it('uses "Details" as label for uuids', () => {
        // GIVEN
        const projectId = v4();

        // WHEN
        const result = createBreadcrumbs(`/projects/${projectId}`);

        // THEN
        expect(result).toHaveLength(3);
        expect(result[0]).toEqual({ text: 'Home', href: '' });
        expect(result[1]).toEqual({ text: 'Projects', href: '/projects' });
        expect(result[2]).toEqual({ text: 'Details', href: `/projects/${projectId}` });
    });
});
