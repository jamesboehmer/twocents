package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jamesboehmer/twocents/handlers"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

var defaultDataDirectory = "./data"
var defaultServiceListenPort = 8080
var defaultAdminListenPort = 8081
var reloadLock uint32 = 0

func main() {
	dataDirectory := flag.String("d", defaultDataDirectory, "Directory where dictionary files are located")
	serviceListenPort := flag.Int("p", defaultServiceListenPort, "Service Listen Port")
	adminListenPort := flag.Int("a", defaultAdminListenPort, "Admin Listen Port")
	allowedOrigin := flag.String("c", handlers.AllowedOrigin, "CORS Allowed Origin")
	useQuicksort := flag.Bool("q", false, "Use Quicksort (default is InsertionSort")
	flag.Parse()
	handlers.UseQuicksort = *useQuicksort

	log.Printf("Using data directory %s", *dataDirectory)
	handlers.DataDirectory = *dataDirectory
	handlers.LoadDictionaries()

	r := mux.NewRouter()
	r.HandleFunc("/twocents", handlers.MetaVersionsHandler).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}", handlers.MetaDictionariesHandler).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}", handlers.TwoCentsHandlerV1).Methods("GET")
	r.HandleFunc("/twocents/{version:v1}/{dictionary}/{query}/{limit:[0-9]+}/{filter}", handlers.TwoCentsHandlerV1).Methods("GET")
	http.Handle("/", r)
	log.Printf("Service listening on port %d", *serviceListenPort)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serviceListenPort), nil))
	}()
	log.Printf("Admin listening on port %d", *adminListenPort)
	if *allowedOrigin != "" {
		log.Printf("Setting CORS Origin to %s", *allowedOrigin)
		handlers.AllowedOrigin = *allowedOrigin
	}

	//TODO: Find a better way to pass a different router to the admin server
	http.ListenAndServe(fmt.Sprintf(":%d", *adminListenPort), &adminHandler{})
}

type adminHandler struct {
}

func (m *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.RequestURI, "/twocents/admin/reload") {
		http.NotFound(w, r)
		return
	}
	// Ensure that multiple reloads never occur simultaneously, and that multiple requests don't queue up.
	// The sync library doesn't seem to have a trylock mechanism, so use atomic CAS
	locked := atomic.CompareAndSwapUint32(&reloadLock, 0, 1)
	if locked {
		w.Write([]byte("Dicionary reload initiated\n"))
		go func() {
			defer atomic.StoreUint32(&reloadLock, 0)
			handlers.LoadDictionaries()
		}()

	} else {
		w.Write([]byte("Dictionary reload already in progress\n"))
	}

}
