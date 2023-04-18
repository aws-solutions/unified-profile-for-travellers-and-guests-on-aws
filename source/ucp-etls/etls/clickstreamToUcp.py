import boto3
import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame

from tah_lib.clickstreamTransform import buildObjectRecord
from tah_lib.etl_utils import argList, ERROR_QUEUE_URL, createDynamicFrame, writeToS3, updateJobPredicates

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, argList)
dynamodbClient = boto3.client('dynamodb')

businessObjects = createDynamicFrame(glueContext, dynamodbClient, args)
count = businessObjects.count()
print("count: ", count)
businessObjects.printSchema()

# repartitioning to obtain 500 records per file
print("nPartitions: ", businessObjects.toDF().rdd.getNumPartitions())
newNPartitions = max(int(count/500), 1)
print("repartitionning in : ", newNPartitions)
businessObjectRepartitionedDF = businessObjects.toDF().coalesce(newNPartitions)
print("nPartitions after: ", businessObjectRepartitionedDF.rdd.getNumPartitions())
businessObjectRepartitionedDF
businessObjectRepartitioned = DynamicFrame.fromDF(
    businessObjectRepartitionedDF, glueContext, "data")

# applying Python transformation function
accpRecords = Map.apply(
    frame=businessObjectRepartitioned,
    f=lambda rec: buildObjectRecord(rec, args[ERROR_QUEUE_URL]))

accpRecordsDF = accpRecords.toDF()
accpRecordsDF.printSchema()

writeToS3(glueContext, accpRecordsDF, "clickstream", args)
updateJobPredicates(dynamodbClient, args, count)
