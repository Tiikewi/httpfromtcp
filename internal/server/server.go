// Package server is server :kuu:
package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	listener      net.Listener
	serverRunning atomic.Bool
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		ln,
		atomic.Bool{},
	}
	server.listen()

	return server, nil
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
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\r\n"))
	conn.Close()
}
