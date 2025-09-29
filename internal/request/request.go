// Package request: package for requests
package request

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"httpFromTcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	State       parsesState
	Headers     headers.Headers
	Body        []byte
}

type parsesState string

const (
	StateInit        parsesState = "init"
	StateHeadersInit parsesState = "parsing headers"
	StateBodyInit    parsesState = "parsing body"
	StateDone        parsesState = "done"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

var (
	bufferSize = 8
	crlf       = "\r\n"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed http request line")

func (r *Request) done() bool {
	return r.State == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		RequestLine{},
		StateInit,
		headers.Headers{},
		nil,
	}

	buf := make([]byte, bufferSize)
	readToIndex := 0
	for !r.done() {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		readN, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				r.State = StateDone
				break
			} else {
				log.Fatal("error when reading", "error", err, readN)
			}
		}

		readToIndex += readN

		parsedN, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error while parsing request, %s", err)
		}

		copy(buf, buf[parsedN:readToIndex])
		readToIndex -= parsedN
	}

	contentLength, _ := r.getContentLegth()
	if len(r.Body) == 0 && contentLength > 0 {
		return nil, fmt.Errorf("content length does not match body, cl: %d, body: %d", len(r.Body), contentLength)
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	fmt.Printf("Parsing... %s, data: %s\n", r.State, data)
	if r.done() {
		return 0, fmt.Errorf("trying to read data in done state")
	}

	parsedBytes := 0
	for r.State != StateDone {
		n, err := r.parseSingle(data[parsedBytes:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}
		parsedBytes += n
	}

	return parsedBytes, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		reqLine, bytes, err := parseRequestLine(string(data))
		if err != nil {
			return bytes, err
		}

		if bytes == 0 {
			return 0, nil
		}

		r.State = StateHeadersInit
		r.RequestLine = *reqLine
		return bytes, nil

	case StateHeadersInit:
		headerN, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.State = StateBodyInit
			return 0, nil
		}

		return headerN, nil
	case StateBodyInit:
		contentLength, err := r.getContentLegth()
		if err != nil {
			return 0, err
		}

		if contentLength == 0 {
			r.State = StateDone
			return 0, nil
		}

		if len(data) >= len(crlf) {
			if string(data[:len(crlf)]) == crlf {
				return len(crlf), nil
			}
		}

		if len(data) < contentLength {
			return 0, nil
		}

		if len(data) > contentLength {
			return 0, fmt.Errorf("body is too long")
		}

		r.State = StateDone
		r.Body = data
		return len(data), nil

	default:
		return 0, fmt.Errorf("unexpected state")
	}
}

func (r *Request) getContentLegth() (int, error) {
	contentLengthStr := r.Headers.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return 0, nil
	}
	return contentLength, nil
}

func parseRequestLine(request string) (*RequestLine, int, error) {
	index := strings.Index(request, "\r\n")
	if index == -1 {
		return nil, 0, nil
	}

	startLine := request[:index]
	read := index + len("\r\n")

	parts := strings.Split(string(startLine), " ")
	if len(parts) != 3 {
		return nil, read, fmt.Errorf("too few parts in request line, parts: %d", len(parts))
	}

	method := parts[0]
	target := parts[1]
	versionParts := strings.Split(parts[2], "/")

	if len(versionParts) != 2 {
		return nil, read, fmt.Errorf("version parts too short; %s", versionParts)
	}
	version := versionParts[1]

	if !IsUpper(method) {
		return nil, read, fmt.Errorf("verb is not uppercase; %s", method)
	}

	if string(version) != "1.1" {
		return nil, read, fmt.Errorf("wrong version number: %s", string(version))
	}

	return &RequestLine{
		HTTPVersion:   string(version),
		RequestTarget: target,
		Method:        method,
	}, read, nil
}

func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
