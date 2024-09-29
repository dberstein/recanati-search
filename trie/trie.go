package trie

import (
	"bufio"
	"io"
	"log"
	"strings"
)

type Node struct {
	c map[byte]*Node
	v any
}

func NewNode(v any) *Node {
	return &Node{
		v: v,
		c: make(map[byte]*Node),
	}
}

type Trie struct {
	root *Node
}

func NewTrie() *Trie {
	return &Trie{
		root: NewNode(""),
	}
}

func (t *Trie) Add(v string) *Trie {
	current := t.root
	word := strings.ToLower(strings.ReplaceAll(v, " ", ""))
	for i := 0; i < len(word); i++ {
		if _, ok := current.c[word[i]]; !ok {
			current.c[word[i]] = NewNode(string(word[i]))
		}
		current = current.c[word[i]]
	}
	return t
}

func (t *Trie) SearchPrefix(q string) bool {
	word := strings.ToLower(strings.ReplaceAll(q, " ", ""))
	current := t.root
	for i := 0; i < len(word); i++ {
		if current == nil || current.c[word[i]] == nil {
			return false
		}
		current = current.c[word[i]]
	}
	return true
}

func (t *Trie) InsertContentWords(r io.Reader) *Trie {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)

	for s.Scan() {
		t.Add(s.Text())
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}

	return t
}
