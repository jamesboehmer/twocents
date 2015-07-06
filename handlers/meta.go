package handlers

import (
	"net/http"
	"encoding/json"
)
type MetaVersions struct {
	Versions []string    `json:"versions"`
}

type MetaDictionaries struct {
	Dictionaries []string    `json:"dictionaries"`
}

func MetaDictionariesHandler(w http.ResponseWriter, r *http.Request) {

	keys := []string{}
	for k, _ := range DictionaryMap {
		keys = append(keys, k)
	}

	dictionaries := MetaDictionaries{
		Dictionaries: keys,
	}

	j, _ := json.Marshal(dictionaries)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if AllowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", AllowedOrigin)
	}
	w.Write(j)

}

func MetaVersionsHandler(w http.ResponseWriter, r *http.Request) {

	versionNumbers := []string{"v1"}

	versions := MetaVersions{
		Versions : versionNumbers,
	}

	j, _ := json.Marshal(versions)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if AllowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", AllowedOrigin)
	}
	w.Write(j)

}
