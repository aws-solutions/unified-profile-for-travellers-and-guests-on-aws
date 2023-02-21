import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
# Change import based on business object
from air_bookingTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(
    sys.argv, ['JOB_NAME', 'GLUE_DB', 'SOURCE_TABLE', 'DEST_BUCKET'])
businessObject = glueContext.create_dynamic_frame.from_catalog(
    database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])
count = businessObject.count()
print("count: ", count)
businessObject.printSchema()
# applying Python transformation function
accpReccords = Map.apply(frame=businessObject, f=buildObjectRecord)
accpReccords.printSchema()
accpReccords.toDF().show(10)
# moving to dataframes
accpReccordsDF = accpReccords.toDF()
# exploding nested array into invividual rows
explodedAccpReccordsDF = accpReccordsDF.select(
    accpReccordsDF.data, explode(accpReccordsDF.data))
# droping original column
explodedAccpReccordsDF = explodedAccpReccordsDF.drop(
    explodedAccpReccordsDF.data)
# back to DynamicFrames
explodedSegments = DynamicFrame.fromDF(
    explodedAccpReccordsDF, glueContext, "explodedAccpReccordsDF")
# Unnesting records
explodedSegmentsUnested = UnnestFrame.apply(frame=explodedSegments)
# displaying new scema and sample data
count = explodedSegmentsUnested.count()
print("count: ", count)
explodedSegmentsUnested.printSchema()
explodedSegmentsUnested.toDF().show(10, truncate=False)

# renaming columns to drop the prefix added by UnnestFrame
mapping = []
for x in range(0, len(explodedSegmentsUnested.schema().fields)):
    # see https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-crawler-pyspark-transforms-ApplyMapping.html for why we need the ``
    before = '`' + explodedSegmentsUnested.schema().fields[x].name + '`'
    after = explodedSegmentsUnested.schema().fields[x].name.split('.')[1]
    mapping.append((before, after))
print("Mappings: ", mapping)

explodedSegmentsUnestedColRenamed = explodedSegmentsUnested.apply_mapping(
    mapping)
# displaying new scema and sample data
count = explodedSegmentsUnestedColRenamed.count()
print("count: ", count)
explodedSegmentsUnestedColRenamed.printSchema()
explodedSegmentsUnestedColRenamed.toDF().show(10, truncate=False)
# writting the content to s3
glueContext.write_dynamic_frame.from_options(frame=explodedSegmentsUnestedColRenamed, connection_type="s3", connection_options={
                                             "path": "s3://"+args["DEST_BUCKET"], "partitionKeys": ["object_type"]}, format="csv")
