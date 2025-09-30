package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpFromTcp/internal/request"
	"httpFromTcp/internal/response"
	"httpFromTcp/internal/server"
)

const port = 42069

const badRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrorHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`

const successHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
	`

func main() {
	var handlerFn server.Handler = func(w *response.Writer, req *request.Request) {
		defaultContentType := "text/html"

		if req.RequestLine.RequestTarget == "/yourproblem" {
			w.WriteStatusLine(400)
			defaultHeaders := response.GetDefaultHeaders(len(badRequestHTML), defaultContentType)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(badRequestHTML))
		}

		if req.RequestLine.RequestTarget == "/myproblem" {
			w.WriteStatusLine(500)
			defaultHeaders := response.GetDefaultHeaders(len(internalErrorHTML), defaultContentType)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(internalErrorHTML))
		}

		w.WriteStatusLine(200)
		defaultHeaders := response.GetDefaultHeaders(len(successHTML), defaultContentType)
		req.Headers = defaultHeaders
		w.WriteHeaders(req.Headers)
		w.Writer.Write([]byte("\r\n"))
		w.WriteBody([]byte(successHTML))
	}

	server, err := server.Serve(port, handlerFn)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
