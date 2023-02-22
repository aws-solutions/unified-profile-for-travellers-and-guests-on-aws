import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
# Change import based on business object
from hotel_bookingTransform import buildObjectRecord
from autoFlatten import flattenWithNestedArrays, cleanNullColumns

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET', 'BUSINESS_OBJECT'])
businessObject = args['BUSINESS_OBJECT']

businessObjectDF = glueContext.create_dynamic_frame.from_catalog(database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

flattenedDF = flattenWithNestedArrays(businessObjectDF.toDF())
flattenedDF = flattenedDF.fillna("")
businessObjectDF = DynamicFrame.fromDF(flattenedDF, glueContext, businessObject)

count = businessObjectDF.count()
print(businessObject + " count: ", count)
businessObjectDF.printSchema()

segments = Map.apply(frame=businessObjectDF, f=buildObjectRecord)
segments.printSchema()
segments.toDF().show(100)

print("nPartitions: ", segments.toDF().rdd.getNumPartitions())
newNPartitions = max(int(count/25), 1)
print("repartitionning in: ", newNPartitions)
repartitionedSegmentsDF = segments.toDF().coalesce(newNPartitions)
repartitionedSegmentsDF = repartitionedSegmentsDF.fillna("")
repartitionedSegmentsDF = cleanNullColumns(repartitionedSegmentsDF)
repartitionedSegmentsDF = repartitionedSegmentsDF.fillna("")
print("nPartitions after: ", repartitionedSegmentsDF.rdd.getNumPartitions())

repartitionedSegmentsDF.write.mode("overwrite").format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject)
