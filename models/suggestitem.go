package models

import (
)

type SuggestItem struct {
	Term   string        `json:"term"`
	Weight int            `json:"weight"`
}

// Comparator function (sort by weights)
func SuggestItemComparator(a, b interface{}) int {

	// Type assertion, program will panic if this is not respected
	c1 := a.(*SuggestItem)
	c2 := b.(*SuggestItem)

	switch {
	case c1.Weight > c2.Weight:
		return -1
	case c1.Weight < c2.Weight:
		return 1
	default:
		switch {
		case c1.Term > c2.Term:
			return 1
		case c1.Term < c2.Term:
			return -1
		default:
			return 0
		}
	}
}