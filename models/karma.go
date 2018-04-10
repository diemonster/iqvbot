package models

import (
	"sort"
)

// KarmaEntry holds information about a specific karma instance
type KarmaEntry struct {
	Upvotes   int
	Downvotes int
}

// The Karma object is used to manage KarmaEntrys in a db.Store
type Karma map[string]KarmaEntry

// SortKeys will return a slice of ordered keys.
// If ascending is true, keys with the lowest karma are returned first.
// If ascending is false, keys with the highest karma are returned first.
func (k Karma) SortKeys(ascending bool) []string {
	sorter := newKarmaSorter(k)
	if ascending {
		sort.Sort(sorter)
	} else {
		sort.Sort(sort.Reverse(sorter))
	}

	return sorter.keys
}

type karmaSorter struct {
	karma Karma
	keys  []string
}

func newKarmaSorter(karma Karma) *karmaSorter {
	keys := make([]string, 0, len(karma))
	for key := range karma {
		keys = append(keys, key)
	}

	return &karmaSorter{
		karma: karma,
		keys:  keys,
	}
}

// Len is a method to satisfy sort.Interface
func (k *karmaSorter) Len() int {
	return len(k.keys)
}

// Swap is a method to satisfy sort.Interface
func (k *karmaSorter) Swap(i, j int) {
	k.keys[i], k.keys[j] = k.keys[j], k.keys[i]
}

// Less is a method to satisfy sort.Interface
func (k *karmaSorter) Less(i, j int) bool {
	entryI := k.karma[k.keys[i]]
	entryJ := k.karma[k.keys[j]]
	return (entryI.Upvotes - entryI.Downvotes) < (entryJ.Upvotes - entryJ.Downvotes)
}
