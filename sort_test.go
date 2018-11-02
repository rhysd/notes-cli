package notes

import (
	"testing"
	"time"
)

func TestSortNotes(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}

	newNote := func(cat, file, created string) *Note {
		t, err := time.Parse(time.RFC3339, created)
		if err != nil {
			panic(err)
		}
		return &Note{
			Config:   cfg,
			Category: cat,
			Created:  t,
			File:     file,
		}
	}

	notes := []*Note{
		newNote("a", "1.md", "2018-10-30T11:37:45+09:00"),
		newNote("a", "2.md", "2018-10-31T11:37:45+09:00"),
		newNote("c", "3.md", "2018-10-30T18:37:45+09:00"),
		newNote("b", "4.md", "2019-10-29T18:37:45+09:00"),
		newNote("c", "0.md", "2018-12-19T18:37:45+09:00"),
	}

	for _, tc := range []struct {
		what string
		sort func(n []*Note)
		want []string
	}{
		{
			what: "created",
			sort: func(n []*Note) { sortByCreated(n) },
			want: []string{"1.md", "3.md", "2.md", "0.md", "4.md"},
		},
		{
			what: "filename",
			sort: func(n []*Note) { sortByFilename(n) },
			want: []string{"0.md", "1.md", "2.md", "3.md", "4.md"},
		},
		{
			what: "category",
			sort: func(n []*Note) { sortByCategory(n) },
			want: []string{"1.md", "2.md", "4.md", "0.md", "3.md"},
		},
	} {
		if len(notes) != len(tc.want) {
			panic("tc.want is invalid: " + tc.what)
		}

		t.Run(tc.what, func(t *testing.T) {
			tc.sort(notes)
			for i, want := range tc.want {
				have := notes[i].File
				if want != have {
					t.Error("mismatch at", i, "want", want, "but have", have)
				}
			}
		})
	}
}
