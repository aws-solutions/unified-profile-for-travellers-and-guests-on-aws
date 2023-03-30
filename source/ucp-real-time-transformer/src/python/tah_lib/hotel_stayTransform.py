import uuid
import traceback
from datetime import datetime
from tah_lib.common import setTimestamp, parseNumber, logErrorToSQS

BIZ_OBJECT_TYPE = "hotel_stay"


def buildObjectRecord(rec, errQueueUrl):
    newRecs = []
    try:
        if len(rec.get("revenue", [])) == 0:
            print("Invalid format Exception")
            raise Exception("Hotel stay must have at least 1 revenue item ")

        for item in rec.get("revenue", []):

            newRec = {}
            newRec["model_version"] = rec.get("modelVersion", "")
            newRec["object_type"] = "hotel_stay_revenue"
            newRec["last_updated"] = rec.get("lastUpdatedOn", "")
            newRec["created_on"] = rec.get("createdOn", "")
            newRec["last_updated_by"] = rec.get("lastUpdatedBy", "")
            newRec["created_by"] = rec.get("createdBy", "")
            newRec["id"] = rec.get("id", "")
            newRec["booking_id"] = rec.get("bookingId", "")
            newRec["traveller_id"] = rec.get("guestId", "")
            newRec["currency_code"] = rec.get("currency", {}).get("code", "")
            newRec["currency_name"] = rec.get("currency", {}).get("name", "")
            newRec["currency_symbol"] = rec.get(
                "currency", {}).get("symbol", "")
            newRec["first_name"] = rec.get("firstName", "")
            newRec["last_name"] = rec.get("lastName", "")
            newRec["email"] = rec.get("email", "")
            newRec["phone"] = rec.get("phone", "")
            newRec["start_date"] = rec.get("startDate", "")
            newRec["hotel_code"] = rec.get("hotelCode", "")
            newRec["type"] = item.get("type", "")
            newRec["description"] = item.get("description", "")
            newRec["amount"] = parseNumber(item.get("amount", ""))
            newRec["date"] = item.get("date", "")
            if newRec["traveller_id"] == "":
                newRec["traveller_id"] = str(
                    uuid.uuid1(node=None, clock_seq=None))
            setTimestamp(newRec)
            newRecs.append(newRec)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        logErrorToSQS(e, rec, errQueueUrl, BIZ_OBJECT_TYPE)
    return {"hotel_stay_revenue_items": newRecs}
