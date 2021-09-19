package configuration

import (
	"context"
	"crypto/tls"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"googlemaps.github.io/maps"
	"log"
	"os"
	"time"
)

func getMongoClient() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://arajpal:" + os.Getenv("MONGO_PASSWORD") + "@active-covid-cases-1.ty5vb.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatalln("fatal error in creating mongo client: ", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalln("fatal error in connecting to mongo: ", err)
	}

	log.Println("Connection to Mongo established......")

	return client
}

func getGoogleMapsClient() *maps.Client {
	googleMapsClient, err := maps.NewClient(maps.WithAPIKey(os.Getenv("GOOGLE_MAPS_API_TOKEN")))
	if err != nil {
		log.Fatalln("fatal error in creating google maps client: ", err)
	}

	log.Println("Connection to GoogleMaps established......")

	return googleMapsClient
}

func getRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "ec2-54-82-190-200.compute-1.amazonaws.com:7950",
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	log.Println("Connection to Redis established......")

	return rdb
}

var MongoClient *mongo.Client
var GoogleMapsClient *maps.Client
var RedisClient *redis.Client

func init() {
	MongoClient = getMongoClient()
	GoogleMapsClient = getGoogleMapsClient()
	RedisClient = getRedisClient()
}
