# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import os
import sys
import uuid

import boto3
from awsglue.context import GlueContext
from awsglue.dynamicframe import DynamicFrame
from awsglue.job import Job
from awsglue.transforms import Map
from awsglue.utils import getResolvedOptions
from botocore.config import Config
from pyspark.context import SparkContext
from tah_lib.common import build_solution_header
from tah_lib.etl_utils import (ERROR_QUEUE_URL, argList, create_dynamic_frame,
                               explode_and_write, update_job_predicates)
from tah_lib.guest_profileTransform import build_object_record

tx_id = str(uuid.uuid4())
print("tx_id: ", tx_id)

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, argList)
append_solution_identifier = build_solution_header(args["METRICS_SOLUTION_ID"], args["METRICS_SOLUTION_VERSION"])
config = Config(**append_solution_identifier)
dynamodb_client = boto3.client('dynamodb', config=config)

businessObjects = create_dynamic_frame(glueContext, dynamodb_client, args)
count = businessObjects.count()
if count == 0:
    print("Biz object count is zero. existing ")
    update_job_predicates(dynamodb_client, args, count)
    os._exit(0)
print("count: ", count)
businessObjects.printSchema()

# applying Python transformation function
accpRecords = Map.apply(
    frame=businessObjects,
    f=lambda rec: build_object_record(rec, args[ERROR_QUEUE_URL], tx_id))

accpRecordsDF = accpRecords.toDF()
accpRecordsDF.printSchema()

# exploding data into individual Dynamic Frames
explode_and_write(glueContext, accpRecordsDF, "guest_profile_recs", "guest_profile",
                args, dynamodb_client, count)
explode_and_write(glueContext, accpRecordsDF, "common_email_recs", "email_history",
                args, dynamodb_client, count)
explode_and_write(glueContext, accpRecordsDF, "common_phone_recs", "phone_history",
                args, dynamodb_client, count)
explode_and_write(glueContext, accpRecordsDF, "hotel_loyalty_recs", "hotel_loyalty",
                args, dynamodb_client, count)
explode_and_write(glueContext, accpRecordsDF, "hotel_loyalty_tx_recs", "loyalty_transaction",
                args, dynamodb_client, count)