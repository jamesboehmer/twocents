package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"log"
)

type TwoCentsV1 struct {
	Suggestions 	[]string	`json:"suggestions"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/twocents/{version:v1}/{query}", TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{query}/{limit:[0-9]+}", TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{query}/{limit:[0-9]+}/{filter}", TwoCentsHandlerV1).Methods("GET")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func TwoCentsHandlerV1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

	// TODO: lookup suggestions from a prefix trie, collate, filter, and sort
	t := TwoCentsV1{
		Suggestions: []string{query},
	}

	j, _ := json.Marshal(t)

	w.Write(j)

}