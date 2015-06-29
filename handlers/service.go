package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
)

type TwoCentsV1 struct {
	Suggestions 	[]string	`json:"suggestions"`
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