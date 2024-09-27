package doc

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"example/trie"
)

type Document struct {
	ID      string
	t       *trie.Trie
	content string
}

type Library struct {
	docs map[string]*Document
}

func NewLibrary(fname ...string) *Library {
	docs := map[string]*Document{}
	lb := Library{docs: docs}

	for _, fn := range fname {
		lb.Add(fn)
	}

	return &lb
}

func (lb *Library) Add(fname string) *Document {
	f, err := os.Open(fname)
	if err != nil {
		panic(f)
	}
	defer f.Close()

	d := NewDocument(fname)
	d.t.InsertContentWords(f)
	lb.docs[d.ID] = d

	return d
}

func (lb *Library) Search(q string) []string {
	res := []string{}

	for _, d := range lb.docs {
		if !d.Search(q) {
			continue
		}
		res = append(res, d.ID)
	}

	return res
}

func NewDocument(fname string) *Document {
	f, err := os.Open(fname)
	if err != nil {
		panic(f)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	d := &Document{
		ID:      getID(content),
		content: string(content),
		t:       trie.NewTrie(),
	}

	return d
}

func getID(content []byte) string {
	h := sha256.New()
	h.Write(content)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (d *Document) Search(q string) bool {
	return d.t.Search(q)
}
