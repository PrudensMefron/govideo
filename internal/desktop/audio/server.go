package audio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Server uses real HTTP for media seeking in native WebKitGTK. Custom Wails
// schemes are not supported consistently by its GStreamer media transport.
// Only loopback is bound; opaque preview tokens are required for every read.
type Server struct {
	server *http.Server
	base   string
}

func NewServer(store PreviewStore) (*Server, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("iniciar reprodução local: %w", err)
	}
	address := listener.Addr().String()
	media := Middleware(store)(http.NotFoundHandler())
	s := &Server{base: "http://" + address}
	s.server = &http.Server{ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != address {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		media.ServeHTTP(w, r)
	})}
	go func() { _ = s.server.Serve(listener) }()
	return s, nil
}

func (s *Server) URL(id string) string { return s.base + "/audio-preview/" + url.PathEscape(id) }
func (s *Server) Close()               { _ = s.server.Close() }
