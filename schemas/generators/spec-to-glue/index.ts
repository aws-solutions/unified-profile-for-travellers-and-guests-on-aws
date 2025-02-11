// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import * as glue from '@aws-cdk/aws-glue-alpha';
import * as fs from 'fs';


schemaToGlue('../../schemas/air/pax_profile.schema.json')
schemaToGlue('../../schemas/air/air_booking.schema.json')
schemaToGlue('../../schemas/lodging/hotel_booking.schema.json')
schemaToGlue('../../schemas/lodging/guest_profile.schema.json')
schemaToGlue('../../schemas/lodging/hotel_stay_revenue.schema.json')
schemaToGlue('../../schemas/common/clickevent.schema.json')
schemaToGlue('../../schemas/common/customer_service_interaction.schema.json')
schemaToGlue('../../schemas/traveller360/traveller.schema.json')


function schemaToGlue(filePath: string): glue.Column[] {
    console.log("Generating Glue Table type for schema file", filePath)
    let data = fs.readFileSync(filePath, 'utf8');
    const schema = JSON.parse(data);
    if (!schema) {
        console.log("Could not parse schema ", schema)
        throw "Could not parse schema: empty file"
    }
    if (!schema["$ref"]) {
        console.log("Invalid schema: no root object ", schema)
        throw "Could not parse schema: Invalid schema"
    }
    const rootObjectName = parseRef(schema["$ref"]);
    console.log("Root object for schema is ", rootObjectName)
    let columns = recursiveGlueColumns(schema["$defs"][rootObjectName], schema["$defs"], 0);
    console.log("Glue columns: ", columns)
    let jsonObject = JSON.stringify({ "columns": columns })
    let object = parseObjectFrommFileName(filePath)
    fs.writeFile('../../glue_schemas/' + object + ".glue.json", jsonObject, function () {
        console.log("Successfully created glue table schema for object", object)
    })
    return columns
}

function parseObjectFrommFileName(path: string) {
    let segments = path.split("/")
    let fileName = segments[segments.length - 1]
    return fileName.split(".")[0]
}

function recursiveGlueColumns(schema: any, definitions: any, level: number): any {
    console.log("Running recursive schema parsing with object", schema)
    if (schema.type) {
        if (schema.type == 'object') {
            let columns: glue.Column[] = [];
            for (let propertyName in schema.properties) {
                const property = schema.properties[propertyName];
                columns.push({
                    name: propertyName,
                    type: recursiveGlueColumns(property, definitions, level + 1),
                });
            }
            if (level === 0) {
                return columns
            }
            return glue.Schema.struct(columns);
        } else if (schema.type == 'string') {
            return glue.Schema.STRING;
        } else if (schema.type == 'boolean') {
            return glue.Schema.BOOLEAN;
        } else if (schema.type == 'integer') {
            return glue.Schema.INTEGER;
        } else if (schema.type == 'number') {
            return glue.Schema.FLOAT;
        } else if (schema.type == 'array') {
            let objSchema = schema.items
            if (schema.items["$ref"]) {
                objSchema = definitions[parseRef(schema.items["$ref"])]
            }
            let glueType = recursiveGlueColumns(objSchema, definitions, level + 1);
            return glue.Schema.array(glueType)
        } else {
            console.log("Unsupported schema type", schema.type);
        }
    } else if (schema["$ref"]) {
        let objSchema = definitions[parseRef(schema["$ref"])]
        let glueType = recursiveGlueColumns(objSchema, definitions, level + 1);
        console.log("level", level, " ref: ", schema["$ref"], "returning glue type: ", glueType)
        return glueType
    }
    return glue.Schema.STRING;
}

//return sobject name from "$ref" key
function parseRef(refstring: string): string {
    if (refstring) {
        return refstring.split("/")[2];
    }
    console.log("No $ref string provided for object",)
    return ""
}