package doc

import (
	"io"
	"log"
	"os"

	"github.com/dberstein/recanati-search/trie"
)

type Document struct {
	ID      string
	T       *trie.Trie
	Content string
}

func NewDocument(r io.Reader) *Document {
	content, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err.Error())
	}

	d := &Document{
		ID:      GetID(content),
		Content: string(content),
		T:       trie.NewTrie(),
	}

	return d
}

func NewFileDocument(fname string) *Document {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	return NewDocument(f)
}

func (d *Document) Search(q string) bool {
	return d.T.Search(q)
}
