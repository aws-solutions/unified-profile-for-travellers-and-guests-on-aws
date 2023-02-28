import sys
import uuid
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
# Change import based on business object
from tah_lib.clickstreamTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])

businessObjectDF = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDF.count()
businessObjectDF.printSchema()

segments = Map.apply(frame=businessObjectDF, f=buildObjectRecord)
segments.printSchema()
# segments.toDF().show(100)

subfolder = str(uuid.uuid1(node=None, clock_seq=None))

segments.toDF().write.format("csv").option("header", "true").save(
    "s3://"+args["DEST_BUCKET"]+"/clickstream/"+subfolder)
