def buildObjectRecord(rec):
    try:
        newRec = {}
        newRec["objectVersion"] = rec["objectVersion"]
        newRec["modelVersion"] = rec["modelVersion"]
        newRec["id"] = rec["id"]
        newRec["externalIDs_id"] = rec["externalIDs_id"]
        newRec["externalIDs_IdName"] = rec["externalIDs_IdName"]
        newRec["externalIDs_originatingSystem"] = rec["externalIDs_originatingSystem"]
        newRec["lastUpdatedOn"] = rec["lastUpdatedOn"]
        newRec["createdOn"] = rec["createdOn"]
        newRec["lastUpdatedBy"] = rec["lastUpdatedBy"]
        newRec["createdBy"] = rec["createdBy"]
        newRec["creationChannelId"] = rec["creationChannelId"]
        newRec["hotelCode"] = rec["hotelCode"]
        newRec["nNights"] = rec["nNights"]
        newRec["nGuests"] = rec["nGuests"]
        newRec["startDate"] = rec["startDate"]
        newRec["endDate"] = rec["endDate"]
        fillHolderDetails("holder_", newRec, rec)
        fillPaymentInformationDetails("paymentInformation_", newRec, rec)
        # contextID
        newRec["contextId"] = rec["contextId"]
        newRec["groupId"] = rec["groupId"]
        newRec["status"] = rec["status"]
        # currency
        newRec["currency_code"] = rec["currency_code"]
        newRec["currency_name"] = rec["currency_name"]
        newRec["currency_symbol"] = rec["currency_symbol"]
        # cancelReason
        newRec["cancelReason_reason"] = rec["cancelReason_reason"]
        newRec["cancelReason_comment_type"] = rec["cancelReason_comment_type"]
        newRec["cancelReason_comment_language_code"] = rec["cancelReason_comment_language_name"]
        newRec["cancelReason_comment_language_name"] = rec["cancelReason_comment_language_code"]
        newRec["cancelReason_comment_title"] = rec["cancelReason_comment_title"]
        newRec["cancelReason_comment_text"] = rec["cancelReason_comment_text"]
        newRec["cancelReason_comment_context"] = rec["cancelReason_comment_context"]
        newRec["cancelReason_comment_createdDateTime"] = rec["cancelReason_comment_createdDateTime"]
        newRec["cancelReason_comment_createdBy"] = rec["cancelReason_comment_createdBy"]
        newRec["cancelReason_comment_lastModifiedDateTime"] = rec["cancelReason_comment_lastModifiedDateTime"]
        newRec["cancelReason_comment_lastModifiedBy"] = rec["cancelReason_comment_lastModifiedBy"]
        newRec["cancelReason_comment_isTravellerViewable"] = rec["cancelReason_comment_isTravellerViewable"]
        # segments
        newRec["segments_id"] = rec["segments_id"]
        newRec["segments_hotelCode"] = rec["segments_hotelCode"]
        newRec["segments_nNights"] = rec["segments_nNights"]
        newRec["segments_nGuests"] = rec["segments_nGuests"]
        newRec["segments_startDate"] = rec["segments_startDate"]
        fillHolderDetails("segments_holder_", newRec, rec)
        fillHolderDetails("segments_additionalGuests_", newRec, rec)
        fillPaymentInformationDetails("segments_paymentInformation_", newRec, rec)
        fillProductsDetails("segments_products_", newRec, rec)
        newRec["segments_groupId"] = rec["segments_groupId"]
        newRec["segments_status"] = rec["segments_status"]
        # pricePerNight
        fillPricePerNightDetails("segments_price_pricePerNight_", newRec, rec)
        # taxePerNight
        fillPricePerNightDetails("segments_price_taxePerNight_", newRec, rec)
        # pricePerStay
        newRec["segments_price_pricePerStay"] = rec["segments_price_pricePerStay"]
        newRec["segments_price_taxPerStay"] = rec["segments_price_taxPerStay"]
        newRec["segments_price_businessRules"] = rec["segments_price_businessRules"]
        newRec["segments_price_totalPricePerNightBeforeTaxes"] = rec["segments_price_totalPricePerNightBeforeTaxes"]
        newRec["segments_price_totalPricePerNightAfterTaxes"] = rec["segments_price_totalPricePerNightAfterTaxes"]
        newRec["segments_price_totalPricePerProductBeforeTaxes"] = rec["segments_price_totalPricePerProductBeforeTaxes"]
        newRec["segments_price_totalPricePerProductAfterTaxes"] = rec["segments_price_totalPricePerProductAfterTaxes"]
        newRec["segments_price_totalBeforeTax"] = float(rec["segments_price_totalBeforeTax_double"])
        newRec["segments_price_totalAfterTax"] = float(rec["segments_price_totalAfterTax_double"])
        # comments
        newRec["comments_type"] = rec["comments_type"]
        newRec["comments_language_code"] = rec["comments_language_code"]
        newRec["comments_language_name"] = rec["comments_language_name"]
        newRec["comments_title"] = rec["comments_title"]
        newRec["comments_text"] = rec["comments_text"]
        newRec["comments_context"] = rec["comments_context"]
        newRec["comments_createdDateTime"] = rec["comments_createdDateTime"]
        newRec["comments_createdBy"] = rec["comments_createdBy"]
        newRec["comments_lastModifiedDateTime"] = rec["comments_lastModifiedDateTime"]
        newRec["comments_lastModifiedBy"] = rec["comments_lastModifiedBy"]
        newRec["comments_isTravellerViewable"] = rec["comments_isTravellerViewable"]
        newRec["partition_0"] = rec["partition_0"]
        newRec["partition_1"] = rec["partition_1"]
        newRec["partition_2"] = rec["partition_2"]
        newRec["partition_3"] = rec["partition_3"]
    except Exception as e:
        newRec["error"] = str(e)
        return newRec
    return newRec


