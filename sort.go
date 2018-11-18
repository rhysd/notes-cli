package notes

import (
	"github.com/pkg/errors"
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
	return a[i].Created.After(a[j].Created)
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
	a []*Note
	t map[*Note]time.Time
}

func (by *byModified) Len() int {
	return len(by.a)
}
func (by *byModified) Swap(i, j int) {
	a := by.a
	a[i], a[j] = a[j], a[i]
}
func (by *byModified) Less(i, j int) bool {
	a := by.a
	l, r := by.t[a[i]], by.t[a[j]]
	return l.After(r)
}

// sortByModified sorts given notes by modified time. When an error occurs, the order of given notes
// is undefined. The latest is the first.
func sortByModified(notes []*Note) error {
	by := &byModified{
		a: notes,
		t: make(map[*Note]time.Time, len(notes)),
	}

	for _, n := range notes {
		s, err := os.Stat(n.FilePath())
		if err != nil {
			return errors.Wrap(err, "Cannot sort by modified time")
		}
		by.t[n] = s.ModTime()
	}

	sort.Sort(by)

	return nil
}
