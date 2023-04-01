import traceback
import uuid
from datetime import datetime
from tah_lib.common import replaceAndjoin, setTimestamp, parseNumber, logErrorToSQS
from tah_lib.common import FIELD_NAME_TRAVELLER_ID, FIELD_NAME_LAST_UPDATED

BIZ_OBJECT_TYPE = "clicstream"


def buildObjectRecord(rec, errQueueUrl):
    try:
        newRec = {}
        newRec["model_version"] = rec.get("modelVersion", "")
        newRec["object_type"] = "clickstream_event"
        newRec["event_type"] = rec.get("event_type", "")
        newRec["event_timestamp"] = rec.get("event_timestamp", "")
        newRec["arrival_timestamp"] = rec.get("arrival_timestamp", "")
        newRec["event_version"] = rec.get("event_version", "")
        newRec["user_agent"] = rec.get("device", {}).get("useragent", "")
        newRec["session_id"] = rec.get("session", {}).get("id", "")
        newRec[FIELD_NAME_TRAVELLER_ID] = rec.get("userId", "")
        newRec["custom_event_name"] = getAttr(
            rec.get("attributes", []), "custom_event_name")
        newRec["destination"] = getAttr(
            rec.get("attributes", []), "destination")
        newRec["num_guests"] = getAttr(rec.get("attributes", []), "num_guests")
        newRec["checkin_date"] = getAttr(
            rec.get("attributes", []), "checkin_date")
        newRec["product_quantities"] = getAttr(
            rec.get("attributes", []), "quantity")
        newRec["products_prices"] = getAttr(
            rec.get("attributes", []), "products_price")
        newRec["products"] = getAttr(rec.get("attributes", []), "products")
        newRec["fare_class"] = getAttr(rec.get("attributes", []), "fare_class")
        newRec["fare_type"] = getAttr(rec.get("attributes", []), "fare_type")
        newRec["flight_segments_departure_date_time"] = getAttr(
            rec.get("attributes", []), "flight_segments_departure_date_time")
        newRec["flight_numbers"] = getAttr(
            rec.get("attributes", []), "flight_segment_sku")
        newRec["flight_market"] = getAttr(
            rec.get("attributes", []), "flight_market")
        newRec["flight_type"] = getAttr(
            rec.get("attributes", []), "flight_type")
        newRec["origin_date"] = getAttr(
            rec.get("attributes", []), "origin_date")
        newRec["origin_date_time"] = getAttr(
            rec.get("attributes", []), "origin_date_time")
        newRec["return_date_time"] = getAttr(
            rec.get("attributes", []), "returning_date_time")
        newRec["return_date"] = getAttr(
            rec.get("attributes", []), "returning_date")
        newRec["return_date"] = getAttr(
            rec.get("attributes", []), "returning_date")
        newRec["return_flight_route"] = getAttr(
            rec.get("attributes", []), "returning_flight_route")
        newRec["num_pax_adults"] = getAttr(
            rec.get("attributes", []), "num_Pax_Adults")
        newRec["num_pax_inf"] = getAttr(
            rec.get("attributes", []), "num_pax_inf")
        newRec["num_pax_children"] = getAttr(
            rec.get("attributes", []), "num_Pax_Children")
        newRec["pax_type"] = getAttr(
            rec.get("attributes", []), "pax_type")
        newRec["pax_type"] = getAttr(
            rec.get("attributes", []), "pax_type")
        newRec["total_passengers"] = getAttr(
            rec.get("attributes", []), "total_passengers")
        newRec["num_nights"] = getAttr(rec.get("attributes", []), "num_nights")
        newRec["room_type"] = getAttr(rec.get("attributes", []), "room_type")
        newRec["rate_plan"] = getAttr(rec.get("attributes", []), "rate_plan")
        newRec["checkin_date"] = getAttr(
            rec.get("attributes", []), "checkin_date")
        newRec["checkout_date"] = getAttr(
            rec.get("attributes", []), "checkout_date")
        newRec["hotel_code"] = getAttr(rec.get("attributes", []), "hotel_code")
        newRec["hotel_code_list"] = getAttr(
            rec.get("attributes", []), "hotel_code_list")
        newRec["num_nights"] = getAttr(rec.get("attributes", []), "num_nights")
        newRec["aws_account_id"] = rec.get("awsAccountId", "")
        # session ID is a mandatory field
        if newRec["session_id"] == "":
            newRec["session_id"] = str(uuid.uuid1(node=None, clock_seq=None))
        # if there is no usser Id provided by the customer, we use the session ID to group all teh events by session
        if newRec[FIELD_NAME_TRAVELLER_ID] == "":
            newRec[FIELD_NAME_TRAVELLER_ID] = newRec["session_id"]
        newRec[FIELD_NAME_LAST_UPDATED] = newRec["event_timestamp"]
        # If no timesstamp is provided we set a timestamp for the object
        setTimestamp(newRec)
        # cleaning up None data
        toDelete = []
        for key in newRec:
            if newRec[key] is None:
                toDelete.append(key)
        for key in toDelete:
            del newRec[key]
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        logErrorToSQS(e, rec, errQueueUrl, BIZ_OBJECT_TYPE)
    return newRec


def getAttr(attributes, name):
    selectedAtt = {}
    for att in attributes:
        if att["name"] == name:
            selectedAtt = att
            break
    if "type" in selectedAtt:
        if selectedAtt["type"] == "string":
            return str(selectedAtt["stringValue"])
        elif selectedAtt["type"] == "strings":
            arr = []
            for val in selectedAtt["stringValues"]:
                arr.append(str(val))
            return replaceAndjoin(arr, "|")
        elif selectedAtt["type"] == "number":
            return parseNumber(selectedAtt["numValue"])
        elif selectedAtt["type"] == "numbers":
            arr = []
            for val in selectedAtt["numValues"]:
                arr.append(str(val))
            return replaceAndjoin(arr, "|")
    return None
