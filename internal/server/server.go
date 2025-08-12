package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	Addr    string
	Running atomic.Bool
	Handler Handler
	ln      net.Listener
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Addr:    ln.Addr().String(),
		Running: atomic.Bool{},
		Handler: handler,
		ln:      ln,
	}
	server.Running.Store(true)

	go server.listen()

	return server, nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	writer := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteStatusLine(response.StatusBadRequest)
		writer.WriteHeaders(response.GetDefaultHeaders(len(err.Error())))
		writer.WriteBody([]byte(err.Error()))
		return
	}

	s.Handler(writer, req)
}

func (s *Server) listen() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if s.Running.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) Close() error {
	s.Running.Store(false)
	return s.ln.Close()
}
