// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//local storage
export const setDomainValueStorage = (domainName: string) => {
    localStorage.setItem('domain', domainName);
};

export const getDomainValueStorage = () => {
    return localStorage.getItem('domain');
};

export const removeDomainValueStorage = () => {
    return localStorage.removeItem('domain');
};
