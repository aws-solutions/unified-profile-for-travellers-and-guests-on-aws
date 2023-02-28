# pyright: reportMissingImports=false, reportUndefinedVariable=false
import sys
import uuid
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
from tah_lib.pax_profileTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjectDYF = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDYF.count()
businessObjectDYF.printSchema()

accpRecords = Map.apply(frame=businessObjectDYF, f=buildObjectRecord)

accpRecordsDF = accpRecords.toDF()
accpRecordsDF.printSchema()
# nestedSegmentsDF.show(10)

profile = accpRecordsDF.select(
    explode(accpRecordsDF.air_profile_recs)).select("col.*")
email = accpRecordsDF.select(
    explode(accpRecordsDF.common_email_recs)).select("col.*")
phone = accpRecordsDF.select(
    explode(accpRecordsDF.common_phone_recs)).select("col.*")
loyalty = accpRecordsDF.select(
    explode(accpRecordsDF.air_loyalty_recs)).select("col.*")

subfolder = str(uuid.uuid1(node=None, clock_seq=None))

profile.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/pax_profile/"+subfolder)
email.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/email_history/"+subfolder)
phone.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/phone_history/"+subfolder)
loyalty.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/air_loyalty/"+subfolder)
