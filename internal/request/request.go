// Package request: package for requests
package request

import (
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	State       parsesState
}

type parsesState string

const (
	StateInit parsesState = "init"
	StateDone parsesState = "done"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

var bufferSize = 8

var ErrMalformedRequestLine = fmt.Errorf("malformed http request line")

func (r *Request) done() bool {
	return r.State == StateDone
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		RequestLine{},
		StateInit,
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
			return nil, fmt.Errorf("error while parsing request")
		}

		copy(buf, buf[parsedN:readToIndex])
		readToIndex -= parsedN
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.done() {
		return 0, fmt.Errorf("trying to read data in done state")
	}
	if r.State == StateInit {
		reqLine, bytes, err := parseRequestLine(string(data))
		if err != nil {
			return bytes, err
		}

		if bytes == 0 {
			return 0, nil
		}

		r.State = StateDone
		r.RequestLine = *reqLine
		fmt.Printf("method is is: %s", reqLine.Method)

		return bytes, nil
	}

	return 0, fmt.Errorf("unexpected state")
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
