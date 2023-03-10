package utils

import (
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
