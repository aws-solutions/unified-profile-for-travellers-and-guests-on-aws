import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
from hotel_stayTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, ['JOB_NAME','GLUE_DB','SOURCE_TABLE','DEST_BUCKET','BUSINESS_OBJECT'])
businessObject = args['BUSINESS_OBJECT']

businessObjectDF = glueContext.create_dynamic_frame.from_catalog(database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDF.count()
print(businessObject + " count: ", count)
businessObjectDF.printSchema()

nestedSegments =  Map.apply(frame = businessObjectDF, f = buildObjectRecord)
segments = nestedSegments.toDF().select(explode(nestedSegments.data))

segments.printSchema()
segments.show(100)

print("nPartitions: ", segments.rdd.getNumPartitions())
newNPartitions = max(int(count/25), 1)
print("repartitionning in: ", newNPartitions)
repartitionedSegmentsDF = segments.coalesce(newNPartitions)
print("nPartitions after: ", repartitionedSegmentsDF.rdd.getNumPartitions())

repartitionedSegmentsDF.write.mode("overwrite").format("csv").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject)