# pyright: reportMissingImports=false, reportUndefinedVariable=false
import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode

from tah_lib.pax_profileTransform import buildObjectRecord
from tah_lib.etl_utils import buildS3SubFolder

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjects = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjects.count()
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

accpRecords = Map.apply(frame=businessObjectRepartitioned, f=buildObjectRecord)

accpRecordsDF = accpRecords.toDF()
accpRecordsDF.printSchema()


profile = accpRecordsDF.select(
    explode(accpRecordsDF.air_profile_recs)).select("col.*")
email = accpRecordsDF.select(
    explode(accpRecordsDF.common_email_recs)).select("col.*")
phone = accpRecordsDF.select(
    explode(accpRecordsDF.common_phone_recs)).select("col.*")
loyalty = accpRecordsDF.select(
    explode(accpRecordsDF.air_loyalty_recs)).select("col.*")

subfolder = buildS3SubFolder()

profile.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/pax_profile/"+subfolder)
email.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/email_history/"+subfolder)
phone.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/phone_history/"+subfolder)
loyalty.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/air_loyalty/"+subfolder)
