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

type WriterStatus string

const (
	StatusLine WriterStatus = "status"
	Headers    WriterStatus = "headers"
	Body       WriterStatus = "body"
)

type Writer struct {
	Writer io.Writer
	State  WriterStatus
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != StatusLine {
		return fmt.Errorf("trying to write status line when writer status is: %s", w.State)
	}

	switch statusCode {
	case Ok:
		_, err := w.Writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case BadRequest:
		_, err := w.Writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case InternalError:
		_, err := w.Writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	}

	w.State = Headers
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != Headers {
		return fmt.Errorf("trying to write headers when writer status is: %s", w.State)
	}

	for k, v := range headers {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	w.State = Body
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != Body {
		return 0, fmt.Errorf("trying to write body when writer status is: %s", w.State)
	}

	return w.Writer.Write(p)
}

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = fmt.Sprint(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = contentType

	return headers
}
