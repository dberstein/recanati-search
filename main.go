package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

func setupRouter(library *doc.Library) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /doc/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		d, ok := library.Docs[id]
		if !ok {
			http.Error(w, "document not found", http.StatusNotFound)
			return
		}
		w.Write([]byte(d.Content))
	})

	mux.HandleFunc("POST /doc", func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		contentId := doc.GetID(bs)
		d, ok := library.Docs[contentId]
		if !ok {
			d = library.Add(strings.NewReader(string(bs)))
		}

		w.Write([]byte(d.ID))
	})

	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("q")
		if search == "" {
			http.Error(w, "missing query", http.StatusBadRequest)
			return
		}

		res := library.Search(search)
		bs, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write(bs)
	})

	return mux
}

func main() {
	library := doc.NewFileLibrary(os.Args[1:]...)

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		search := readLine(os.Stdin) // search term from stdin
		fmt.Println("Search:", search)
		fmt.Println("Found:", library.Search(search))
	}

	mux := setupRouter(library)
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal(err)
	}
}
