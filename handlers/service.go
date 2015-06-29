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

	itemMap["default"] = []*models.SuggestItem{suggestItem, suggestItem2}

	return itemMap
}

func LoadDictionaries() {
	var newDictionaryMap = make(map[string]*patricia.Trie)
	var itemMap = ImportDictionaries()

	for dictionaryName, suggestItems := range itemMap {
		log.Print("Loading dictionary " + dictionaryName)
		// First see if the trie already exists
		var trie, ok = newDictionaryMap[dictionaryName]
		if !ok {
			trie = patricia.NewTrie()
		}

		// Great, we have a trie, now let's see if prefixes for the
		// suggestItems exist in the trie
		for _, suggestItem := range suggestItems {
			//Tokenize the suggested term by whitespace.  Each token will become a prefix in the trie
			var tokens = strings.Fields(suggestItem.Term)
			for _, token := range tokens {
				lowerToken := strings.ToLower(token)
				// The values in the trie are sorted sets of SuggestItems
				trieItem := trie.Get([]byte(lowerToken))
				if trieItem != nil {
					suggestItemSet := trieItem.(treeset.Set)
					//If the set already exists, add the new suggestion to the set
					suggestItemSet.Add(patricia.Prefix([]byte(lowerToken)))

				} else {
					// Otherwise create a new set, add the SuggestItem, and insert it into
					// the trie using the lowercase token for the prefix
					suggestItemSet := treeset.NewWith(models.SuggestItemComparator)
					suggestItemSet.Add(suggestItem)
					trie.Insert(patricia.Prefix([]byte(lowerToken)), suggestItemSet)
				}
			}
		}
		newDictionaryMap[dictionaryName] = trie
		log.Print("Dictionary " + dictionaryName + " loaded")
	}
	//Atomic swap
	DictionaryMap = newDictionaryMap
	log.Print("All dictionaries updated")
}

type TwoCentsV1 struct {
	Suggestions []string    `json:"suggestions"`
}

func TwoCentsHandlerV1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dictionaryName := vars["dictionary"]
	dictionary, found := DictionaryMap[dictionaryName]
	if !found {
		http.NotFound(w, r)
		return
	}
	query := vars["query"]
	if &dictionary == nil {

	}

	// TODO: lookup suggestions from a prefix trie, collate, filter, and sort
//	someItems := []*patricia.Item{}
//
//	someFunc := func(prefix patricia.Prefix, item patricia.Item) error {
//		someItems = append(someItems, &item)
//		return nil
//	}
//	trie.Visit(someFunc)
//
//	log.Print(someItems)

	t := TwoCentsV1{
		Suggestions: []string{query},
	}

	j, _ := json.Marshal(t)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)

}