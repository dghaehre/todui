package main

import (
	"strings"
)

func Contains[T comparable](list []T, x T) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

// THE filter function
func filterContents(list []Todo, filter string) []Todo {
	var newList = make([]Todo, 0, len(list))
	for _, t := range list {
		if strings.Contains(t.desc, filter) {
			newList = append(newList, t)
		}
	}
	return newList
}
