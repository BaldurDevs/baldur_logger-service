package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"log-service/cmd/api/data"
	"net/http"
	"os"
)

type Config struct {
	Models data.Models
}

func main() {
	// Connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	// Ping to mongo
	log.Println("Pinging to mongo...")
	mongoErr := mongoClient.Ping(context.TODO(), nil)
	if mongoErr != nil {
		log.Println(mongoErr)
		return
	}
	log.Println("Ping successful!!")

	// Create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), data.TimeOutInterval)
	defer cancel()

	// Close connection
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(mongoClient),
	}

	// Start web server
	app.serve()

}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Println("Starting service on port", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}
}

func connectToMongo() (*mongo.Client, error) {
	var mongoUrl string
	godotenv.Load(".env")

	mongoUrl = os.Getenv("ATLAS_URL")
	if mongoUrl == "" {
		mongoUrl = fmt.Sprintf("mongodb://%s:%s@%s:%d", data.DbUser, data.DbPassword, data.Host, data.MongoPort)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Println("Error creating new mongo client...")
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), data.TimeOutInterval)
	err = client.Connect(ctx)
	if err != nil {
		log.Println("Error connecting mongo...")
		return nil, err
	}

	return client, nil
}
