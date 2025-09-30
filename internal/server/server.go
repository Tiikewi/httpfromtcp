// Package server is server :kuu:
package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"httpFromTcp/internal/request"
	"httpFromTcp/internal/response"
)

type HandlerError struct {
	Status  int
	Message string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener      net.Listener
	serverRunning atomic.Bool
	handler       Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		ln,
		atomic.Bool{},
		handler,
	}
	server.listen()

	return server, nil
}

func (he *HandlerError) writeError(w *response.Writer) error {
	err := w.WriteStatusLine(response.StatusCode(he.Status))
	return err
}

func (s *Server) Close() error {
	s.serverRunning.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	s.serverRunning.Store(true)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.serverRunning.Load() {
				panic("error when starting listening")
			}
		}

		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := &response.Writer{
		Writer: conn,
		State:  response.StatusLine,
	}

	req, err := request.RequestFromReader(conn)
	if err != nil {
		panic("Foo faa")
	}

	s.handler(w, req)

	conn.Write([]byte("\r\n"))
	conn.Close()
}
