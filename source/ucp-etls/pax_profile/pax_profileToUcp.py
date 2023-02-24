# pyright: reportMissingImports=false
import sys
from awsglue.transforms import Map
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.sql.functions import explode, col
from pax_profileTransform import buildObjectRecord

glueContext = GlueContext(SparkContext.getOrCreate())
args = getResolvedOptions(sys.argv, ['JOB_NAME','GLUE_DB','SOURCE_TABLE','DEST_BUCKET','BUSINESS_OBJECT'])
businessObject = args['BUSINESS_OBJECT']

businessObjectDYF = glueContext.create_dynamic_frame.from_catalog(database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDYF.count()
print("count: ", count)
businessObjectDYF.printSchema()

nestedSegments =  Map.apply(frame = businessObjectDYF, f = buildObjectRecord)
nestedSegmentsDF = nestedSegments.toDF()

nestedSegmentsDF.printSchema()
nestedSegmentsDF.show(10)

profileDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs).alias("data")).select(col("data.*"))
emailDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs).alias("data")).select(col("data.*"))
phoneDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs).alias("data")).select(col("data.*"))
loyaltyDF = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs).alias("data")).select(col("data.*"))

# TODO: decide on optimal partition strategy
profileDF = profileDF.coalesce(max(1, profileDF.count() / 25))
emailDF = emailDF.coalesce(max(1, emailDF.count() / 25))
phoneDF = phoneDF.coalesce(max(1, phoneDF.count() / 25))
loyaltyDF = loyaltyDF.coalesce(max(1, loyaltyDF.count() / 25))

profileDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/pax_profile")
emailDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/email_history")
phoneDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/phone_history")
loyaltyDF.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/air_loyalty")
