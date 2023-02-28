# this file shopuld eventyally contain common functions between all etls
# this function set the mandatory ACCP traveller ID using the following rule
# if the guest has an ID provided by customer we use this one
# if not we will use the unique generated UUID for all ACCP records associated with this guest


def setTravellerId(rec, guestBizObject, uid):
    rec["traveller_id"] = guestBizObject["id"]
    if guestBizObject["id"] == "":
        rec["traveller_id"] = uid


# this function set the primary email according to the following logicc
# 1- if one or multiple emails are flagged as Primary, we choose the first one
# 2- if not we choose the first email in the list.
# 3- then we use teh email type to chosse whether we set the biz_email key or 'email' key
# 3- if the list is empty we set both keys to empty strings to allow creation of the column in the CSV
def setPrimaryEmail(rec, emails):
    if len(emails) == 0:
        rec['email_business'] = ""
        rec['email'] = ""
        return
    primaryEmail = emails[0]
    for email in emails:
        if email['primary'] == True:
            primaryEmail = email
            break
    if primaryEmail['type'] == 'business':
        rec['email_business'] = primaryEmail['address']
    else:
        rec['email'] = primaryEmail['address']
    # we keep the email type in case the value is outside the enum
    rec['email_type'] = primaryEmail['type']


# this function sets the primary phone using the same logic than the email function
# we have 4 types of phones: default, business, mobile and home
def setPrimaryPhone(rec, phones):
    if len(phones) == 0:
        rec['phone'] = ""
        rec['phone_home'] = ""
        rec['phone_mobile'] = ""
        rec['phone_business'] = ""
        return
    primaryPhone = phones[0]
    for phone in phones:
        if phone['primary'] == True:
            primaryPhone = phone
            break
    if primaryPhone['type'] == 'business':
        rec['phone_business'] = primaryPhone['number']
    elif primaryPhone['type'] == 'home':
        rec['phone_home'] = primaryPhone['number']
    elif primaryPhone['type'] == 'mobile':
        rec['phone_business'] = primaryPhone['number']
    else:
        rec['phone'] = primaryPhone['number']
    rec['phone_type'] = primaryPhone['type']


# this function sets the primary address  using the same logic than the email function
def setPrimaryAddress(rec, addresses):
    if len(addresses) == 0:
        rec['address_type'] = ""
        rec['address_line1'] = ""
        rec['address_line2'] = ""
        rec['address_line3'] = ""
        rec['address_line4'] = ""
        rec['address_city'] = ""
        rec['address_state_province'] = ""
        rec['address_postal_code'] = ""
        rec['address_country'] = ""
        return
    primaryAddress = addresses[0]
    for address in addresses:
        if address['primary'] == True:
            primaryAddress = address
            break
    if primaryAddress['type'] == 'business':
        setAddress(rec, 'business_', primaryAddress)
    elif primaryAddress['type'] == 'shipping':
        setAddress(rec, 'shipping_', primaryAddress)
    elif primaryAddress['type'] == 'mailing':
        setAddress(rec, 'mailing_', primaryAddress)
    else:
        setAddress(rec, '', primaryAddress)
    rec['address_type'] = primaryAddress['type']


def setBillingAddress(rec, paymentInfo):
    if "ccInfo" in paymentInfo and "address" in paymentInfo["ccInfo"]:
        setAddress(rec, "billing_", paymentInfo["ccInfo"]["address"])
    elif "address" in paymentInfo:
        setAddress(rec, "billing_", paymentInfo["address"])
    else:
        rec['address_billing_line1'] = ""
        rec['address_billing_line2'] = ""
        rec['address_billing_line3'] = ""
        rec['address_billing_ine4'] = ""
        rec['address_billing_city'] = ""
        rec['address_billing_state_province'] = ""
        rec['address_billing_postal_code'] = ""
        rec['address_billing_country'] = ""


def setAddress(rec, addType, address):
    rec["address_" + addType + "line1"] = address["line1"]
    rec["address_" + addType + "line2"] = address["line2"]
    rec["address_" + addType + "line3"] = address["line3"]
    rec["address_" + addType + "line4"] = address["line4"]
    rec["address_" + addType + "postal_codes"] = address["postalCode"]
    rec["address_" + addType + "city"] = address["city"]
    if address['stateProvince']:
        rec["address_" + addType +
            "_state_province"] = address['stateProvince']['code']
    if address['country']:
        rec["address_" + addType + "_country"] = address['country']['code']


def setPaymentInfo(rec, paymentInfo):
    rec["payment_type"] = paymentInfo["paymentType"]
    if "ccInfo" in paymentInfo:
        rec["cc_token"] = paymentInfo["ccInfo"]["token"]
        rec["cc_type"] = paymentInfo["ccInfo"]["cardType"]
        rec["cc_exp"] = paymentInfo["ccInfo"]["cardExp"]
        rec["cc_cvv"] = paymentInfo["ccInfo"]["cardCvv"]
        rec["cc_name"] = paymentInfo["ccInfo"]["name"]


def getExternalId(externalIds, system):
    for exId in externalIds:
        if exId["originatingSystem"] == system:
            return exId["id"]
    return ""


def buildSerializedLists(list, key, sep):
    parts = []
    for item in list:
        if key in item:
            item[key].replace(sep, "")
            parts.append(item[key])
    return sep.join(parts)


def replaceAndjoin(array, sep):
    parts = []
    for item in array:
        item.replace(sep, "")
        parts.append(item)
    return sep.join(parts)
