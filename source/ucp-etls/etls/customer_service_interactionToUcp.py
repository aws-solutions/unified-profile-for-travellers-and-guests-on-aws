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
from tah_lib.customer_service_interactionTransform import build_object_record
from tah_lib.etl_utils import (ERROR_QUEUE_URL, argList, create_dynamic_frame,
                               explode_and_write, update_job_predicates)

tx_id = str(uuid.uuid4())
print("tx_id: ", tx_id)

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, argList)
append_solution_identifier = build_solution_header(args["METRICS_SOLUTION_ID"], args["METRICS_SOLUTION_VERSION"])
config = Config(**append_solution_identifier)
dynamodb_client = boto3.client('dynamodb', config=config)

businessObjects = create_dynamic_frame(glueContext, dynamodb_client, args)
count = businessObjects.count()
print("count: ", count)
if count == 0:
    print("Biz object count is zero. existing ")
    update_job_predicates(dynamodb_client, args, count)
    os._exit(0)
businessObjects.printSchema()

# applying Python transformation function
accpRecords = Map.apply(
    frame=businessObjects,
    f=lambda rec: build_object_record(rec, args[ERROR_QUEUE_URL], tx_id))

accpRecordsDF = accpRecords.toDF()
accpRecordsDF.printSchema()

# exploding data into individual Dynamic Frames
explode_and_write(glueContext, accpRecordsDF, "conversation_recs", "customer_service_interaction",
                args, dynamodb_client, count)
