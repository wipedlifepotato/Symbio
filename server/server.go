package server

import (
    "net/http"
)


type Server struct {
    mux *http.ServeMux
}


func New() *Server {
    s := &Server{
        mux: http.NewServeMux(),
    }
    return s
}

func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

func (s *Server) Start(port string) error {
    return http.ListenAndServe("127.0.0.1:"+port, s.mux)
}


func (s *Server) Handle(path string, handler http.HandlerFunc) {
    s.mux.HandleFunc(path, handler)
}
