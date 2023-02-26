import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
from tah_lib.hotel_stayTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjects = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjects.count()
print("count: ", count)
businessObjects.printSchema()

accpReccords = Map.apply(frame=businessObjects, f=buildObjectRecord)

accpReccordsDF = accpReccords.toDF()


revenue_items = accpReccordsDF.select(
    explode(accpReccordsDF.hotel_stay_revenue_items)).select("col.*")

revenue_items.printSchema()

revenue_items.write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/hotel_stay_revenue_items")
