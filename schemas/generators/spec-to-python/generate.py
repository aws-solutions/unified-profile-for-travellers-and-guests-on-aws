# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

from jsonschemacodegen import python as pygen
from jsonschemacodegen.resolver import SimpleResolver
import json

files = ['../../schemas/lodging/booking.schema.json',
'../../schemas/lodging/loyalty.schema.json']
for file in files:
    with open(file) as fp:
        generator = pygen.GeneratorFromSchema('output_dir', resolver=SimpleResolver("#"))
        decoded = json.load(fp)
        print(decoded)
        for obj in decoded["definitions"]:
            filename = obj[0].lower() + obj[1:]
            generator.Generate(decoded["definitions"][obj], None, obj,filename)