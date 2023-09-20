package main

import (
	"log"

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
	// t := &router.Template{
	// 	Templates: template.Must(template.ParseGlob("./public/*.html")),
	// }
	// app.Template = t

	mux := app.Router()
	log.Fatal(mux.Start(":8080"))
}
