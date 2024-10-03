package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3" // Import driver (blank import for registration)
)

func ensureSchema(db *sql.DB) error {
	if _, err := db.Exec(
		"CREATE VIRTUAL TABLE IF NOT EXISTS docs USING FTS5(body);",
	); err != nil {
		return err
	}
	return nil
}

func NewDb(dsn string) *sql.DB {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	err = ensureSchema(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

var db *sql.DB

func init() {
	db = NewDb(":memory:")
}

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Create new document
	mux.HandleFunc("POST /doc", func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := db.Exec("INSERT INTO docs (body) VALUES (?)", bs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		docId, err := res.LastInsertId()
		w.Write([]byte(strconv.FormatInt(docId, 10)))
	})

	// Retrieve document
	mux.HandleFunc("GET /doc/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body string
		err := db.QueryRow(
			"SELECT body FROM docs WHERE rowid = ?", r.PathValue("id")).Scan(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write([]byte(body))
	})

	// Delete document
	mux.HandleFunc("DELETE /doc/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := db.Exec("DELETE FROM docs WHERE rowid = ?", r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write([]byte(strconv.FormatInt(affected, 10)))
	})

	// Search documents
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("q")
		if search == "" {
			http.Error(w, "missing query", http.StatusBadRequest)
			return
		}

		rows, err := db.Query("SELECT rowid, body FROM docs WHERE body MATCH ?", search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer rows.Close()

		var docIds []int64
		for rows.Next() {
			var id int64
			var body string

			err = rows.Scan(&id, &body)
			if err != nil {
				log.Fatal(err)
			}

			docIds = append(docIds, id)
		}

		bs, err := json.Marshal(docIds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bs)
	})

	return mux
}

func main() {
	mux := setupRouter()
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal(err)
	}
}
