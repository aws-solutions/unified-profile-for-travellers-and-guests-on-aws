def buildObjectRecord(rec):
    try:
        newRec = {'data': []}

        # Passenger Profile
        profileRec = {
            'object_type': 'pax_profile',
            'model_version': rec['modelVersion'],
            'id': rec['id'],
            'last_updated_on': rec['lastUpdatedOn'],
            'created_on': rec['createdOn'],
            'last_updated_by': rec['lastUpdatedBy'],
            'created_by': rec['createdBy'],
            'honorific': rec['honorific'],
            'first_name': rec['firstName'],
            'middle_name': rec['middleName'],
            'last_name': rec['lastName'],
            'gender': rec['gender'],
            'pronoun': rec['pronoun'],
            'date_of_birth': rec['dateOfBirth'],
            'language': rec['language'],
            'nationality': rec['nationality'],
            'job_title': rec['jobTitle'],
            'company': rec['parentCompany'],

            # TODO: decide on strategy to set address
            'mail_address_type': rec['addresses'][0]['type'],
            'mail_address_line1': rec['addresses'][0]['line1'],
            'mail_address_line2': rec['addresses'][0]['line2'],
            'mail_address_line3': rec['addresses'][0]['line3'],
            'mail_address_line4': rec['addresses'][0]['line4'],
            'mail_address_city': rec['addresses'][0]['city'],
            'mail_address_state_province': rec['addresses'][0]['stateProvince']['code'],
            'mail_address_postal_code': rec['addresses'][0]['postalCode'],
            'mail_address_country': rec['addresses'][0]['country']['code'],

            # TODO: decide on how to store multiple loyalty and identity records
        }
        if rec['emails'][0]['type'] == 'business':
            profileRec['biz_email'] = rec['emails'][0]['address']
        else:
            profileRec['email'] = rec['emails'][0]['address']

        number = buildPhoneNumber(rec['phones'][0]['countryCode'], rec['phones'][0]['number'])
        if rec['phones'][0]['type'] == 'business':
            profileRec['biz_phone'] = number
        else:
            profileRec['phone'] = number

        # TODO: decide on adding primary fields
        # Set primary option for phone/email/address
        setPrimaryEmail(profileRec, rec['emails'])

        # Email Addresses
        for email in rec['emails']:
            historicalEmail = {
                'object_type': 'email_history',
                'model_version': rec['modelVersion'],
                'last_updated': rec['lastUpdatedOn'],
                'last_updated_by': rec['lastUpdatedBy'],
                'address': email['address'],
                'type': email['type'],
            }
            newRec['data'].append(historicalEmail)
        
        # Phone Numbers
        for phone in rec['phones']:
            historicalPhone = {
                'object_type': 'phone_history',
                'model_version': rec['modelVersion'],
                'last_updated': rec['lastUpdatedOn'],
                'last_updated_by': rec['lastUpdatedBy'],
                'number': phone['number'],
                'country_code': phone['countryCode'],
                'type': phone['type'],
            }
            newRec['data'].append(historicalPhone)

        # Loyalty Programs
        for loyalty in rec['loyaltyPrograms']:
            loyaltyRec = {
                'object_type': 'air_loyalty',
                'model_version': rec['modelVersion'],
                'last_updated': rec['lastUpdatedOn'],
                'last_updated_by': rec['lastUpdatedBy'],
                'id': loyalty['id'],
                'program_name': loyalty['programName'],
                'miles': loyalty['id'],
                'miles_to_next_level': loyalty['milesToNextLevel'],
                'level': loyalty['level'],
                'joined': loyalty['joined']
            }
            newRec['data'].append(loyaltyRec)

        # Identity Proofs??

        newRec['data'].append(profileRec)
    except Exception as e:
        newRec['error'] = str(e)
        return newRec
    return newRec

def setPrimaryEmail(profileRec, emails):
    for email in emails:
        if email['primary'] == True:
            profileRec['address'] = email['address']
            profileRec['type'] = email['type']
            return

def buildPhoneNumber(countryCode, number):
    return '+' + str(countryCode) + str(number)
    