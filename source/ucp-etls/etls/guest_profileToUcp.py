import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
# Change import based on business object
from tah_lib.guest_profileTransform import buildObjectRecord
from tah_lib.etl_utils import writeToS3


glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObject = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObject.count()
businessObject.printSchema()

# repartitioning to obtain 500 records per file
print("nPartitions: ", businessObject.toDF().rdd.getNumPartitions())
newNPartitions = max(int(count/500), 1)
print("repartitionning in : ", newNPartitions)
businessObjectRepartitionedDF = businessObject.toDF().coalesce(newNPartitions)
print("nPartitions after: ", businessObjectRepartitionedDF.rdd.getNumPartitions())
businessObjectRepartitionedDF
businessObjectRepartitioned = DynamicFrame.fromDF(
    businessObjectRepartitionedDF, glueContext, "data")

nestedSegments = Map.apply(
    frame=businessObjectRepartitioned, f=buildObjectRecord)
nestedSegmentsDF = nestedSegments.toDF()

nestedSegmentsDF.printSchema()

profile = nestedSegmentsDF.select(
    explode(nestedSegmentsDF.guest_profile_recs)).select("col.*")
email = nestedSegmentsDF.select(
    explode(nestedSegmentsDF.common_email_recs)).select("col.*")
phone = nestedSegmentsDF.select(
    explode(nestedSegmentsDF.common_phone_recs)).select("col.*")
loyalty = nestedSegmentsDF.select(
    explode(nestedSegmentsDF.hotel_loyalty_recs)).select("col.*")

profile.printSchema()
email.printSchema()
phone.printSchema()
loyalty.printSchema()

writeToS3(glueContext, profile, args["DEST_BUCKET"], "guest_profile")
writeToS3(glueContext, email, args["DEST_BUCKET"], "email_history")
writeToS3(glueContext, phone, args["DEST_BUCKET"], "phone_history")
writeToS3(glueContext, loyalty, args["DEST_BUCKET"], "hotel_loyalty")
