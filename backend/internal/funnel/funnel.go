package funnel

import "net/http"

type FunnelServer struct {
}

func NewServer() (*FunnelServer, error) {
	return &FunnelServer{}, nil
}

func (s *FunnelServer) Handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/immis", s.immiHandler)
	return r
}

func (s *FunnelServer) immiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World 1"))
}
