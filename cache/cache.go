package cache

import (
	"app/configuration"
	"app/service"
	"context"
	"encoding/json"
	// "errors"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type Cache struct {
	TotalCases  int
	LastUpdated time.Time
}

func Get(state string) (error, []service.ActiveCovidCasesResponse) {
	var activeCovidCasesResponseList = make([]service.ActiveCovidCasesResponse, 0)

	var stateWideData Cache
	var countryWideData Cache

	stateResult, stateErr := configuration.RedisClient.Get(context.Background(), state).Result()

	countryResult, countryError := configuration.RedisClient.Get(context.Background(), "Entire Country").Result()

	if stateErr != nil && stateErr != redis.Nil || countryError != nil && countryError != redis.Nil {
		if stateErr != nil {
			log.Println("find: redis error: ", stateErr)
			return stateErr, activeCovidCasesResponseList
		}

		log.Println("find: redis error: ", countryError)
		return countryError, activeCovidCasesResponseList
	}

	if stateResult == "" || countryResult == "" {
		log.Println("cache miss")
		return nil, activeCovidCasesResponseList
	}

	if err := json.Unmarshal([]byte(stateResult), &stateWideData); err != nil {
		return err, activeCovidCasesResponseList
	}

	if err := json.Unmarshal([]byte(countryResult), &countryWideData); err != nil {
		return err, activeCovidCasesResponseList
	}

	activeCovidCasesResponseList = append(activeCovidCasesResponseList, service.ActiveCovidCasesResponse{
		State:       state,
		TotalCases:  stateWideData.TotalCases,
		LastUpdated: stateWideData.LastUpdated,
	})

	activeCovidCasesResponseList = append(activeCovidCasesResponseList, service.ActiveCovidCasesResponse{
		State:       "Entire Country",
		TotalCases:  countryWideData.TotalCases,
		LastUpdated: countryWideData.LastUpdated,
	})

	log.Println("cache hit")

	return nil, activeCovidCasesResponseList
}

func Set(activeCovidCasesResponseList []service.ActiveCovidCasesResponse) error {

	var stateWideData, stateWideDataErr = json.Marshal(Cache{
		TotalCases:  activeCovidCasesResponseList[0].TotalCases,
		LastUpdated: activeCovidCasesResponseList[0].LastUpdated,
	})

	var countryWideData, countryWideDataErr = json.Marshal(Cache{
		TotalCases:  activeCovidCasesResponseList[1].TotalCases,
		LastUpdated: activeCovidCasesResponseList[1].LastUpdated,
	})

	if stateWideDataErr != nil || countryWideDataErr != nil {
		if stateWideDataErr != nil {
			log.Println("error in marshalling :", stateWideDataErr)
			return stateWideDataErr
		}

		log.Println("error in marshalling :", countryWideDataErr)
		return countryWideDataErr
	}

	var err = configuration.RedisClient.Set(context.Background(), activeCovidCasesResponseList[0].State, stateWideData, 30*time.Minute).Err()
	if err != nil {
		log.Println("error in setting state data in redis: ", err)
		return err
	}

	err = configuration.RedisClient.Set(context.Background(), activeCovidCasesResponseList[1].State, countryWideData, 30*time.Minute).Err()
	if err != nil {
		log.Println("error in setting country data in redis: ", err)
		return err
	}

	return nil
}
