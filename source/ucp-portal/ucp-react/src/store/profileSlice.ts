// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Common, DEFAULT_ACCP_PAGE_SIZE } from '../models/common';
import { ConfigResponse, HyperLinkMapping } from '../models/config';
import { MultiPaginationProp } from '../models/pagination';
import {
    AirBooking,
    AirLoyalty,
    AncillaryService,
    Clickstream,
    CustomerServiceInteraction,
    EmailHistory,
    HotelBooking,
    HotelLoyalty,
    HotelStay,
    LoyaltyTx,
    MergeContext,
    PhoneHistory,
} from '../models/traveller';
import { configApiSlice } from './configApiSlice';
import { profileApiSlice } from './profileApiSlice';
import { RootState } from './store';

export interface ProfileRequest {
    connectId: string;
    travellerId: string;
    lastName: string;
    email: string;
    phone: string;
    addressAddress1: string;
    addressAddress2: string;
    addressCity: string;
    addressState: string;
    addressProvince: string;
    addressPostalCode: string;
    addressCountry: string;
    addressType: string;
}

export interface ProfileTableItem extends ProfileRequest {
    firstName: string;
}

export enum DialogType {
    NONE = 'none',
    RECORD = 'record',
    CHAT = 'chat',
    HISTORY = 'Interaction history',
}

export interface PortalMappings {
    air_booking: HyperLinkMapping[];
    email_history: HyperLinkMapping[];
    phone_history: HyperLinkMapping[];
    air_loyalty: HyperLinkMapping[];
    loyalty_transaction: HyperLinkMapping[];
    clickstream: HyperLinkMapping[];
    guest_profile: HyperLinkMapping[];
    hotel_loyalty: HyperLinkMapping[];
    hotel_booking: HyperLinkMapping[];
    pax_profile: HyperLinkMapping[];
    hotel_stay_revenue_items: HyperLinkMapping[];
    customer_service_interaction: HyperLinkMapping[];
    ancillary_service: HyperLinkMapping[];
}

export type RecordType =
    | AirBooking
    | AirLoyalty
    | AncillaryService
    | CustomerServiceInteraction
    | Clickstream
    | EmailHistory
    | HotelBooking
    | HotelLoyalty
    | HotelStay
    | LoyaltyTx
    | PhoneHistory;

export interface ProfileState extends ProfileRequest {
    searchResults: ConfigResponse | null;
    viewProfile: ConfigResponse;
    dialogType: DialogType;
    dialogRecord: RecordType | null;
    portalConfig: PortalMappings;
    multiPaginationParam: MultiPaginationProp;
    interactionHistory: MergeContext[];
    profileSummary: string;
}

const initialState: ProfileState = {
    connectId: '',
    travellerId: '',
    lastName: '',
    email: '',
    phone: '',
    addressAddress1: '',
    addressAddress2: '',
    addressCity: '',
    addressState: '',
    addressProvince: '',
    addressPostalCode: '',
    addressCountry: '',
    addressType: '',
    searchResults: null,
    viewProfile: {} as ConfigResponse,
    dialogType: DialogType.NONE,
    dialogRecord: null,
    portalConfig: {
        air_booking: [],
        email_history: [],
        phone_history: [],
        air_loyalty: [],
        loyalty_transaction: [],
        clickstream: [],
        guest_profile: [],
        hotel_loyalty: [],
        hotel_booking: [],
        pax_profile: [],
        hotel_stay_revenue_items: [],
        customer_service_interaction: [],
        ancillary_service: [],
    },
    multiPaginationParam: {
        objects: Object.values(Common),
        pages: new Array(11).fill(0),
        pageSizes: new Array(11).fill(DEFAULT_ACCP_PAGE_SIZE),
    },
    interactionHistory: [],
    profileSummary: '',
};

