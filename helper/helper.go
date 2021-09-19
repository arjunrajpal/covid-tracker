package helper

import (
	"log"
	"strconv"
	"time"
)

func ConvertToFloat64(lat string, lng string) (bool, float64, float64) {
	latitude, latitudeErr := strconv.ParseFloat(lat, 64)
	longitude, longitudeErr := strconv.ParseFloat(lng, 64)

	return latitudeErr == nil && longitudeErr == nil, latitude, longitude
}

func GetCurrentEpochTime() int64 {
	now := time.Now()
	secs := now.Unix()

	log.Println("Time ", secs)
	return secs
}
