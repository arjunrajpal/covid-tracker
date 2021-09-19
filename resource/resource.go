package resource

import (
	"app/cache"
	"app/configuration"
	"app/helper"
	"app/service"
	"context"
	"github.com/labstack/echo/v4"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
)

type Response struct {
	Error                        string                             `json:"error"`
	ActiveCovidCasesResponseList []service.ActiveCovidCasesResponse `json:"activeCovidCases"`
}

// @Description Takes in the lat/lng of the user and returns the total active cases of that state along with the total active cases of the entire country
// @Accept  json
// @Produce  json
// @Param lat query float32 true "latitude" Format(float)
// @Param lng query float32 true "longitude" Format(float)
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/getActiveCases [get]
func GetActiveCases(c echo.Context) error {

	var activeCovidCasesResponseList = make([]service.ActiveCovidCasesResponse, 0)

	lat := c.QueryParam("lat")
	lng := c.QueryParam("lng")

	// Validate inputs
	var isValid, latitude, longitude = helper.ConvertToFloat64(lat, lng)

	if !isValid {
		log.Println("latitude and longitude cannot be converted into float64 error")
		return c.JSON(http.StatusBadRequest, Response{Error: "Invalid latitude and longitude", ActiveCovidCasesResponseList: activeCovidCasesResponseList})
	}

	r := &maps.GeocodingRequest{
		LatLng: &maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		},
	}

	// Fetch Reverse geocoding details
	route, err := configuration.GoogleMapsClient.ReverseGeocode(context.Background(), r)
	if err != nil {
		log.Println("fatal error: %s", err)
		return c.JSON(http.StatusInternalServerError, Response{Error: "Internal Error occurred", ActiveCovidCasesResponseList: activeCovidCasesResponseList})
	}

	var state, country string
	var countryFound, stateFound bool = false, false

	// Extract state and country
	for i := 0; i < len(route); i++ {
		for j := 0; j < len(route[i].AddressComponents); j++ {
			if route[i].AddressComponents[j].Types[0] == "administrative_area_level_1" {
				state = route[i].AddressComponents[j].LongName
				log.Println("state: ", state)
				stateFound = true
			}

			if route[i].AddressComponents[j].Types[0] == "country" {
				country = route[i].AddressComponents[j].LongName
				log.Println("country: ", country)
				countryFound = true
			}

			if stateFound && countryFound {
				break
			}
		}

		if stateFound && countryFound {
			break
		}
	}

	if country != "India" {
		return c.JSON(http.StatusBadRequest, Response{Error: "Coordinates do not lie in India", ActiveCovidCasesResponseList: activeCovidCasesResponseList})
	}

	// Fetch details from cache
	err, activeCovidCasesResponseList = cache.Get(state)
	if err != nil || len(activeCovidCasesResponseList) == 0 {
		// Fetch covid details for specific state from DB in case of cache miss
		err, covidResult := service.GetCovidDetails(c, state)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Error: "Internal error occurred", ActiveCovidCasesResponseList: activeCovidCasesResponseList})
		}

		activeCovidCasesResponseList = service.MapToCovidResponse(covidResult)

		cache.Set(activeCovidCasesResponseList)
	}

	return c.JSON(http.StatusOK, Response{Error: "", ActiveCovidCasesResponseList: activeCovidCasesResponseList})
}