export const profileSlice = createSlice({
    name: 'profiles',
    initialState,
    reducers: {
        setConnectId: (state: ProfileState, action: PayloadAction<string>) => {
            state.connectId = action.payload;
        },
        setTravellerId: (state: ProfileState, action: PayloadAction<string>) => {
            state.travellerId = action.payload;
        },
        setLastName: (state: ProfileState, action: PayloadAction<string>) => {
            state.lastName = action.payload;
        },
        setEmailId: (state: ProfileState, action: PayloadAction<string>) => {
            state.email = action.payload;
        },
        setPhone: (state: ProfileState, action: PayloadAction<string>) => {
            state.phone = action.payload;
        },
        setAddress1: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressAddress1 = action.payload;
        },
        setAddress2: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressAddress2 = action.payload;
        },
        setCity: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressCity = action.payload;
        },
        setState: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressState = action.payload;
        },
        setProvince: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressProvince = action.payload;
        },
        setPostalCode: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressPostalCode = action.payload;
        },
        setCountry: (state: ProfileState, action: PayloadAction<string>) => {
            state.addressCountry = action.payload;
        },
        setSearchResults: (state: ProfileState, action: PayloadAction<ConfigResponse>) => {
            state.searchResults = action.payload;
        },
        setViewProfile: (state: ProfileState, action: PayloadAction<ConfigResponse>) => {
            state.viewProfile = action.payload;
        },
        setDialogType: (state: ProfileState, action: PayloadAction<DialogType>) => {
            state.dialogType = action.payload;
        },
        setDialogRecord: (state: ProfileState, action: PayloadAction<RecordType>) => {
            state.dialogRecord = action.payload;
        },
        setPaginationParam: (state: ProfileState, action: PayloadAction<any>) => {
            state.multiPaginationParam = action.payload;
        },
        resetState: () => {
            return initialState;
        },
    },
    extraReducers: builder => {
        builder.addMatcher(profileApiSlice.endpoints.getProfilesByParam.matchFulfilled, (state, { payload }) => {
            state.searchResults = payload;
        }),
            builder.addMatcher(profileApiSlice.endpoints.getProfilesById.matchFulfilled, (state, { payload }) => {
                state.viewProfile = payload;
            }),
            builder.addMatcher(configApiSlice.endpoints.getPortalConfig.matchFulfilled, (state, { payload }) => {
                if (payload.portalConfig.hyperlinkMappings === undefined) return;
                const mappings = payload.portalConfig.hyperlinkMappings;
                for (const mapping of mappings) {
                    state.portalConfig[mapping.accpObject as keyof PortalMappings].push(mapping);
                }
            }),
            builder.addMatcher(profileApiSlice.endpoints.getInteractionHistory.matchFulfilled, (state, { payload }) => {
                state.interactionHistory = payload.interactionHistory;
            }),
            builder.addMatcher(profileApiSlice.endpoints.getProfilesSummary.matchFulfilled, (state, { payload }) => {
                state.profileSummary = payload;
            });
    },
});

export const {
    setConnectId,
    setTravellerId,
    setLastName,
    setEmailId,
    setPhone,
    setAddress1,
    setAddress2,
    setCity,
    setState,
    setProvince,
    setPostalCode,
    setCountry,
    setSearchResults,
    setViewProfile,
    setDialogType,
    setDialogRecord,
    setPaginationParam,
    resetState,
} = profileSlice.actions;

export const selectConnectId = (state: RootState) => state.profiles.connectId;
export const selectTravellerId = (state: RootState) => state.profiles.travellerId;
export const selectLastName = (state: RootState) => state.profiles.lastName;
export const selectEmailId = (state: RootState) => state.profiles.email;
export const selectPhone = (state: RootState) => state.profiles.phone;
export const selectAddress1 = (state: RootState) => state.profiles.addressAddress1;
export const selectAddress2 = (state: RootState) => state.profiles.addressAddress2;
export const selectCity = (state: RootState) => state.profiles.addressCity;
export const selectState = (state: RootState) => state.profiles.addressState;
export const selectProvince = (state: RootState) => state.profiles.addressProvince;
export const selectPostalCode = (state: RootState) => state.profiles.addressPostalCode;
export const selectCountry = (state: RootState) => state.profiles.addressCountry;
export const selectSearchResults = (state: RootState) => state.profiles.searchResults;
export const selectViewProfile = (state: RootState) => state.profiles.viewProfile;
export const selectDialogType = (state: RootState) => state.profiles.dialogType;
export const selectDialogRecord = (state: RootState) => state.profiles.dialogRecord;
export const selectPortalConfig = (state: RootState) => state.profiles.portalConfig;
export const selectPaginationParam = (state: RootState) => state.profiles.multiPaginationParam;
export const selectInteractionHistory = (state: RootState) => state.profiles.interactionHistory;
export const selectProfileSummary = (state: RootState) => state.profiles.profileSummary;
