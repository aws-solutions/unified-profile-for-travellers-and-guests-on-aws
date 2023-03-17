import uuid
from datetime import datetime
from pyspark.sql.functions import explode


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
    # https://stackoverflow.com/questions/36164914/prevent-dataframe-partitionby-from-removing-partitioned-columns-from-schema
    duplicated = dataFrame.withColumn("_model_version", dataFrame["model_version"]).withColumn(
        "_traveller_id", dataFrame["traveller_id"])
    duplicated.write.format("csv").option("header", "true").partitionBy("_model_version", "_traveller_id", "last_updated_partition").mode(
        "overwrite").save("s3://"+bucketName + "/" + folder)


def explodeAndWrite(glueContext, df, accpType, bucketName, folder):
    print("[explodeAndWrite] exploding acptype ", accpType)
    dfExploded = df.select(explode(df[accpType]))
    dfExploded.printSchema()
    if dfExploded.schema["col"].dataType.typeName() == "struct":
        accpObjectDf = dfExploded.select("col.*")
        accpObjectDf.printSchema()
        print("[explodeAndWrite] writting to S3 at ", bucketName, "/", folder)
        writeToS3(glueContext, accpObjectDf, bucketName, folder)
    else:
        print("[explodeAndWrite] No ACCP records to write for type " + accpType)
