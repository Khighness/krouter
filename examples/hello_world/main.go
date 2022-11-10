package main

import (
	"log"
	"net/http"

	"github.com/khighness/krouter"
)

// @Author KHighness
// @Update 2022-11-10

func main() {
	router := krouter.New()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	log.Fatalln(http.ListenAndServe(":3333", router))
}
