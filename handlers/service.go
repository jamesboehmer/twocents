package handlers

import (
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/tchap/go-patricia/patricia"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/jamesboehmer/twocents/models"
	"strings"
	"log"
)

// This maps dictionary names to prefix tries.  All threads can read this,
// but only the admin thread will be updating it by building a new one and
// atomically swapping
var DictionaryMap = make(map[string]*patricia.Trie)

func ImportDictionaries() map[string][]*models.SuggestItem {
	var itemMap = make(map[string][]*models.SuggestItem)

	//TODO: import dictionaries from files or database
	var suggestItem = new(models.SuggestItem)
	suggestItem.Term = "Foo Bar"
	suggestItem.Weight = 100

	var suggestItem2 = new(models.SuggestItem)
	suggestItem2.Term = "Foobar"
	suggestItem2.Weight = 101

	var suggestItem3 = new(models.SuggestItem)
	suggestItem3.Term = "Foø Bar Baz"
	suggestItem3.Weight = 99

	itemMap["default"] = []*models.SuggestItem{suggestItem, suggestItem2, suggestItem3}

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
	//TODO: parse int from mux vars
	limit := 10

	//TODO: use filter from mux vars

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
	for finalSuggestionSetPosition < limit {
		for _, suggestionSetItem := range trieItems {
			if suggestionSetItem.Size() > finalSuggestionSetPosition {
				thisItem := suggestionSetItem.Values()[finalSuggestionSetPosition]
				//TODO: use filter parameter
				collatedSuggestionSet.Add(thisItem)
			}
		}
		finalSuggestionSetPosition++
	}

	suggestions := []string{}
	for _, suggestion := range collatedSuggestionSet.Values() {
		suggestions = append(suggestions, suggestion.(*models.SuggestItem).Term)
	}

	t := TwoCentsV1{
		Suggestions: suggestions,
	}

	j, _ := json.Marshal(t)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)

}