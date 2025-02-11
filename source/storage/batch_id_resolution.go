// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package customerprofileslcs

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"tah/upt/source/tah-core/ecs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/google/uuid"
)

type UuidPartition struct{ LowerBound, UpperBound uuid.UUID }

// partition the space of v4 (random) UUIDs into 2^numBits partitions.
// returns a closure yielding pairs of v4 UUIDs representing the lower and upper bounds of a closed interval for each partition.
// the second (bool) return value is true until iteration completes or an error is encountered.
// bits must be between zero (a single partition) and 31 (2^31 or 2,147,483,648 partitions), inclusive.
func PartitionUuid4Space(numBits int) (func() (UuidPartition, bool, error), error) {
	// take numBits as a parameter instead of a number of partitions to make enforcing a power of two easy

	// 2^31 partitions should be sufficient for all practical purposes

	// see RFC 4122, 9562 for details on UUID v4 format
	// UUIDs are encoded as 128 bit integers in network byte order (most significant byte first)
	// use the high 32 bits (random_a) to partition the space

	// the usefulness of this approach to partition a space of random identifiers depends on the assumption
	// ... that UUIDs conform to the UUID v4 format and are uniformly-distributed
	if numBits < 0 || numBits > 31 {
		return func() (UuidPartition, bool, error) { return UuidPartition{}, false, nil }, fmt.Errorf(
			"bits must be between zero and 31: %v",
			numBits,
		)
	}

	// closure captures the value of the high numBits, which is all that is needed to generate the next partition
	var i uint32
	return func() (UuidPartition, bool, error) {
		// 1 << numBits is the number of partitions (2^numBits)
		if i < 1<<numBits {
			part, err := createUuid4Partition(i, numBits)
			if err != nil {
				return UuidPartition{}, false, err
			}
			i++
			return part, true, nil
		} else {
			return UuidPartition{}, false, nil
		}
	}, nil
}

// see RFC 4122, 9562 for details on UUID v4 format
// UUIDs are encoded as 128 bit integers in network byte order (most significant byte first)
// these constants are the low 12 bytes of the minimum and maximum v4 UUIDs possible given a particular
// ... value for the high 32 bits
// the version field (most significant 4 bits of octet 6) must be set to 0x4 for both
// the variant field (variable number of the most significant bits of octet 8) must have the most
// ... significant bits set to 0b10, leaving a minimum value for the nibble of 0x8 and a maximum of 0xB
var minUuidLowBytes = [12]byte{0x00, 0x00, 0x40, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
var maxUuidLowBytes = [12]byte{0xFF, 0xFF, 0x4F, 0xFF, 0xBF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

// create a partition (pair) of v4 UUIDs representing the smallest and largest UUIDs possible with
// the low numBits of value as the high numBits of the UUID.
func createUuid4Partition(value uint32, numBits int) (UuidPartition, error) {
	// create a mask for the numBits bits that we care about
	var mask uint32 = (1 << numBits) - 1
	// calculate how far to shift these bits to reach the highest bits
	shift := 32 - numBits
	highBits := (value & mask) << shift
	lower, err := createUuid4(highBits, minUuidLowBytes)
	if err != nil {
		return UuidPartition{}, err
	}
	// the lower 32 - numBits bits must be set to one to represent the largest UUID in this partition
	upperHighBits := highBits | ^(mask << shift)
	upper, err := createUuid4(upperHighBits, maxUuidLowBytes)
	if err != nil {
		return UuidPartition{}, err
	}
	return UuidPartition{LowerBound: lower, UpperBound: upper}, nil
}

// create a UUID v4 using highBits as the high 32 bits and lowBytes as the remaining bytes
func createUuid4(highBits uint32, lowBytes [12]byte) (uuid.UUID, error) {
	uuidBytes := [16]byte{}
	// see RFC 4122, 9562 for details on UUID v4 format
	// UUIDs are encoded as 128 bit integers in network byte order (most significant byte first)
	binary.BigEndian.PutUint32(uuidBytes[0:4], highBits)
	copy(uuidBytes[4:16], lowBytes[:])
	lower, err := uuid.FromBytes(uuidBytes[:])
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error generating UUID v4 from %v: %w", uuidBytes, err)
	}
	return lower, nil
}

type EcsApi interface {
	RunTask(context.Context, ecs.TaskConfig, []types.KeyValuePair, string) (*awsecs.RunTaskOutput, error)
}

func StartRuleBasedJob(
	ctx context.Context,
	cfg EcsApi,
	taskConfig ecs.TaskConfig,
	domain string,
	objectTypeCounts map[string]int64,
	ruleSet RuleSet,
	startedBy string,
) error {
	var allErrors error
	for objectTypeName, count := range objectTypeCounts {
		if count == 0 || objectTypeName == countTableProfileObject {
			continue
		}
		// until we do load testing, clamp to between 1 (2^0) and 16 (2^4) partitions
		// until we do load testing, partition the table if it has more than one million (1e6) rows
		partitionBits := int(min(4, max(0, count/1e6-1)))

		for _, rule := range ruleSet.Rules {
			getPartition, err := PartitionUuid4Space(partitionBits)
			if err != nil {
				allErrors = errors.Join(allErrors, err)
				continue
			}
			for partition, ok, err := getPartition(); ok && err == nil; partition, ok, err = getPartition() {
				envOverrides := []types.KeyValuePair{
					{Name: aws.String("DOMAIN_NAME"), Value: &domain},
					{Name: aws.String("OBJECT_TYPE_NAME"), Value: &objectTypeName},
					{Name: aws.String("RULE_SET_VERSION"), Value: aws.String(strconv.Itoa(ruleSet.LatestVersion))},
					{Name: aws.String("RULE_ID"), Value: aws.String(strconv.Itoa(rule.Index))},
					{Name: aws.String("PARTITION_LOWER_BOUND"), Value: aws.String(partition.LowerBound.String())},
					{Name: aws.String("PARTITION_UPPER_BOUND"), Value: aws.String(partition.UpperBound.String())},
				}
				response, err := cfg.RunTask(ctx, taskConfig, envOverrides, startedBy)
				if err != nil {
					allErrors = errors.Join(allErrors, err)
				}
				if response != nil {
					for _, failure := range response.Failures {
						allErrors = errors.Join(
							allErrors,
							fmt.Errorf(
								"error running task on %v (%v): %v",
								aws.ToString(failure.Arn),
								aws.ToString(failure.Reason),
								aws.ToString(failure.Detail),
							),
						)
					}
				}
			}
		}
	}
	return allErrors
}
