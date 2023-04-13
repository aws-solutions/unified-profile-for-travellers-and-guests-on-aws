package utils

import (
	"errors"
	"log"
	"strconv"
	"time"
)

func ParseFloat64(val string) float64 {
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing float: %s", err)
		return 0
	}
	return parsed
}

func ParseInt64(val string) int64 {
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("[WARNING] error while parsing int: %s", err)
		return 0
	}
	return parsed
}

func ParseInt(val string) int {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("[WARNING] error while parsing int: %s", err)
		return 0
	}
	return parsed
}

func ParseTime(t string, layout string) time.Time {
	parsed, err := time.Parse(layout, t)
	if err != nil {
		log.Printf("[WARNING] error while parsing date: %s", err)
		return time.Time{}
	}
	return parsed
}

// Convert string to time.
//
// Returns type's default value if the value is empty
// or there is an error while parsing.
func TryParseTime(t, layout string) (time.Time, error) {
	if t == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(layout, t)
	if err != nil {
		return time.Time{}, errors.New("error parsing time: \"%s\"")
	}
	return parsed, nil
}

// Convert string to int.
//
// Returns type's default value if the value is empty
// or there is an error while parsing.
func TryParseInt(val string) (int, error) {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.New("error parsing value \"%s\" to int")
	}
	return parsed, nil
}

func ParseIntArray(array []string) []int {
	res := []int{}
	for _, val := range array {
		res = append(res, ParseInt(val))
	}
	return res
}
