package main

import (
	"log"
	"os"
	"time"

	"github.com/NikhilSharmaWe/quasar/router"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("vars.env")
	if err != nil {
		log.Fatal("failed to load .env file")
	}
}

func main() {
	app := router.NewApplication()

	mux := app.Router()

	go func() {
		for range time.NewTicker(time.Second * 3).C {
			app.DispatchKeyFrame()
		}
	}()

	go app.HandleMessages()

	log.Fatal(mux.Start(os.Getenv("ADDR")))
}
