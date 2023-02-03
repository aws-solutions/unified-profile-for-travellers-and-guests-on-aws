def buildObjectRecord(rec):
    try:
        newRec = {}
        newRec["modelVersion"] = rec["modelVersion"]
        newRec["event_type"] = rec["event_type"]
        newRec["event_timestamp"] = rec["event_timestamp"]
        newRec["arrival_timestamp"] = rec["arrival_timestamp"]
        newRec["event_version"] = rec["event_version"]
        # newRec["application"] = rec[""]
        # newRec["client"] = rec[""]
        newRec["useragent"] = rec["device"]["useragent"]
        newRec["session_id"] = rec["session"]["id"]
        newRec["SelectNNights"] = float(getAttr(rec["attributes"], "num_nights"))
        newRec["customer_event_name"] = getAttr(rec["attributes"], "customer_event_name")
        newRec["SelectDestination"] = getAttr(rec["attributes"], "destination")
        newRec["SelectNGuests"] = float(getAttr(rec["attributes"], "num_guests"))
        newRec["SelectStartDate"] = getAttr(rec["attributes"], "checkin_date")
        newRec["Quantity"] = float(getAttr(rec["attributes"], "quantity"))
        # newRec["endpoint"] = rec[""]
        newRec["awsAccountId"] = rec["awsAccountId"]
        # newRec[""] = rec[""]
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