def fillHolderDetails(startingString, dict, element):
    dict[startingString+"modelVersion"] = element[startingString+"modelVersion"]
    dict[startingString+"id"] = element[startingString+"id"]
    dict[startingString+"lastUpdatedOn"] = element[startingString+"lastUpdatedOn"]
    dict[startingString+"createdOn"] = element[startingString+"createdOn"]
    dict[startingString+"lastUpdatedBy"] = element[startingString+"lastUpdatedBy"]
    dict[startingString+"createdBy"] = element[startingString+"createdBy"]
    # addresses, emails, phones
    fillEmailsDetails(startingString+"emails_", dict, element)
    fillPhonesDetails(startingString+"phones_", dict, element)
    fillAddressesDetails(startingString+"addresses_", dict, element)
    dict[startingString+"honorific"] = element[startingString+"honorific"]
    dict[startingString+"firstName"] = element[startingString+"firstName"]
    dict[startingString+"middleName"] = element[startingString+"middleName"]
    dict[startingString+"lastName"] = element[startingString+"lastName"]
    dict[startingString+"gender"] = element[startingString+"gender"]
    dict[startingString+"pronoun"] = element[startingString+"pronoun"]
    dict[startingString+"dateOfBirth"] = element[startingString+"dateOfBirth"]
    dict[startingString+"language_code"] = element[startingString+"language_code"]
    dict[startingString+"language_name"] = element[startingString+"language_name"]
    dict[startingString+"nationality_code"] = element[startingString+"nationality_code"]
    dict[startingString+"nationality_name"] = element[startingString+"nationality_name"]
    dict[startingString+"jobTitle"] = element[startingString+"jobTitle"]
    dict[startingString+"parentCompany"] = element[startingString+"parentCompany"]
    fillLoyaltyProgramsDetails(startingString+"loyaltyPrograms_", dict, element)
    return dict


def fillAddressesDetails(startingString, dict, element):
    dict[startingString+"type"] = element[startingString+"type"]
    dict[startingString+"line1"] = element[startingString+"line1"]
    dict[startingString+"line2"] = element[startingString+"line2"]
    dict[startingString+"line2"] = element[startingString+"line3"]
    dict[startingString+"line2"] = element[startingString+"line4"]
    dict[startingString+"city"] = element[startingString+"city"]
    dict[startingString+"stateProvince_code"] = element[startingString + "stateProvince_code"]
    dict[startingString+"stateProvince_name"] = element[startingString + "stateProvince_name"]
    dict[startingString+"postalCode"] = element[startingString+"postalCode"]
    dict[startingString+"country_code"] = element[startingString+"country_code"]
    dict[startingString+"country_name"] = element[startingString+"country_name"]
    dict[startingString+"primary"] = element[startingString+"primary"]
    return dict


