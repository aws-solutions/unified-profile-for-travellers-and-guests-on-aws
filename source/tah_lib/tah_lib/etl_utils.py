# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import uuid
from datetime import datetime
from typing import TYPE_CHECKING, Any, Final, Literal, TypedDict

from pyspark.sql.functions import explode

from tah_lib.common import get_job_predicates, send_solution_metrics

if TYPE_CHECKING:
    from mypy_boto3_dynamodb.client import DynamoDBClient
    from pyspark.sql import SparkSession
else:
    DynamoDBClient = object
    SparkContext = object

# List of job args
JOB_NAME: Final = "JOB_NAME"
GLUE_DB: Final = "GLUE_DB"
SOURCE_TABLE: Final = "SOURCE_TABLE"
DEST_BUCKET: Final = "DEST_BUCKET"
ERROR_QUEUE_URL: Final = "ERROR_QUEUE_URL"
ACCP_DOMAIN: Final = "ACCP_DOMAIN"
DYNAMO_TABLE: Final = "DYNAMO_TABLE"
BIZ_OBJECT_NAME: Final = "BIZ_OBJECT_NAME"
METRICS_SOLUTION_ID: Final = "METRICS_SOLUTION_ID"
METRICS_SOLUTION_VERSION: Final = "METRICS_SOLUTION_VERSION"
METRICS_UUID: Final = "METRICS_UUID"
SEND_ANONYMIZED_DATA: Final = "SEND_ANONYMIZED_DATA"
argList: Final = [
    JOB_NAME,
    GLUE_DB,
    SOURCE_TABLE,
    DEST_BUCKET,
    ERROR_QUEUE_URL,
    ACCP_DOMAIN,
    DYNAMO_TABLE,
    BIZ_OBJECT_NAME,
    METRICS_SOLUTION_ID,
    METRICS_SOLUTION_VERSION,
    METRICS_UUID,
    SEND_ANONYMIZED_DATA,
]


def build_s3_sub_folder() -> str:
    return datetime.now().strftime("%Y/%m/%d/%H/") + str(uuid.uuid4())


class S3ConnectorOptions(TypedDict):
    header: Literal["true"]
    partitionKeys: list[Literal["model_version"]]


def build_s3_connector_options() -> S3ConnectorOptions:
    return {
        "header": "true",
        "partitionKeys": ["model_version"],
    }


class Args(TypedDict, total=False):
    JOB_NAME: str
    GLUE_DB: str
    SOURCE_TABLE: str
    DEST_BUCKET: str
    ERROR_QUEUE_URL: str
    ACCP_DOMAIN: str
    DYNAMO_TABLE: str
    BIZ_OBJECT_NAME: str
    METRICS_SOLUTION_ID: str
    METRICS_SOLUTION_VERSION: str
    METRICS_UUID: str
    SEND_ANONYMIZED_DATA: str


def write_to_s3(glue_context: Any, data_frame: Any, folder: str, args: Args) -> None:
    # By default, spark dataframe overwites all data (even partitions that do not have new data).
    # We use the partitionOverwriteMode=dynamic to only overwrite new partitions.
    spark: SparkSession = glue_context.spark_session
    spark.conf.set("spark.sql.sources.partitionOverwriteMode", "dynamic")
    # addressing similar issue then the one mentioned at
    # https://stackoverflow.com/questions/36164914/prevent-dataframe-partitionby-from-removing-partitioned-columns-from-schema
    duplicated = data_frame.withColumn(
        "_model_version", data_frame["model_version"]
    ).withColumn("_tx_id", data_frame["tx_id"])
    # we partition by tx_id to avoid 2 concurent job writting to the same folder (ex hotel and air booking writting to email_history) and erasing eachother's data
    duplicated.write.format("csv").option("header", "true").partitionBy(
        "_model_version"
    ).partitionBy("_tx_id").mode("overwrite").save(
        "s3://"
        + args[DEST_BUCKET]
        + "/"
        + args[ACCP_DOMAIN]
        + "/"
        + folder
        + "/"
        + datetime.now().strftime("%Y/%m/%d/%H/")
    )


def explode_and_write(
    glue_context: Any,
    df: Any,
    accp_type: str,
    folder: str,
    args: Args,
    dynamodb_client: DynamoDBClient,
    record_count: int,
) -> None:
    print("[explode_and_write] exploding acptype ", accp_type)
    df_exploded = df.select(explode(df[accp_type]))
    df_exploded.printSchema()
    if df_exploded.schema["col"].dataType.typeName() == "struct":
        accp_object_df = df_exploded.select("col.*")
        accp_object_df.printSchema()
        count = accp_object_df.count()
        print("[explode_and_write] ", accp_type, " count=", count)
        print(
            "[explode_and_write] writting to S3 at ",
            args[DEST_BUCKET],
            "/",
            args[ACCP_DOMAIN],
            "/",
            accp_type,
        )
        write_to_s3(glue_context, accp_object_df, folder, args)
    else:
        print("[explode_and_write] No ACCP records to write for type " + accp_type)
    update_job_predicates(dynamodb_client, args, record_count)


# technically this returns an AWS Glue DynamicFrame, for which there are no type hints
def create_dynamic_frame(
    glue_context: Any, dynamodb_client: DynamoDBClient, args: Args
) -> Any:
    partitions = get_job_predicates(
        dynamodb_client, args[DYNAMO_TABLE], args[BIZ_OBJECT_NAME], args[ACCP_DOMAIN]
    )
    print("[create_dynamic_frame] Applying partition predicate: ", partitions)
    # If there are no partitions created by previous job run, process all available data.
    if partitions["pdp"] is None or partitions["cpp"] is None:
        return glue_context.create_dynamic_frame.from_catalog(
            database=args[GLUE_DB],
            table_name=args[SOURCE_TABLE],
            additional_options={"recurse": True},
        )
    # Otherwise, process new data since last job run only
    else:
        return glue_context.create_dynamic_frame.from_catalog(
            database=args[GLUE_DB],
            table_name=args[SOURCE_TABLE],
            push_down_predicate=partitions["pdp"],
            additional_options={"recurse": True},
        )
        # removing the additional option to avoid transient issue with catalogPartitionPredicate
        # additional_options={"catalogPartitionPredicate": partitions['cpp'], "recurse": True})


# save job run bookmark so data is not reprocessed on next run
def update_job_predicates(
    dynamodb_client: DynamoDBClient, args: Args, record_count: int
) -> None:
    # Update bookmark record
    dynamodb_client.put_item(
        TableName=args[DYNAMO_TABLE],
        Item={
            "item_id": {"S": "glue_job_bookmark"},
            "item_type": {"S": args[BIZ_OBJECT_NAME] + "_" + args[ACCP_DOMAIN]},
            "bookmark": {"S": datetime.now().strftime("%Y/%m/%d/%H")},
            "records_processed": {"N": str(record_count)},
        },
    )
    # Log to AWS Solution Metrics
    payload = {
        "solution": "SO0244",
        "business_object": args[BIZ_OBJECT_NAME],
        "records_processed": record_count,
    }
    send_solution_metrics(
        args[METRICS_SOLUTION_ID],
        args[METRICS_SOLUTION_VERSION],
        args[METRICS_UUID],
        args[SEND_ANONYMIZED_DATA],
        payload,
    )
