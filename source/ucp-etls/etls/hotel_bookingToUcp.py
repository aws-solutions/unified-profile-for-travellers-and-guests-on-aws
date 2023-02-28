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
from tah_lib.hotel_bookingTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjects = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjects.count()
print("count: ", count)
businessObjects.printSchema()

accpReccords = Map.apply(frame=businessObjects, f=buildObjectRecord)
accpReccords.printSchema()
# accpReccords.toDF().show(10)
# moving to dataframes
accpReccordsDF = accpReccords.toDF()
# exploding data into individual Dynamic Frames
bookings = accpReccordsDF.select(
    explode(accpReccordsDF.hotel_booking_recs)).select("col.*")
email = accpReccordsDF.select(
    explode(accpReccordsDF.common_email_recs)).select("col.*")
phone = accpReccordsDF.select(
    explode(accpReccordsDF.common_phone_recs)).select("col.*")
loyalty = accpReccordsDF.select(
    explode(accpReccordsDF.guest_loyalty_recs)).select("col.*")

bookings.printSchema()
email.printSchema()
phone.printSchema()
loyalty.printSchema()

subfolder = str(uuid.uuid1(node=None, clock_seq=None))

bookings.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/hotel_booking/"+subfolder)
email.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/email_history/"+subfolder)
phone.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/phone_history/"+subfolder)
loyalty.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/hotel_loyalty/"+subfolder)
