package models

import (
	"sort"
	"strings"
)

// A Pipeline has a name and series of steps
type Pipeline struct {
	Name        string
	CurrentStep int
	Steps       []string
}

// The Pipelines object is used to manage Pipelines in a db.Store
type Pipelines []*Pipeline

// Get will return the pipeline with the matching name.
// The name is not case sensitive.
// A bool is also returned denoting if the pipeline exists or not.
func (p Pipelines) Get(name string) (*Pipeline, bool) {
	name = strings.ToLower(name)
	for _, pipeline := range p {
		if strings.ToLower(pipeline.Name) == name {
			return pipeline, true
		}
	}

	return nil, false
}

// Delete will delete the pipeline with the matching name.
// The name is not case sensitive.
// A bool is also returned denoting if the pipeline existed or not.
func (p *Pipelines) Delete(name string) bool {
	name = strings.ToLower(name)
	for i := 0; i < len((*p)); i++ {
		if strings.ToLower((*p)[i].Name) == name {
			*p = append((*p)[:i], (*p)[i+1:]...)
			return true
		}
	}

	return false
}

// Sort will sort the pipelines by their name.
// If ascending is true, names are sorted by alphabetical order.
// If ascending is false, names are sorted by reverse alphabetical order.
func (p Pipelines) Sort(ascending bool) {
	if ascending {
		sort.Sort(p)
	} else {
		sort.Sort(sort.Reverse(p))
	}
}

// Len is a method to satisfy sort.Interface
func (p Pipelines) Len() int {
	return len(p)
}

// Swap is a method to satisfy sort.Interface
func (p Pipelines) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less is a method to satisfy sort.Interface
func (p Pipelines) Less(i, j int) bool {
	return p[i].Name < p[j].Name
}
