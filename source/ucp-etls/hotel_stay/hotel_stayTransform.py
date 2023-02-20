def buildObjectRecord(rec):
    newRecs = []
    for item in rec["revenue"]:
        try:
            newRec = {}
            newRec["modelVersion"] = rec["modelVersion"]
            newRec["id"] = rec["id"]
            newRec["bookingId"] = rec["bookingId"]
            newRec["guestId"] = rec["guestId"]
            newRec["lastUpdatedOn"] = rec["lastUpdatedOn"]
            newRec["createdOn"] = rec["createdOn"]
            newRec["lastUpdatedBy"] = rec["lastUpdatedBy"]
            newRec["createdBy"] = rec["createdBy"]
            newRec["currency_code"] = rec["currency"]["code"]
            newRec["currency_name"] = rec["currency"]["name"]
            newRec["currency_symbol"] = rec["currency"]["symbol"]
            newRec["firstName"] = rec["firstName"]
            newRec["lastName"] = rec["lastName"]
            newRec["email"] = rec["email"]
            newRec["phone"] = rec["phone"]
            newRec["startDate"] = rec["startDate"]
            newRec["hotelCode"] = rec["hotelCode"]
            newRec["type"] = item["type"]
            newRec["description"] = item["description"]
            newRec["currency_code"] = rec["currency"]["code"]
            newRec["currency_name"] = rec["currency"]["name"]
            newRec["currency_symbol"] = rec["currency"]["symbol"]
            newRec["amount"] = item["amount"]
            newRec["date"] = item["date"]
            newRecs.append(newRec)
        except Exception as e:
            newRec["error"] = str(e)
        newRecs.append(newRec)

    return {"data": newRecs}