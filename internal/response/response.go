// Package response todo
package response

import (
	"fmt"
	"io"

	"httpFromTcp/internal/headers"
)

type StatusCode int

const (
	Ok            StatusCode = 200
	BadRequest    StatusCode = 400
	InternalError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	fmt.Println("write tatusline")
	switch statusCode {
	case Ok:
		n, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))

		fmt.Println(n)
		if err != nil {
			return err
		}
	case BadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case InternalError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["content-length"] = fmt.Sprint(contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
