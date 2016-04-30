package aqua

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"
)

func ModAccessLog(path string) func(http.Handler) http.Handler {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	l := log.New(f, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			l.Printf("%s %s %.3f", r.Method, r.RequestURI, time.Since(start).Seconds())
		})
	}
}

func ModSlowLog(path string, msec int) func(http.Handler) http.Handler {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	l := log.New(f, "", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start).Seconds() - float64(msec)/1000.0
			if dur > 0 {
				l.Printf("%s %s %.3f", r.Method, r.RequestURI, time.Since(start).Seconds())
			}
		})
	}
}

func ModRecorder() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			for k, v := range rec.Header() {
				w.Header()[k] = v
			}

			len := rec.Body.Len()
			w.Header().Set("Content-Length", strconv.Itoa(len))
			w.WriteHeader(rec.Code)
			w.Write(rec.Body.Bytes())
		})
	}
}
