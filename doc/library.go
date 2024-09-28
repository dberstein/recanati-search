package doc

import (
	"io"
	"os"
	"strings"
)

type Library struct {
	Docs map[string]*Document
}

func NewFileLibrary(fname ...string) *Library {
	docs := map[string]*Document{}
	lb := Library{Docs: docs}

	for _, fn := range fname {
		lb.AddFile(fn)
	}

	return &lb
}

func (lb *Library) Add(r io.Reader) *Document {
	d := NewDocument(r)
	d.T.InsertContentWords(strings.NewReader(d.Content))
	lb.Docs[d.ID] = d
	return d
}

func (lb *Library) AddFile(fname string) *Document {
	f, err := os.Open(fname)
	if err != nil {
		panic(f)
	}
	defer f.Close()

	d := NewFileDocument(fname)
	d.T.InsertContentWords(f)
	lb.Docs[d.ID] = d

	return d
}

func (lb *Library) Search(q string) []string {
	res := []string{}

	for _, d := range lb.Docs {
		if !d.Search(q) {
			continue
		}
		res = append(res, d.ID)
	}

	return res
}
