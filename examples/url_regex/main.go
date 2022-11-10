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
	router.Get("/user/{name:\\w+}", func(w http.ResponseWriter, r *http.Request) {
		name := krouter.GetParam(r, "name")
		w.Write([]byte("hello " + name))
	})
	log.Fatalln(http.ListenAndServe(":3333", router))
}
