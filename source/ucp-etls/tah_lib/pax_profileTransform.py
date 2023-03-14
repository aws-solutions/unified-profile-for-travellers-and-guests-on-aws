import uuid
import traceback
from datetime import datetime

from tah_lib.common import setPrimaryEmail, setPrimaryPhone, setPrimaryAddress, setTravellerId, setTimestamp


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
            'model_version': rec.get('modelVersion', ''),
            'id': rec.get('id', ''),
            'last_updated': rec.get('lastUpdatedOn', ''),
            'created_on': rec.get('createdOn', ''),
            'last_updated_by': rec.get('lastUpdatedBy', ''),
            'created_by': rec.get('createdBy', ''),
            'honorific': rec.get('honorific', ''),
            'first_name': rec.get('firstName', ''),
            'middle_name': rec.get('middleName', ''),
            'last_name': rec.get('lastName', ''),
            'gender': rec.get('gender', ''),
            'pronoun': rec.get('pronoun', ''),
            'date_of_birth': rec.get('dateOfBirth', ''),
            'job_title': rec.get('jobTitle', ''),
            'company': rec.get('parentCompany', ''),
        }
        profileRec["nationality_code"] = rec.get(
            'nationality', {}).get("code", "")
        profileRec["nationality_name"] = rec.get(
            'nationality', {}).get("name", "")
        profileRec["language_code"] = rec.get('language', {}).get("code", "")
        profileRec["language_name"] = rec.get('language', {}).get("name", "")

        # Set primary option for phone/email/address
        setPrimaryEmail(profileRec, rec.get('emails', []))
        setPrimaryPhone(profileRec, rec.get('phones', []))
        setTravellerId(profileRec, rec, cid)
        setTimestamp(profileRec)
        setPrimaryAddress(profileRec, rec.get('addresses', []))

        # Email Addresses
        for email in rec['emails']:
            historicalEmail = {
                'model_version': rec.get("modelVersion", ""),
                'object_type': "email_history",
                'last_updated': rec.get("lastUpdatedOn", ""),
                'last_updated_by': rec.get("lastUpdatedBy", ""),
                'address': email.get("address", ""),
                'type': email.get("type", ""),
            }
            setTravellerId(historicalEmail, rec, cid)
            setTimestamp(historicalEmail)
            emailRecs.append(historicalEmail)

        # Phone Numbers
        for phone in rec['phones']:
            historicalPhone = {
                'model_version': rec.get("modelVersion", ""),
                'object_type': "phone_history",
                'last_updated': rec.get("lastUpdatedOn", ""),
                'last_updated_by': rec.get("lastUpdatedBy", ""),
                'number': phone.get("number", ""),
                'country_code': phone.get("countryCode", ""),
                'type': phone.get("type", ""),
            }
            setTravellerId(historicalPhone, rec, cid)
            setTimestamp(historicalPhone)
            phoneRecs.append(historicalPhone)

        # Loyalty Programs
        for loyalty in rec['loyaltyPrograms']:
            loyaltyRec = {
                'model_version': rec.get("modelVersion", ""),
                'object_type': "air_loyalty",
                'last_updated': rec.get("lastUpdatedOn", ""),
                'last_updated_by': rec.get("lastUpdatedBy", ""),
                "id": loyalty.get("id", ""),
                "program_name": loyalty.get("programName", ""),
                "miles": loyalty.get("miles", ""),
                "miles_to_next_level": loyalty.get("milesToNextLevel", ""),
                "level": loyalty.get("level", ""),
                "joined": loyalty.get("joined", "")
            }
            setTravellerId(loyaltyRec, rec, cid)
            setTimestamp(loyaltyRec)
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
