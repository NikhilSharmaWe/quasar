package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	DATABASE_NAME   string
	MONGO_URL       string
	SESSION_SEC_KEY string
)

type application struct {
	ctx      context.Context
	logger   *log.Logger
	userRepo *GenericRepo[User]
}

func main() {
	err := godotenv.Load("vars.env")
	if err != nil {
		log.Fatal("failed to load .env file")
	}

	DATABASE_NAME = os.Getenv("DATABASE_NAME")
	MONGO_URL = os.Getenv("MONGO_URL")
	SESSION_SEC_KEY = os.Getenv("SESSION_SECRET_KEY")

	var (
		ctx      = context.Background()
		logger   = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		userRepo = GenericRepo[User]{
			Collection: "user",
		}
	)

	app := &application{
		ctx:      ctx,
		logger:   logger,
		userRepo: &userRepo,
	}

	// err := userRepo.SaveAccount(&User{Username: "Tushti"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// res, err := userRepo.Find(bson.M{"username": "Tushti"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("%+v\n", res)

	mux := app.router()

	log.Fatal(mux.Start(":8080"))
}
