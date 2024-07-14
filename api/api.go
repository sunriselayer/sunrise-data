package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {})
	http.Handle("/", r)
}
