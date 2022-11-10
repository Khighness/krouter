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

	r1 := "user_profile"
	router.GetAndName("/user/:user/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r1))
	}, r1)
	param1 := make(map[string]string)
	param1["user"] = "KHighness"

	if route1, err := router.Generate(http.MethodGet, r1, param1); err != nil {
		panic(err)
	} else {
		log.Println(route1)
	}

	r2 := "user_repository"
	router.GetAndName("/user/{user:\\w+}/{repository:\\w+}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r2))
	}, r2)
	param2 := make(map[string]string)
	param2["user"] = "KHighness"
	param2["repository"] = "krouter"
	if route2, err := router.Generate(http.MethodGet, r2, param2); err != nil {
		panic(err)
	} else {
		log.Println(route2)
	}
}
