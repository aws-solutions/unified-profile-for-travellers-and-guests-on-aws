def buildObjectRecord(rec):
    try:
        newRec = {"data": []}
        for pax in rec["passengerInfo"]["passengers"]:
            airBookingRec = {
                'model_version': rec["modelVersion"],
                'object_type': "air_booking",
                'last_updated': rec["lastUpdatedOn"],
                'last_updated_by': rec["lastUpdatedBy"],
                'from': rec["itinerary"]["from"],
                'to': rec["itinerary"]["to"],
                'departure_date': rec["itinerary"]["departureDate"],
                'departure_time': rec["itinerary"]["departureTime"],
                'channel': rec["creationChannelId"],
                'status': rec["status"],
                "honorific": "Ms",
                "first_name": pax["firstName"],
                "middle_name": pax["middleName"],
                "last_name": pax["lastName"],
                "gender": pax["gender"],
                "pronoun": pax["pronoun"],
                "date_of_birth": pax["dateOfBirth"],
                "language": pax["language"]["code"],
                "nationality": pax["nationality"]["code"],
                "job_title": pax["jobTitle"],
                "company": pax["parentCompany"],
                "price": rec["price"]["grandTotal"],
                "mail_address_type": pax["addresses"][0]["type"],
                "mail_address_line1": pax["addresses"][0]["line1"],
                "mail_address_line2": pax["addresses"][0]["line2"],
                "mail_address_line3": pax["addresses"][0]["line3"],
                "mail_address_line4": pax["addresses"][0]["line4"],
                "mail_address_city": pax["addresses"][0]["city"],
                "mail_address_stateProvince": pax["addresses"][0]["stateProvince"]["code"],
                "mail_address_postalCode": pax["addresses"][0]["postalCode"],
                "mail_address_country": pax["addresses"][0]["country"]["code"],
                "bill_address_type": rec["paymentInformation"]["ccInfo"]["address"]["type"],
                "bill_address_line1": rec["paymentInformation"]["ccInfo"]["address"]["line1"],
                "bil_address_line2": rec["paymentInformation"]["ccInfo"]["address"]["line2"],
                "bil_address_line3": rec["paymentInformation"]["ccInfo"]["address"]["line3"],
                "bil_address_line4": rec["paymentInformation"]["ccInfo"]["address"]["line4"],
                "bil_address_city": rec["paymentInformation"]["ccInfo"]["address"]["city"],
                "bil_address_stateProvince": rec["paymentInformation"]["ccInfo"]["address"]["stateProvince"]["code"],
                "bil_address_postalCode": rec["paymentInformation"]["ccInfo"]["address"]["postalCode"],
                "bil_address_country": rec["paymentInformation"]["ccInfo"]["address"]["country"]["code"],
                "payment_type": rec["paymentInformation"]["paymentType"],
                "cc_number": rec["paymentInformation"]["paymentType"],
                "cc_token": rec["paymentInformation"]["ccInfo"]["token"],
                "cc_type": rec["paymentInformation"]["ccInfo"]["cardType"],
                "cc_exp": rec["paymentInformation"]["ccInfo"]["cardExp"],
                "cc_cvv": rec["paymentInformation"]["ccInfo"]["cardCvv"],
                "cc_name": rec["paymentInformation"]["ccInfo"]["name"],
            }
            if pax["emails"][0]["type"] == "business":
                airBookingRec['biz_email'] = pax["emails"][0]["address"]
            else:
                airBookingRec['email'] = pax["emails"][0]["address"]
            if pax["phones"][0]["type"] == "business":
                airBookingRec['biz_phone'] = pax["phones"][0]["number"]
            else:
                airBookingRec['phone'] = pax["phones"][0]["number"]

            for email in pax["emails"]:
                historicalEmail = {
                    'model_version': rec["modelVersion"],
                    'object_type': "email_history",
                    'last_updated': rec["lastUpdatedOn"],
                    'last_updated_by': rec["lastUpdatedBy"],
                    'address': email["address"],
                    'type': email["type"],
                }
                newRec["data"].append(historicalEmail)
            for phone in pax["phones"]:
                historicalPhone = {
                    'model_version': rec["modelVersion"],
                    'object_type': "phone_history",
                    'last_updated': rec["lastUpdatedOn"],
                    'last_updated_by': rec["lastUpdatedBy"],
                    'number': phone["number"],
                    'country_code': phone["countryCode"],
                    'type': email["type"],
                }
                newRec["data"].append(historicalPhone)
            for loyalty in pax["loyaltyPrograms"]:
                loyaltyRec = {
                    'model_version': rec["modelVersion"],
                    'object_type': "air_loyalty",
                    'last_updated': rec["lastUpdatedOn"],
                    'last_updated_by': rec["lastUpdatedBy"],
                    "id": loyalty["id"],
                    "program_name": loyalty["programName"],
                    "miles": loyalty["id"],
                    "miles_to_next_level": loyalty["milesToNextLevel"],
                    "level": loyalty["level"],
                    "joined": loyalty["joined"]
                }
                newRec["data"].append(loyaltyRec)

            newRec["data"].append(airBookingRec)
    except Exception as e:
        newRec["error"] = str(e)
        return newRec
    return newRec
