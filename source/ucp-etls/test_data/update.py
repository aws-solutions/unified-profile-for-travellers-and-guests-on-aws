import json
from tah_lib.air_bookingTransform import buildObjectRecord as air_bookingTransform
from tah_lib.hotel_bookingTransform import buildObjectRecord as hotel_bookingTransform
from tah_lib.pax_profileTransform import buildObjectRecord as pax_profileTransform
from tah_lib.guest_profileTransform import buildObjectRecord as guest_profileTransform
from tah_lib.hotel_stayTransform import buildObjectRecord as hotel_stayTransform
from tah_lib.clickstreamTransform import buildObjectRecord as clickstreamTransform


transfoms = {
    "air_booking": air_bookingTransform,
    "clickstream": clickstreamTransform,
    "guest_profile": guest_profileTransform,
    "hotel_booking": hotel_bookingTransform,
    "hotel_stay": hotel_stayTransform,
    "pax_profile": pax_profileTransform
}


def loadTestRecord(data_type, rec):
    path = 'test_data/' + data_type + "/" + rec + '.json'
    f = open(path)
    data = json.load(f)
    f.close()
    return transfoms[data_type](data, "")


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


for data_path in ["air_booking", "clickstream", "guest_profile", "hotel_booking", "hotel_stay", "pax_profile"]:
    for rec in ["data1", "data2"]:
        print("Updating expected test result for ", data_path, "/", rec)
        actual = loadTestRecord(data_path, rec)
        print(actual)
        with open('test_data/' + data_path + "/" + rec + "_expected.json", "w") as outfile:
            json.dump(actual, outfile)
