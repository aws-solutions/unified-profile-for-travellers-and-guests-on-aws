
from datetime import datetime
import json
import boto3

SUPPORTED_ADRESS_TYPES = ["", "billing", "business", "mailing"]
SUPPORTED_ADRESS_FIELDS = ["line1", "line2", "line3",
                           "line4", "city", "state_province", "postal_code", "country"]
FIELD_NAME_TRAVELLER_ID = "traveller_id"
FIELD_NAME_LAST_UPDATED = "last_updated"
# field added to be used in S3 partitining which requires removing some special char
FIELD_NAME_LAST_UPDATED_PARTITION = "last_updated_partition"
# 2021-12-09T07:59:14.873255297Z
DATETIME_FORMAT = "%Y-%m-%dT%H:%M:%S.%f000Z"

SQS_MES_ATTR_UCP_ERROR_TYPE = "UcpErrorType"
SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME = "BusinessObjectTypeName"
SQS_MES_ATTR_MESSAGE = "Message"


# resilient float parsing to be used to avoid cases where both int and float would come in for the same value which would lead to
# CSV data source does not support struct<int:int,double:double> data type.
# pyspark.sql.utils.AnalysisException: CSV data source does not support struct<int:int,double:double> data type.
def parseNumber(num):
    try:
        return float(num)
    except:
        return 0.0


def setTravellerId(rec, guestBizObject, uid):
    rec[FIELD_NAME_TRAVELLER_ID] = guestBizObject.get("id", "")
    if rec[FIELD_NAME_TRAVELLER_ID] == "":
        rec[FIELD_NAME_TRAVELLER_ID] = uid


# seting default timestamp if does not exist
def setTimestamp(rec):
    if rec.get(FIELD_NAME_LAST_UPDATED, "") == "":
        rec[FIELD_NAME_LAST_UPDATED] = datetime.now().strftime(DATETIME_FORMAT)
    ts = rec[FIELD_NAME_LAST_UPDATED].replace("-", "/")
    ts = ts.replace("T", "/")
    ts = ts.replace(":", "/")
    ts = ts.replace(".", "/")
    rec[FIELD_NAME_LAST_UPDATED_PARTITION] = ts

# this function set the primary email according to the following logic
# 1- if one or multiple emails are flagged as Primary, we choose the first one
# 2- if not we choose the first email in the list.
# 3- then we use the email type to choose whether we set the biz_email key or 'email' key
# 3- if the list is empty we set both keys to empty strings to allow creation of the column in the CSV


def setPrimaryEmail(rec, emails):
    rec['email_business'] = ""
    rec['email'] = ""
    rec['email_type'] = ""
    if len(emails) == 0:
        return
    primaryEmail = emails[0]
    for email in emails:
        if email.get('primary', False) == True:
            primaryEmail = email
            break
    if primaryEmail.get('type', "") == 'business':
        rec['email_business'] = primaryEmail.get('address', "")
    else:
        rec['email'] = primaryEmail.get('address', "")
    # we keep the email type in case the value is outside the enum
    rec['email_type'] = primaryEmail.get('type', "")


# this function sets the primary phone using the same logic than the email function
# we have 4 types of phones: default, business, mobile and home
def setPrimaryPhone(rec, phones):
    rec['phone'] = ""
    rec['phone_home'] = ""
    rec['phone_mobile'] = ""
    rec['phone_business'] = ""
    if len(phones) == 0:
        return
    primaryPhone = phones[0]
    for phone in phones:
        if phone.get('primary', False) == True:
            primaryPhone = phone
            break
    if primaryPhone['type'] == 'business':
        rec['phone_business'] = primaryPhone.get('number', "")
    elif primaryPhone['type'] == 'home':
        rec['phone_home'] = primaryPhone.get('number', "")
    elif primaryPhone['type'] == 'mobile':
        rec['phone_business'] = primaryPhone.get('number', "")
    else:
        rec['phone'] = primaryPhone.get('number', "")
    rec['phone_type'] = primaryPhone.get('type', "")


# this function sets the primary address  using the same logic than the email function
def setPrimaryAddress(rec, addresses):
    # initializing all address types if they do not exist on the record
    for addrType in SUPPORTED_ADRESS_TYPES:
        if addrType != "":
            addrType += "_"
        for field in SUPPORTED_ADRESS_FIELDS:
            if 'address_'+addrType+field not in rec:
                rec['address_'+addrType+field] = ""
    if len(addresses) == 0:
        return
    primaryAddress = addresses[0]
    for address in addresses:
        if address['primary'] == True:
            primaryAddress = address
            break
    if primaryAddress.get('type', "") == 'business':
        setAddress(rec, 'business', primaryAddress)
    elif primaryAddress.get('type', "") == 'shipping':
        setAddress(rec, 'shipping', primaryAddress)
    elif primaryAddress.get('type', "") == 'mailing':
        setAddress(rec, 'mailing', primaryAddress)
    else:
        setAddress(rec, '', primaryAddress)
    rec['address_type'] = primaryAddress.get('type', "")


def setBillingAddress(rec, paymentInfo):
    if "ccInfo" in paymentInfo and "address" in paymentInfo["ccInfo"]:
        setAddress(rec, "billing", paymentInfo["ccInfo"]["address"])
    elif "address" in paymentInfo:
        setAddress(rec, "billing", paymentInfo["address"])
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
    if addType not in SUPPORTED_ADRESS_TYPES:
        return
    if address is None:
        address = {}
    if addType != "":
        addType += "_"
    rec["address_" + addType + "line1"] = address.get("line1", "")
    rec["address_" + addType + "line2"] = address.get("line2", "")
    rec["address_" + addType + "line3"] = address.get("line3", "")
    rec["address_" + addType + "line4"] = address.get("line4", "")
    rec["address_" + addType + "postal_code"] = address.get("postalCode", "")
    rec["address_" + addType + "city"] = address.get("city", "")
    rec["address_" + addType +
        "state_province"] = address.get('stateProvince', {}).get('code', "")
    rec["address_" + addType +
        "country"] = address.get('country', {}).get('code', "")


def setPaymentInfo(rec, paymentInfo):
    rec["payment_type"] = paymentInfo.get("paymentType", "")
    rec["cc_token"] = paymentInfo.get("ccInfo", {}).get("token")
    rec["cc_type"] = paymentInfo.get("ccInfo", {}).get("cardType")
    rec["cc_exp"] = paymentInfo.get("ccInfo", {}).get("cardExp")
    rec["cc_cvv"] = paymentInfo.get("ccInfo", {}).get("cardCvv")
    rec["cc_name"] = paymentInfo.get("ccInfo", {}).get("name")


def getExternalId(externalIds, system):
    if externalIds is None:
        return ""
    for exId in externalIds:
        if exId.get("originatingSystem", "") == system:
            return exId.get("id", "")
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


def noneToList(val):
    if val is None:
        return []
    return val


def logErrorToSQS(e: Exception, rec: dict, errQueueUrl: str, bizObjType: str):
    client = boto3.client('sqs')
    client.send_message(
        QueueUrl=errQueueUrl,
        MessageBody=json.Dump(rec),
        MessageAttributes={
            SQS_MES_ATTR_UCP_ERROR_TYPE: {
                'StringValue': 'transformer_exception',
                'DataType': 'string'
            },
            SQS_MES_ATTR_BUSINESS_OBJECT_TYPE_NAME: {
                'StringValue': bizObjType,
                'DataType': 'string'
            },
            SQS_MES_ATTR_MESSAGE: {
                'StringValue': str(e),
                'DataType': 'string'
            }
        },
    )
