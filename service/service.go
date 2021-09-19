package service

import (
	"app/dao"
	"app/helper"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MyGovAPIResponse struct {
	State      map[int]string `json:"Name of State / UT"`
	TotalCases map[int]string `json:"Total Confirmed cases"`
}

type ActiveCovidCasesResponse struct {
	State       string    `json:"state"`
	TotalCases  int       `json:"totalCases"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type StateWiseCovidCases struct {
	State       string
	TotalCases  int
	LastUpdated int64
}

func MapToCovidResponse(stateWiseCovidCasesData []dao.StateWiseCovidCasesData) []ActiveCovidCasesResponse {
	var activeCovidCasesResponseList []ActiveCovidCasesResponse

	for i := 0; i < len(stateWiseCovidCasesData); i++ {
		activeCovidCasesResponseList = append(activeCovidCasesResponseList, ActiveCovidCasesResponse{
			State:       stateWiseCovidCasesData[i].State,
			TotalCases:  stateWiseCovidCasesData[i].TotalCases,
			LastUpdated: time.Unix(stateWiseCovidCasesData[i].LastUpdated, 0),
		})
	}

	return activeCovidCasesResponseList
}

func getHttpClientAndRequestForCovidData(time int64) (*http.Client, *http.Request) {
	req, err := http.NewRequest("GET", "https://www.mygov.in/sites/default/files/covid/covid_state_counts_ver1.json", nil)

	if err != nil {
		log.Println("error in http client request creation: ", err)
	}

	httpClient := &http.Client{}

	q := req.URL.Query()
	q.Add("timestamp", strconv.FormatInt(time, 10))
	req.URL.RawQuery = q.Encode()

	return httpClient, req
}

func fetchCovidData() (int, []StateWiseCovidCases) {
	var myGovAPIResponse MyGovAPIResponse
	var currentTime int64 = helper.GetCurrentEpochTime()
	client, req := getHttpClientAndRequestForCovidData(currentTime)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("error in making api req to www.mygov.in: ", err)
		// return c.String(http.StatusInternalServerError, "Error occurred")
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&myGovAPIResponse); err != nil {
		log.Println("error in decoding api response from www.mygov.in: ", err)
	}

	var StateWiseCovidCasesList []StateWiseCovidCases
	var totalActiveCases int

	for k, v := range myGovAPIResponse.State {

		totalCases, err := strconv.Atoi(myGovAPIResponse.TotalCases[k])
		if err != nil {
			log.Println("error in converting myGovAPIResponse.TotalCases to integer: ", err)
		}

		var activeCaseResponse = StateWiseCovidCases{
			State:       v,
			TotalCases:  totalCases,
			LastUpdated: currentTime,
		}

		totalActiveCases += totalCases
		StateWiseCovidCasesList = append(StateWiseCovidCasesList, activeCaseResponse)
	}

	return totalActiveCases, StateWiseCovidCasesList
}

func GetCovidDetails(c echo.Context, state string) (error, []dao.StateWiseCovidCasesData) {

	var result []dao.StateWiseCovidCasesData

	err, stateDetails := dao.FetchRecordsForSpecificState(c, state)

	if err != nil {
		log.Println("fatal error in fetching state specific record: ", err)
		return err, result
	}

	result = append(result, stateDetails)

	err, countryWideDetails := dao.FetchRecordsForSpecificState(c, "Entire Country")
	if err != nil {
		log.Println("fatal error in fetching record for entire country: ", err)
		return err, result
	}

	result = append(result, countryWideDetails)

	return err, result
}

// Cron job to refresh data in Mongo every 1 hr
func createCronJob() {
	c := cron.New()
	c.AddFunc("@every 60m", func() {
		log.Println("Every 1 hr job\n")
		totalActiveCases, stateWiseCovidCasesList := fetchCovidData()

		var stateWiseCovidCasesData []dao.StateWiseCovidCasesData

		for i := 0; i < len(stateWiseCovidCasesList); i++ {
			var covidCases = dao.StateWiseCovidCasesData{
				State:       stateWiseCovidCasesList[i].State,
				TotalCases:  stateWiseCovidCasesList[i].TotalCases,
				LastUpdated: stateWiseCovidCasesList[i].LastUpdated,
			}

			stateWiseCovidCasesData = append(stateWiseCovidCasesData, covidCases)
		}

		if len(stateWiseCovidCasesList) >= 0 {
			var countryWide = dao.StateWiseCovidCasesData{
				State:       "Entire Country",
				TotalCases:  totalActiveCases,
				LastUpdated: stateWiseCovidCasesList[0].LastUpdated,
			}

			stateWiseCovidCasesData = append(stateWiseCovidCasesData, countryWide)
		}

		var allRecords []dao.StateWiseCovidCasesData = dao.FetchRecordsForAllStates()
		dao.StoreInDB(stateWiseCovidCasesData, allRecords)
	})

	// Start cron with one scheduled job
	log.Println("Start cron")
	c.Start()
}

func init() {
	createCronJob()
}
