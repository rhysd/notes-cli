package notes

import (
	"os"
	"sort"
	"strings"
	"time"
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

type byModified struct {
	a    []*Note
	memo map[*Note]time.Time
	err  error
}

func (by *byModified) modTime(i int) time.Time {
	n := by.a[i]
	if t, ok := by.memo[n]; ok {
		return t
	}
	s, err := os.Stat(n.FilePath())
	if err != nil {
		by.err = err
		return time.Time{}
	}
	t := s.ModTime()
	by.memo[n] = t
	return t
}
func (by *byModified) Len() int {
	return len(by.a)
}
func (by *byModified) Swap(i, j int) {
	a := by.a
	a[i], a[j] = a[j], a[i]
}
func (by *byModified) Less(i, j int) bool {
	if by.err != nil {
		return true
	}
	l, r := by.modTime(i), by.modTime(j)
	return l.After(r)
}

// sortByModified sorts given notest by modified time. When an error occurs, the order of given notes
// are undefined.
func sortByModified(n []*Note) error {
	by := &byModified{
		a:    n,
		memo: make(map[*Note]time.Time, len(n)),
	}
	sort.Sort(by)
	return by.err
}
