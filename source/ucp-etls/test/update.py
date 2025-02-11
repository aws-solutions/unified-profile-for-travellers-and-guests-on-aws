# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import json

from tah_lib.air_bookingTransform import \
    build_object_record as air_bookingTransform
from tah_lib.clickstreamTransform import \
    build_object_record as clickstreamTransform
from tah_lib.customer_service_interactionTransform import \
    build_object_record as customer_service_interactionTransform
from tah_lib.guest_profileTransform import \
    build_object_record as guest_profileTransform
from tah_lib.hotel_bookingTransform import \
    build_object_record as hotel_bookingTransform
from tah_lib.hotel_stayTransform import \
    build_object_record as hotel_stayTransform
from tah_lib.pax_profileTransform import \
    build_object_record as pax_profileTransform

test_data_path = '../test_data/'

transfoms = {
    "air_booking": air_bookingTransform,
    "clickstream": clickstreamTransform,
    "guest_profile": guest_profileTransform,
    "hotel_booking": hotel_bookingTransform,
    "hotel_stay": hotel_stayTransform,
    "pax_profile": pax_profileTransform,
    "customer_service_interaction": customer_service_interactionTransform
}


def loadTestRecord(data_type, rec):
    path = test_data_path + data_type + "/" + rec + '.jsonl'
    f = open(path)
    first_line = f.readline().strip()
    data = json.loads(first_line)
    f.close()
    return transfoms[data_type](data, "", "test_tx_id")


def loadExpectedRecord(data_file):
    f = open(data_file)
    data = json.load(f)
    f.close()
    return data


for data_path in ["air_booking",  "guest_profile", "hotel_booking", "hotel_stay", "pax_profile", "clickstream", "customer_service_interaction"]:
    for rec in ["data1", "data2"]:
        print("Updating expected test result for ",
              test_data_path + data_path, "/", rec)
        actual = loadTestRecord(data_path, rec)
        with open(test_data_path + data_path + "/" + rec + "_expected.json", "w") as outfile:
            json.dump(actual, outfile)


customer_test_files = {
    "air_booking": [], 
    "guest_profile": ["customer_1"], 
    "hotel_booking": [ "customer_1"],
     "hotel_stay": [], 
     "pax_profile": [], 
     "clickstream": ["customer_1", "full"], 
     "customer_service_interaction": []
}

def loadTestRecords(data_file, transform, queueUrl=""):
    f = open(data_file)
    #iterate on all lines
    recs = []
    for line in f:
        data = json.loads(line)
        recs.append(transform(data, queueUrl, "test_tx_id"))
    f.close()
    return recs

for data_path in ["air_booking",  "guest_profile", "hotel_booking", "hotel_stay", "pax_profile", "clickstream", "customer_service_interaction"]:
    for rec in customer_test_files[data_path]:
        print("Updating expected test result for ",
              test_data_path + data_path, "/", rec)
        expected_recs = loadTestRecords(test_data_path + data_path+ "/" + rec + ".jsonl",transfoms[data_path])
        with open(test_data_path + data_path + "/" + rec + "_expected.jsonl", "w") as outfile:
            for expected in expected_recs:
                json.dump(expected, outfile)
                outfile.write('\n')
