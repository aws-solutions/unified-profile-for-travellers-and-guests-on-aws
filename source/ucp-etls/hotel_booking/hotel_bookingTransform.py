def buildObjectRecord(rec):
    try:
        arrayRec = []
        newRec = {}
        newRec["objectVersion"] = rec["objectVersion"]
        newRec["modelVersion"] = rec["modelVersion"]
        newRec["id"] = rec["id"]
        #external ids
        generateExternalIdArray(arrayRec, rec)
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
        #holders
        newRec["holder.modelVersion"] = rec["holder"]["modelVersion"]
        newRec["holder.id"] = rec["holder"]["id"]
        newRec["holder.lastUpdatedOn"] = rec["holder"]["lastUpdatedOn"]
        newRec["holder.createdOn"] = rec["holder"]["createdOn"]
        newRec["holder.lastUpdatedBy"] = rec["holder"]["lastUpdatedBy"]
        newRec["holder.createdBy"] = rec["holder"]["createdBy"]
        generateEmailsArray(arrayRec, rec)
        generatePhonesArray(arrayRec, rec)
        generateAdressesArray(arrayRec, rec)
        newRec["holder.honorific"] = rec["holder"]["honorific"]
        newRec["holder.firstName"] = rec["holder"]["firstName"]
        newRec["holder.middleName"] = rec["holder"]["middleName"]
        newRec["holder.lastName"] = rec["holder"]["lastName"]
        newRec["holder.gender"] = rec["holder"]["gender"]
        newRec["holder.pronoun"] = rec["holder"]["pronoun"]
        newRec["holder.dateOfBirth"] = rec["holder"]["dateOfBirth"]
        newRec["holder.language.code"] = rec["holder"]["language"]["code"]
        newRec["holder.language.name"] = rec["holder"]["language"]["name"]
        newRec["holder.nationality.code"] = rec["holder"]["nationality"]["code"]
        newRec["holder.nationality.name"] = rec["holder"]["nationality"]["name"]
        newRec["holder.jobTitle"] = rec["holder"]["jobTitle"]
        newRec["holder.parentCompany"] = rec["holder"]["parentCompany"]
        generateLoyaltyProgramsArray(arrayRec, rec)
        #paymentInformation
        newRec["paymentInformation.paymentType"] = rec["paymentInformation"]["paymentType"]
        newRec["paymentInformation.ccInfo.token"] = rec["paymentInformation"]["ccInfo"]["token"]
        newRec["paymentInformation.ccInfo.cardType"] = rec["paymentInformation"]["ccInfo"]["cardType"]
        newRec["paymentInformation.ccInfo.cardExp"] = rec["paymentInformation"]["ccInfo"]["cardExp"]
        newRec["paymentInformation.ccInfo.cardCvv"] = rec["paymentInformation"]["ccInfo"]["cardCvv"]
        newRec["paymentInformation.ccInfo.expiration"] = rec["paymentInformation"]["ccInfo"]["expiration"]
        newRec["paymentInformation.ccInfo.name"] = rec["paymentInformation"]["ccInfo"]["name"]
        newRec = fillAddressDetails("paymentInformation.address.", newRec, rec["paymentInformation"]["address"])
        newRec["paymentInformation.routingNumber"] = rec["paymentInformation"]["routingNumber"]
        newRec["paymentInformation.accountNumber"] = rec["paymentInformation"]["accountNumber"]
        newRec["paymentInformation.voucherID"] = rec["paymentInformation"]["voucherID"]
        #contextID
        newRec["contextId"] = rec["contextId"]
        newRec["groupId"] = rec["groupId"]
        newRec["status"] = rec["status"]
        #currency
        newRec["currency.code"] = rec["currency"]["code"]
        newRec["currency.name"] = rec["currency"]["name"]
        newRec["currency.symbol"] = rec["currency"]["symbol"]
        #cancelReason
        newRec["cancelReason.reason"] = rec["cancelReason"]["reason"]
        newRec["cancelReason.comment.type"] = rec["cancelReason"]["comment"]["type"]
        newRec["cancelReason.comment.language.code"] = rec["cancelReason"]["comment"]["language"]["code"]
        newRec["cancelReason.comment.language.name"] = rec["cancelReason"]["comment"]["language"]["name"]
        newRec["cancelReason.comment.title"] = rec["cancelReason"]["comment"]["title"]
        newRec["cancelReason.comment.text"] = rec["cancelReason"]["comment"]["text"]
        newRec["cancelReason.comment.context"] = rec["cancelReason"]["comment"]["context"]
        newRec["cancelReason.comment.createdDateTime"] = rec["cancelReason"]["comment"]["createdDataTime"]
        newRec["cancelReason.comment.createdBy"] = rec["cancelReason"]["comment"]["createdBy"]
        newRec["cancelReason.comment.lastModifiedDateTime"] = rec["cancelReason"]["comment"]["lastModifiedDateTime"]
        newRec["cancelReason.comment.lastModifiedBy"] = rec["cancelReason"]["comment"]["lastModifiedBy"]
        newRec["cancelReason.comment.isTravellerViewable"] = rec["cancelReason"]["comment"]["isTravellerViewable"]
        #segments
        #comments
    except Exception as e:
            newRec["error"] = str(e)
            return newRec
    return newRec

