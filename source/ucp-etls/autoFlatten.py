from pyspark.sql import DataFrame
import pyspark.sql.functions as F
from pyspark.sql.types import *

def explode_cols(data, cols):
    data = data.withColumn('exp_combo', F.arrays_zip(*cols))
    data = data.withColumn('exp_combo', F.explode('exp_combo'))
    for col in cols:
        data = data.withColumn(col, F.col('exp_combo.' + col))

    return data.drop(F.col('exp_combo'))

# Create outer method to return the flattened Data Frame
def flatten_json_df(_df: DataFrame) -> DataFrame:
    # List to hold the dynamically generated column names
    flattened_col_list = []
    
    # Inner method to iterate over Data Frame to generate the column list
    def get_flattened_cols(df: DataFrame, struct_col: str = None) -> None:
        for col in df.columns:
            if df.schema[col].dataType.typeName() != 'struct':
                if struct_col is None:
                    flattened_col_list.append(f"{col} as {col.replace('.','_')}")
                else:
                    t = struct_col + "." + col
                    flattened_col_list.append(f"{t} as {t.replace('.','_')}")
            else:
                chained_col = struct_col +"."+ col if struct_col is not None else col
                get_flattened_cols(df.select(col+".*"), chained_col)
    
    # Call the inner Method
    get_flattened_cols(_df)
    
    # Return the flattened Data Frame
    return _df.selectExpr(flattened_col_list)

def get_top_level_arrays(df):
    #Given a PySpark dataframe, returns a list of all elements in the dataframe
    #that is an array at the top level schema.
    
    array_columns = []
    for column in df.schema:
        if isinstance(column.dataType, ArrayType):
            array_columns.append(column.name)

    return array_columns

def flattenWithNestedArrays(dataframe: DataFrame):
    sparkDF = dataframe.toDF()
    flattenedDF = flatten_json_df(sparkDF)
    arrNames = get_top_level_arrays(sparkDF)

    while arrNames != []:
        explodedDF = explode_cols(flattenedDF, arrNames)
        flattenedDF = flatten_json_df(explodedDF)
        arrNames = get_top_level_arrays(flattenedDF)
    return flattenedDF