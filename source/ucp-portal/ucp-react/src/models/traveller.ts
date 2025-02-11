// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

interface BaseAttributes {
    accpObjectID: string;
    overallConfidenceScore: string;
    travellerId: string;
    lastUpdated: string; // time.Time
    lastUpdatedBy: string;
    interactionType: string;
    connectId?: string;
}

export interface AirBooking extends BaseAttributes {
    bookingId: string;
    segmentId: string;
    from: string;
    to: string;
    flightNumber: string;
    departureDate: string;
    departureTime: string;
    arrivalDate: string;
    arrivalTime: string;
    channel: string;
    status: string;
    totalPrice: number; // Porting from GO float64
    travellerPrice: number; // Porting from GO float64
    bookerId: string;
    creationChannelId: string;
    lastUpdateChannelId: string;
}

export interface AirLoyalty extends BaseAttributes {
    loyaltyId: string;
    programName: string;
    miles: string;
    milesToNextLevel: string;
    level: string;
    joined: string; // time.Time
}

export interface AncillaryService extends BaseAttributes {
    ancillaryType: string;
    bookingId: string;
    flightNumber: string;
    departureDate: string;
    baggageType: string;
    paxIndex: bigint; // Porting from GO int64
    quantity: bigint; // Porting from GO int64
    weight: number; // Porting from GO float64
    dimentionsLength: number; // Porting from GO float64
    dimentionsWidth: number; // Porting from GO float64
    dimentionsHeight: number; // Porting from GO float64
    priorityBagDrop: boolean;
    priorityBagReturn: boolean;
    lotBagInsurance: boolean;
    valuableBaggageInsurance: boolean;
    handsFreeBaggage: boolean;
    seatNumber: string;
    seatZone: string;
    neighborFreeSeat: boolean;
    upgradeAuction: boolean;
    changeType: string;
    otherAncilliaryType: string;
    priorityServiceType: string;
    loungeAccess: boolean;
    price: number; // Porting from GO float64
    currency: string;
}

export interface EmailHistory extends BaseAttributes {
    address: string;
    type: string;
}

export interface PhoneHistory extends BaseAttributes {
    number: string;
    countryCode: string;
    type: string;
}

export interface HotelBooking extends BaseAttributes {
    bookingId: string;
    hotelCode: string;
    numNights: number;
    numGuests: number;
    productId: string;
    checkInDate: string; // time.Time
    roomTypeCode: string;
    roomTypeName: string;
    roomTypeDescription: string;
    ratePlanCode: string;
    ratePlanName: string;
    ratePlanDescription: string;
    attributeCodes: string;
    attributeNames: string;
    attributeDescriptions: string;
    addOnCodes: string;
    addOnNames: string;
    addOnDescriptions: string;
    pricePerTraveller: string;
    bookerId: string;
    status: string;
    totalSegmentBeforeTax: number; // Porting from GO float64
    totalSegmentAfterTax: number; // Porting from GO float64
    creationChannelId: string;
    lastUpdateChannelId: string;
}

export interface HotelLoyalty extends BaseAttributes {
    loyaltyId: string;
    programName: string;
    points: string;
    units: string;
    pointsToNextLevel: string;
    level: string;
    joined: string; // time.Time
}

export interface HotelStay extends BaseAttributes {
    stayId: string;
    bookingId: string;
    currencyCode: string;
    currencyName: string;
    currencySymbol: string;
    firstName: string;
    lastName: string;
    email: string;
    phone: string;
    startDate: string; // time.Time
    hotelCode: string;
    type: string;
    description: string;
    amount: string;
    date: string; // time.Time
}

export interface CustomerServiceInteraction extends BaseAttributes {
    channel: string;
    conversation: string;
    duration: number;
    endTime: Date;
    interactionType: string;
    languageCode: string;
    languageName: string;
    sentimentScore: number;
    sessionId: string;
    startTime: Date;
    status: string;
}

export interface ConversationItem {
    content: string;
    from: string;
    to: string;
    sentiment: string;
    start_time: string;
    end_time: string;
}

export interface Clickstream extends BaseAttributes {
    sessionId: string;
    eventTimestamp: string; // time.Time
    eventType: string;
    eventVersion: string;
    arrivalTimestamp: string; // time.Time
    userAgent: string;
    customEventName: string;
    customerBirthdate: string; // time.Time
    customerCountry: string;
    customerEmail: string;
    customerFirstName: string;
    customerGender: string;
    customerId: string;
    customerLastName: string;
    customerNationality: string;
    customerPhone: string;
    customerType: string;
    customerLoyaltyId: string;
    languageCode: string;
    currency: string;
    products: string;
    quantities: string;
    productsPrices: string;
    ecommerceAction: string;
    orderPaymentType: string;
    orderPromoCode: string;
    pageName: string;
    pageTypeEnvironment: string;
    transactionId: string;
    bookingId: string;
    geofenceLatitude: string;
    geofenceLongitude: string;
    geofenceId: string;
    geofenceName: string;
    poiId: string;
    url: string;
    custom: string;
    fareClass: string;
    fareType: string;
    flightSegmentsDepartureDateTime: string;
    flightSegmentsArrivalDateTime: string;
    flightSegments: string;
    flightSegmentsSku: string;
    flightRoute: string;
    flightNumbers: string;
    flightMarket: string;
    flightType: string;
    originDate: string;
    originDateTime: string; // time.Time
    returnDate: string;
    returnDateTime: string; // time.Time
    returnFlightRoute: string;
    numPaxAdults: number;
    numPaxInf: number;
    numPaxChildren: number;
    paxType: string;
    totalPassengers: number;
    roomType: string;
    ratePlan: string;
    checkinDate: string; // time.Time
    checkoutDate: string; // time.Time
    numNights: number;
    numGuests: number;
    numGuestsAdult: number;
    numGuestsChildren: number;
    hotelCode: string;
    hotelCodeList: string;
    hotelName: string;
    destination: string;
}

export interface LoyaltyTx extends BaseAttributes {
    pointsOffset: number; // Porting from GO float64
    pointUnit: string;
    originPointsOffset: number; // Porting from GO float64
    qualifyingPointsOffset: number; // Porting from GO float64
    source: string;
    category: string;
    bookingDate: string; // time.Time
    orderNumber: string;
    productId: string;
    expireInDays: number;
    amount: number; // Porting from GO float64
    amountType: string;
    voucherQuantity: number;
    corporateReferenceNumber: string;
    promotions: string;
    activityDay: string; // time.Time
    location: string;
    toLoyaltyId: string;
    fromLoyaltyId: string;
    organizationCode: string;
    eventName: string;
    documentNumber: string;
    corporateId: string;
    programName: string;
}

export interface MergeContext {
    timestamp: string;
    mergeType: string;
    confidenceUpdateFactor: number;
    ruleId: string;
    ruleSetVersion: string;
    operatorId: string;
    toMergeConnectId: string;
    mergeIntoConnectId: string;
}

export interface UnmergeRq {
    toUnmergeConnectID: string;
    mergedIntoConnectID: string;
    interactionToUnmerge: string;
    interactionType: string;
}

export interface MergeRq {
    source: string;
    target: string;
    falsePositive: boolean;
    mergeContext: MergeContext;
}

// imported from storage/customerprofiles_lowcost_constant.go
export enum MergeType {
    RULE = 'rule',
    AI = 'ai',
    MANUAL = 'manual',
    UNMERGE = 'unmerge',
}

export const mergeTypePrefix = 'MERGE_TYPE_';