def getAttr(attributes, name):
    obj = next(x for x in attributes if x["name"] == name) # loop through each attribute in list once for better performance
    if obj is None:
        return ""
    type = obj["type"]
    if type == "string":
        return obj["stringvValue"] # is this a typo?
    elif type == "strings":
        return obj["stringValues"]
    elif type == "number":
        return obj["numValue"]
    elif type == "numbers":
        return obj["numValues"]
    else:
        raise Exception("Unknown attribute type")

def getHolder(holder, name):
    return 0

def getPaymentInformation(paymentInfo, name):
    return 0

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

def fillHolderDetails(startingString, dict, element):
    dict[startingString+"modelVersion"] = element["modelVersion"]
    dict[startingString+"id"] = element["id"]
    dict[startingString+"lastUpdatedOn"] = element["lastUpdatedOn"]
    dict[startingString+"createdOn"] = element["createdOn"]
    dict[startingString+"lastUpdatedBy"] = element["lastUpdatedBy"]
    dict[startingString+"createdBy"] = element["createdBy"]
    #addresses, emails, phones
    #dict = fillAddressDetails(startingString, dict, element)
    #TODO: Create programatic way of adding addresses, emails, and phone arrays to holder object 
    dict[startingString+"honorific"] = element["honorific"]
    dict[startingString+"firstName"] = element["firstName"]
    dict[startingString+"middleName"] = element["middleName"]
    dict[startingString+"lastName"] = element["lastName"]
    dict[startingString+"gender"] = element["gender"]
    dict[startingString+"pronoun"] = element["pronoun"]
    dict[startingString+"dateOfBirth"] = element["dateOfBirth"]
    dict[startingString+"language.code"] = element["language"]["code"]
    dict[startingString+"language.name"] = element["language"]["name"]
    dict[startingString+"nationality.code"] = element["nationality"]["code"]
    dict[startingString+"nationality.name"] = element["nationality"]["name"]
    dict[startingString+"jobTitle"] = element["jobTitle"]
    dict[startingString+"parentCompany"] = element["parentCompany"]
    return dict

def generateExternalIdArray(arrayRec, rec):
    arr = rec["externalIds"]
    for element in arr:
        newRecExternalId = {}
        newRecExternalId["externalIds"]["id"] = element["id"]
        newRecExternalId["externalIds"]["IdName"] = element["IdName"]
        newRecExternalId["externalIds"]["originatingSystem"] = element["originatingSystem"]
        arrayRec.append(newRecExternalId)

def generateEmailsArray(arrayRec, rec):
    arr = rec["holder"]["emails"]
    for element in arr:
        newRecEmail = {}
        newRecEmail["holder.emails.type"] = element["type"]
        newRecEmail["holder.emails.address"] = element["address"]
        newRecEmail["holder.emails.primary"] = element["primary"]
        arrayRec.append(newRecEmail)

