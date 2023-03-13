import sys
import uuid
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode


# Change import based on business object
from tah_lib.air_bookingTransform import buildObjectRecord
from tah_lib.etl_utils import writeToS3

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])
businessObject = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObject.count()
print("count: ", count)
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

# applying Python transformation function
accpReccords = Map.apply(
    frame=businessObjectRepartitioned, f=buildObjectRecord)
accpReccords.printSchema()
# accpReccords.toDF().show(10)
# moving to dataframes
accpReccordsDF = accpReccords.toDF()
# exploding data into individual Dynamic Frames
bookings = accpReccordsDF.select(
    explode(accpReccordsDF.air_booking_recs)).select("col.*")
email = accpReccordsDF.select(
    explode(accpReccordsDF.common_email_recs)).select("col.*")
phone = accpReccordsDF.select(
    explode(accpReccordsDF.common_phone_recs)).select("col.*")
loyalty = accpReccordsDF.select(
    explode(accpReccordsDF.air_loyalty_recs)).select("col.*")

bookings.printSchema()
email.printSchema()
phone.printSchema()
loyalty.printSchema()

writeToS3(glueContext, bookings, args["DEST_BUCKET"], "air_booking")
writeToS3(glueContext, email, args["DEST_BUCKET"], "email_history")
writeToS3(glueContext, phone, args["DEST_BUCKET"], "phone_history")
writeToS3(glueContext, loyalty, args["DEST_BUCKET"], "air_loyalty")
