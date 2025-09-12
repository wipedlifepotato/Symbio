package server

import "net/http"

func NewSubrouterWithMiddleware(mux *http.ServeMux, middleware func(http.Handler) http.Handler) *http.ServeMux {
	newMux := http.NewServeMux()
	newMux.Handle("/", middleware(mux))
	return newMux
}
