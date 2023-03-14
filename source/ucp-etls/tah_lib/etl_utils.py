import uuid
from datetime import datetime


def buildS3SubFolder():
    return datetime.now().strftime("%Y/%m/%d/%H/") + str(uuid.uuid1(node=None, clock_seq=None))


def buildS3ConnectorOptions():
    return {
        "header": "true",
        "partitionKeys": ["model_version"],
    }


def writeToS3(glueContext, dataFrame, bucketName: str, folder: str):
    # By default, spark dataframe overwites all data (even partitions that do not have new data).
    # We use the partitionOverwriteMode=dynamic to only overwrite new partitions.
    spark = glueContext.spark_session
    spark.conf.set('spark.sql.sources.partitionOverwriteMode', 'dynamic')
    dataFrame.write.format("csv").option("header", "true").partitionBy("model_version", "traveller_id", "last_updated_partition").mode(
        "overwrite").save("s3://"+bucketName + "/" + folder)
