def buildObjectRecord(rec):
    try:
        # remove the word holder from each of the below lines please
        arrayRec = []
        newRec = {}
        newRec["modelVersion"] = rec["modelVersion"]
        newRec["id"] = rec["id"]
        newRec["lastUpdatedOn"] = rec["lastUpdatedOn"]
        newRec["createdOn"] = rec["createdOn"]
        newRec["lastUpdatedBy"] = rec["lastUpdatedBy"]
        newRec["createdBy"] = rec["createdBy"]
        generateEmailsArray(arrayRec, rec)
        generatePhonesArray(arrayRec, rec)
        generateAdressesArray(arrayRec, rec)
        newRec["honorific"] = rec["honorific"]
        newRec["firstName"] = rec["firstName"]
        newRec["middleName"] = rec["middleName"]
        newRec["lastName"] = rec["lastName"]
        newRec["gender"] = rec["gender"]
        newRec["pronoun"] = rec["pronoun"]
        newRec["dateOfBirth"] = rec["dateOfBirth"]
        newRec["language.code"] = rec["language"]["code"]
        newRec["language.name"] = rec["language"]["name"]
        newRec["nationality.code"] = rec["nationality"]["code"]
        newRec["nationality.name"] = rec["nationality"]["name"]
        newRec["jobTitle"] = rec["jobTitle"]
        newRec["parentCompany"] = rec["parentCompany"]
        generateLoyaltyProgramsArray(arrayRec, rec)
    except Exception as e:
        newRec["error"] = str(e)
        return newRec
    return newRec

def fillAddressDetails(startingString, dict, element):
    dict[startingString+"type"] = element["type"]
    dict[startingString+"line1"] = element["line1"]
    dict[startingString+"line2"] = element["line2"]
    dict[startingString+"line2"] = element["line3"]
    dict[startingString+"line2"] = element["line4"]
    dict[startingString+"city"] = element["city"]
    dict[startingString+"stateProvince.code"] = element["stateProvince"]["code"]
    dict[startingString+"stateProvince.name"] = element["stateProvince"]["name"]
    dict[startingString+"postalCode"] = element["postalCode"]
    dict[startingString+"country.code"] = element["country"]["code"]
    dict[startingString+"country.name"] = element["country"]["name"]
    dict[startingString+"primary"] = element["primary"]
    return dict

def generateEmailsArray(arrayRec, rec):
    arr = rec["emails"]
    for element in arr:
        newRecEmail = {}
        newRecEmail["emails.type"] = element["type"]
        newRecEmail["emails.address"] = element["address"]
        newRecEmail["emails.primary"] = element["primary"]
        arrayRec.append(newRecEmail)

def generatePhonesArray(arrayRec, rec):
    arr = rec["phones"]
    for element in arr:
        newRecPhone = {}
        newRecPhone["phones.type"] = element["type"]
        newRecPhone["phones.number"] = element["number"]
        newRecPhone["phones.primary"] = element["primary"]
        newRecPhone["phones.countryCode"] = element["countryCODE"]
        arrayRec.append(newRecPhone)

def generateAdressesArray(arrayRec, rec):
    arr = rec["addresses"]
    for element in arr:
        newRecAddresses = {}
        newRecAddresses = fillAddressDetails("addresses.", newRecAddresses, element)
        arrayRec.append(newRecAddresses)

def generateLoyaltyProgramsArray(arrayRec, rec):
    arr = rec["loyaltyPrograms"]
    for element in arr:
        newRecLoyaltyPrograms = {}
        newRecLoyaltyPrograms["loyaltyPrograms.id"] = element["id"]
        newRecLoyaltyPrograms["loyaltyPrograms.programName"] = element["programName"]
        newRecLoyaltyPrograms["loyaltyPrograms.points"] = element["points"]
        newRecLoyaltyPrograms["loyaltyPrograms.pointUnit"] = element["pointUnit"]
        newRecLoyaltyPrograms["loyaltyPrograms.pointsToNextLevel"] = element["pointsToNextLevel"]
        newRecLoyaltyPrograms["loyaltyPrograms.level"] = element["level"]
        newRecLoyaltyPrograms["loyaltyPrograms.joined"] = element["joined"]
        arrayRec.append(newRecLoyaltyPrograms)
