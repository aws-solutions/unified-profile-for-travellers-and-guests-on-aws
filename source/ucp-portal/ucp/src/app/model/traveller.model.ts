
export class Traveller {
    // Metadata
    modelVersion: string;
    errors: string[];

    // Profile IDs
    connectId: string;
    travellerId: string;
    pssid: string;
    gdsid: string;
    pmsid: string;
    crsid: string;

    // Profile Data
    honorific: string;
    firstName: string;
    middleName: string;
    lastName: string;
    gender: string;
    pronoun: string[];
    birthDate: Date;
    jobTitle: string;
    companyName: string;

    // Contact Info
    phoneNumber: string;
    mobilePhoneNumber: string;
    homePhoneNumber: string;
    businessPhoneNumber: string;
    personalEmailAddress: string;
    businessEmailAddress: string;
    nationalityCode: string;
    nationalityName: string;
    languageCode: string;
    languageName: string;

    // Addresses
    homeAddress: Address;
    businessAddress: Address;
    mailingAddress: Address;
    billingAddress: Address;

    // Payment Info
    // TODO: should payment data be for an order, profile, separate history object or combination?

    // Object Type Records
    airBookingRecords: AirBooking[];
    airLoyaltyRecords: AirLoyalty[];
    clickstreamRecords: Clickstream[];
    emailHistoryRecords: EmailHistory[];
    hotelBookingRecords: HotelBooking[];
    hotelLoyaltyRecords: HotelLoyalty[];
    hotelStayRecords: HotelStay[];
    phoneHistoryRecords: PhoneHistory[];

    parsingErrors: string[];

    constructor() { }
}

export class Address {
    line1: string;
    line2?: string;
    city: string;
    state?: string;
    zip: string;
    country: string;
}

export class AirBooking {
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
    price: string;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class AirLoyalty {
    loyaltyId: string;
    programName: string;
    miles: string;
    milesToNextLevel: string;
    level: string;
    joined: Date;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class Clickstream {
    sessionId: string;
    eventTimestamp: Date;
    eventType: string;
    eventVersion: string;
    arrivalTimestamp: Date;
    userAgent: string;
    products: string;
    fareClass: string;
    fareType: string;
    flightSegmentsDepartureDateTime: Date;
    flightNumbers: string;
    flightMarket: string;
    flightType: string;
    originDate: string;
    originDateTime: Date;
    returnDate: string;
    returnDateTime: Date;
    returnFlightRoute: string;
    numPaxAdults: number;
    numPaxInf: number;
    numPaxChildren: number;
    paxType: string;
    totalPassengers: number;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class EmailHistory {
    address: string;
    type: string;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class HotelBooking {
    bookingId: string;
    hotelCode: string;
    numNights: number;
    numGuests: number;
    productId: string;
    checkInDate: Date;
    roomTypeCode: string;
    roomTypeName: string;
    roomTypeDescription: string;
    attributeCodes: string;
    attributeNames: string;
    attributeDescriptions: string;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class HotelLoyalty {
    loyaltyId: string;
    programName: string;
    points: string;
    units: string;
    pointsToNextLevel: string;
    level: string;
    joined: Date;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class HotelStay {
    id: string;
    bookingId: string;
    currencyCode: string;
    currencyName: string;
    currencySymbol: string;
    firstName: string;
    lastName: string;
    email: string;
    phone: string;
    startDate: Date;
    hotelCode: string;
    type: string;
    description: string;
    amount: string;
    date: Date;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

export class PhoneHistory {
    number: string;
    countryCode: string;
    type: string;
    lastUpdated: Date;
    lastUpdatedBy: string;
}

