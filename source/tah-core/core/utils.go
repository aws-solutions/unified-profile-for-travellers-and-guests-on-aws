// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	geohash "github.com/mmcloughlin/geohash"
)

var CLOUDRACK_DATEFORMAT string = "20060102"

func SliceContains(s string, slce []string) bool {
	for _, v := range slce {
		if v == s {
			return true
		}
	}
	return false
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func Chunk(array []interface{}, chunkSize int) [][]interface{} {
	var divided [][]interface{}
	for i := 0; i < len(array); i += chunkSize {
		end := i + chunkSize
		if end > len(array) {
			end = len(array)
		}
		divided = append(divided, array[i:end])
	}
	return divided
}

// Build a double array of integer from a max value and a chunkSize
func ChunkArray(size int64, chunkSize int64) [][]interface{} {
	toChunk := make([]interface{}, 0, 0)
	for i := int64(0); i < size; i++ {
		toChunk = append(toChunk, i)
	}
	return Chunk(toChunk, int(chunkSize))
}

// Generates a unique ID based timestamp
func GenerateUniqueId() string {
	return UniqueIdOfLength(10)
}

// Generates a unique ID based timestamp
const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func UniqueIdOfLength(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func UUID() string {
	return uuid.New().String()
}

func EmptyUUID() string {
	return uuid.Nil.String()
}

func ParseEpochMs(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// TODO: to evaluate th value of this tradeoff: this is probably a little slow but abstract the complexity for all uses of
// the save many function(and actually any core operation on array of interface)
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

/**********
*Geo Hasshing
**************/
func GeoHash(lat, lng float64) string {
	return geohash.Encode(lat, lng)
}
func GeoHashWithPrecision(lat, lng, radius float64) string {
	prec := builCaracterLength(radius)
	return geohash.EncodeWithPrecision(lat, lng, prec)
}

// geohasg precision: https://en.wikipedia.org/wiki/Geohash
// 1   ±2500
// 2   ±630
// 3   ±78
// 4   ±20
// 5   ±2.4
// 6   ±0.61
// 7   ±0.076
// 8   ±0.019
func builCaracterLength(radius float64) uint {
	precisions := []float64{2500, 630, 78, 20, 2.4, 0.61, 0.076, 0.019}
	ind := 0
	if radius >= precisions[ind] {
		return uint(ind + 1)
	}
	ind = ind + 1
	for ind < len(precisions)-2 {
		if radius >= precisions[ind+1] && radius < precisions[ind] {
			return uint(ind + 1)
		}
		ind = ind + 1
	}
	log.Printf("[UTILS] Geohash precision for %f KM is  %v \n", radius, ind)
	return uint(ind)
}

func GeterateDateRange(startDate string, endDate string) []string {
	start, _ := time.Parse(CLOUDRACK_DATEFORMAT, startDate)
	end, _ := time.Parse(CLOUDRACK_DATEFORMAT, endDate)
	diff := int(end.Sub(start).Hours() / 24.0)
	log.Printf("[UTILS] Diff between %v and %v => %+v \n", start, end, diff)
	dates := make([]string, 0, 0)
	for i := 0; i < diff; i++ {
		t := start.AddDate(0, 0, i)
		dates = append(dates, t.Format(CLOUDRACK_DATEFORMAT))
	}
	log.Printf("[UTILS] Generated date range from %s to %s => %+v \n", startDate, endDate, dates)
	return dates
}

/********************
* Utility Functions
*********************/

func IsEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ErrorsToError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msgs := []string{}
	for _, err := range errs {
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	//only nil provided
	if len(msgs) == 0 {
		return nil
	}
	return errors.New("Multiple errors: " + strings.Join(msgs, ", "))
}

func MapStringInterfaceToMapStringString(m map[string]interface{}) map[string]string {
	result := make(map[string]string)

	for key, value := range m {
		switch v := value.(type) {
		case string:
			result[key] = v
		case int:
			result[key] = strconv.Itoa(v)
		case int64:
			result[key] = strconv.FormatInt(v, 10)
		case bool:
			result[key] = strconv.FormatBool(v)
		case float64:
			result[key] = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}

	return result
}

func S3BucketNameToARN(name string) string {
	return fmt.Sprintf("arn:aws:s3:::%s", name)
}

func ContainsString(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length])
}

func IsSnakeCaseLower(input string) bool {
	for _, char := range input {
		if !((char >= 'a' && char <= 'z') || char == '_' || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// generates a password for Cognito and RDS
// satisfies common requirements
// - atleast 8 characters
// - 1 lowercase, uppercase, number and special chars
func RandomPassword(pwLen int) string {
	length := max(pwLen, 8)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// characters that are valid for both Cognito and RDS
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	spl := "!#$%^&*()_+"
	allChars := lowercase + uppercase + numbers + spl

	password := make([]byte, length)

	// at least one character from each required set
	password[0] = lowercase[r.Intn(len(lowercase))]
	password[1] = uppercase[r.Intn(len(uppercase))]
	password[2] = numbers[r.Intn(len(numbers))]
	password[3] = spl[r.Intn(len(spl))]

	for i := 4; i < length; i++ {
		password[i] = allChars[r.Intn(len(allChars))]
	}

	return string(password)
}

/********************
* Pointer Converters
*********************/

func ToMapString(in map[string]*string) map[string]string {
	out := map[string]string{}
	for key, val := range in {
		out[key] = PtToString(val)
	}
	return out
}

func ToMapPtString(in map[string]string) map[string]*string {
	out := map[string]*string{}
	for key, val := range in {
		out[key] = &val
	}
	return out
}

func PtToString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func PtToInt64(in *int64) int64 {
	if in != nil {
		return *in
	}
	return 0
}

func PtToFloat64(in *float64) float64 {
	if in != nil {
		return *in
	}
	return float64(0.0)
}

func ConvertFloatStringToIntString(floatStr string) (string, error) {
	// Convert the float string to a float64 value
	floatVal, err := strconv.ParseFloat(floatStr, 64)
	if err != nil {
		return "", err
	}

	// Convert the float64 value to an integer
	intVal := int(floatVal)

	// Convert the integer to a string
	intStr := strconv.Itoa(intVal)

	return intStr, nil
}

/********
* Test Utils
**********/

func GetTestRegion() string {
	//getting region for local testing
	region := os.Getenv("TAH_CORE_REGION")
	if region == "" {
		//getting region for codeBuild project
		return os.Getenv("AWS_REGION")
	}
	return region
}
