import uuid
import traceback
from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setTravellerId


def buildObjectRecord(rec):
    try:
        profileRecs = []
        emailRecs = []
        phoneRecs = []
        loyaltyRecs = []
        cid = str(uuid.uuid1(node=None, clock_seq=None))
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
        }

        # Set primary option for phone/email/address
        setPrimaryEmail(profileRec, rec['emails'])
        setPrimaryPhone(profileRec, rec['phones'])
        setPrimaryAddress(profileRec, rec['addresses'])
        setTravellerId(profileRec, rec, cid)

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
            setTravellerId(historicalEmail, rec, cid)
            emailRecs.append(historicalEmail)

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
            setTravellerId(historicalPhone, rec, cid)
            phoneRecs.append(historicalPhone)

        # Loyalty Programs
        for loyalty in rec['loyaltyPrograms']:
            loyaltyRec = {
                'object_type': 'air_loyalty',
                'model_version': rec['modelVersion'],
                'last_updated': rec['lastUpdatedOn'],
                'last_updated_by': rec['lastUpdatedBy'],
                'id': loyalty['id'],
                'program_name': loyalty['programName'],
                'miles': loyalty['miles'],
                'miles_to_next_level': loyalty['milesToNextLevel'],
                'level': loyalty['level'],
                'joined': loyalty['joined']
            }
            setTravellerId(loyaltyRec, rec, cid)
            loyaltyRecs.append(loyaltyRec)

        profileRecs.append(profileRec)
    except Exception as e:
        profileRecs['error'] = str(e)
    return {
        'air_profile_recs': profileRecs,
        'common_email_recs': emailRecs,
        'common_phone_recs': phoneRecs,
        'air_loyalty_recs': loyaltyRecs,
    }
