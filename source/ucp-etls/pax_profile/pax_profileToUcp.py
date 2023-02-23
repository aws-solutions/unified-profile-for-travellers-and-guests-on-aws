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

businessObjectDYF = glueContext.create_dynamic_frame.from_catalog(database=args["GLUE_DB"], table_name=args["SOURCE_TABLE"])

count = businessObjectDYF.count()
print(businessObject + " count: ", count)
businessObjectDYF.printSchema()

nestedSegments =  Map.apply(frame = businessObjectDYF, f = buildObjectRecord)
nestedSegmentsDF = nestedSegments.toDF()

nestedSegmentsDF.printSchema()
nestedSegmentsDF.show(10)

profile = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
email = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
phone = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))
loyalty = nestedSegmentsDF.select(explode(nestedSegmentsDF.profileRecs))

profile.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/pax_profile")
email.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/email_history")
phone.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/phone_history")
loyalty.write.format("csv").option("header", "true").save("s3://"+args["DEST_BUCKET"]+"/"+businessObject+"/air_loyalty")