def fillEmailsDetails(startingString, dict, element):
    dict[startingString+"type"] = element[startingString+"type"]
    dict[startingString+"address"] = element[startingString+"address"]
    dict[startingString+"primary"] = element[startingString+"primary"]
    return dict


def fillPhonesDetails(startingString, dict, element):
    dict[startingString+"type"] = element[startingString+"type"]
    dict[startingString+"number"] = element[startingString+"number"]
    dict[startingString+"primary"] = element[startingString+"primary"]
    dict[startingString+"countryCode"] = element[startingString+"countryCode"]
    return dict


def fillLoyaltyProgramsDetails(startingString, dict, element):
    dict[startingString+"id"] = element[startingString+"id"]
    dict[startingString+"programName"] = element[startingString+"programName"]
    dict[startingString+"points"] = element[startingString+"points"]
    dict[startingString+"pointUnit"] = element[startingString+"pointUnit"]
    dict[startingString+"pointsToNextLevel"] = element[startingString+"pointsToNextLevel"]
    dict[startingString+"level"] = element[startingString+"level"]
    dict[startingString+"joined"] = element[startingString+"joined"]
    return dict


def fillPaymentInformationDetails(startingString, dict, element):
    dict[startingString+"paymentType"] = element[startingString+"paymentType"]
    dict[startingString+"ccInfo_token"] = element[startingString+"ccInfo_token"]
    dict[startingString+"ccInfo_cardType"] = element[startingString+"ccInfo_cardType"]
    dict[startingString+"ccInfo_cardExp"] = element[startingString+"ccInfo_cardExp"]
    dict[startingString+"ccInfo_cardCvv"] = element[startingString+"ccInfo_cardCvv"]
    dict[startingString+"ccInfo_expiration"] = element[startingString+"ccInfo_expiration"]
    dict[startingString+"ccInfo_name"] = element[startingString+"ccInfo_name"]
    fillAddressesDetails(startingString+"ccInfo_address_", dict, element)
    dict[startingString+"routingNumber"] = element[startingString+"routingNumber"]
    dict[startingString+"accountNumber"] = element[startingString+"accountNumber"]
    dict[startingString+"voucherID"] = element[startingString+"voucherID"]
    return dict


def fillProductsDetails(startingString, dict, element):
    dict[startingString+"id"] = element[startingString+"id"]
    dict[startingString+"roomType_Code"] = element[startingString+"roomType_Code"]
    dict[startingString+"roomType_Name"] = element[startingString+"roomType_Name"]
    dict[startingString+"roomType_Description"] = element[startingString + "roomType_Description"]
    dict[startingString+"ratePlan_Code"] = element[startingString+"ratePlan_Code"]
    dict[startingString+"ratePlan_Name"] = element[startingString+"ratePlan_Name"]
    dict[startingString+"ratePlan_Description"] = element[startingString + "ratePlan_Description"]
    # attributes
    fillAttributesAddOnDetails(startingString+"attributes_", dict, element)
    # addOn
    fillAttributesAddOnDetails(startingString+"addOn_", dict, element)
    return dict


def fillAttributesAddOnDetails(startingString, dict, element):
    dict[startingString+"Code"] = element[startingString+"Code"]
    dict[startingString+"Name"] = element[startingString+"Name"]
    dict[startingString+"Description"] = element[startingString+"Description"]
    return dict


def fillPricePerNightDetails(startingString, dict, element):
    dict[startingString+"date"] = element[startingString+"date"]
    dict[startingString+"label"] = element[startingString+"label"]
    dict[startingString+"amountPerProduct_productId"] = element[startingString + "amountPerProduct_productId"]
    dict[startingString+"amountPerProduct_amount"] = element[startingString + "amountPerProduct_amount"]
    dict[startingString+"amountPerProduct_label"] = element[startingString + "amountPerProduct_label"]
    dict[startingString+"amountPerProduct_currency_code"] = element[startingString + "amountPerProduct_currency_code"]
    dict[startingString+"amountPerProduct_currency_name"] = element[startingString + "amountPerProduct_currency_name"]
    dict[startingString+"amountPerProduct_currency_symbol"] = element[startingString + "amountPerProduct_currency_symbol"]
    dict[startingString+"currency_code"] = element[startingString+"currency_code"]
    dict[startingString+"currency_name"] = element[startingString+"currency_name"]
    dict[startingString+"currency_symbol"] = element[startingString+"currency_symbol"]
    return dict
