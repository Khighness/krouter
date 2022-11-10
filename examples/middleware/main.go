package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/khighness/krouter"
)

// @Author KHighness
// @Update 2022-11-10

func main() {
	router := krouter.New()
	router.NotFoundFunc(notFoundFunc)
	router.Use(withLogging, withTracing, withStatusRecord)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	log.Fatal(http.ListenAndServe(":3333", router))
}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Not found page !")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func withStatusRecord(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := statusRecorder{w, http.StatusOK}
		next.ServeHTTP(&rec, r)
		log.Printf("[status] response status: %v\n", rec.status)
	}
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[logging] connection from: %s", r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}

func withTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[tracing] request for: %s", r.RequestURI)
		next.ServeHTTP(w, r)
	}
}
