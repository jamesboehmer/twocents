package handlers

import (
	"net/http"
	"os"
	"io/ioutil"
	"strconv"
	"encoding/csv"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/tchap/go-patricia/patricia"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/jamesboehmer/twocents/models"
	"strings"
	"log"
	"fmt"
)

// This maps dictionary names to prefix tries.  All threads can read this,
// but only the admin thread will be updating it by building a new one and
// atomically swapping
var DictionaryMap = make(map[string]*patricia.Trie)

var DataDirectory = "."

func ImportDictionaries() map[string][]*models.SuggestItem {
	var itemMap = make(map[string][]*models.SuggestItem)

	fileInfo, err := ioutil.ReadDir(DataDirectory)
	if err != nil {
		 log.Print(err)
		 return itemMap
	}
	numberOfDictionaries := 0
	for _, file := range fileInfo {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			dictionaryFile := fmt.Sprintf("%s%s%s", DataDirectory, string(os.PathSeparator), file.Name())
			dictionaryName := strings.TrimSuffix(file.Name(), ".txt")
			log.Printf("Importing dictionary %s from file %s",dictionaryName, dictionaryFile)

			csvfile, err := os.Open(dictionaryFile)
			if err != nil {
				 log.Print(err)
				 continue
			}
			defer csvfile.Close()
			reader := csv.NewReader(csvfile)
			reader.FieldsPerRecord = 2
			reader.Comma = '|'

			rawCSVdata, err := reader.ReadAll()
			if err != nil {
				log.Print(err)
				continue
			}

			for _, each := range rawCSVdata {
				var suggestItem = new(models.SuggestItem)
				suggestItem.Term = each[0]
				weight, err := strconv.Atoi(each[1])
				if err == nil {
					suggestItem.Weight = weight
					itemMap[dictionaryName] = append(itemMap[dictionaryName], suggestItem)
				}

			}
			numberOfDictionaries++
		}
	}

	log.Printf("Imported %d dictionaries", numberOfDictionaries)
	return itemMap
}

func LoadDictionaries() {
	var newDictionaryMap = make(map[string]*patricia.Trie)
	var itemMap = ImportDictionaries()

	numPrefixes := 0
	numSuggestions := 0
	numDictionaries := 0

	for dictionaryName, suggestItems := range itemMap {
		numDictionaries++
		log.Print("Loading dictionary " + dictionaryName)
		// First see if the trie already exists
		var trie, ok = newDictionaryMap[dictionaryName]
		if !ok {
			trie = patricia.NewTrie()
		}

		// Great, we have a trie, now let's see if prefixes for the
		// suggestItems exist in the trie
		for _, suggestItem := range suggestItems {
			numSuggestions++
			//Tokenize the suggested term by whitespace.  Each token will become a prefix in the trie
			var tokens = strings.Fields(suggestItem.Term)
			for _, token := range tokens {
				numPrefixes++
				//TODO: use ascii folding
				lowerToken := strings.ToLower(token)
				// The values in the trie are sorted sets of SuggestItems
				trieItem := trie.Get([]byte(lowerToken))
				if trieItem != nil {
					suggestItemSet := trieItem.(treeset.Set)
					//If the set already exists, add the new suggestion to the set
					suggestItemSet.Add(suggestItem)

				} else {
					// Otherwise create a new set, add the SuggestItem, and insert it into
					// the trie using the lowercase token for the prefix
					suggestItemSet := treeset.NewWith(models.SuggestItemComparator)
//					log.Printf("Inserting suggestion item %s (%s)", lowerToken, suggestItem.Term)
					suggestItemSet.Add(suggestItem)
					trie.Insert(patricia.Prefix([]byte(lowerToken)), *suggestItemSet)
				}
			}
		}
		newDictionaryMap[dictionaryName] = trie
		log.Print("Dictionary " + dictionaryName + " loaded")
	}
	//Atomic swap
	DictionaryMap = newDictionaryMap
	log.Printf("All dictionaries updated")
}

type TwoCentsV1 struct {
	Suggestions []string    `json:"suggestions"`
}

func TwoCentsHandlerV1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dictionaryName := vars["dictionary"]
	limit := 10
	limitParam := vars["limit"]
	if limitParam != "" {
		limit, _ = strconv.Atoi(limitParam)
	}

	//case-insensitive filter
	filter := vars["filter"]
	if filter != "" {
		filter = strings.ToLower(filter)
	}

	dictionaryTrie, found := DictionaryMap[dictionaryName]
	if !found {
		http.NotFound(w, r)
		return
	}
	//TODO: use ascii folding
	query := vars["query"]

	/*
	The values in the patricia-trie are sets of SuggestItems.  patricia-trie won't return the list of nodes
	for you, but will invoke a function on all visited nodes.  This []treeset.Set will hold the results of the
	visited nodes.  visitorFunc will actually add those sets to that array.
	 */
	trieItems := []treeset.Set{}
	visitorFunc := func(prefix patricia.Prefix, item patricia.Item) error {
		trieItems = append(trieItems, item.(treeset.Set))
		return nil
	}
	dictionaryTrie.VisitSubtree(patricia.Prefix([]byte(strings.ToLower(query))), visitorFunc)

	/*
	This set will hold the SuggestItems we pull from the front of every set retrieve from the patricia-trie.  Since
	it's tree set, the items are sorted using the SuggestItemComparator, which compares by weight and string,
	guaranteeing the items within a set are ordered
	 */
	collatedSuggestionSet := treeset.NewWith(models.SuggestItemComparator)

	//If there were fewer suggestions than the requested limit, lower the limit
	totalSuggestions := 0
	for _, suggestionSetItem := range trieItems {
		totalSuggestions += suggestionSetItem.Size()
	}
	if totalSuggestions < limit {
		limit = totalSuggestions
	}

	/*
	The results from the patrica-trie visit are all sorted sets.  However, they're only sorted within the set.  Since
	we know that they're in weight-descending order, we can reliably pick the first element from each set, and insert
	them into another sorted result set.  After <limit> iterations, we're guaranteed to have the top weighted items
	in weight-descending order, and we only need to slice the array
	 */
	finalSuggestionSetPosition := 0
	for finalSuggestionSetPosition < limit && collatedSuggestionSet.Size() < limit {
		for _, suggestionSetItem := range trieItems {
			if suggestionSetItem.Size() > finalSuggestionSetPosition {
				thisItem := suggestionSetItem.Values()[finalSuggestionSetPosition].(*models.SuggestItem)
				//case-insensitive filter
				if filter != "" {
					if strings.Contains(strings.ToLower(thisItem.Term), filter) {
						collatedSuggestionSet.Add(thisItem)
					}
				} else {
					collatedSuggestionSet.Add(thisItem)
				}

			}
		}
		finalSuggestionSetPosition++
	}

	if len(collatedSuggestionSet.Values()) < limit {
		limit = len(collatedSuggestionSet.Values())
	}
	suggestions := []string{}
	for _, suggestion := range collatedSuggestionSet.Values()[0:limit] {
		suggestions = append(suggestions, suggestion.(*models.SuggestItem).Term)
	}

	t := TwoCentsV1{
		Suggestions: suggestions,
	}

	j, _ := json.Marshal(t)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)

}