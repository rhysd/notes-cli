package notes

import (
	"sort"
	"strings"
)

type byCreated []*Note

func (a byCreated) Len() int {
	return len(a)
}
func (a byCreated) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byCreated) Less(i, j int) bool {
	return a[i].Created.Before(a[j].Created)
}

func sortByCreated(n []*Note) {
	sort.Sort(byCreated(n))
}

type byFilename []*Note

func (a byFilename) Len() int {
	return len(a)
}
func (a byFilename) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byFilename) Less(i, j int) bool {
	return strings.Compare(a[i].File, a[j].File) < 0
}

func sortByFilename(n []*Note) {
	sort.Sort(byFilename(n))
}

type byCategory []*Note

func (a byCategory) Len() int {
	return len(a)
}
func (a byCategory) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byCategory) Less(i, j int) bool {
	l, r := a[i], a[j]
	cmp := strings.Compare(l.Category, r.Category)
	if cmp != 0 {
		return cmp < 0
	}
	return strings.Compare(l.File, r.File) < 0
}

func sortByCategory(n []*Note) {
	sort.Sort(byCategory(n))
}
