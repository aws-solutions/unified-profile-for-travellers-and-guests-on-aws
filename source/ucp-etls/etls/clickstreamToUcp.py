import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame


# Change import based on business object
from tah_lib.clickstreamTransform import buildObjectRecord
from tah_lib.etl_utils import writeToS3

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET', 'ERROR_QUEUE_URL', 'ACCP_DOMAIN'])

businessObject = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"], additional_options={"recurse": True})

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

accpReccords = Map.apply(
    frame=businessObjectRepartitioned,
    f=lambda rec: buildObjectRecord(rec, args['ERROR_QUEUE_URL']))
accpReccords.printSchema()

writeToS3(glueContext, accpReccords.toDF(),
          args["DEST_BUCKET"], args["ACCP_DOMAIN"] + "/" + "clickstream")
