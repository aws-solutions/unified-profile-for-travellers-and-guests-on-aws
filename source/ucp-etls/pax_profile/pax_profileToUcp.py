# pyright: reportMissingImports=false, reportUndefinedVariable=false
import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode
from pax_profileTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, ['JOB_NAME','GLUE_DB','SOURCE_TABLE','DEST_BUCKET','BUSINESS_OBJECT'])
businessObject = args['BUSINESS_OBJECT']

# TODO: decide on solution to get data from business object partition
# https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-crawler-pyspark-extensions-glue-context.html#aws-glue-api-crawler-pyspark-extensions-glue-context-create-sample-dynamic-frame-from-catalog
businessObjectDYF = glueContext.create_dynamic_frame.from_catalog(database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDYF.count()
print("count: ", count)
businessObjectDYF.printSchema()

nestedSegments =  Map.apply(frame = businessObjectDYF, f = buildObjectRecord)
nestedSegmentsDF = nestedSegments.toDF()

nestedSegmentsDF.printSchema()
nestedSegmentsDF.show(10)

profileDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
emailDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
phoneDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
loyaltyDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))

profileDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/pax_profile")
emailDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/email_history")
phoneDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/phone_history")
loyaltyDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/air_loyalty")
