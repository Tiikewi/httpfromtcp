package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpFromTcp/internal/headers"
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

		s := req.RequestLine.RequestTarget
		if s == "/yourproblem" {
			w.WriteStatusLine(400)
			defaultHeaders := response.GetDefaultHeaders(len(badRequestHTML), defaultContentType, false)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(badRequestHTML))
		}
		if s == "/myproblem" {
			w.WriteStatusLine(500)
			defaultHeaders := response.GetDefaultHeaders(len(internalErrorHTML), defaultContentType, false)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(internalErrorHTML))
		}
		if strings.HasPrefix(s, "/httpbin/") {
			// defaultHeaders := response.GetDefaultHeaders(0, defaultContentType, true)
			// req.Headers = defaultHeaders

			url := "https://httpbin.org" + strings.TrimPrefix(s, "/httpbin")
			res, err := http.Get(url)
			if err != nil {
				panic("TODO")
			}
			defer res.Body.Close()

			w.WriteStatusLine(response.StatusCode(res.StatusCode))

			hdrs := headers.NewHeaders()
			for k, v := range res.Header {
				hdrs[strings.ToLower(k)] = strings.Join(v, ",")
			}
			delete(hdrs, "content-length")
			hdrs["transfer-encoding"] = "chunked"
			hdrs["trailer"] = "X-Content-SHA256,X-Content-Length"

			w.WriteHeaders(hdrs)
			w.Writer.Write([]byte("\r\n"))

			buf := make([]byte, 1024)
			fullResp := make([]byte, 60000) // TODO:
			for {
				n, err := res.Body.Read(buf)
				if n > 0 {
					w.WriteChunkedBody(buf[:n])
					fullResp = append(fullResp, buf[:n]...)
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					// handle error (and likely break)
					break
				}
			}
			w.WriteChunkedBodyDone()
			sha := sha256.Sum256(fullResp)
			tr := headers.NewHeaders()
			tr["X-Content-SHA256"] = fmt.Sprintf("%x", sha[:])
			tr["X-Content-Length"] = fmt.Sprint(len(fullResp))
			w.WriteTrailers(tr)

		}
		if s == "/video" {
			w.WriteStatusLine(200)
			video, _ := os.ReadFile("../../assets/vim.mp4")
			defaultHeaders := response.GetDefaultHeaders(len(video), "video/mp4", false)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(video))
		} else {
			w.WriteStatusLine(200)
			defaultHeaders := response.GetDefaultHeaders(len(successHTML), defaultContentType, false)
			req.Headers = defaultHeaders
			w.WriteHeaders(req.Headers)
			w.Writer.Write([]byte("\r\n"))
			w.WriteBody([]byte(successHTML))
		}
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
