package main

import "time"

// Implements sort.Interface for []Todo based on priority

type ByPriority []Todo

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

type ByDueThenPriority []Todo

func (a ByDueThenPriority) Len() int      { return len(a) }
func (a ByDueThenPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDueThenPriority) Less(i, j int) bool {
	if a[i].Due.Date == "" && a[j].Due.Date == "" {
		return a[i].Priority > a[j].Priority
	}
	ti, erri := time.Parse("2006-01-02", a[i].Due.Date)
	tj, errj := time.Parse("2006-01-02", a[j].Due.Date)
	if erri != nil && errj != nil {
		return a[i].Priority > a[j].Priority
	}
	if errj != nil {
		return true
	}
	if erri != nil {
		return false
	}
	if ti.Equal(tj) {
		return a[i].Priority > a[j].Priority
	}
	return ti.Before(tj)
}
