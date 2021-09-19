package dao

import (
	"app/configuration"
	"context"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
)

type StateWiseCovidCasesData struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	State       string             `bson:"state,omitempty"`
	TotalCases  int                `bson:"totalCases,omitempty"`
	LastUpdated int64              `bson:"lastUpdated,omitempty"`
}

func FetchRecordsForAllStates() []StateWiseCovidCasesData {
	database := configuration.MongoClient.Database("covid-cases")
	stateWiseCovidCasesDataCollection := database.Collection("state-wise-active-cases")

	var stateWiseCovidCasesData []StateWiseCovidCasesData

	cursor, err := stateWiseCovidCasesDataCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Println("error in fetching all records from mongo: ", err)
	}
	if err = cursor.All(context.Background(), &stateWiseCovidCasesData); err != nil {
		log.Println("error in converting to []StateWiseCovidCasesData: ", err)
	}

	return stateWiseCovidCasesData
}

func FetchRecordsForSpecificState(c echo.Context, state string) (error, StateWiseCovidCasesData) {
	database := configuration.MongoClient.Database("covid-cases")
	stateWiseCovidCasesDataCollection := database.Collection("state-wise-active-cases")

	var result StateWiseCovidCasesData

	err := stateWiseCovidCasesDataCollection.FindOne(
		context.Background(),
		bson.D{{"state", state}},
	).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			log.Println("No match found: ", err)
		}
		log.Println("error on searching: ", err)
	} else {
		log.Println("Found document: ", result)
	}

	return err, result
}

func StoreInDB(stateWiseCovidCasesData []StateWiseCovidCasesData, allRecords []StateWiseCovidCasesData) {
	database := configuration.MongoClient.Database("covid-cases")
	stateWiseCovidCasesDataCollection := database.Collection("state-wise-active-cases")

	var stateWiseCovidCasesDataMapByState map[string]StateWiseCovidCasesData = make(map[string]StateWiseCovidCasesData)

	for i := 0; i < len(allRecords); i++ {
		stateWiseCovidCasesDataMapByState[allRecords[i].State] = allRecords[i]
	}

	for i := 0; i < len(stateWiseCovidCasesData); i++ {

		value, found := stateWiseCovidCasesDataMapByState[stateWiseCovidCasesData[i].State]

		if !found {
			insertResult, err := stateWiseCovidCasesDataCollection.InsertOne(context.Background(), stateWiseCovidCasesData[i])
			if err != nil {
				log.Println("fatal error in inserting a document in mongo: ", err)
			}
			log.Println("Inserted a new document for "+stateWiseCovidCasesData[i].State+": ", insertResult.InsertedID)
		} else {
			documentId := value.ID
			value = stateWiseCovidCasesData[i]
			value.ID = documentId
			updateResult, err := stateWiseCovidCasesDataCollection.UpdateByID(context.Background(), value.ID, bson.D{{"$set", bson.D{{"totalCases", value.TotalCases}, {"lastUpdated", value.LastUpdated}}}})
			if err != nil {
				log.Println("fatal error in updating a document in mongo: ", err)
			}

			log.Println("No of documents matched: " + strconv.FormatInt(updateResult.MatchedCount, 10))
			log.Println("No of documents updated: " + strconv.FormatInt(updateResult.ModifiedCount, 10))
		}
	}

}

var database *mongo.Database
var stateWiseCovidCasesDataCollection *mongo.Collection

func init() {
	database = configuration.MongoClient.Database("covid-cases")
	stateWiseCovidCasesDataCollection = database.Collection("state-wise-active-cases")
}
