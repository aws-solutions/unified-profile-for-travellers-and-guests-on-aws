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
        profileRec = {
            'object_type': 'pax_profile',
            'model_version': rec.get('modelVersion', ""),
            'last_updated_on': rec.get('lastUpdatedOn', ""),
            'created_on': rec.get('createdOn', ""),
            'last_updated_by': rec.get('lastUpdatedBy', ""),
            'created_by': rec.get('createdBy', ""),
            'honorific': rec.get('honorific', ""),
            'first_name': rec.get('firstName', ""),
            'middle_name': rec.get('middleName', ""),
            'last_name': rec.get('lastName', ""),
            'gender': rec.get('gender', ""),
            'pronoun': rec.get('pronoun', ""),
            'date_of_birth': rec.get('dateOfBirth', ""),
            'job_title': rec.get('jobTitle', ""),
            'company': rec.get('parentCompany', ""),
        }
        if "nationality" in rec:
            profileRec["nationality_code"] = rec.get(
                'nationality', {}).get("code", "")
            profileRec["nationality_name"] = rec.get(
                'nationality', {}).get("name", "")
        if "language" in rec:
            profileRec["language_code"] = rec.get(
                'language', {}).get("code", "")
            profileRec["language_name"] = rec.get(
                'language', {}).get("name", "")

        # Set primary option for phone/email/address
        setPrimaryEmail(profileRec, rec.get('emails', []))
        setPrimaryPhone(profileRec, rec.get('phones', []))
        setPrimaryAddress(profileRec, rec.get('addresses', []))
        setTravellerId(profileRec, rec, cid)

        # Email Addresses
        for email in rec.get('emails', []):
            historicalEmail = {
                'object_type': 'email_history',
                'model_version': rec.get('modelVersion', ""),
                'last_updated': rec.get('lastUpdatedOn', ""),
                'last_updated_by': rec.get('lastUpdatedBy', ""),
                'address': email.get('address', ""),
                'type': email.get('type', ""),
            }
            setTravellerId(historicalEmail, rec, cid)
            emailRecs.append(historicalEmail)

        # Phone Numbers
        for phone in rec.get('phones', []):
            historicalPhone = {
                'object_type': 'phone_history',
                'model_version': rec.get('modelVersion', ""),
                'last_updated': rec.get('lastUpdatedOn', ""),
                'last_updated_by': rec.get('lastUpdatedBy', ""),
                'number': phone.get('number', ""),
                'country_code': phone.get('countryCode', ""),
                'type': phone.get('type', ""),
            }
            setTravellerId(historicalPhone, rec, cid)
            phoneRecs.append(historicalPhone)

        # Loyalty Programs
        for loyalty in rec.get('loyaltyPrograms', []):
            loyaltyRec = {
                'object_type': 'air_loyalty',
                'model_version': rec.get('modelVersion', ""),
                'last_updated': rec.get('lastUpdatedOn', ""),
                'last_updated_by': rec.get('lastUpdatedBy', ""),
                'id': loyalty.get('id', ""),
                'program_name': loyalty.get('programName', ""),
                'points': loyalty.get('points', ""),
                'units': loyalty.get('pointUnit', ""),
                'points_to_next_level': loyalty.get('pointsToNextLevel', ""),
                'level': loyalty.get('level', ""),
                'joined': loyalty.get('joined'),
            }
            setTravellerId(loyaltyRec, rec, cid)
            loyaltyRecs.append(loyaltyRec)

        profileRecs.append(profileRec)
    except Exception as e:
        traceback_info = traceback.format_exc()
        print(traceback_info)
        profileRecs.append({"error": str(e), "trace": traceback_info})
    return {
        'guest_profile_recs': profileRecs,
        'common_email_recs': emailRecs,
        'common_phone_recs': phoneRecs,
        'hotel_loyalty_recs': loyaltyRecs,
    }
