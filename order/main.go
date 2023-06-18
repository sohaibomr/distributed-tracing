package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla/v2"
)

func main() {
	fmt.Println("Starting server...")
	apmUrl := os.Getenv("ELASTIC_APM_SERVER_URL")
	fmt.Println(apmUrl)
	r := mux.NewRouter()
	apmgorilla.Instrument(r)
	r.HandleFunc("/order", placeNewOrder)
	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), r))
}
