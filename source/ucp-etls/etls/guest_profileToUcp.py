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
from tah_lib.guest_profileTransform import buildObjectRecord


glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjectDYF = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDYF.count()
businessObjectDYF.printSchema()

nestedSegments = Map.apply(frame=businessObjectDYF, f=buildObjectRecord)
nestedSegmentsDF = nestedSegments.toDF()

nestedSegmentsDF.printSchema()
nestedSegmentsDF.show(10)

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

subfolder = str(uuid.uuid1(node=None, clock_seq=None))

profile.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/guest_profile/"+subfolder)
email.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/email_history/"+subfolder)
phone.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/phone_history/"+subfolder)
loyalty.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/hotel_loyalty/"+subfolder)
