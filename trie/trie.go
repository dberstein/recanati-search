package trie

import (
	"bufio"
	"io"
	"log"
	"strings"
)

type Node struct {
	c map[rune]*Node
	v rune
}

func NewNode(v rune) *Node {
	return &Node{
		v: v,
		c: make(map[rune]*Node),
	}
}

type Trie struct {
	root *Node
}

func NewTrie() *Trie {
	return &Trie{
		root: NewNode('.'),
	}
}

func (t *Trie) Add(v string) *Trie {
	current := t.root
	for _, r := range strings.ToLower(strings.ReplaceAll(v, " ", "")) {
		if _, ok := current.c[r]; !ok {
			current.c[r] = NewNode(r)
		}
		current = current.c[r]
	}
	return t
}

func (t *Trie) Delete(key string) *Trie {
	// todo: delete from trie, ie. t.c[key]
	return t
}

func (t *Trie) InsertWords(r io.Reader) *Trie {
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

func (t *Trie) SearchPrefix(q string) bool {
	current := t.root
	for _, r := range strings.ToLower(strings.ReplaceAll(q, " ", "")) {
		if current == nil || current.c[r] == nil {
			return false
		}
		current = current.c[r]

	}
	return true
}
