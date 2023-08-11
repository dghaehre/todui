package main

import "time"

// Implements sort.Interface for []Todo based on priority

type ByPriority []Todo

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

type ByDue []Todo

func (a ByDue) Len() int      { return len(a) }
func (a ByDue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDue) Less(i, j int) bool {
	if a[i].Due.Date == "" && a[j].Due.Date == "" {
		return a[i].Priority > a[j].Priority
	}
	ti, err := time.Parse("2006-01-02", a[i].Due.Date)
	if err != nil {
		return false
	}
	tj, err := time.Parse("2006-01-02", a[j].Due.Date)
	if err != nil {
		return true
	}
	return ti.Before(tj)
}

// TODO: add default sorter that sort by due date, then priority, then time created
