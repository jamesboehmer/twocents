package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"github.com/jamesboehmer/twocents/handlers"
	"flag"
)

var defaultDataDirectory = "."
var defaultServiceListenPort = 8080

func main() {
	dataDirectory := flag.String("d", defaultDataDirectory, "Directory where dictionary files are located")
	serviceListenPort := flag.Int("p", defaultServiceListenPort, "Service Listen Port")
	flag.Parse()

	log.Printf("Using data directory %s", *dataDirectory)
	handlers.DataDirectory = *dataDirectory
	handlers.LoadDictionaries()

	r := mux.NewRouter()
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}/{filter}", handlers.TwoCentsHandlerV1).Methods("GET")
	http.Handle("/", r)
	log.Printf("Service listening on port %d", *serviceListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serviceListenPort), nil))
}

