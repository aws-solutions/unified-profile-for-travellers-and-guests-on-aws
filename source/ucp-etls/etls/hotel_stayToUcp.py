import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
from pyspark.sql.functions import size

from tah_lib.hotel_stayTransform import buildObjectRecord
from tah_lib.etl_utils import explodeAndWrite

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET', 'ERROR_QUEUE_URL', 'ACCP_DOMAIN'])

businessObjects = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"], additional_options={"recurse": True})

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

accpReccords = Map.apply(
    frame=businessObjectRepartitioned,
    f=lambda rec: buildObjectRecord(rec, args['ERROR_QUEUE_URL']))
accpReccordsDF = accpReccords.toDF()

explodeAndWrite(glueContext, accpReccordsDF, "hotel_stay_revenue_items",
                args["DEST_BUCKET"], "hotel_stay_revenue_items", args["ACCP_DOMAIN"])
