import uuid


def buildObjectRecord(rec):
    newRecs = []
    for item in rec["revenue"]:
        try:
            newRec = {}
            newRec["model_version"] = rec["modelVersion"]
            newRec["last_updated_on"] = rec["lastUpdatedOn"]
            newRec["created_on"] = rec["createdOn"]
            newRec["last_updated_by"] = rec["lastUpdatedBy"]
            newRec["created_by"] = rec["createdBy"]
            newRec["id"] = rec["id"]
            newRec["bookingId"] = rec["bookingId"]
            newRec["traveller_id"] = rec["guestId"]
            newRec["currency_code"] = rec["currency"]["code"]
            newRec["currency_name"] = rec["currency"]["name"]
            newRec["currency_symbol"] = rec["currency"]["symbol"]
            newRec["first_name"] = rec["firstName"]
            newRec["last_name"] = rec["lastName"]
            newRec["email"] = rec["email"]
            newRec["phone"] = rec["phone"]
            newRec["start_date"] = rec["startDate"]
            newRec["hotel_code"] = rec["hotelCode"]
            newRec["type"] = item["type"]
            newRec["description"] = item["description"]
            if "currency" in rec:
                newRec["currency_code"] = rec["currency"]["code"]
                newRec["currency_name"] = rec["currency"]["name"]
                newRec["currency_symbol"] = rec["currency"]["symbol"]
            newRec["amount"] = item["amount"]
            newRec["date"] = item["date"]

            if newRec["traveller_id"] == "":
                newRec["traveller_id"] = str(
                    uuid.uuid1(node=None, clock_seq=None))

            newRecs.append(newRec)
        except Exception as e:
            newRec["error"] = str(e)
        newRecs.append(newRec)

    return {"hotel_stay_revenue_items": newRecs}
