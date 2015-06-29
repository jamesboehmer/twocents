package models

import (
	"log"
)

type SuggestItem struct {
	Term   string        `json:"term"`
	Weight int            `json:"weight"`
}

// Comparator function (sort by weights)
func SuggestItemComparator(a, b interface{}) int {

	// Type assertion, program will panic if this is not respected
	c1 := a.(SuggestItem)
	c2 := b.(SuggestItem)

	log.Print("Item 1:" + c1.Term)
	log.Print("Item 2:" + c2.Term)
	switch {
	case c1.Weight > c2.Weight:
		return 1
	case c1.Weight < c2.Weight:
		return -1
	default:
		return 0
	}
}