import uuid
from datetime import datetime
from pyspark.sql.functions import explode
from awsglue.dynamicframe import DynamicFrame
from tah_lib.common import getJobPredicates

# List of job args
JOB_NAME = 'JOB_NAME'
GLUE_DB = 'GLUE_DB'
SOURCE_TABLE = 'SOURCE_TABLE'
DEST_BUCKET = 'DEST_BUCKET'
ERROR_QUEUE_URL = 'ERROR_QUEUE_URL'
ACCP_DOMAIN = 'ACCP_DOMAIN'
DYNAMO_TABLE = 'DYNAMO_TABLE'
REGION = 'REGION'
BIZ_OBJECT_NAME = 'BIZ_OBJECT_NAME'
argList = [
    JOB_NAME,
    GLUE_DB,
    SOURCE_TABLE,
    DEST_BUCKET,
    ERROR_QUEUE_URL,
    ACCP_DOMAIN,
    DYNAMO_TABLE,
    REGION,
    BIZ_OBJECT_NAME,
]


def buildS3SubFolder():
    return datetime.now().strftime("%Y/%m/%d/%H/") + str(uuid.uuid1(node=None, clock_seq=None))


def buildS3ConnectorOptions():
    return {
        "header": "true",
        "partitionKeys": ["model_version"],
    }


def writeToS3(glueContext, dataFrame, folder, args):
    # By default, spark dataframe overwites all data (even partitions that do not have new data).
    # We use the partitionOverwriteMode=dynamic to only overwrite new partitions.
    spark = glueContext.spark_session
    spark.conf.set('spark.sql.sources.partitionOverwriteMode', 'dynamic')
    # addressing similar issue then the one mentioned at
    # https://stackoverflow.com/questions/36164914/prevent-dataframe-partitionby-from-removing-partitioned-columns-from-schema
    duplicated = dataFrame.withColumn("_model_version", dataFrame["model_version"]).withColumn(
        "_traveller_id", dataFrame["traveller_id"])
    duplicated.write.format("csv").option("header", "true").partitionBy("_model_version", "_traveller_id", "last_updated_partition").mode(
        "overwrite").save("s3://" + args[DEST_BUCKET] + "/" + args[ACCP_DOMAIN] + "/" + folder)


def explodeAndWrite(glueContext, df, accpType, folder, args, dynamodbClient, recordCount):
    print("[explodeAndWrite] exploding acptype ", accpType)
    dfExploded = df.select(explode(df[accpType]))
    dfExploded.printSchema()
    if dfExploded.schema["col"].dataType.typeName() == "struct":
        accpObjectDf = dfExploded.select("col.*")
        accpObjectDf.printSchema()
        print("[explodeAndWrite] writting to S3 at ",
              args[DEST_BUCKET], "/", args[ACCP_DOMAIN], "/", accpType)
        writeToS3(glueContext, accpObjectDf, folder, args)
    else:
        print("[explodeAndWrite] No ACCP records to write for type " + accpType)
    updateJobPredicates(dynamodbClient, args, recordCount)


def createDynamicFrame(glueContext, dynamodbClient, args):
    partitions = getJobPredicates(dynamodbClient, args[DYNAMO_TABLE], args[BIZ_OBJECT_NAME], args[ACCP_DOMAIN])
    # If there are no partitions created by previous job run, process all available data.
    if partitions['pdp'] is None or partitions['cpp'] is None:
        return glueContext.create_dynamic_frame.from_catalog(
            database=args[GLUE_DB], table_name=args[SOURCE_TABLE], additional_options={"recurse": True})
    # Otherwise, process new data since last job run only
    else:
        return glueContext.create_dynamic_frame.from_catalog(
            database=args[GLUE_DB], table_name=args[SOURCE_TABLE], push_down_predicate=partitions['pdp'],
            additional_options={"catalogPartitionPredicate": partitions['cpp'], "recurse": True})


# save job run bookmark so data is not reprocessed on next run
def updateJobPredicates(dynamodbClient, args, recordCount):
    dynamodbClient.put_item(
        TableName=args[DYNAMO_TABLE],
        Item={
            'item_id': {'S': 'glue_job_bookmark'},
            'item_type': {'S': args[BIZ_OBJECT_NAME] + "_" + args[ACCP_DOMAIN]},
            'bookmark': {'S': datetime.now().strftime("%Y/%m/%d/%H")},
            'records_processed': {'N': str(recordCount)},
        },
    )
