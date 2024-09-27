package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"example/doc"
)

func readLine(r io.Reader) string {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		return s.Text()
	}
	return ""
}

func main() {
	library := doc.NewLibrary(os.Args[1:]...)

	// search term from stdin
	search := readLine(os.Stdin)

	fmt.Println("Search:", search)
	fmt.Println("Found:", library.Search(search))
}
