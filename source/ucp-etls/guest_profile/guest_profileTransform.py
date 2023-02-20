def buildObjectRecord(rec):
    try:
        newRec = {}
        newRec["modelVersion"] = rec["modelVersion"]
        newRec["id"] = rec["id"]
        newRec["lastUpdatedOn"] = rec["lastUpdatedOn"]
        newRec["createdOn"] = rec["createdOn"]
        newRec["lastUpdatedBy"] = rec["lastUpdatedBy"]
        newRec["createdBy"] = rec["createdBy"]
        fillEmailsDetails("emails_", newRec, rec)
        fillPhonesDetails("phones_", newRec, rec)
        fillAddressesDetails("addresses_", newRec, rec)
        newRec["honorific"] = rec["honorific"]
        newRec["firstName"] = rec["firstName"]
        newRec["middleName"] = rec["middleName"]
        newRec["lastName"] = rec["lastName"]
        newRec["gender"] = rec["gender"]
        newRec["pronoun"] = rec["pronoun"]
        newRec["dateOfBirth"] = rec["dateOfBirth"]
        newRec["language_code"] = rec["language_code"]
        newRec["language_name"] = rec["language_name"]
        newRec["nationality_code"] = rec["nationality_code"]
        newRec["nationality_name"] = rec["nationality_name"]
        newRec["jobTitle"] = rec["jobTitle"]
        newRec["parentCompany"] = rec["parentCompany"]
        fillLoyaltyProgramsDetails("loyaltyPrograms_", newRec, rec)
        newRec["partition_0"] = rec["partition_0"]
        newRec["partition_1"] = rec["partition_1"]
        newRec["partition_2"] = rec["partition_2"]
    except Exception as e:
        newRec["error"] = str(e)
        return newRec
    return newRec


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


def fillLoyaltyProgramsDetails(startingString, dict, element):
    dict[startingString+"id"] = element[startingString+"id"]
    dict[startingString+"programName"] = element[startingString+"programName"]
    dict[startingString+"points"] = element[startingString+"points"]
    dict[startingString+"pointUnit"] = element[startingString+"pointUnit"]
    dict[startingString+"pointsToNextLevel"] = element[startingString+"pointsToNextLevel"]
    dict[startingString+"level"] = element[startingString+"level"]
    dict[startingString+"joined"] = element[startingString+"joined"]
    return dict
