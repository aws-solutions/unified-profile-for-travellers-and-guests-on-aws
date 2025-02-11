// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"log"
	"os"
	ir "tah/upt/source/ucp-fargate/identity-resolution"
	cache "tah/upt/source/ucp-fargate/ucp-cache"
)

var TASK_REQUEST_TYPE = os.Getenv("TASK_REQUEST_TYPE")

func main() {
	log.Printf("Starting Fargate Task for Request Type: %v", TASK_REQUEST_TYPE)
	switch TASK_REQUEST_TYPE {
	case "id-res":
		err := ir.IdentityResolutionMain()
		if err != nil {
			log.Printf("Error running identity resolution: %v", err)
			os.Exit(1)
		}
	case "rebuild-cache":
		cache.CacheMain()
	default:
		log.Println("Invalid request type")
		os.Exit(1)
	}
}
