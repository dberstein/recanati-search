package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"

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

func setupRouter(dsn string) *http.ServeMux {
	db = NewDb(dsn)
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
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		bs, err = json.Marshal(struct {
			Document int64 `json:"document"`
		}{
			Document: docId,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bs)
	})

	// Retrieve document
	mux.HandleFunc("GET /doc/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body string
		err := db.QueryRow(
			"SELECT body FROM docs WHERE rowid = ?", r.PathValue("id")).Scan(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
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

		if affected == 0 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		bs, err := json.Marshal(struct {
			Deleted int64 `json:"deleted"`
		}{
			Deleted: affected,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	})

	// Search documents
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("q")
		if search == "" {
			http.Error(w, "missing query", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(
			"SELECT rowid, body FROM docs WHERE body MATCH ? ORDER BY rank", search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer rows.Close()

		var docIds []int64 = []int64{}
		for rows.Next() {
			var id int64
			var body string

			err = rows.Scan(&id, &body)
			if err != nil {
				log.Fatal(err)
			}

			docIds = append(docIds, id)
		}

		bs, err := json.Marshal(struct {
			Matches []int64 `json:"matches"`
		}{
			Matches: docIds,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bs)
	})

	return mux
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{w, 200}
		h.ServeHTTP(recorder, r)

		log.Print(strings.Join([]string{
			color.MagentaString(getRemoteAddress(r)),
			getColorCode(recorder.statusCode),
			color.YellowString(r.Method),
			"\"" + color.CyanString(r.URL.String()) + "\"",
			"\"" + color.BlueString(r.Header.Get("User-Agent")) + "\"",
			time.Now().Sub(start).String(),
		}, " "))
	}

	return http.HandlerFunc(fn)
}

func getColorCode(code int) string {
	colorFn := color.GreenString
	if code > 399 {
		colorFn = color.YellowString
	}
	if code > 499 {
		colorFn = color.RedString
	}
	return colorFn(strconv.Itoa(code))
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}

func main() {
	dsn := flag.String("dsn", ":memory:", "Sqlite DSN")
	port := flag.Int("port", 8080, "Listen port")
	flag.Parse()

	mux := setupRouter(*dsn)
	addr := "0.0.0.0:" + strconv.Itoa(*port)

	srv := &http.Server{
		Addr:           addr,
		IdleTimeout:    0,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        logRequestHandler(mux),
	}

	fmt.Println(color.GreenString("Listening:"), color.YellowString(addr))
	log.Fatal(srv.ListenAndServe())
}
