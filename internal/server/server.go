// Package server is server :kuu:
package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"httpFromTcp/internal/request"
	"httpFromTcp/internal/response"
)

type HandlerError struct {
	Status  int
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

func (he *HandlerError) writeError(w io.Writer) error {
	err := response.WriteStatusLine(w, response.StatusCode(he.Status))
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		panic("Foo faa")
	}

	var buffer bytes.Buffer
	handlerError := s.handler(&buffer, req)
	if handlerError != nil {
		handlerError.writeError(conn)
		defaultHeaders := response.GetDefaultHeaders(len(handlerError.Message))
		response.WriteHeaders(conn, defaultHeaders)
		conn.Write([]byte("\r\n"))
		conn.Write([]byte(handlerError.Message))
	} else {
		response.WriteStatusLine(conn, response.Ok)
		defaultHeaders := response.GetDefaultHeaders(buffer.Len())
		response.WriteHeaders(conn, defaultHeaders)
		conn.Write([]byte("\r\n"))
		fmt.Println("BUFFER: ", buffer.String(), buffer.Len())
		conn.Write(buffer.Bytes())
		conn.Close()
	}
}