def generatePhonesArray(arrayRec, rec):
    arr = rec["holder"]["phones"]
    for element in arr:
        newRecPhone = {}
        newRecPhone["holder.phones.type"] = element["type"]
        newRecPhone["holder.phones.number"] = element["number"]
        newRecPhone["holder.phones.primary"] = element["primary"]
        newRecPhone["holder.phones.countryCode"] = element["countryCODE"]
        arrayRec.append(newRecPhone)

def generateAdressesArray(arrayRec, rec):
    arr = rec["holder"]["addresses"]
    for element in arr:
        newRecAddresses = {}
        newRecAddresses = fillAddressDetails("holder.addresses.", newRecAddresses, element)
        arrayRec.append(newRecAddresses)

def generateLoyaltyProgramsArray(arrayRec, rec):
    arr = rec["holder"]["loyaltyPrograms"]
    for element in arr:
        newRecLoyaltyPrograms = {}
        newRecLoyaltyPrograms["holder.loyaltyPrograms.id"] = element["id"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.programName"] = element["programName"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.points"] = element["points"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.pointUnit"] = element["pointUnit"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.pointsToNextLevel"] = element["pointsToNextLevel"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.level"] = element["level"]
        newRecLoyaltyPrograms["holder.loyaltyPrograms.joined"] = element["joined"]
        arrayRec.append(newRecLoyaltyPrograms)

def generateSegmentsArray(arrayRec, rec):
    arr = rec["segments"]
    for element in arr:
        newRecSegments = {}
        newRecSegments["segments.id"] = element["id"]
        newRecSegments["segments.hotelCode"] = element["hotelCode"]
        newRecSegments["segments.nNights"] = element["nNights"]
        newRecSegments["segments.nGuests"] = element["nGuests"]
        newRecSegments["segments.startDate"] = element["startDate"]
        #products
        generateProductsArray(arrayRec, element)
        #holder
        #additionalGuests
        #paymentInformation
        arrayRec.append(newRecSegments)

def generateProductsArray(arrayRec, rec):
    arr = rec["products"]
    for element in arr:
        newRecProducts = {}
        newRecProducts["segments.products.id"] = element["id"]
        newRecProducts["segments.products.roomType.Code"] = element["roomType"]["Code"]
        newRecProducts["segments.products.roomType.Name"] = element["roomType"]["Name"]
        newRecProducts["segments.products.roomType.Description"] = element["roomType"]["Description"]
        newRecProducts["segments.products.ratePlan.Code"] = element["ratePlan"]["Code"]
        newRecProducts["segments.products.ratePlan.Name"] = element["ratePlan"]["Name"]
        newRecProducts["segments.products.ratePlan.Description"] = element["ratePlan"]["Description"]
        #attributes
        generateAttributesArray(arrayRec, element)
        #addOn
        generateAddOnArray(arrayRec, element)
        #holder
        #additionalGuests
        arrayRec.append(newRecProducts)

def generateAttributesArray(arrayRec, rec):
    arr = rec["attributes"]
    for element in arr:
        newRecAttributes = {}
        newRecAttributes["segments.products.ratePlan.attributes.Code"] = element["Code"]
        newRecAttributes["segments.products.ratePlan.attributes.Name"] = element["Name"]
        newRecAttributes["segments.products.ratePlan.attributes.Description"] = element["Description"]
        arrayRec.append(newRecAttributes)

def generateAddOnArray(arrayRec, rec):
    arr = rec["attributes"]
    for element in arr:
        newRecAddOn = {}
        newRecAddOn["segments.products.ratePlan.addOn.Code"] = element["Code"]
        newRecAddOn["segments.products.ratePlan.addOn.Name"] = element["Name"]
        newRecAddOn["segments.products.ratePlan.addOn.Description"] = element["Description"]
        arrayRec.append(newRecAddOn)

