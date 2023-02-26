import traceback


def buildObjectRecord(rec):
    try:
        newRec = {}
        newRec["model_version"] = rec["modelVersion"]
        newRec["event_type"] = rec["event_type"]
        newRec["event_timestamp"] = rec["event_timestamp"]
        newRec["arrival_timestamp"] = rec["arrival_timestamp"]
        newRec["event_version"] = rec["event_version"]
        newRec["user_agent"] = rec["device"]["useragent"]
        if "session" in rec:
            newRec["session_id"] = rec["session"]["id"]
        newRec["custom_event_name"] = getAttr(
            rec["attributes"], "custom_event_name")
        newRec["destination"] = getAttr(rec["attributes"], "destination")
        newRec["num_guests"] = getAttr(rec["attributes"], "num_guests")
        newRec["checkin_date"] = getAttr(rec["attributes"], "checkin_date")
        newRec["product_quantities"] = getAttr(rec["attributes"], "quantity")
        newRec["products_prices"] = getAttr(
            rec["attributes"], "products_price")
        newRec["products"] = getAttr(rec["attributes"], "products")
        newRec["fare_class"] = getAttr(rec["attributes"], "fare_class")
        newRec["fare_type"] = getAttr(rec["attributes"], "fare_type")
        newRec["flight_segments_departure_date_time"] = getAttr(
            rec["attributes"], "flight_segments_departure_date_time")
        newRec["flight_numbers"] = getAttr(
            rec["attributes"], "flight_segment_sku")
        newRec["flight_market"] = getAttr(
            rec["attributes"], "flight_market")
        newRec["flight_type"] = getAttr(
            rec["attributes"], "flight_type")
        newRec["origin_date"] = getAttr(
            rec["attributes"], "origin_date")
        newRec["origin_date_time"] = getAttr(
            rec["attributes"], "origin_date_time")
        newRec["return_date_time"] = getAttr(
            rec["attributes"], "returning_date_time")
        newRec["return_date"] = getAttr(
            rec["attributes"], "returning_date")
        newRec["return_date"] = getAttr(
            rec["attributes"], "returning_date")
        newRec["return_flight_route"] = getAttr(
            rec["attributes"], "returning_flight_route")
        newRec["num_pax_adults"] = getAttr(
            rec["attributes"], "num_Pax_Adults")
        newRec["num_pax_inf"] = getAttr(
            rec["attributes"], "num_pax_inf")
        newRec["num_pax_children"] = getAttr(
            rec["attributes"], "num_Pax_Children")
        newRec["pax_type"] = getAttr(
            rec["attributes"], "pax_type")
        newRec["pax_type"] = getAttr(
            rec["attributes"], "pax_type")
        newRec["total_passengers"] = getAttr(
            rec["attributes"], "total_passengers")
        newRec["num_nights"] = getAttr(rec["attributes"], "num_nights")
        newRec["room_type"] = getAttr(rec["attributes"], "room_type")
        newRec["rate_plan"] = getAttr(rec["attributes"], "rate_plan")
        newRec["checkin_date"] = getAttr(rec["attributes"], "checkin_date")
        newRec["checkout_date"] = getAttr(rec["attributes"], "checkout_date")
        newRec["hotel_code"] = getAttr(rec["attributes"], "hotel_code")
        newRec["hotel_code_list"] = getAttr(
            rec["attributes"], "hotel_code_list")
        newRec["num_nights"] = getAttr(rec["attributes"], "num_nights")
        newRec["aws_account_id"] = rec["awsAccountId"]
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        newRec["error"] = str(e)
    return newRec


def getAttr(attributes, name):
    selectedAtt = {}
    for att in attributes:
        if att["name"] == name:
            selectedAtt = att
            break
    # loop through each attribute in list once for better performance
    # obj = next(x for x in attributes if x["name"] == name)
    # if obj is None:
    #    return ""
    if "type" in selectedAtt:
        if selectedAtt["type"] == "string":
            return selectedAtt["stringValue"]
        elif selectedAtt["type"] == "strings":
            return selectedAtt["stringValues"]
        elif selectedAtt["type"] == "number":
            try:
                return float(selectedAtt["numValue"])
            except Exception as e:
                return 0.0
        elif selectedAtt["type"] == "numbers":
            arr = []
            for val in selectedAtt["numValues"]:
                try:
                    arr.append(float(val))
                except Exception as e:
                    arr.append(0.0)
            return arr
    return ""
