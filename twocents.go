package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"github.com/jamesboehmer/twocents/handlers"
)

func main() {
	handlers.LoadDictionaries()
	r := mux.NewRouter()
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}/{filter}", handlers.TwoCentsHandlerV1).Methods("GET")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

