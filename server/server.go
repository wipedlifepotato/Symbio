package server

import (
    "net/http"
)

type Server struct {
    mux *http.ServeMux
}

func New() *Server {
    return &Server{
        mux: http.NewServeMux(),
    }
}

func (s *Server) GetMux() *http.ServeMux {
    return s.mux
}

func (s *Server) Start(addr, port string) error {
    return http.ListenAndServe(addr+":"+port, s.mux)
}

func (s *Server) Handle(path string, handler http.HandlerFunc) {
    s.mux.HandleFunc(path, handler)
}

func (s *Server) HandleHandler(path string, handler http.Handler) {
    s.mux.Handle(path, handler)
}
