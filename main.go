package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
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
		if len(bytes.TrimSpace(bs)) == 0 {
			http.Error(w, "empty document", http.StatusBadRequest)
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
		w.Header().Add("Content-Type", "application/json")
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

		w.Header().Add("Content-Type", "application/json")
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

		w.Header().Add("Content-Type", "application/json")
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
	fgColorRed := color.New(color.FgRed)
	bgWhiteFgColorRed := fgColorRed.Add(color.BgWhite)

	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{w, 200}
		h.ServeHTTP(recorder, r)

		log.Print(strings.Join([]string{
			color.MagentaString(r.Host),
			bgWhiteFgColorRed.Sprint(getRemoteAddress(r)),
			getColoredStatusCode(recorder.statusCode),
			r.Method,
			"\"" + color.CyanString(r.URL.String()) + "\"",
			"\"" + color.CyanString(r.Header.Get("User-Agent")) + "\"",
			time.Now().Sub(start).String(),
		}, " "))
	}

	return http.HandlerFunc(fn)
}

func getColoredStatusCode(code int) string {
	var colorFn func(string, ...interface{}) string
	if code < http.StatusMultipleChoices { // sucesses
		colorFn = color.HiGreenString
	} else if code < http.StatusInternalServerError { // client errors
		colorFn = color.HiBlueString
	} else { // server errors
		colorFn = color.HiRedString
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

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
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
			// Returns first non-private IP address...
			parts[i] = strings.TrimSpace(p)
			if !isPrivateIP(net.ParseIP(parts[i])) {
				return parts[i]
			}
		}
	}
	return hdrRealIP
}

func main() {
	dsn := flag.String("dsn", ":memory:", "Sqlite DSN")
	addr := flag.String("addr", ":8080", "Listen address")
	flag.Parse()

	mux := setupRouter(*dsn)

	srv := &http.Server{
		Addr:              *addr,
		IdleTimeout:       0,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		Handler:           logRequestHandler(mux),
	}

	fmt.Println(color.HiGreenString("Listening:"), *addr)
	log.Fatal(srv.ListenAndServe())
}
